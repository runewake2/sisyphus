package main

import (
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
				equal(t, 1, reads["issues/"+state])
			}
			for path, count := range reads {
				if count > 1 {
					t.Errorf("%s was read %d times", path, count)
				}
			}
		})
	}
}
