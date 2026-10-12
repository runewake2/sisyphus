package mq

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeClock is a clock that a test moves.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Add(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func openServer(t *testing.T, dir string, clock *fakeClock) *Server {
	t.Helper()
	s, err := Open(dir, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if clock != nil {
		s.now = clock.Now
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func send(t *testing.T, s *Server, from string, to Address, kind, body string) Message {
	t.Helper()
	m, err := s.Send(Message{From: from, To: to, Kind: kind, Body: body})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func receive(t *testing.T, s *Server, worker, role string, kinds ...string) []string {
	t.Helper()
	messages, err := s.Receive(context.Background(), Receiver{Worker: worker, Role: role, Kinds: kinds})
	if err != nil {
		t.Fatal(err)
	}
	var bodies []string
	for _, m := range messages {
		bodies = append(bodies, m.Body)
	}
	return bodies
}

func ack(t *testing.T, s *Server, worker string, messages ...Message) {
	t.Helper()
	var ids []string
	for _, m := range messages {
		ids = append(ids, m.ID)
	}
	if err := s.Ack(worker, ids); err != nil {
		t.Fatal(err)
	}
}

func equalBodies(t *testing.T, want, got []string) {
	t.Helper()
	if strings.Join(want, "|") != strings.Join(got, "|") {
		t.Errorf("want %q, got %q", want, got)
	}
}

func newClock() *fakeClock { return &fakeClock{now: time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)} }

func TestADirectMessageReachesItsWorkerOnly(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	receive(t, s, "engineer-1", "engineer")
	m := send(t, s, "alice", Address{Worker: "merge-queue-1"}, "task", "merge fix-login-bug")

	equalBodies(t, nil, receive(t, s, "engineer-1", "engineer"))
	equalBodies(t, []string{"merge fix-login-bug"}, receive(t, s, "merge-queue-1", "merge-queue"))
	ack(t, s, "merge-queue-1", m)
}

func TestAMessageWaitsForAWorkerThatHasNotStarted(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	send(t, s, "alice", Address{Worker: "merge-queue-1"}, "task", "first")

	equalBodies(t, []string{"first"}, receive(t, s, "merge-queue-1", ""))
}

func TestAnUnacknowledgedMessageComesAgainAfterTheLease(t *testing.T) {
	clock := newClock()
	s := openServer(t, t.TempDir(), clock)
	m := send(t, s, "alice", Address{Worker: "lint-1"}, "task", "lint it")

	equalBodies(t, []string{"lint it"}, receive(t, s, "lint-1", ""))
	equalBodies(t, nil, receive(t, s, "lint-1", "")) // leased
	clock.Add(2 * time.Minute)
	again, _ := s.Receive(context.Background(), Receiver{Worker: "lint-1"})
	if len(again) != 1 || again[0].ID != m.ID {
		t.Fatalf("want the same message again, with the same id, got %v", again)
	}
	ack(t, s, "lint-1", m)
	clock.Add(2 * time.Minute)
	equalBodies(t, nil, receive(t, s, "lint-1", ""))
}

func TestARoleMessageReachesOneWorkerOfTheRole(t *testing.T) {
	clock := newClock()
	s := openServer(t, t.TempDir(), clock)
	receive(t, s, "engineer-1", "engineer")
	receive(t, s, "engineer-2", "engineer")
	m := send(t, s, "alice", Address{Role: "engineer"}, "task", "align the work")

	first := receive(t, s, "engineer-1", "engineer")
	second := receive(t, s, "engineer-2", "engineer")
	equalBodies(t, []string{"align the work"}, first)
	equalBodies(t, nil, second)

	// The first worker does not acknowledge it: after the lease, another worker of the role gets it.
	clock.Add(2 * time.Minute)
	equalBodies(t, []string{"align the work"}, receive(t, s, "engineer-2", "engineer"))
	ack(t, s, "engineer-2", m)
	clock.Add(2 * time.Minute)
	equalBodies(t, nil, receive(t, s, "engineer-1", "engineer"))
}

func TestARoleMessageToAllReachesEveryWorkerOfTheRole(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	receive(t, s, "engineer-1", "engineer")
	receive(t, s, "engineer-2", "engineer")
	receive(t, s, "docs-1", "technical-docs")
	send(t, s, "alice", Address{Role: "engineer", All: true}, "info", "main moved")

	equalBodies(t, []string{"main moved"}, receive(t, s, "engineer-1", "engineer"))
	equalBodies(t, []string{"main moved"}, receive(t, s, "engineer-2", "engineer"))
	equalBodies(t, nil, receive(t, s, "docs-1", "technical-docs"))
}

func TestABroadcastReachesEveryListenerOfItsKind(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	receive(t, s, "lint-1", "", "bookmark-ready")
	receive(t, s, "docs-1", "", "issue-closed")
	receive(t, s, "human-alice", "")
	send(t, s, "merge-queue-1", Address{Broadcast: true}, "bookmark-ready", "ai/fix is ready")

	equalBodies(t, []string{"ai/fix is ready"}, receive(t, s, "lint-1", "", "bookmark-ready"))
	equalBodies(t, nil, receive(t, s, "docs-1", "", "issue-closed"))
	equalBodies(t, []string{"ai/fix is ready"}, receive(t, s, "human-alice", ""))
	equalBodies(t, nil, receive(t, s, "merge-queue-1", "")) // not to its sender
}

func TestABroadcastBeforeAListenerStartedDoesNotReachIt(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	send(t, s, "alice", Address{Broadcast: true}, "info", "old news")
	receive(t, s, "late-1", "")
	send(t, s, "alice", Address{Broadcast: true}, "info", "new news")

	equalBodies(t, []string{"new news"}, receive(t, s, "late-1", ""))
}

func TestTheMessagesOfOneSenderKeepTheirOrder(t *testing.T) {
	clock := newClock()
	s := openServer(t, t.TempDir(), clock)
	first := send(t, s, "alice", Address{Worker: "w-1"}, "task", "first")
	equalBodies(t, []string{"first"}, receive(t, s, "w-1", ""))
	second := send(t, s, "alice", Address{Worker: "w-1"}, "task", "second")
	other := send(t, s, "bob", Address{Worker: "w-1"}, "task", "from bob")

	// first is leased and not acknowledged: second waits behind it, but bob's message does not.
	equalBodies(t, []string{"from bob"}, receive(t, s, "w-1", ""))
	ack(t, s, "w-1", other)
	clock.Add(2 * time.Minute)
	equalBodies(t, []string{"first", "second"}, receive(t, s, "w-1", ""))
	ack(t, s, "w-1", first, second)
}

func TestARepeatedSendWithTheSameIDAddsNothing(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	m := Message{ID: "task-42", From: "alice", To: Address{Worker: "w-1"}, Kind: "task", Body: "once"}

	first, err := s.Send(m)
	if err != nil {
		t.Fatal(err)
	}
	copy, err := s.Send(m)
	if err != nil {
		t.Fatal(err)
	}

	if first.Seq != copy.Seq {
		t.Errorf("want the first message back, got seq %d and %d", first.Seq, copy.Seq)
	}
	equalBodies(t, []string{"once"}, receive(t, s, "w-1", ""))
}

func TestTheLogRebuildsTheStateAfterARestart(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	handled := send(t, s, "alice", Address{Worker: "w-1"}, "task", "handled")
	send(t, s, "alice", Address{Worker: "w-1"}, "task", "pending")
	receive(t, s, "w-1", "")
	ack(t, s, "w-1", handled)
	s.Close()

	again := openServer(t, dir, nil)

	// The lease is gone with the old service, so the pending message comes again.
	equalBodies(t, []string{"pending"}, receive(t, again, "w-1", ""))
	log, err := os.ReadFile(filepath.Join(dir, LogFile))
	if err != nil {
		t.Fatal(err)
	}
	// Append-only: the handled message is still in the log, with its acknowledgement.
	if n := strings.Count(string(log), `"type":"message"`); n != 2 {
		t.Errorf("want 2 messages in the log, got %d", n)
	}
	if n := strings.Count(string(log), `"type":"ack"`); n != 1 {
		t.Errorf("want 1 acknowledgement in the log, got %d", n)
	}
}

func TestAnInvalidMessageIsRefused(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())

	_, err := s.Send(Message{From: "Alice", To: Address{Worker: "w-1", Broadcast: true}, Kind: ""})

	if err == nil {
		t.Fatal("want an error")
	}
	for _, part := range []string{`from "Alice" is not a worker name`, `kind "" is not a kind`, "exactly one receiver"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("want %q in %q", part, err)
		}
	}
}

func TestAReceiveWaitsForAMessage(t *testing.T) {
	s := openServer(t, t.TempDir(), nil)
	done := make(chan []Message)
	go func() {
		messages, _ := s.Receive(context.Background(), Receiver{Worker: "w-1", Wait: 5 * time.Second})
		done <- messages
	}()
	time.Sleep(100 * time.Millisecond)
	send(t, s, "alice", Address{Worker: "w-1"}, "task", "now")

	select {
	case messages := <-done:
		if len(messages) != 1 {
			t.Fatalf("got %v", messages)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the receive did not end when the message came")
	}
}

func TestTheServiceNeedsTheToken(t *testing.T) {
	dir := t.TempDir()
	s := openServer(t, dir, nil)
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get("http://" + s.Address() + "/health")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	client, err := Connect(dir)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401 without the token, got %d", resp.StatusCode)
	}
	if _, err := client.Send(Message{From: "alice", To: Address{Worker: "w-1"}, Kind: "task"}); err != nil {
		t.Errorf("with the token: %v", err)
	}
	if err := (func() error { second := openServer(t, dir, nil); return second.Start() })(); err == nil || !strings.Contains(err.Error(), "already runs") {
		t.Errorf("want a second service refused, got %v", err)
	}
}

func TestARoleOfAFamilyGetsItsMessages(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	receive(t, s, "formatter-golang-1", "formatter/golang")
	receive(t, s, "formatter-yaml-1", "formatter/yaml")
	receive(t, s, "engineer-1", "engineer")
	send(t, s, "alice", Address{Role: "formatter/golang"}, "task", "format the Go files")
	send(t, s, "alice", Address{Role: "formatter", All: true}, "info", "to every formatter")

	equalBodies(t, []string{"format the Go files", "to every formatter"}, receive(t, s, "formatter-golang-1", "formatter/golang"))
	equalBodies(t, []string{"to every formatter"}, receive(t, s, "formatter-yaml-1", "formatter/yaml"))
	equalBodies(t, nil, receive(t, s, "engineer-1", "engineer"))
}

func TestAMessageToOneWorkerOfAFamilyNeedsTheRole(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	receive(t, s, "formatter-golang-1", "formatter/golang")
	send(t, s, "alice", Address{Role: "formatter"}, "task", "which formatter?")

	equalBodies(t, nil, receive(t, s, "formatter-golang-1", "formatter/golang"))
}

func TestABroadcastReachesEachListeningRoleOnce(t *testing.T) {
	clock := newClock()
	s := openServer(t, t.TempDir(), clock)
	receive(t, s, "check-1", "repo-check", "bookmark-ready")
	receive(t, s, "check-2", "repo-check", "bookmark-ready")
	receive(t, s, "reviewer-1", "code-reviewer", "bookmark-ready")
	m := send(t, s, "implementer-1", Address{Broadcast: true}, "bookmark-ready", "ai/fix is ready")

	equalBodies(t, []string{"ai/fix is ready"}, receive(t, s, "check-1", "repo-check", "bookmark-ready"))
	equalBodies(t, nil, receive(t, s, "check-2", "repo-check", "bookmark-ready")) // one worker of the role
	equalBodies(t, []string{"ai/fix is ready"}, receive(t, s, "reviewer-1", "code-reviewer", "bookmark-ready"))
	ack(t, s, "check-1", m)
	clock.Add(2 * time.Minute)
	equalBodies(t, nil, receive(t, s, "check-2", "repo-check", "bookmark-ready"))
}

func TestABroadcastWaitsForAWorkerOfARoleThatListenedBefore(t *testing.T) {
	s := openServer(t, t.TempDir(), newClock())
	receive(t, s, "check-1", "repo-check", "bookmark-ready")
	send(t, s, "alice", Address{Broadcast: true}, "bookmark-ready", "before any role")
	// check-1 stops; a new worker of the role starts after the send.
	equalBodies(t, []string{"before any role"}, receive(t, s, "check-9", "repo-check", "bookmark-ready"))
	// A role that never listened before the send does not get it.
	equalBodies(t, nil, receive(t, s, "docs-1", "technical-docs", "bookmark-ready"))
}

func TestAKeepAliveKeepsALongTaskFromOtherWorkers(t *testing.T) {
	clock := newClock()
	s := openServer(t, t.TempDir(), clock)
	receive(t, s, "slow-1", "slow")
	receive(t, s, "slow-2", "slow")
	m := send(t, s, "alice", Address{Role: "slow"}, "task", "a long task")
	equalBodies(t, []string{"a long task"}, receive(t, s, "slow-1", "slow"))

	for range 5 { // five leases long, renewed within each
		clock.Add(40 * time.Second)
		renewal, err := s.Renew("slow-1", []string{m.ID})
		if err != nil || len(renewal.Renewed) != 1 || renewal.Lease != time.Minute {
			t.Fatalf("want the lease renewed, got %+v, %v", renewal, err)
		}
		equalBodies(t, nil, receive(t, s, "slow-2", "slow"))
	}
	status := s.Workers()
	equal(t, "slow-1", status[0].Worker)
	equal(t, clock.Now(), status[0].LastSeen)
	equal(t, m.ID, status[0].Holds[0].ID)

	// A worker that stops renewing loses the message to another worker, and is told so.
	clock.Add(2 * time.Minute)
	equalBodies(t, []string{"a long task"}, receive(t, s, "slow-2", "slow"))
	renewal, _ := s.Renew("slow-1", []string{m.ID})
	equal(t, 0, len(renewal.Renewed))
	equal(t, m.ID, renewal.Lost[0])
}

func equal[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}
