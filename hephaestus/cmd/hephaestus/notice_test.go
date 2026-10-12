package main

import (
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
