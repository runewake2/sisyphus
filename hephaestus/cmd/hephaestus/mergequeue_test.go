package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/runewake2/sisyphus/hephaestus/internal/mergequeue"
)

func (r *testRepo) gitOut(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.root
	out, err := cmd.Output()
	if err != nil {
		r.t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func (r *testRepo) mergeQueue(args ...string) mergequeue.Report {
	r.t.Helper()
	res := r.mustRun(append([]string{"merge-queue", "--json"}, args...)...)
	var report mergequeue.Report
	if err := json.Unmarshal([]byte(res.output), &report); err != nil {
		r.t.Fatalf("%v\n%s\n%s", err, res.output, res.errors)
	}
	return report
}

func mergedBookmarks(report mergequeue.Report) []string {
	var names []string
	for _, m := range report.Merged {
		names = append(names, m.Bookmark)
	}
	return names
}

func TestTheMergeQueueMakesALinearHistory(t *testing.T) {
	r := newMergeRepo(t, "first-task-work", "second-task-work", "third-task-work")
	r.branch("ai/second", "main", "2026-10-09T10:00:00", map[string]string{"second.txt": "2\n"})
	r.branch("ai/first", "main", "2026-10-10T10:00:00", map[string]string{"first.txt": "1\n"})
	r.branch("ai/third", "main", "2026-10-11T10:00:00", map[string]string{"third.txt": "3\n"})
	r.task("first-task-work", "ai/first")
	r.task("second-task-work", "ai/second", "first-task-work")
	r.task("third-task-work", "ai/third")
	mainBefore, secondBefore := r.gitOut("rev-parse", "main"), r.gitOut("rev-parse", "ai/second")

	report := r.mergeQueue("first-task-work", "second-task-work", "third-task-work", "--name", "mq/test")

	// ai/second waits for ai/first; then it is older than ai/third.
	equalStrings(t, []string{"ai/first", "ai/second", "ai/third"}, mergedBookmarks(report))
	isTrue(t, report.Created, "the bookmark is created")
	equal(t, "", r.gitOut("rev-list", "--merges", "main..mq/test"))
	equal(t, "ai/third\nai/second\nai/first", r.gitOut("log", "--format=%s", "main..mq/test"))
	equal(t, mainBefore, r.gitOut("rev-parse", "main"))
	equal(t, secondBefore, r.gitOut("rev-parse", "ai/second"))
	isTrue(t, !strings.Contains(r.gitOut("worktree", "list"), "hephaestus-merge-queue"), "the scratch worktree is gone")
}

func TestVersionsAreGivenInTheOrderOfTheMerge(t *testing.T) {
	r := newMergeRepo(t, "dark-mode-work", "login-fix-work")
	r.write("VERSION", "0.0.5\n")
	r.write("CHANGELOG.md", "# Changelog\n\n## [[0.0.5]] - 2026-10-01\n- Older.\n")
	r.write("changelog/0.0.5.md", "# 0.0.5\n")
	r.commit("versions")
	r.branch("ai/login", "main", "2026-10-11T10:00:00", map[string]string{
		"changelog/unreleased/unreleased-login-fix.md": entry("Fix the login loop", "Fix the login loop.", "2026-10-11"),
		"notes.md": "See [[unreleased-login-fix]].\n",
	})
	r.branch("ai/dark", "main", "2026-10-09T10:00:00", map[string]string{
		"changelog/unreleased/unreleased-dark-mode.md": entry("Add a dark mode", "Add a dark mode.", "2026-10-09"),
	})
	r.task("login-fix-work", "ai/login")
	r.task("dark-mode-work", "ai/dark", "login-fix-work")

	report := r.mergeQueue("dark-mode-work", "login-fix-work", "--name", "mq/test")

	equalStrings(t, []string{"ai/login", "ai/dark"}, mergedBookmarks(report))
	equal(t, "0.0.6", report.Merged[0].Versions[0].Version)
	equal(t, "0.0.7", report.Merged[1].Versions[0].Version)
	equal(t, "0.0.7", r.gitOut("show", "mq/test:VERSION"))
	contains(t, r.gitOut("show", "mq/test:changelog/0.0.6.md"), "# 0.0.6")
	contains(t, r.gitOut("show", "mq/test:changelog/0.0.6.md"), "Fix the login loop")
	equal(t, "See [[0.0.6]].", r.gitOut("show", "mq/test:notes.md"))
	contains(t, r.gitOut("show", "mq/test:CHANGELOG.md"), "## [[0.0.7]]")
	contains(t, r.gitOut("log", "--format=%s", "main..mq/test"), "merge-queue: give versions to ai/login")
	equal(t, "0.0.5", r.gitOut("show", "main:VERSION"))
}

func TestCollidingDecisionRecordsAreRenumbered(t *testing.T) {
	r := newMergeRepo(t, "alpha-task-work", "beta-task-work")
	r.write("design/decisions/cli/0001-first.md", "# 0001: First\n")
	r.commit("decisions")
	r.branch("ai/alpha", "main", "2026-10-09T10:00:00", map[string]string{"design/decisions/cli/0002-alpha.md": "# 0002: Alpha\n"})
	r.branch("ai/beta", "main", "2026-10-10T10:00:00", map[string]string{"design/decisions/cli/0002-beta.md": "# 0002: Beta\n", "beta.md": "See [[0002-beta]].\n"})
	r.task("alpha-task-work", "ai/alpha")
	r.task("beta-task-work", "ai/beta")

	report := r.mergeQueue("alpha-task-work", "beta-task-work", "--name", "mq/test")

	equal(t, "design/decisions/cli/0003-beta.md", report.Merged[1].Renumbered[0].To)
	equal(t, "# 0003: Beta", r.gitOut("show", "mq/test:design/decisions/cli/0003-beta.md"))
	equal(t, "See [[0003-beta]].", r.gitOut("show", "mq/test:beta.md"))
	equal(t, "# 0002: Alpha", r.gitOut("show", "mq/test:design/decisions/cli/0002-alpha.md"))
}

func TestAConflictLeavesABookmarkOut(t *testing.T) {
	r := newMergeRepo(t, "left-task-work", "right-task-work")
	startQueue(t, r)
	r.mustRun("messages", "receive", "--as", "engineer-1", "--role", "engineer") // the engineer listens
	r.branch("ai/left", "main", "2026-10-09T10:00:00", map[string]string{"README.md": "# Left\n"})
	r.branch("ai/right", "main", "2026-10-10T10:00:00", map[string]string{"README.md": "# Right\n"})
	r.task("left-task-work", "ai/left")
	r.task("right-task-work", "ai/right")

	report := r.mergeQueue("left-task-work", "right-task-work", "--name", "mq/test")

	equalStrings(t, []string{"ai/left"}, mergedBookmarks(report))
	equal(t, 1, len(report.LeftOut))
	equal(t, "conflict", report.LeftOut[0].Reason)
	equalStrings(t, []string{"README.md"}, report.LeftOut[0].Files)
	equal(t, "# Left", r.gitOut("show", "mq/test:README.md"))
	contains(t, r.showIssue("right-task-work")["body"].(string), "the merge queue left the work of this task (ai/right) out of mq/test")
	conflict := receivedJSON(t, r, "--as", "engineer-1", "--role", "engineer")
	equal(t, 1, len(conflict))
	equal(t, "conflict", conflict[0].Kind)
	equal(t, "right-task-work", conflict[0].Issue)
}

func TestFailedChecksLeaveABookmarkAndItsDependentsOut(t *testing.T) {
	r := newMergeRepo(t, "good-task-work", "bad-task-work", "after-bad-work")
	r.branch("ai/good", "main", "2026-10-09T10:00:00", map[string]string{"good.txt": "ok\n"})
	r.branch("ai/bad", "main", "2026-10-10T10:00:00", map[string]string{"bad.txt": "broken\n"})
	r.branch("ai/after", "main", "2026-10-11T10:00:00", map[string]string{"after.txt": "ok\n"})
	r.task("good-task-work", "ai/good")
	r.task("bad-task-work", "ai/bad")
	r.task("after-bad-work", "ai/after", "bad-task-work")

	report := r.mergeQueue("good-task-work", "bad-task-work", "after-bad-work", "--name", "mq/test", "--checks", `test ! -e bad.txt || { echo "bad.txt is not allowed"; exit 1; }`)

	equalStrings(t, []string{"ai/good"}, mergedBookmarks(report))
	reasons := map[string]string{}
	for _, left := range report.LeftOut {
		reasons[left.Bookmark] = left.Reason
		if left.Reason == "checks" {
			contains(t, left.Detail, "bad.txt is not allowed")
		}
	}
	equal(t, "checks", reasons["ai/bad"])
	equal(t, "dependency", reasons["ai/after"])
	isTrue(t, !strings.Contains(r.gitOut("ls-tree", "-r", "--name-only", "mq/test"), "bad.txt"), "the failed work is not in the line")
}

func TestNothingToMergeCreatesNoBookmark(t *testing.T) {
	r := newMergeRepo(t, "done-task-work")
	r.git("branch", "ai/done", "main")
	r.task("done-task-work", "ai/done")

	report := r.mergeQueue("done-task-work", "--name", "mq/test")

	isTrue(t, !report.Created, "no bookmark")
	equal(t, "nothing", report.LeftOut[0].Reason)
	equal(t, "", r.gitOut("branch", "--list", "mq/test"))
}

func TestTheNameOfTheResultIsFree(t *testing.T) {
	r := newMergeRepo(t, "first-task-work")
	r.branch("ai/first", "main", "2026-10-10T10:00:00", map[string]string{"first.txt": "1\n"})
	r.task("first-task-work", "ai/first")
	r.git("branch", "mq/taken", "main")

	taken := r.run("merge-queue", "first-task-work", "--name", "mq/taken")

	equal(t, 1, taken.exit)
	contains(t, taken.errors, "The bookmark mq/taken already exists")
}

func TestTheMergeQueueInAJJRepo(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj is not installed")
	}
	r := newMergeRepo(t, "base-task-work", "top-task-work")
	jj := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("jj", append([]string{"--config", "user.name=test", "--config", "user.email=test@example.com"}, args...)...)
		cmd.Dir = r.root
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("jj %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	jj("git", "init", "--colocate")
	jj("new", "main")
	r.write("base.txt", "base\n")
	jj("bookmark", "create", "ai/base", "-r", "@")
	jj("new")
	r.write("top.txt", "top\n")
	jj("bookmark", "create", "ai/top", "-r", "@")
	jj("new", "main")
	r.task("base-task-work", "ai/base")
	r.task("top-task-work", "ai/top")

	report := r.mergeQueue("top-task-work", "base-task-work", "--name", "mq/test")

	equalStrings(t, []string{"ai/base", "ai/top"}, mergedBookmarks(report))
	contains(t, jj("bookmark", "list"), "mq/test")
}

// buildHephaestusOnce builds this command once and puts it first on PATH, for a role whose command
// runs hephaestus.
var buildHephaestusOnce = sync.OnceValue(func() error {
	dir, err := os.MkdirTemp("", "hephaestus-test-self")
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "hephaestus"), ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		return &buildError{string(out)}
	}
	return os.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
})

