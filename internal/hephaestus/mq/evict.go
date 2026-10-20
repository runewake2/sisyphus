package mq

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"
)

// eviction is the mark of an evicted worker. It expires after one lease: by then every lease that
// the worker held before the eviction would have ended anyway, and the name is free for a new
// worker. Evictions are in memory only, as the leases are: when the service restarts, the leases
// are gone, which is what an eviction makes happen.
type eviction struct {
	reason string
	until  time.Time
}

// Evicted is the answer of an eviction: the leases that it dropped.
type Evicted struct {
	Worker string `json:"worker"`
	// Dropped are the ids of the messages whose leases the worker held. They are deliverable to
	// another worker at once.
	Dropped []string `json:"dropped"`
	// Issues are the issues of the dropped messages, for the human to check with doctor.
	Issues []string `json:"issues,omitempty"`
}

// Evict drops every lease that the worker holds, so that its messages are deliverable to another
// worker at once, and marks the worker evicted: its next keep-alive tells it to stop its task, and
// its next receive is refused once, so the worker process ends instead of taking more work. The
// break glass for a worker that loops and never releases its lease.
func (s *Server) Evict(worker, reason string) (Evicted, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.workers[worker] == nil {
		return Evicted{}, fmt.Errorf("No worker %q has received a message: there is nothing to evict.", worker)
	}
	result := Evicted{Worker: worker, Dropped: []string{}}
	for key, held := range s.leases {
		if held.worker != worker {
			continue
		}
		result.Dropped = append(result.Dropped, held.id)
		if index, ok := s.state.byID[held.id]; ok {
			if issue := s.state.messages[index].Issue; issue != "" && !slices.Contains(result.Issues, issue) {
				result.Issues = append(result.Issues, issue)
			}
		}
		delete(s.leases, key)
	}
	slices.Sort(result.Dropped)
	slices.Sort(result.Issues)
	if reason == "" {
		reason = "a human evicted the worker"
	}
	s.evictions[worker] = eviction{reason: reason, until: s.now().Add(s.lease)}
	// Wake the receivers that wait: the dropped messages are deliverable, and the evicted worker's
	// own receive must end with the refusal.
	close(s.changed)
	s.changed = make(chan struct{})
	return result, nil
}

// evictionOf returns the reason of the eviction of the worker, or "" when it is not evicted. clear
// removes the mark, so that the refusal of a receive happens once. The caller holds mu.
func (s *Server) evictionOf(worker string, clear bool) string {
	e, ok := s.evictions[worker]
	if !ok {
		return ""
	}
	if !s.now().Before(e.until) {
		delete(s.evictions, worker)
		return ""
	}
	if clear {
		delete(s.evictions, worker)
	}
	return e.reason
}

func (s *Server) handleEvict(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Worker string `json:"worker"`
		Reason string `json:"reason"`
	}
	if !decode(w, r, &req) {
		return
	}
	evicted, err := s.Evict(req.Worker, req.Reason)
	reply(w, evicted, err)
}

// Evict tells the queue to evict a worker: drop its leases, stop its task through its keep-alive,
// and refuse its next receive.
func (c *Client) Evict(worker, reason string) (Evicted, error) {
	var evicted Evicted
	err := c.call(context.Background(), http.MethodPost, "/evict", map[string]string{"worker": worker, "reason": reason}, &evicted)
	return evicted, err
}
