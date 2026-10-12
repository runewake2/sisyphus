package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// newNoticeRepo makes a tracker repo with listening tool roles, a family role, and a slow role, and a
// queue with lease.
func newNoticeRepo(t *testing.T, lease time.Duration) *testRepo {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the commands of the test roles are sh")
	}
	r := newTrackerRepo(t)
	t.Setenv(ownerVariable, "")
	r.write("hephaestus/roles/check-ok.md", toolRoleFile("check-ok", `echo "checked $HEPHAESTUS_MESSAGE_ISSUE"`, "", "[bookmark-ready]"))
	r.write("hephaestus/roles/check-no.md", toolRoleFile("check-no", `echo "2 problems"; exit 1`, "", "[bookmark-ready]"))
	r.write("hephaestus/roles/formatter/upper/role.md", toolRoleFile("formatter/upper", `echo "formatted by $SISYPHUS_OWNER" >> stamps.txt`, "", "[bookmark-ready]"))
	r.write("hephaestus/roles/slow.md", toolRoleFile("slow", `echo "run by $SISYPHUS_OWNER" >> slow.txt; sleep 2`, "", ""))
	r.commit("roles")
	startQueueWithLease(t, r, lease)
	return r
}

func TestABroadcastAboutAClosedIssueReachesEveryListeningRoleAndLeavesItClosed(t *testing.T) {
	r := newNoticeRepo(t, time.Minute)
	for _, role := range []string{"check-ok", "check-no", "formatter/upper"} {
		r.runAgent(role) // the roles listen before the broadcast
	}
	r.mustRunSisyphus("start", "fix-login-bug", "--owner", "implementer-1")
	r.mustRunSisyphus("close", "fix-login-bug", "-m", "Fixed.")
	r.mustRun("messages", "send", "--from", "implementer-1", "--broadcast", "--kind", "bookmark-ready", "--issue", "fix-login-bug", "ai/fix is ready")

	ok, no, family := r.runAgent("check-ok"), r.runAgent("check-no"), r.runAgent("formatter/upper")

	equal(t, "done", ok[0].Status)
	equal(t, "failed", no[0].Status)
	equal(t, "needs-human", no[0].Next)
	equal(t, "done", family[0].Status)
	equal(t, "formatted by formatter-upper-1\n", r.read("stamps.txt"))
	issue := r.showIssue("fix-login-bug")
	equal[any](t, "closed", issue["state"])
	equal[any](t, "implementer-1", issue["owner"])
	contains(t, strings.Join(toStrings(issue["tags"]), ","), "needs-human")
	body := issue["body"].(string)
	contains(t, body, "check-ok-1 (role check-ok) did the bookmark-ready notice of message")
	contains(t, body, "checked fix-login-bug")
	contains(t, body, "check-no-1 (role check-no) failed at the bookmark-ready notice")
	contains(t, body, "2 problems")
}

func TestAKeepAliveKeepsALongTaskFromAnotherWorker(t *testing.T) {
	r := newNoticeRepo(t, 600*time.Millisecond)
	r.mustRun("messages", "send", "--from", "alice", "--role", "slow", "--kind", "task", "a long task")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.run("agent", "run", "slow", "--as", "slow-1", "--once")
	}()
	time.Sleep(1200 * time.Millisecond) // two leases into the task of slow-1
	second := r.run("agent", "run", "slow", "--as", "slow-2", "--once", "--json")
	workers := r.mustRun("messages", "workers").output
	wg.Wait()

	equal(t, "[]\n", second.output)
	equal(t, "run by slow-1\n", r.read("slow.txt"))
	contains(t, workers, "slow-1")
}

func TestAMessageToOneWorkerOfAFamilyWarns(t *testing.T) {
	r := newNoticeRepo(t, time.Minute)
	r.write("hephaestus/roles/formatter/lower/role.md", toolRoleFile("formatter/lower", "true", "", ""))

	family := r.mustRun("messages", "send", "--from", "alice", "--role", "formatter", "--kind", "task", "format it")
	all := r.mustRun("messages", "send", "--from", "alice", "--role", "formatter", "--all", "--kind", "info", "to every formatter")
	one := r.mustRun("messages", "send", "--from", "alice", "--role", "formatter/upper", "--kind", "task", "format it")

	contains(t, family.errors, "Warning: formatter is a family of roles (formatter/lower, formatter/upper), not a role. A message to one worker of a family reaches no worker")
	contains(t, family.errors, "--role formatter/lower, or to every worker of every role of the family with --role formatter --all.")
	isTrue(t, family.output != "", "the message is sent all the same")
	equal(t, "", all.errors)
	equal(t, "", one.errors)
}