func TestTheMergeQueueRunsAsAToolAgent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the test role runs sh")
	}
	if err := buildHephaestusOnce(); err != nil {
		t.Fatal(err)
	}
	r := newMergeRepo(t, "first-task-work", "land-the-work")
	r.write("hephaestus/roles/merge-queue.md", toolRoleFile("merge-queue", "hephaestus merge-queue --from-message", "", ""))
	r.commit("role")
	startQueue(t, r)
	r.branch("ai/first", "main", "2026-10-10T10:00:00", map[string]string{"first.txt": "1\n"})
	r.task("first-task-work", "ai/first")
	r.mustRun("messages", "send", "--from", "alice", "--role", "merge-queue", "--kind", "task", "--issue", "land-the-work",
		"--data", `{"tasks": ["first-task-work"], "name": "mq/agent"}`, "Land the work.")

	results := r.runAgent("merge-queue")

	equal(t, "done", results[0].Status)
	contains(t, results[0].Output, "mq/agent at ")
	equal(t, "closed", r.showIssue("land-the-work")["state"])
	contains(t, r.showIssue("land-the-work")["body"].(string), "1. ai/first (first-task-work): 1 change")
	equal(t, "1", r.gitOut("rev-list", "--count", "main..mq/agent"))
}

func TestTheMergeQueueFromAJJWorkspaceThatIsNotColocated(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj is not installed")
	}
	r := newMergeRepo(t, "work-task-here")
	jj := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("jj", append([]string{"--config", "user.name=test", "--config", "user.email=test@example.com"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("jj %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	jj(r.root, "git", "init", "--colocate")
	second := &testRepo{t: t, root: filepath.Join(t.TempDir(), "second")}
	jj(r.root, "workspace", "add", "--name", "second", "-r", "main", second.root)
	second.write("work.txt", "work\n")
	jj(second.root, "commit", "-m", "work in the second workspace")
	jj(second.root, "bookmark", "create", "ai/work", "-r", "@-")
	// The task of the work, in the tracker of the second workspace.
	if out, err := exec.Command("sisyphus", "-C", second.root, "update", "work-task-here", "closed", "--bookmark", "ai/work").CombinedOutput(); err != nil {
		t.Fatal(string(out))
	}

	report := second.mergeQueue("work-task-here", "--name", "mq/test")

	equalStrings(t, []string{"ai/work"}, mergedBookmarks(report))
	equal(t, "work", r.gitOut("show", "mq/test:work.txt"))
}
