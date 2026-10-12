package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
)

// startQueue runs the message queue of the repo of r for the test.
func startQueue(t *testing.T, r *testRepo) string {
	t.Helper()
	return startQueueWithLease(t, r, time.Minute)
}

func startQueueWithLease(t *testing.T, r *testRepo, lease time.Duration) string {
	t.Helper()
	found, err := repo.Find(r.root)
	if err != nil {
		t.Fatal(err)
	}
	server, err := mq.Open(found.SharedDir(), lease)
	if err != nil {
		t.Fatal(err)
	}
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { server.Close() })
	return found.SharedDir()
}

func newQueueRepo(t *testing.T) *testRepo {
	t.Helper()
	t.Setenv(ownerVariable, "")
	r := newGitRepo(t)
	r.write("README.md", "# Readme\n")
	r.commit("main")
	startQueue(t, r)
	return r
}

func receivedJSON(t *testing.T, r *testRepo, args ...string) []mq.Message {
	t.Helper()
	res := r.mustRun(append([]string{"messages", "receive", "--json"}, args...)...)
	var messages []mq.Message
	if err := json.Unmarshal([]byte(res.output), &messages); err != nil {
		t.Fatalf("%v\n%s", err, res.output)
	}
	return messages
}

func TestSendAndReceiveADirectMessage(t *testing.T) {
	r := newQueueRepo(t)

	id := strings.TrimSpace(r.mustRun("messages", "send", "--from", "alice", "--to", "merge-queue-1", "--kind", "task", "--issue", "fix-login-bug", "--data", `{"tasks":["fix-login-bug"]}`, "Merge the work.").output)
	got := receivedJSON(t, r, "--as", "merge-queue-1")

	equal(t, 1, len(got))
	equal(t, id, got[0].ID)
	equal(t, "alice", got[0].From)
	equal(t, "task", got[0].Kind)
	equal(t, "fix-login-bug", got[0].Issue)
	equal(t, "Merge the work.", got[0].Body)
	var data bytes.Buffer
	if err := json.Compact(&data, got[0].Data); err != nil {
		t.Fatal(err)
	}
	equal(t, `{"tasks":["fix-login-bug"]}`, data.String())
}

func TestReceiveTextShowsTheMessage(t *testing.T) {
	r := newQueueRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--to", "w-1", "--kind", "info", "Line one\nLine two")

	res := r.mustRun("messages", "receive", "--as", "w-1")

	contains(t, res.output, "  alice -> w-1  [info]\n    Line one\n    Line two\n")
}

func TestAckStopsAMessageFromComingAgain(t *testing.T) {
	r := newQueueRepo(t)
	t.Setenv(ownerVariable, "w-1")
	id := strings.TrimSpace(r.mustRun("messages", "send", "--from", "alice", "--to", "w-1", "Do it.").output)
	receivedJSON(t, r)

	r.mustRun("messages", "ack", id)

	equal(t, 0, len(receivedJSON(t, r)))
}

func TestRoleAndBroadcastMessagesFromTheCommandLine(t *testing.T) {
	r := newQueueRepo(t)
	receivedJSON(t, r, "--as", "engineer-1", "--role", "engineer")
	receivedJSON(t, r, "--as", "engineer-2", "--role", "engineer")
	receivedJSON(t, r, "--as", "lint-1", "--kind", "bookmark-ready")

	r.mustRun("messages", "send", "--from", "alice", "--role", "engineer", "--all", "main moved")
	r.mustRun("messages", "send", "--from", "merge-queue-1", "--broadcast", "--kind", "bookmark-ready", "ai/fix is ready")
	r.mustRun("messages", "send", "--from", "merge-queue-1", "--broadcast", "--kind", "other", "not for lint")

	// engineer-1 listens for every kind: the role message and both broadcasts.
	equal(t, 3, len(receivedJSON(t, r, "--as", "engineer-1", "--role", "engineer")))
	lint := receivedJSON(t, r, "--as", "lint-1", "--kind", "bookmark-ready")
	equal(t, 1, len(lint))
	equal(t, "ai/fix is ready", lint[0].Body)
}

