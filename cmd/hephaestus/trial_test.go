package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// newTrialRepo makes a tracker repo with a committed tool role, stamp, whose command changes a file.
func newTrialRepo(t *testing.T) *testRepo {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the commands of the test roles are sh")
	}
	r := newTrackerRepo(t)
	t.Setenv(ownerVariable, "")
	t.Setenv("HEPHAESTUS_AI_RUNNER", "")
	r.write("hephaestus/roles/stamp.md", toolRoleFile("stamp", `echo "$HEPHAESTUS_MESSAGE_ISSUE by $SISYPHUS_OWNER" > stamps.txt && echo "Stamped $HEPHAESTUS_MESSAGE_ISSUE."`, "", ""))
	r.write("hephaestus/roles/refuse.md", toolRoleFile("refuse", `echo "Cannot do it." && exit 3`, "code-fixer", ""))
	r.commit("roles")
	return r
}

func (r *testRepo) trial(args ...string) Trial {
	r.t.Helper()
	res := r.mustRun(append([]string{"agent", "trial", "--json"}, args...)...)
	var trial Trial
	if err := json.Unmarshal([]byte(res.output), &trial); err != nil {
		r.t.Fatalf("%v\n%s\n%s", err, res.output, res.errors)
	}
	return trial
}

func TestATrialRunsARoleInACopyAndChangesNothing(t *testing.T) {
	r := newTrialRepo(t)
	head := r.gitOut("rev-parse", "HEAD")

	trial := r.trial("stamp", "fix-login-bug")

	equal(t, "done", trial.Result.Status)
	equal(t, "open", trial.Before)
	equal(t, "closed", trial.After)
	contains(t, strings.Join(trial.Changed, "\n"), "stamps.txt")
	contains(t, trial.Result.Output, "Stamped fix-login-bug.")
	equal(t, 0, len(trial.Messages))
	// Nothing reaches the repo: the issue, the files, the commits, and the worktrees are the same.
	equal(t, "open", r.showIssue("fix-login-bug")["state"])
	isTrue(t, !r.exists("stamps.txt"), "the trial wrote to the repo")
	equal(t, head, r.gitOut("rev-parse", "HEAD"))
	isTrue(t, !strings.Contains(r.gitOut("worktree", "list"), "hephaestus-trial"), "the copy is gone")

	listed := r.mustRun("agent", "trials")
	contains(t, listed.output, "stamp")
	contains(t, listed.output, "fix-login-bug")
	contains(t, listed.output, "open -> closed")
}

func TestATrialRecordsTheMessagesThatAFailureWouldSend(t *testing.T) {
	r := newTrialRepo(t)

	trial := r.trial("refuse", "fix-login-bug", "--keep")
	t.Cleanup(func() { r.git("worktree", "remove", "--force", trial.Kept) })

	equal(t, "failed", trial.Result.Status)
	equal(t, 3, trial.Result.Exit)
	equal(t, "code-fixer", trial.Result.Next)
	equal(t, "open", trial.After)
	equal(t, 1, len(trial.Messages))
	equal(t, "code-fixer", trial.Messages[0].To.Role)
	// The copy is kept, with the issue stopped in it.
	contains(t, string(must(os.ReadFile(filepath.Join(trial.Kept, "issues", "open", "fix-login-bug.md")))), `delegate: "code-fixer"`)
	equal(t, "", r.metadataOf("fix-login-bug")["delegate"])
}

func TestATrialOfAnAIRoleRecordsItsUsage(t *testing.T) {
	r := newTrialRepo(t)
	runner := filepath.Join(t.TempDir(), "model.sh")
	if err := os.WriteFile(runner, []byte(streamRunner), 0o755); err != nil {
		t.Fatal(err)
	}

	refused := r.run("agent", "trial", "code-fixer", "fix-login-bug")
	trial := r.trial("code-fixer", "fix-login-bug", "--runner", runner)

	equal(t, 1, refused.exit)
	contains(t, refused.errors, "is an AI role")
	equal(t, "done", trial.Result.Status)
	equal(t, "closed", trial.After)
	equal(t, 2200, trial.Result.Usage.Tokens)
	equal(t, 0.25, trial.Result.Usage.CostUSD)
	equal(t, "open", r.showIssue("fix-login-bug")["state"])
	listed := r.mustRun("agent", "trials", "code-fixer")
	contains(t, listed.output, "2200")
	contains(t, listed.output, "$0.25")
	notContains(t, listed.output, "stamp")
}

func TestATrialNeedsAnIssueOfTheCopy(t *testing.T) {
	r := newTrialRepo(t)

	res := r.run("agent", "trial", "stamp", "no-such-issue")

	equal(t, 1, res.exit)
	contains(t, res.errors, "The issue no-such-issue is not in main")
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
