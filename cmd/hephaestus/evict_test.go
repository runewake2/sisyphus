package main

import (
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestEvictStopsALoopingWorkerAndFreesItsTask(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the command of the test role is sh")
	}
	r := newTrackerRepo(t)
	t.Setenv(ownerVariable, "")
	r.write("hephaestus/roles/looper.md", toolRoleFile("looper", "sleep 30", "", ""))
	r.sisyphus("new", "loop-task-work")
	r.commit("the looper role")
	startQueueWithLease(t, r, time.Second) // a short lease, so the keep-alive renews often and hears the eviction fast
	r.mustRun("messages", "send", "--from", "alice", "--role", "looper", "--kind", "task", "--issue", "loop-task-work", "Loop.")

	// The worker takes the task and loops in it.
	type agentDone struct{ res result }
	done := make(chan agentDone, 1)
	go func() {
		done <- agentDone{res: r.run("agent", "run", "looper", "--as", "looper-1", "--once", "--json")}
	}()
	waitFor(t, "the worker claims the issue", func() bool {
		issue := r.showIssue("loop-task-work")
		return issue["state"] == "in-progress"
	})

	// Without --force, evict refuses and only tells.
	res := r.run("evict", "looper-1")
	if res.exit != 1 || !strings.Contains(res.errors, "--force") {
		t.Fatalf("evict without --force: want a refusal that names --force, got exit %d: %s", res.exit, res.errors)
	}
	contains(t, res.output, "looper-1")

	// With --force, the task stops, the worker ends, and the issue is free again.
	forced := r.mustRun("evict", "looper-1", "--force", "--reason", "it loops")
	contains(t, forced.output, "looper-1 is evicted")
	contains(t, forced.output, "hephaestus doctor loop-task-work")
	agent := <-done
	if agent.res.exit == 0 {
		t.Fatalf("the evicted worker ended with exit code 0:\n%s", agent.res.errors)
	}
	contains(t, agent.res.output, `"status": "evicted"`)
	issue := r.showIssue("loop-task-work")
	equal(t, "open", issue["state"])
	if tags, ok := issue["tags"].([]any); ok && slices.Contains(tags, any("needs-human")) {
		t.Fatal("the eviction tagged the issue needs-human: a human already acts")
	}
	contains(t, issue["body"].(string), "a human evicted the worker looper-1")
	contains(t, issue["body"].(string), "it loops")

	// The task message was not acknowledged: another worker gets it at once.
	messages := receivedJSON(t, r, "--as", "looper-2", "--role", "looper", "--kind", "task")
	equal(t, 1, len(messages))
	equal(t, "loop-task-work", messages[0].Issue)
}

// waitFor polls until the condition holds, or fails the test after a few seconds.
func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("waiting for %s: it did not happen", what)
}
