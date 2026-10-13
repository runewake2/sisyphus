// Package mq is the message queue of hephaestus: a service that relays messages between workers
// (humans, AI agents, and tool agents) in every workspace of one repo
// ([[agent-delegation#Messages between workers]]).
//
//   - Every event (a message, a worker that registers, an acknowledgement) is appended to a log that
//     nothing changes or removes. The service rebuilds its state from the log when it starts.
//   - A message goes to one worker, to one worker of a role, to every worker of a role, or to every
//     listener as a broadcast. A listener can filter broadcasts by kind. A role can be in a family,
//     for example formatter/golang; a message to every worker of the family formatter reaches the
//     workers of each role in it.
//   - A broadcast reaches each listening role once: one worker of the role handles it, also a worker
//     that starts after the send, if a worker of its role listened before. A listener with no role
//     gets it as a worker of its own.
//   - Delivery is at least once: a message stays pending for its receiver until the receiver
//     acknowledges it. A message that is not acknowledged within the lease is given again. Each
//     message has an id, so that a receiver can ignore a copy.
//   - The messages of one sender reach each receiver in the order that the sender sent them: a
//     message is not given while an earlier message of its sender to the same receiver waits for
//     its acknowledgement.
package mq

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Address is the receiver of a message: exactly one of Worker, Role, and Broadcast.
type Address struct {
	// Worker is the owner name of one worker, for example merge-queue-1.
	Worker string `json:"worker,omitempty"`
	// Role is a role, for example engineer or formatter/golang. The message goes to one worker of the
	// role, or with All to every worker of the role. With All, a family, for example formatter, also
	// reaches every worker of each role in it.
	Role string `json:"role,omitempty"`
	All  bool   `json:"all,omitempty"`
	// Broadcast sends the message to every listener.
	Broadcast bool `json:"broadcast,omitempty"`
}

// String returns the address as people write it: a worker name, role:<role>, role:<role>:all, or
// broadcast.
func (a Address) String() string {
	switch {
	case a.Broadcast:
		return "broadcast"
	case a.Role != "" && a.All:
		return "role:" + a.Role + ":all"
	case a.Role != "":
		return "role:" + a.Role
	}
	return a.Worker
}

// Message is one message. The service sets Seq and Time; the sender can set ID, so that a send that
// it repeats after a failure does not make a second message.
type Message struct {
	ID    string          `json:"id"`
	Seq   int64           `json:"seq"`
	From  string          `json:"from"`
	To    Address         `json:"to"`
	Kind  string          `json:"kind"`
	Issue string          `json:"issue,omitempty"`
	Time  time.Time       `json:"time"`
	Body  string          `json:"body,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

var (
	namePattern = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*$`)
	// rolePattern is a role name: a name, or a path of names for a role of a family.
	rolePattern = regexp.MustCompile(`^[a-z0-9]+([._-][a-z0-9]+)*(/[a-z0-9]+([._-][a-z0-9]+)*)*$`)
)

// roleMatches reports whether a message to role reaches a worker of workerRole: the same role, or
// with all, a role in the family role.
func roleMatches(role, workerRole string, all bool) bool {
	return workerRole == role || (all && strings.HasPrefix(workerRole, role+"/"))
}

// validate checks a message before the service accepts it.
func (m Message) validate() error {
	var problems []string
	if !namePattern.MatchString(m.From) {
		problems = append(problems, fmt.Sprintf("from %q is not a worker name: lowercase words, for example merge-queue-1", m.From))
	}
	if !namePattern.MatchString(m.Kind) {
		problems = append(problems, fmt.Sprintf("kind %q is not a kind: lowercase words, for example task or info", m.Kind))
	}
	targets := 0
	if m.To.Worker != "" {
		targets++
		if !namePattern.MatchString(m.To.Worker) {
			problems = append(problems, fmt.Sprintf("worker %q is not a worker name", m.To.Worker))
		}
	}
	if m.To.Role != "" {
		targets++
		if !rolePattern.MatchString(m.To.Role) {
			problems = append(problems, fmt.Sprintf("role %q is not a role name: lowercase words, or a path of them for a role of a family, for example formatter/golang", m.To.Role))
		}
	}
	if m.To.Broadcast {
		targets++
	}
	if targets != 1 {
		problems = append(problems, "give exactly one receiver: a worker, a role, or broadcast")
	}
	if m.To.All && m.To.Role == "" {
		problems = append(problems, "all is only for a role: every worker of the role")
	}
	if len(m.Data) > 0 && !json.Valid(m.Data) {
		problems = append(problems, "data is not valid JSON")
	}
	if len(problems) > 0 {
		return errors.New("The message is not valid: " + joinSentences(problems))
	}
	return nil
}

