package main

import (
	"os/exec"
	"testing"
)

// newDecisionsRepo makes a repo whose main branch has the decision 0003-shared-cache, and whose
// working copy, a line of work, also added 0003-local-cache, as two lines of work that merged do.
func newDecisionsRepo(t *testing.T) *testRepo {
	t.Helper()
	r := newGitRepo(t)
	r.write("design/decisions/cli/0001-first.md", "# 0001: First\n")
	r.write("design/decisions/cli/0002-second.md", "# 0002: Second\n")
	r.write("design/decisions/cli/0003-shared-cache.md", "# 0003: Shared cache\n")
	r.commit("main")
	r.write("design/decisions/cli/0003-local-cache.md", "# 0003: Local cache\n\n- Status: Accepted\n")
	r.write("changelog/unreleased/unreleased-cache.md", "See [[0003-local-cache]] and [[0003-shared-cache]].\n")
	return r
}

func TestRenumberGivesTheNewRecordTheNextNumber(t *testing.T) {
	r := newDecisionsRepo(t)

	res := r.mustRun("decisions", "renumber", "--onto", "main")

	equal(t, "design/decisions/cli/0003-local-cache.md -> design/decisions/cli/0004-local-cache.md\nchangelog/unreleased/unreleased-cache.md: 1 link\n", res.output)
	equal(t, "# 0004: Local cache\n\n- Status: Accepted\n", r.read("design/decisions/cli/0004-local-cache.md"))
	isTrue(t, !r.exists("design/decisions/cli/0003-local-cache.md"), "the old name is gone")
	equal(t, "# 0003: Shared cache\n", r.read("design/decisions/cli/0003-shared-cache.md"))
	equal(t, "See [[0004-local-cache]] and [[0003-shared-cache]].\n", r.read("changelog/unreleased/unreleased-cache.md"))
}

func TestRenumberDryRunChangesNothing(t *testing.T) {
	r := newDecisionsRepo(t)

	r.mustRun("decisions", "renumber", "--dry-run")

	isTrue(t, r.exists("design/decisions/cli/0003-local-cache.md"), "nothing moved")
}

func TestRenumberWithNoCollisionSaysSo(t *testing.T) {
	r := newGitRepo(t)
	r.write("design/decisions/cli/0001-first.md", "# 0001: First\n")
	r.commit("main")

	res := r.mustRun("decisions", "renumber")

	equal(t, "", res.output)
	contains(t, res.errors, "No two decision records have the same number.")
}

func TestRenumberLeavesACollisionOfTheBaseToAHuman(t *testing.T) {
	r := newGitRepo(t)
	r.write("design/decisions/cli/0001-first.md", "# 0001: First\n")
	r.write("design/decisions/cli/0001-other.md", "# 0001: Other\n")
	r.commit("main")

	res := r.mustRun("decisions", "renumber")

	contains(t, res.errors, "Warning: main already has design/decisions/cli/0001-first.md and design/decisions/cli/0001-other.md with the same number.")
	isTrue(t, r.exists("design/decisions/cli/0001-other.md"), "nothing moved")
}

func TestRenumberInAJJRepo(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj is not installed")
	}
	r := &testRepo{t: t, root: t.TempDir()}
	jj := func(args ...string) {
		t.Helper()
		cmd := exec.Command("jj", append([]string{"--config", "user.name=test", "--config", "user.email=test@example.com"}, args...)...)
		cmd.Dir = r.root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("jj %v: %v\n%s", args, err, out)
		}
	}
	jj("git", "init")
	r.write("design/decisions/cli/0001-shared.md", "# 0001: Shared\n")
	jj("commit", "-m", "main")
	jj("bookmark", "create", "main", "-r", "@-")
	r.write("design/decisions/cli/0001-local.md", "# 0001: Local\n")

	res := r.mustRun("decisions", "renumber", "--onto", "main")

	equal(t, "design/decisions/cli/0001-local.md -> design/decisions/cli/0002-local.md\n", res.output)
}
