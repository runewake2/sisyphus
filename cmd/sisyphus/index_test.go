package main

import (
	"fmt"
	"strings"
	"testing"
)

// readsOf runs one command and returns how many times it read each state directory and each issue file.
func readsOf(r *testRepo, args ...string) map[string]int {
	r.t.Helper()
	reads := map[string]int{}
	traceRead = func(path string) { reads[relative(r.root, path)]++ }
	defer func() { traceRead = nil }()
	r.mustRun(args...)
	return reads
}

func TestEveryCommandReadsEachStateDirectoryAndIssueFileOnce(t *testing.T) {
	commands := [][]string{
		{"list", "--blocked"},
		{"list", "--parent", "login-epic"},
		{"search", "login", "--state", "open,in-progress,closed"},
		{"graph", "--full", "login-epic"},
		{"show", "split-login-form"},
		{"update", "login-epic", "closed"},
		{"parent", "fix-login-bug", "login-epic"},
		{"depends-on", "add-dark-mode", "split-login-form"},
		{"depends-on", "split-login-form", "fix-login-bug", "--clear"},
		{"new", "web/add-remember-me", "--parent", "login-epic", "--depends-on", "split-login-form", "--deferred-from", "fix-login-bug"},
		{"remote", "add-dark-mode", "https://example.com/issues/1"},
	}
	for _, args := range commands {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			r := newTestRepo(t)
			r.mustRun("new", "web/auth/login-epic")
			r.mustRun("new", "web/auth/fix-login-bug", "--parent", "login-epic")
			r.mustRun("new", "web/split-login-form", "--parent", "login-epic", "--depends-on", "fix-login-bug")
			r.mustRun("new", "mobile/add-dark-mode", "--depends-on", "web/split-login-form")

			reads := readsOf(r, args...)

			for _, state := range states {
				isTrue(t, reads["issues/"+state] <= 1, "each state directory is read at most once")
			}
			for path, count := range reads {
				if count > 1 {
					t.Errorf("%s was read %d times", path, count)
				}
			}
		})
	}
}

func TestCommandsReadManyIssuesInParallelOnceEach(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "big-epic-test")
	for i := 0; i < 60; i++ {
		name := fmt.Sprintf("parallel-issue-%d", i)
		args := []string{"new", name, "--parent", "big-epic-test"}
		if i > 0 && i%3 == 0 {
			args = append(args, "--depends-on", fmt.Sprintf("parallel-issue-%d", i-1))
		}
		r.mustRun(args...)
	}

	for _, args := range [][]string{{"list", "--ready"}, {"search", "parallel"}, {"lint"}, {"update", "big-epic-test", "closed"}} {
		reads := readsOf(r, args...)
		for path, n := range reads {
			if n != 1 {
				t.Errorf("%v read %s %d times", args, path, n)
			}
		}
	}
	ready := r.mustRun("list", "--ready", "--state", "open", "--json").output
	isTrue(t, strings.Contains(ready, "parallel-issue-1\"") && !strings.Contains(ready, "parallel-issue-3\""), "dependencies are seen")
	contains(t, r.mustRun("search", "Parallel issue 59").output, "parallel-issue-59")
}

func TestCommandsFindAFullNameWithoutReadingTheStateDirectories(t *testing.T) {
	commands := [][]string{
		{"show", "web/fix-login-bug"},
		{"show", "[[web/fix-login-bug]]"},
		{"update", "web/fix-login-bug", "--priority", "high"},
		{"edit", "web/fix-login-bug", "Notes", "a note"},
		{"remote", "web/fix-login-bug", "https://example.com/issues/1"},
	}
	for _, args := range commands {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			r := newTestRepo(t)
			r.mustRun("new", "web/fix-login-bug")
			r.mustRun("new", "other-issue-test")

			reads := readsOf(r, args...)

			for _, state := range states {
				equal(t, 0, reads["issues/"+state])
			}
		})
	}
}

func TestShortNamesAndClosingStillReadTheStateDirectories(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/fix-login-bug")

	equal(t, 1, readsOf(r, "show", "fix-login-bug")["issues/open"])
	equal(t, 1, readsOf(r, "update", "web/fix-login-bug", "closed")["issues/open"])
}

func TestAFullNameInTwoStateDirectoriesIsStillAnError(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/fix-login-bug")
	r.write("issues/closed/web/fix-login-bug.md", strings.Replace(r.read("issues/open/web/fix-login-bug.md"), "state: open", "state: closed", 1))

	res := r.run("show", "web/fix-login-bug")

	equal(t, 1, res.exit)
	contains(t, res.error, "exists in more than one state directory")
}