func TestStatusShowsWhatNeedsAHuman(t *testing.T) {
	r := newNoticeRepo(t, time.Minute)
	r.mustRunSisyphus("update", "fix-login-bug", "--add-tags", "needs-human")
	r.mustRunSisyphus("update", "add-dark-mode", "--metadata", "referendum=no-result")
	r.mustRun("messages", "send", "--from", "alice", "--role", "engineer", "--kind", "task", "align it")
	main := strings.TrimSpace(r.git("rev-parse", "main"))
	r.mustRun("messages", "send", "--from", "merge-queue", "--broadcast", "--kind", "merge-queue-done",
		"--data", `{"bookmark": "merge-queue/2026-10-12", "base": "main", "head": "`+main+`", "merged": ["ai/a"], "left-out": []}`, "made")
	r.mustRun("messages", "send", "--from", "release-manager-1", "--broadcast", "--kind", "release-recommendation",
		"--data", `{"bookmark": "merge-queue/2026-10-12", "recommend": "move", "summary": "One version, no problems."}`, "move it")

	res := r.mustRun("status")

	for _, part := range []string{
		"Issues that wait for a human (1)\n  fix-login-bug  open",
		"add-dark-mode  no-result. A human decides it",
		"role:engineer  1 (task)",
		"no live worker: start one",
		"merge-queue/2026-10-12 on main",
		"main has it already.",
		"release manager (release-manager-1, ",
		"move. One version, no problems.",
	} {
		contains(t, res.output, part)
	}
}

func TestATaskThatRunsTooLongStopsAndGoesToAHuman(t *testing.T) {
	r := newNoticeRepo(t, time.Minute)
	marker := filepath.Join(t.TempDir(), "finished")
	r.write("hephaestus/roles/hang.md", toolRoleFile("hang", "sleep 3; touch "+marker, "", ""))
	r.commit("hang")
	r.mustRun("messages", "send", "--from", "alice", "--role", "hang", "--kind", "task", "--issue", "fix-login-bug", "hang")
	started := time.Now()

	results := r.runAgent("hang", "--task-timeout", "300ms")

	isTrue(t, time.Since(started) < 2*time.Second, "the worker waited for the task after its limit")
	equal(t, "error", results[0].Status)
	contains(t, results[0].Detail, "the task ran longer than its limit of 300ms, so it was stopped")
	issue := r.showIssue("fix-login-bug")
	equal[any](t, "open", issue["state"])
	contains(t, strings.Join(toStrings(issue["tags"]), ","), "needs-human")
	time.Sleep(3 * time.Second)
	_, err := os.Stat(marker)
	isTrue(t, os.IsNotExist(err), "the command went on after the limit")
}

// streamRunner is a fake model that reports usage as JSON lines, as claude -p --output-format
// stream-json --verbose does, and then closes the issue.
const streamRunner = `#!/bin/sh
cat >/dev/null
i=0
while [ $i -lt ${TURNS:-2} ]; do
  i=$((i+1))
  echo '{"type":"assistant","message":{"id":"m'$i'","usage":{"input_tokens":1000,"output_tokens":100}}}'
  sleep 0.2
done
sisyphus close "$HEPHAESTUS_MESSAGE_ISSUE" -m "Done." >/dev/null
echo '{"type":"result","result":"Done.","total_cost_usd":0.25,"num_turns":2}'
`

func TestAnAIRunRecordsItsUsageAndStopsAtItsTokenLimit(t *testing.T) {
	r := newAIRepo(t)
	runner := filepath.Join(t.TempDir(), "model.sh")
	if err := os.WriteFile(runner, []byte(streamRunner), 0o755); err != nil {
		t.Fatal(err)
	}
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "Review it.")

	done := r.runAgent("reviewer", "--runner", runner)

	equal(t, "done", done[0].Status)
	equal(t, 2200, done[0].Usage.Tokens)
	equal(t, 0.25, done[0].Usage.CostUSD)
	contains(t, r.showIssue("fix-login-bug")["body"].(string), "used 2200 tokens, $0.25, 2 turns, in ")

	r.mustRunSisyphus("new", "long-review")
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "long-review", "Review it.")
	t.Setenv("TURNS", "20")
	over := r.runAgent("reviewer", "--runner", runner, "--max-tokens", "3000")

	equal(t, "error", over[0].Status)
	contains(t, over[0].Detail, "the task passed its limit of 3000 tokens, so it was stopped")
	equal(t, 3300, over[0].Usage.Tokens)
	equal[any](t, "open", r.showIssue("long-review")["state"])
}
