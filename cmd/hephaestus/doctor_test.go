package main

import (
	"strings"
	"testing"
)

func TestDoctorFindsAStuckTaskAndKickRestartsIt(t *testing.T) {
	r := newMergeRepo(t, "stuck-task-work")
	startQueue(t, r)
	r.sisyphus("update", "stuck-task-work", "--metadata", "delegate=implementer")
	r.sisyphus("start", "stuck-task-work", "--owner", "implementer-1")

	// The owner never spoke to the queue: the doctor finds the stuck task, with exit code 2.
	res := r.run("doctor", "stuck-task-work")
	if res.exit != 2 {
		t.Fatalf("doctor on a stuck task: want exit code 2, got %d: %s%s", res.exit, res.output, res.errors)
	}
	contains(t, res.output, "does not know its owner implementer-1")
	contains(t, res.output, "hephaestus kick stuck-task-work")

	// The kick reopens the issue and sends the task back to the role.
	kick := r.mustRun("kick", "stuck-task-work", "--reason", "The worker ran out of API keys.")
	contains(t, kick.output, "reopened the issue")
	contains(t, kick.output, "role:implementer")
	equal(t, "open", r.showIssue("stuck-task-work")["state"])
	contains(t, r.showIssue("stuck-task-work")["body"].(string), "human kicked the task to implementer")
	contains(t, r.showIssue("stuck-task-work")["body"].(string), "The worker ran out of API keys.")

	// A worker of the role gets the task.
	messages := receivedJSON(t, r, "--as", "implementer-2", "--role", "implementer", "--kind", "task")
	equal(t, 1, len(messages))
	equal(t, "task", messages[0].Kind)
	equal(t, "stuck-task-work", messages[0].Issue)
	contains(t, messages[0].Body, "kicked")

	// Now a live worker holds the task: the doctor is calm, and a kick without --force refuses.
	res = r.mustRun("doctor", "stuck-task-work")
	contains(t, res.output, "none: the task is not stuck")
	res = r.run("kick", "stuck-task-work")
	if res.exit != 1 || !strings.Contains(res.errors, "--force") {
		t.Fatalf("kick of a live task: want a refusal that names --force, got exit %d: %s", res.exit, res.errors)
	}
}

func TestKickClearsNeedsHumanAndKeepsAWaitingMessage(t *testing.T) {
	r := newMergeRepo(t, "kicked-task-work")
	startQueue(t, r)
	// A worker stopped the task for a human, and its task message still waits for the role.
	r.mustRun("messages", "send", "--from", "alice", "--role", "implementer", "--kind", "task", "--issue", "kicked-task-work", "Do the task.")
	r.sisyphus("update", "kicked-task-work", "--metadata", "delegate=implementer", "--add-tags", "needs-human")

	kick := r.mustRun("kick", "kicked-task-work", "--as", "alice")
	contains(t, kick.output, "removed needs-human")
	contains(t, kick.errors, "already waits")
	issue := r.showIssue("kicked-task-work")
	if tags, ok := issue["tags"].([]any); ok {
		for _, tag := range tags {
			if tag == "needs-human" {
				t.Fatal("the kick did not remove needs-human")
			}
		}
	}
	// No second message: the one that waited is the one a worker gets.
	messages := receivedJSON(t, r, "--as", "implementer-1", "--role", "implementer", "--kind", "task")
	equal(t, 1, len(messages))
	equal(t, "Do the task.", messages[0].Body)
}

func TestDoctorWithoutAQueue(t *testing.T) {
	r := newMergeRepo(t, "quiet-task-work")
	r.sisyphus("update", "quiet-task-work", "--metadata", "delegate=implementer")
	res := r.run("doctor", "quiet-task-work")
	if res.exit != 2 {
		t.Fatalf("doctor without a queue: want exit code 2, got %d: %s%s", res.exit, res.output, res.errors)
	}
	contains(t, res.output, "No message queue runs")
	contains(t, res.output, "hephaestus messages serve")
}
