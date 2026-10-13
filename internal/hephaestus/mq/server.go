package mq

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	// LogFile is the name of the append-only log in the shared directory.
	LogFile = "messages.log"
	// DiscoveryFile tells a client where the service listens, and its token.
	DiscoveryFile = "mq.json"
	// DefaultLease is how long a receiver has to acknowledge a message before it is given again.
	DefaultLease = 60 * time.Second
	// maxWait is the longest that a receive waits for a message.
	maxWait = 5 * time.Minute
)

// Discovery is the content of the discovery file.
type Discovery struct {
	Address string    `json:"address"`
	Token   string    `json:"token"`
	PID     int       `json:"pid"`
	Started time.Time `json:"started"`
}

// Server is the message queue service of one repo.
type Server struct {
	dir   string
	lease time.Duration
	token string

	mu     sync.Mutex
	log    *os.File
	state  *state
	leases map[string]delivery
	// seen is the last time that each worker received, acknowledged, or renewed: the keep-alive that
	// tells a scheduler that a worker still works. It is not in the log.
	seen    map[string]time.Time
	changed chan struct{} // closed and replaced when a message arrives

	listener net.Listener
	http     *http.Server
	now      func() time.Time
}

// delivery is the lease of a message to a worker, until a time.
type delivery struct {
	worker string
	id     string
	until  time.Time
}

