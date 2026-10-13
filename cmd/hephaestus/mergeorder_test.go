package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

// branch makes a git branch name from base with one commit that writes files, dated date.
func (r *testRepo) branch(name, base, date string, files map[string]string) {
	r.t.Helper()
	r.git("checkout", "-q", "-b", name, base)
	for file, content := range files {
		r.write(file, content)
	}
	r.git("add", "-A")
	cmd := exec.Command("git", "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", name)
	cmd.Dir = r.root
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("commit %s: %v\n%s", name, err, out)
	}
	r.git("checkout", "-q", "main")
}

func (r *testRepo) sisyphus(args ...string) {
	r.t.Helper()
	cmd := exec.Command("sisyphus", append([]string{"-C", r.root}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		r.t.Fatalf("sisyphus %v: %v\n%s", args, err, out)
	}
}

// task makes the issue name a finished task: its dependencies, then closed with its bookmark, as
// the tasks of the merge queue are.
func (r *testRepo) task(name, bookmark string, dependsOn ...string) {
	r.t.Helper()
	for _, dependency := range dependsOn {
		r.sisyphus("depends-on", name, dependency)
	}
	r.sisyphus("update", name, "closed", "--bookmark", bookmark)
}

// newMergeRepo makes a repo with a sisyphus tracker on main, and an issue for each name, open.
func newMergeRepo(t *testing.T, issues ...string) *testRepo {
	t.Helper()
	if err := buildSisyphusOnce(); err != nil {
		t.Fatal(err)
	}
	r := newGitRepo(t)
	r.sisyphus("init")
	r.write("README.md", "# Readme\n")
	for _, issue := range issues {
		r.sisyphus("new", issue)
	}
	r.commit("main")
	return r
}

type planJSON struct {
	Order []struct {
		Bookmark string
		Tasks    []string
		After    []struct{ Bookmark, Why string }
		Overlaps []struct {
			Bookmark string
			Files    []string
		}
	}
	Skipped  []struct{ Bookmark, Why string }
	Warnings []string
}

func (r *testRepo) mergeOrder(args ...string) planJSON {
	r.t.Helper()
	res := r.mustRun(append([]string{"merge-order", "--json"}, args...)...)
	var plan planJSON
	if err := json.Unmarshal([]byte(res.output), &plan); err != nil {
		r.t.Fatalf("%v\n%s", err, res.output)
	}
	return plan
}

func bookmarksOf(plan planJSON) []string {
	var names []string
	for _, place := range plan.Order {
		names = append(names, place.Bookmark)
	}
	return names
}

func equalStrings(t *testing.T, want, got []string) {
	t.Helper()
	if len(want) != len(got) {
		t.Errorf("want %q, got %q", want, got)
		return
	}
	for i := range want {
		if want[i] != got[i] {
			t.Errorf("want %q, got %q", want, got)
			return
		}
	}
}

func TestADependencyGoesFirst(t *testing.T) {
	r := newMergeRepo(t, "fix-login-bug", "split-login-form")
	r.branch("ai/split", "main", "2026-10-11T10:00:00", map[string]string{"split.txt": "split\n"})
	r.branch("ai/fix", "main", "2026-10-10T10:00:00", map[string]string{"fix.txt": "fix\n"})
	r.task("fix-login-bug", "ai/fix", "split-login-form")
	r.task("split-login-form", "ai/split")

	plan := r.mergeOrder("fix-login-bug", "split-login-form")

	// ai/fix is older, but its task depends on the task of ai/split.
	equalStrings(t, []string{"ai/split", "ai/fix"}, bookmarksOf(plan))
	equal(t, "fix-login-bug depends on split-login-form", plan.Order[1].After[0].Why)
}

func TestAStackedBookmarkGoesAfterItsBase(t *testing.T) {
	r := newMergeRepo(t, "base-task-work", "top-task-work")
	r.branch("ai/base", "main", "2026-10-11T10:00:00", map[string]string{"base.txt": "base\n"})
	r.branch("ai/top", "ai/base", "2026-10-11T11:00:00", map[string]string{"top.txt": "top\n"})
	r.task("base-task-work", "ai/base")
	r.task("top-task-work", "ai/top")

	plan := r.mergeOrder("top-task-work", "base-task-work")

	equalStrings(t, []string{"ai/base", "ai/top"}, bookmarksOf(plan))
	equal(t, "ai/top is based on ai/base", plan.Order[1].After[0].Why)
}

func TestTheFewestOverlapsGoFirstThenTheOldest(t *testing.T) {
	r := newMergeRepo(t, "alpha-task-work", "beta-task-work", "gamma-task-work")
	r.branch("ai/alpha", "main", "2026-10-09T10:00:00", map[string]string{"shared.txt": "a\n", "alpha.txt": "a\n"})
	r.branch("ai/beta", "main", "2026-10-10T10:00:00", map[string]string{"shared.txt": "b\n"})
	r.branch("ai/gamma", "main", "2026-10-11T10:00:00", map[string]string{"gamma.txt": "g\n"})
	r.task("alpha-task-work", "ai/alpha")
	r.task("beta-task-work", "ai/beta")
	r.task("gamma-task-work", "ai/gamma")

	plan := r.mergeOrder("alpha-task-work", "beta-task-work", "gamma-task-work")

	// gamma shares no file; alpha and beta share shared.txt, and alpha is older.
	equalStrings(t, []string{"ai/gamma", "ai/alpha", "ai/beta"}, bookmarksOf(plan))
	equal(t, "ai/beta", plan.Order[1].Overlaps[0].Bookmark)
	equalStrings(t, []string{"shared.txt"}, plan.Order[1].Overlaps[0].Files)
}

func TestTasksWithOneBookmarkMergeTogether(t *testing.T) {
	r := newMergeRepo(t, "first-task-work", "second-task-work")
	r.branch("ai/both", "main", "2026-10-11T10:00:00", map[string]string{"both.txt": "x\n"})
	r.task("first-task-work", "ai/both")
	r.task("second-task-work", "ai/both", "first-task-work")

	plan := r.mergeOrder("first-task-work", "second-task-work")

	equal(t, 1, len(plan.Order))
	equalStrings(t, []string{"first-task-work", "second-task-work"}, plan.Order[0].Tasks)
}

func TestABookmarkWithNoChangeIsLeftOut(t *testing.T) {
	r := newMergeRepo(t, "done-task-work", "real-task-work")
	r.git("branch", "ai/done", "main")
	r.branch("ai/real", "main", "2026-10-11T10:00:00", map[string]string{"real.txt": "x\n"})
	r.task("done-task-work", "ai/done")
	r.task("real-task-work", "ai/real")

	res := r.mustRun("merge-order", "done-task-work", "real-task-work")

	contains(t, res.output, "1. ai/real (real-task-work)\n")
	contains(t, res.output, "left out: ai/done (done-task-work): it changes nothing from main")
}

func TestACycleIsAnError(t *testing.T) {
	r := newMergeRepo(t, "first-task-work", "second-task-work")
	r.branch("ai/one", "main", "2026-10-11T10:00:00", map[string]string{"one.txt": "1\n"})
	r.branch("ai/two", "main", "2026-10-11T10:00:00", map[string]string{"two.txt": "2\n"})
	r.task("first-task-work", "ai/one")
	r.task("second-task-work", "ai/two")
	// sisyphus refuses a cycle, so make it by hand, as a bad merge could.
	r.write("issues/closed/first-task-work.md", replaceLine(r.read("issues/closed/first-task-work.md"), "depends-on:", `depends-on: ["[[second-task-work]]"]`))
	r.write("issues/closed/second-task-work.md", replaceLine(r.read("issues/closed/second-task-work.md"), "depends-on:", `depends-on: ["[[first-task-work]]"]`))

	res := r.run("merge-order", "first-task-work", "second-task-work")

	equal(t, 1, res.exit)
	contains(t, res.errors, "form a cycle")
	contains(t, res.errors, "ai/one (first-task-work)")
	contains(t, res.errors, "ai/two (second-task-work)")
}

func TestATaskNeedsAnExistingBookmark(t *testing.T) {
	r := newMergeRepo(t, "no-bookmark-task", "gone-bookmark-task")
	r.task("gone-bookmark-task", "ai/gone")

	res := r.run("merge-order", "no-bookmark-task", "gone-bookmark-task")

	equal(t, 1, res.exit)
	contains(t, res.errors, "task no-bookmark-task has no bookmark")
	contains(t, res.errors, "the bookmark ai/gone of task gone-bookmark-task does not exist")
}

func TestADependencyOutsideTheSetWarns(t *testing.T) {
	r := newMergeRepo(t, "fix-login-bug", "split-login-form", "old-done-work")
	r.sisyphus("update", "old-done-work", "closed")
	r.branch("ai/fix", "main", "2026-10-11T10:00:00", map[string]string{"fix.txt": "x\n"})
	r.task("fix-login-bug", "ai/fix", "split-login-form")

	res := r.mustRun("merge-order", "fix-login-bug")

	contains(t, res.errors, "Warning: task fix-login-bug depends on split-login-form, which is not in the set and not closed.")
}

func TestMergeOrderInAJJRepo(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj is not installed")
	}
	r := newMergeRepo(t, "base-task-work", "top-task-work")
	jj := func(args ...string) {
		t.Helper()
		cmd := exec.Command("jj", append([]string{"--config", "user.name=test", "--config", "user.email=test@example.com"}, args...)...)
		cmd.Dir = r.root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("jj %v: %v\n%s", args, err, out)
		}
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

	plan := r.mergeOrder("top-task-work", "base-task-work")

	equalStrings(t, []string{"ai/base", "ai/top"}, bookmarksOf(plan))
}

func replaceLine(text, prefix, line string) string {
	lines := splitKeep(text)
	for i, l := range lines {
		if len(l) >= len(prefix) && l[:len(prefix)] == prefix {
			lines[i] = line + "\n"
		}
	}
	out := ""
	for _, l := range lines {
		out += l
	}
	return out
}

func splitKeep(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i+1])
			start = i + 1
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}