func joinSentences(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "; "
		}
		out += p
	}
	return out + "."
}

// Event is one line of the log.
type Event struct {
	Type    string    `json:"type"` // message, register, or ack
	Seq     int64     `json:"seq"`
	Time    time.Time `json:"time"`
	Message *Message  `json:"message,omitempty"`
	Worker  string    `json:"worker,omitempty"`
	Role    string    `json:"role,omitempty"`
	Kinds   []string  `json:"kinds,omitempty"`
	IDs     []string  `json:"ids,omitempty"`
}

// worker is what the service knows of one receiver.
type worker struct {
	name  string
	role  string
	kinds []string
	// since is the seq of the first registration of the worker. Broadcasts and messages to all
	// workers of a role reach the workers that listened when they were sent.
	since int64
	acked map[string]bool
}

// state is the state of the queue, which the log rebuilds.
type state struct {
	seq      int64
	messages []Message
	byID     map[string]int
	workers  map[string]*worker
	// roleDone holds the messages to one worker of a role that a worker of the role acknowledged.
	roleDone map[string]bool
	// roleSince is the seq of the first registration of a worker of each role: a broadcast after it
	// reaches the role, also when no worker of the role runs at the send.
	roleSince map[string]int64
	// roleHeard holds the broadcasts that a worker of a role acknowledged, as role + "\x00" + id.
	roleHeard map[string]bool
}

func newState() *state {
	return &state{byID: map[string]int{}, workers: map[string]*worker{}, roleDone: map[string]bool{}, roleSince: map[string]int64{}, roleHeard: map[string]bool{}}
}

// apply changes the state for one event of the log.
func (s *state) apply(e Event) {
	if e.Seq > s.seq {
		s.seq = e.Seq
	}
	switch e.Type {
	case "message":
		m := *e.Message
		s.byID[m.ID] = len(s.messages)
		s.messages = append(s.messages, m)
	case "register":
		w, ok := s.workers[e.Worker]
		if !ok {
			w = &worker{name: e.Worker, since: e.Seq, acked: map[string]bool{}}
			s.workers[e.Worker] = w
		}
		w.role, w.kinds = e.Role, e.Kinds
		if _, ok := s.roleSince[e.Role]; e.Role != "" && !ok {
			s.roleSince[e.Role] = e.Seq
		}
	case "ack":
		w := s.workers[e.Worker]
		for _, id := range e.IDs {
			index, ok := s.byID[id]
			if !ok {
				continue
			}
			switch m := s.messages[index]; {
			case m.To.Role != "" && !m.To.All:
				s.roleDone[id] = true
			case m.To.Broadcast && w != nil && w.role != "":
				s.roleHeard[w.role+"\x00"+id] = true
			case w != nil:
				w.acked[id] = true
			}
		}
	}
}

// pendingFor reports whether message m waits for worker w: it is addressed to w, and w has not
// acknowledged it.
func (s *state) pendingFor(w *worker, m Message) bool {
	switch {
	case m.To.Worker != "":
		return m.To.Worker == w.name && !w.acked[m.ID]
	case m.To.Role != "" && !m.To.All:
		return m.To.Role == w.role && !s.roleDone[m.ID]
	case m.To.Role != "":
		return roleMatches(m.To.Role, w.role, true) && m.Seq > w.since && !w.acked[m.ID]
	}
	// A broadcast: to every listener but its sender, filtered by kind; once for each role.
	if m.From == w.name || (len(w.kinds) > 0 && !slices.Contains(w.kinds, m.Kind)) {
		return false
	}
	if w.role != "" {
		return m.Seq > s.roleSince[w.role] && !s.roleHeard[w.role+"\x00"+m.ID]
	}
	return m.Seq > w.since && !w.acked[m.ID]
}

// leaseKey identifies the delivery of message m to worker w. A message to one worker of a role, and
// a broadcast to a worker with a role, have one lease for the role, so that only one worker of the
// role has it at a time.
func leaseKey(w *worker, m Message) string {
	if m.To.Role != "" && !m.To.All {
		return "role:" + m.To.Role + ":" + m.ID
	}
	if m.To.Broadcast && w.role != "" {
		return "role:" + w.role + ":" + m.ID
	}
	return "worker:" + w.name + ":" + m.ID
}