// Open reads the log of dir, or makes an empty one, and returns a server that is not started yet.
func Open(dir string, lease time.Duration) (*Server, error) {
	if lease <= 0 {
		lease = DefaultLease
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Server{dir: dir, lease: lease, state: newState(), leases: map[string]delivery{}, seen: map[string]time.Time{}, changed: make(chan struct{}), now: time.Now}
	path := filepath.Join(dir, LogFile)
	if file, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
		line := 0
		for scanner.Scan() {
			line++
			var e Event
			if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
				file.Close()
				return nil, fmt.Errorf("%s:%d is not a valid event: %v. The log is append-only: correct or remove the line by hand only if a crash cut it.", path, line, err)
			}
			s.state.apply(e)
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	log, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	s.log = log
	return s, nil
}

// append writes one event to the log, and applies it. The caller holds mu.
func (s *Server) append(e Event) error {
	e.Seq = s.state.seq + 1
	e.Time = s.now().UTC()
	if e.Message != nil {
		e.Message.Seq, e.Message.Time = e.Seq, e.Time
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if _, err := s.log.Write(append(line, '\n')); err != nil {
		return err
	}
	if err := s.log.Sync(); err != nil {
		return err
	}
	s.state.apply(e)
	return nil
}

// Start listens on a free port of 127.0.0.1, and writes the discovery file. It fails when another
// service of the repo already answers.
func (s *Server) Start() error {
	if d, err := ReadDiscovery(s.dir); err == nil {
		if (&Client{discovery: d, http: &http.Client{Timeout: 2 * time.Second}}).Health() == nil {
			return fmt.Errorf("A message queue already runs for this repo, at %s (pid %d).", d.Address, d.PID)
		}
	}
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return err
	}
	s.token = hex.EncodeToString(token)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.listener = listener
	mux := http.NewServeMux()
	mux.HandleFunc("POST /send", s.handleSend)
	mux.HandleFunc("POST /receive", s.handleReceive)
	mux.HandleFunc("POST /ack", s.handleAck)
	mux.HandleFunc("POST /renew", s.handleRenew)
	mux.HandleFunc("GET /workers", s.handleWorkers)
	mux.HandleFunc("GET /pending", func(w http.ResponseWriter, r *http.Request) {
		pending := s.Pending()
		if pending == nil {
			pending = []Pending{}
		}
		reply(w, pending, nil)
	})
	mux.HandleFunc("GET /log", s.handleLog)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	s.http = &http.Server{Handler: s.authorize(mux), ReadHeaderTimeout: 10 * time.Second}
	d := Discovery{Address: listener.Addr().String(), Token: s.token, PID: os.Getpid(), Started: s.now().UTC()}
	data, _ := json.MarshalIndent(d, "", "  ")
	if err := os.WriteFile(filepath.Join(s.dir, DiscoveryFile), data, 0o600); err != nil {
		listener.Close()
		return err
	}
	go s.http.Serve(listener)
	return nil
}

// Address returns the address that the server listens on.
func (s *Server) Address() string { return s.listener.Addr().String() }

// Close stops the server, removes the discovery file, and closes the log.
func (s *Server) Close() error {
	if s.http != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.http.Shutdown(ctx)
		if d, err := ReadDiscovery(s.dir); err == nil && d.Token == s.token {
			os.Remove(filepath.Join(s.dir, DiscoveryFile))
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.log.Close()
}

func (s *Server) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+s.token {
			http.Error(w, "The token is wrong: read it from the discovery file of the repo.", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Send accepts a message, and returns it with its seq and time. A message with the id of a message
// that the log has already is a repeated send: Send returns the first one, and adds nothing.
func (s *Server) Send(m Message) (Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m.ID != "" {
		if index, ok := s.state.byID[m.ID]; ok {
			return s.state.messages[index], nil
		}
	} else {
		m.ID = newID(s.now())
	}
	if err := m.validate(); err != nil {
		return Message{}, err
	}
	if err := s.append(Event{Type: "message", Message: &m}); err != nil {
		return Message{}, err
	}
	close(s.changed)
	s.changed = make(chan struct{})
	return s.state.messages[len(s.state.messages)-1], nil
}

// Receiver is a worker that asks for its messages.
type Receiver struct {
	Worker string   `json:"worker"`
	Role   string   `json:"role,omitempty"`
	Kinds  []string `json:"kinds,omitempty"`
	// Wait is how long to wait for a message when none waits now.
	Wait time.Duration `json:"wait,omitempty"`
	// Max is the largest number of messages to return; 0 means 100.
	Max int `json:"max,omitempty"`
}

// Receive returns the messages that wait for the worker, in log order, and leases them: they are not
// given again until the lease ends, unless the worker acknowledges them. It registers the worker
// first, or records a new role or kinds. With Wait, it waits for a message up to that time.
func (s *Server) Receive(ctx context.Context, r Receiver) ([]Message, error) {
	if !namePattern.MatchString(r.Worker) {
		return nil, fmt.Errorf("worker %q is not a worker name: lowercase words, for example merge-queue-1.", r.Worker)
	}
	if r.Role != "" && !rolePattern.MatchString(r.Role) {
		return nil, fmt.Errorf("role %q is not a role name: lowercase words, or a path of them for a role of a family, for example formatter/golang.", r.Role)
	}
	kinds := slices.Clone(r.Kinds)
	slices.Sort(kinds)
	kinds = slices.Compact(kinds)
	wait := min(r.Wait, maxWait)
	deadline := s.now().Add(wait)
	for {
		s.mu.Lock()
		w := s.state.workers[r.Worker]
		if w == nil || w.role != r.Role || !slices.Equal(w.kinds, kinds) {
			if err := s.append(Event{Type: "register", Worker: r.Worker, Role: r.Role, Kinds: kinds}); err != nil {
				s.mu.Unlock()
				return nil, err
			}
			w = s.state.workers[r.Worker]
		}
		s.seen[r.Worker] = s.now()
		messages := s.deliverable(w, r.Max)
		changed := s.changed
		s.mu.Unlock()
		if len(messages) > 0 || wait <= 0 || !s.now().Before(deadline) {
			return messages, nil
		}
		// Wait for a new message, the end of a lease, or the deadline.
		timer := time.NewTimer(min(time.Until(deadline), time.Second))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-changed:
		case <-timer.C:
		}
		timer.Stop()
	}
}

// deliverable returns the messages that wait for w and that no lease holds, in log order, and
// leases them. A message of a sender whose earlier message to w is leased waits, so that the
// messages of one sender keep their order. The caller holds mu.
func (s *Server) deliverable(w *worker, max int) []Message {
	if max <= 0 {
		max = 100
	}
	now := s.now()
	blocked := map[string]bool{}
	var result []Message
	for _, m := range s.state.messages {
		if !s.state.pendingFor(w, m) {
			continue
		}
		key := leaseKey(w, m)
		if held, ok := s.leases[key]; ok && now.Before(held.until) {
			blocked[m.From] = true
			continue
		}
		if blocked[m.From] {
			continue
		}
		if len(result) >= max {
			blocked[m.From] = true // keep the order of this sender for the next receive
			continue
		}
		s.leases[key] = delivery{worker: w.name, id: m.ID, until: now.Add(s.lease)}
		result = append(result, m)
	}
	return result
}

// Ack records that the worker handled the messages with ids. An id that the worker already
// acknowledged, or that does not exist, changes nothing.
func (s *Server) Ack(worker string, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	w := s.state.workers[worker]
	if w == nil {
		return fmt.Errorf("No worker %q has received a message. Receive first.", worker)
	}
	s.seen[worker] = s.now()
	var fresh []string
	for _, id := range ids {
		index, ok := s.state.byID[id]
		if !ok {
			continue
		}
		m := s.state.messages[index]
		if !s.state.pendingFor(w, m) {
			continue
		}
		fresh = append(fresh, id)
		delete(s.leases, leaseKey(w, m))
	}
	if len(fresh) == 0 {
		return nil
	}
	if err := s.append(Event{Type: "ack", Worker: worker, IDs: fresh}); err != nil {
		return err
	}
	close(s.changed)
	s.changed = make(chan struct{})
	return nil
}

// Renewal is the answer to a keep-alive: the messages whose leases the worker still holds, now
// longer, and the messages whose leases it lost, because they ended before the renewal and another
// worker may have them now.
type Renewal struct {
	Renewed []string      `json:"renewed"`
	Lost    []string      `json:"lost"`
	Lease   time.Duration `json:"lease"`
}

// Renew is the keep-alive of a worker that still handles messages: it records that the worker is
// alive, and gives each lease on the ids that the worker holds a new full lease. A worker renews
// well within the lease (Lease in the answer), so that a long task is not given to another worker.
func (s *Server) Renew(worker string, ids []string) (Renewal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w := s.state.workers[worker]
	if w == nil {
		return Renewal{}, fmt.Errorf("No worker %q has received a message. Receive first.", worker)
	}
	now := s.now()
	s.seen[worker] = now
	renewal := Renewal{Renewed: []string{}, Lost: []string{}, Lease: s.lease}
	for _, id := range ids {
		index, ok := s.state.byID[id]
		if !ok {
			renewal.Lost = append(renewal.Lost, id)
			continue
		}
		key := leaseKey(w, s.state.messages[index])
		held, ok := s.leases[key]
		if !ok || held.worker != worker || !now.Before(held.until) {
			renewal.Lost = append(renewal.Lost, id)
			continue
		}
		held.until = now.Add(s.lease)
		s.leases[key] = held
		renewal.Renewed = append(renewal.Renewed, id)
	}
	return renewal, nil
}

// WorkerStatus is what the service knows of a worker: when it last spoke, and the messages that it
// holds now.
type WorkerStatus struct {
	Worker   string    `json:"worker"`
	Role     string    `json:"role,omitempty"`
	Kinds    []string  `json:"kinds,omitempty"`
	LastSeen time.Time `json:"last-seen,omitzero"`
	Holds    []Held    `json:"holds"`
}

// Held is a message that a worker holds, until its lease ends.
type Held struct {
	ID    string    `json:"id"`
	Until time.Time `json:"until"`
}

// Workers returns every worker that the log knows, by name, with its keep-alive and its leases.
func (s *Server) Workers() []WorkerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	var result []WorkerStatus
	for _, w := range s.state.workers {
		status := WorkerStatus{Worker: w.name, Role: w.role, Kinds: w.kinds, LastSeen: s.seen[w.name], Holds: []Held{}}
		for _, held := range s.leases {
			if held.worker == w.name && now.Before(held.until) {
				status.Holds = append(status.Holds, Held{ID: held.id, Until: held.until})
			}
		}
		slices.SortFunc(status.Holds, func(a, b Held) int { return strings.Compare(a.ID, b.ID) })
		result = append(result, status)
	}
	slices.SortFunc(result, func(a, b WorkerStatus) int { return strings.Compare(a.Worker, b.Worker) })
	return result
}

// Log returns the messages of the log with a seq after since.
func (s *Server) Log(since int64) []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []Message
	for _, m := range s.state.messages {
		if m.Seq > since {
			result = append(result, m)
		}
	}
	return result
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	var m Message
	if !decode(w, r, &m) {
		return
	}
	sent, err := s.Send(m)
	reply(w, sent, err)
}

func (s *Server) handleReceive(w http.ResponseWriter, r *http.Request) {
	var rec Receiver
	if !decode(w, r, &rec) {
		return
	}
	messages, err := s.Receive(r.Context(), rec)
	if messages == nil {
		messages = []Message{}
	}
	reply(w, messages, err)
}

func (s *Server) handleAck(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Worker string   `json:"worker"`
		IDs    []string `json:"ids"`
	}
	if !decode(w, r, &req) {
		return
	}
	reply(w, struct{}{}, s.Ack(req.Worker, req.IDs))
}

func (s *Server) handleRenew(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Worker string   `json:"worker"`
		IDs    []string `json:"ids"`
	}
	if !decode(w, r, &req) {
		return
	}
	renewal, err := s.Renew(req.Worker, req.IDs)
	reply(w, renewal, err)
}

func (s *Server) handleWorkers(w http.ResponseWriter, r *http.Request) {
	workers := s.Workers()
	if workers == nil {
		workers = []WorkerStatus{}
	}
	reply(w, workers, nil)
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	var since int64
	fmt.Sscan(r.URL.Query().Get("since"), &since)
	messages := s.Log(since)
	if messages == nil {
		messages = []Message{}
	}
	reply(w, messages, nil)
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		reply(w, nil, fmt.Errorf("The request is not valid JSON: %v.", err))
		return false
	}
	return true
}

func reply(w http.ResponseWriter, v any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, context.Canceled) {
			status = 499
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(v)
}

// newID returns a new message id: the time, then random hex, so that ids sort by time.
func newID(now time.Time) string {
	random := make([]byte, 6)
	rand.Read(random)
	return strings.ToLower(fmt.Sprintf("%s-%s", now.UTC().Format("20060102T150405.000000"), hex.EncodeToString(random)))
}
