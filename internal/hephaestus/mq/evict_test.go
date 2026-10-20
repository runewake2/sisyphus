package mq

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestEvictDropsTheLeasesAndStopsTheWorker(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)}
	s := openServer(t, t.TempDir(), clock)
	m := send(t, s, "human", Address{Role: "implementer"}, "task", "Do the task.")
	if got := receive(t, s, "implementer-1", "implementer"); len(got) != 1 {
		t.Fatalf("the worker did not get the task: %v", got)
	}

	evicted, err := s.Evict("implementer-1", "it loops")
	if err != nil {
		t.Fatal(err)
	}
	if len(evicted.Dropped) != 1 || evicted.Dropped[0] != m.ID {
		t.Fatalf("the eviction dropped %v, not the lease of %s", evicted.Dropped, m.ID)
	}

	// The keep-alive of the evicted worker says so, and renews nothing.
	renewal, err := s.Renew("implementer-1", []string{m.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !renewal.Evicted || renewal.Reason != "it loops" || len(renewal.Renewed) != 0 {
		t.Fatalf("the renewal of the evicted worker is %+v", renewal)
	}

	// The message is deliverable to another worker at once, without waiting for the lease.
	if got := receive(t, s, "implementer-2", "implementer"); len(got) != 1 {
		t.Fatalf("another worker did not get the dropped task: %v", got)
	}
}

func TestAnEvictedReceiveIsRefusedOnce(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)}
	s := openServer(t, t.TempDir(), clock)
	send(t, s, "human", Address{Role: "implementer"}, "task", "Do the task.")
	receive(t, s, "implementer-1", "implementer")
	if _, err := s.Evict("implementer-1", ""); err != nil {
		t.Fatal(err)
	}

	// The eviction mark is consumed by the first contact: here a receive, refused once.
	if _, err := s.Receive(context.Background(), Receiver{Worker: "implementer-1", Role: "implementer"}); err == nil || !strings.Contains(err.Error(), "evicted") {
		t.Fatalf("the receive of the evicted worker is not refused: %v", err)
	}
	if _, err := s.Receive(context.Background(), Receiver{Worker: "implementer-1", Role: "implementer"}); err != nil {
		t.Fatalf("the second receive is still refused: %v", err)
	}
}

func TestAnEvictionExpiresAfterOneLease(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)}
	s := openServer(t, t.TempDir(), clock)
	send(t, s, "human", Address{Role: "implementer"}, "task", "Do the task.")
	receive(t, s, "implementer-1", "implementer")
	if _, err := s.Evict("implementer-1", ""); err != nil {
		t.Fatal(err)
	}
	clock.Add(2 * time.Minute) // past the lease of one minute
	if _, err := s.Receive(context.Background(), Receiver{Worker: "implementer-1", Role: "implementer"}); err != nil {
		t.Fatalf("the eviction did not expire: %v", err)
	}
}

func TestEvictNeedsAKnownWorker(t *testing.T) {
	s := openServer(t, t.TempDir(), nil)
	if _, err := s.Evict("nobody", ""); err == nil || !strings.Contains(err.Error(), "nothing to evict") {
		t.Fatalf("an unknown worker is not refused: %v", err)
	}
}