func TestSendNeedsOneReceiverAndASender(t *testing.T) {
	r := newQueueRepo(t)

	none := r.run("messages", "send", "--from", "alice", "text")
	two := r.run("messages", "send", "--from", "alice", "--to", "w-1", "--broadcast", "text")
	noSender := r.run("messages", "send", "--to", "w-1", "text")

	contains(t, none.errors, "Give exactly one receiver")
	contains(t, two.errors, "Give exactly one receiver")
	contains(t, noSender.errors, "Give --from <worker>, or set SISYPHUS_OWNER.")
}

func TestARepeatedSendWithAnIDIsOneMessage(t *testing.T) {
	r := newQueueRepo(t)

	r.mustRun("messages", "send", "--from", "alice", "--to", "w-1", "--id", "task-42", "once")
	r.mustRun("messages", "send", "--from", "alice", "--to", "w-1", "--id", "task-42", "once")

	equal(t, 1, len(receivedJSON(t, r, "--as", "w-1")))
}

func TestWithoutTheServiceCommandsSaySo(t *testing.T) {
	r := newGitRepo(t)

	res := r.run("messages", "send", "--from", "alice", "--to", "w-1", "text")

	equal(t, 1, res.exit)
	contains(t, res.errors, "No message queue runs for this repo. Start it with: hephaestus messages serve")
}

func TestLogReadsTheFileAlsoWithoutTheService(t *testing.T) {
	r := newGitRepo(t)
	found, _ := repo.Find(r.root)
	server, err := mq.Open(found.SharedDir(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	server.Send(mq.Message{From: "alice", To: mq.Address{Worker: "w-1"}, Kind: "info", Body: "kept"})
	server.Close()

	res := r.mustRun("messages", "log")

	contains(t, res.output, "alice -> w-1  [info]\n    kept\n")
}

func TestMessagesReachAnotherWorktreeOfTheRepo(t *testing.T) {
	r := newQueueRepo(t)
	other := &testRepo{t: t, root: filepath.Join(t.TempDir(), "other")}
	r.git("worktree", "add", "-q", other.root)

	r.mustRun("messages", "send", "--from", "alice", "--to", "w-1", "from the main worktree")
	got := receivedJSON(t, other, "--as", "w-1")

	equal(t, 1, len(got))
	equal(t, "from the main worktree", got[0].Body)
}

func TestListenRunsTheCommandForEachMessage(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the command is a shell script")
	}
	r := newQueueRepo(t)
	out := filepath.Join(t.TempDir(), "handled")
	script := filepath.Join(t.TempDir(), "agent.sh")
	r.write(relativeTo(t, r.root, script), "#!/bin/sh\nread message\necho \"$HEPHAESTUS_MESSAGE_KIND $SISYPHUS_OWNER $HEPHAESTUS_MESSAGE_ISSUE\" >> "+out+"\ncase \"$message\" in *fail*) exit 1;; esac\n")
	if err := os.Chmod(script, 0o755); err != nil {
		t.Fatal(err)
	}
	r.mustRun("messages", "receive", "--as", "lint-1", "--kind", "bookmark-ready") // listen from now on
	r.mustRun("messages", "send", "--from", "merge-queue-1", "--broadcast", "--kind", "bookmark-ready", "--issue", "fix-login-bug", "ai/fix is ready")
	r.mustRun("messages", "send", "--from", "merge-queue-1", "--broadcast", "--kind", "bookmark-ready", "--issue", "other-task", "this will fail")

	res := r.mustRun("messages", "listen", "--as", "lint-1", "--kind", "bookmark-ready", "--once", "--", script)

	handled, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "bookmark-ready lint-1 fix-login-bug\nbookmark-ready lint-1 other-task\n", string(handled))
	contains(t, res.errors, "failed for message")
	// The failed message stays pending: it comes again after the lease, not now.
	equal(t, 0, len(receivedJSON(t, r, "--as", "lint-1", "--kind", "bookmark-ready")))
}

// relativeTo returns path relative to root, for testRepo.write, which takes a path from the root.
func relativeTo(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}
