package mq

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Pending is one delivery that waits: a message that its receiver, a worker or a role, has not
// acknowledged yet.
type Pending struct {
	ID    string    `json:"id"`
	Seq   int64     `json:"seq"`
	Kind  string    `json:"kind"`
	From  string    `json:"from"`
	To    string    `json:"to"`
	Issue string    `json:"issue,omitempty"`
	Sent  time.Time `json:"sent"`
	// For is the receiver that it waits for: worker:<name> or role:<role>.
	For string `json:"for"`
	// LeasedBy is the worker that has it now, if a lease holds it.
	LeasedBy   string    `json:"leased-by,omitempty"`
	LeaseUntil time.Time `json:"lease-until,omitzero"`
}

// Pending returns every delivery that waits, oldest first. A message to one worker or to one worker
// of a role waits also when that worker or no worker of that role has started yet. A broadcast and a
// message to every worker of a role wait only for the receivers that they reach: the roles and
// workers that listened before the send.
func (s *Server) Pending() []Pending {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	st := s.state
	names := make([]string, 0, len(st.workers))
	for name := range st.workers {
		names = append(names, name)
	}
	slices.Sort(names)
	var result []Pending
	add := func(m Message, waitsFor, key string) {
		p := Pending{ID: m.ID, Seq: m.Seq, Kind: m.Kind, From: m.From, To: m.To.String(), Issue: m.Issue, Sent: m.Time, For: waitsFor}
		if held, ok := s.leases[key]; ok && now.Before(held.until) {
			p.LeasedBy, p.LeaseUntil = held.worker, held.until
		}
		result = append(result, p)
	}
	for _, m := range st.messages {
		switch {
		case m.To.Worker != "":
			if w := st.workers[m.To.Worker]; w == nil || !w.acked[m.ID] {
				add(m, "worker:"+m.To.Worker, "worker:"+m.To.Worker+":"+m.ID)
			}
		case m.To.Role != "" && !m.To.All:
			if !st.roleDone[m.ID] {
				add(m, "role:"+m.To.Role, "role:"+m.To.Role+":"+m.ID)
			}
		default:
			heard := map[string]bool{}
			for _, name := range names {
				w := st.workers[name]
				if !st.pendingFor(w, m) {
					continue
				}
				if m.To.Broadcast && w.role != "" {
					if !heard[w.role] {
						heard[w.role] = true
						add(m, "role:"+w.role, leaseKey(w, m))
					}
					continue
				}
				add(m, "worker:"+w.name, leaseKey(w, m))
			}
		}
	}
	return result
}

// ReadLog reads the messages of the log in the shared directory dir, oldest first. It works also
// when the service does not run. A log that does not exist has no message.
func ReadLog(dir string) ([]Message, error) {
	file, err := os.Open(filepath.Join(dir, LogFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var messages []Message
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for scanner.Scan() {
		var e Event
		if json.Unmarshal(scanner.Bytes(), &e) == nil && e.Type == "message" && e.Message != nil {
			messages = append(messages, *e.Message)
		}
	}
	if err := scanner.Err(); err != nil {
		return messages, fmt.Errorf("%s: %v", strings.TrimSpace(LogFile), err)
	}
	return messages, nil
}
