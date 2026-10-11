package main

import (
	"strings"
	"testing"
)

// completions runs the hidden __complete command, which the completion scripts call, and returns the
// candidates without the directive line.
func completions(t *testing.T, r *testRepo, args ...string) []string {
	t.Helper()
	res := r.run(append([]string{"__complete"}, args...)...)
	equal(t, 0, res.exit)
	var result []string
	for _, line := range strings.Split(strings.TrimSpace(res.output), "\n") {
		if line != "" && !strings.HasPrefix(line, ":") {
			result = append(result, line)
		}
	}
	return result
}

func TestCompletionOfIssueNamesHasTheirTitles(t *testing.T) {
	r := setUpListIssues(t)

	equalSlices(t, []string{"open-critical-test\tOpen critical test", "open-low-test\tOpen low test"},
		completions(t, r, "show", "open-"))
}

func TestCompletionOfIssueNamesFollowsTheCommand(t *testing.T) {
	r := setUpListIssues(t)

	equalSlices(t, []string{"open-low-test\tOpen low test"}, completions(t, r, "start", "open-l"))
	equalSlices(t, nil, completions(t, r, "start", "in"))
	equalSlices(t, []string{"in-progress-test\tIn progress test"}, completions(t, r, "close", "in"))
	equalSlices(t, nil, completions(t, r, "close", "clo"))
	equalSlices(t, []string{"closed-test\tClosed test"}, completions(t, r, "update", "clo"))
}

func TestCompletionOfStatesAndSections(t *testing.T) {
	r := setUpListIssues(t)

	equalSlices(t, []string{"in-progress"}, completions(t, r, "update", "open-low-test", "i"))
	equalSlices(t, []string{"Resolution"}, completions(t, r, "edit", "open-low-test", "res"))
}

func TestCompletionOfFlagValues(t *testing.T) {
	r := setUpListIssues(t)

	equalSlices(t, []string{"high"}, completions(t, r, "new", "--priority", "h"))
	equalSlices(t, []string{"open,in-progress", "open,closed"}, completions(t, r, "list", "--state", "open,"))
	equalSlices(t, []string{"all"}, completions(t, r, "list", "--state", "a"))
	equalSlices(t, []string{"scheduler"}, completions(t, r, "update", "open-low-test", "--add-tags", "sch"))
	equalSlices(t, []string{"open-low-test\tOpen low test"}, completions(t, r, "new", "--parent", "open-l"))
}

func TestCompletionScriptsForEveryShell(t *testing.T) {
	r := newTestRepo(t)

	for shell, marker := range map[string]string{
		"bash":       "__start_sisyphus",
		"zsh":        "#compdef sisyphus",
		"fish":       "complete -c sisyphus",
		"powershell": "Register-ArgumentCompleter",
		"nushell":    "sisyphus __complete",
	} {
		res := r.run("completion", shell)
		equal(t, 0, res.exit)
		if !strings.Contains(res.output, marker) {
			t.Errorf("completion %s: no %q in the script", shell, marker)
		}
	}
	res := r.run("completion", "bsh")
	equal(t, 1, res.exit)
	if !strings.Contains(res.error, "bash") {
		t.Errorf("completion bsh: no suggestion in %q", res.error)
	}
}
