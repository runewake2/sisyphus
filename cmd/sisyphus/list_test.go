package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func setUpListIssues(t *testing.T) *testRepo {
	t.Helper()
	r := newTestRepo(t)
	r.mustRun("new", "open-low-test", "--priority", "low", "--tags", "plan")
	r.mustRun("new", "open-critical-test", "--priority", "critical", "--tags", "scheduler")
	r.mustRun("new", "in-progress-test", "--state", "in-progress", "--bookmark", "ai/work", "--owner", "alice", "--priority", "high")
	r.mustRun("new", "closed-test", "--state", "closed", "--resolution", "completed")
	r.mustRun("new", "sub-issue-test", "--parent", "open-low-test")
	return r
}

func TestListDefaultsToOpenAndInProgress(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	// sorted by state (open before in-progress), then priority (critical, then medium, then low)
	equalSlices(t, []string{
		"open-critical-test",
		"sub-issue-test",
		"open-low-test",
		"in-progress-test",
	}, namesOf(t, res.output))
}

func TestListFiltersByState(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list", "--state", "closed")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"closed-test"}, namesOf(t, res.output))
}

func TestListFiltersByPriority(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list", "--priority", "critical,high")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"open-critical-test", "in-progress-test"}, namesOf(t, res.output))
}

func TestListFiltersByTags(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list", "--tags", "scheduler")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"open-critical-test"}, namesOf(t, res.output))
}

func TestListFiltersByOwner(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list", "--owner", "alice")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"in-progress-test"}, namesOf(t, res.output))
}

func TestListFiltersByParent(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list", "--parent", "open-low-test")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"sub-issue-test"}, namesOf(t, res.output))
}

func TestListCombinesFiltersWithAnd(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list", "--state", "open", "--priority", "low")

	equal(t, 0, res.exit)
	equalSlices(t, []string{"open-low-test"}, namesOf(t, res.output))
}

func TestListRejectsAnInvalidStateOrPriority(t *testing.T) {
	cases := []struct{ flag, value string }{
		{"--state", "done"},
		{"--priority", "urgent"},
	}
	for _, c := range cases {
		t.Run(c.flag, func(t *testing.T) {
			r := setUpListIssues(t)

			res := r.run("list", c.flag, c.value)

			equal(t, 1, res.exit)
		})
	}
}

func TestListPrintsJSON(t *testing.T) {
	r := setUpListIssues(t)

	res := r.run("list", "--state", "closed", "--json")

	equal(t, 0, res.exit)
	var rows []listRow
	if err := json.Unmarshal([]byte(res.output), &rows); err != nil {
		t.Fatalf("invalid JSON: %s\n%s", err, res.output)
	}
	equal(t, 1, len(rows))
	equal(t, "closed-test", rows[0].Name)
	equal(t, "closed", rows[0].State)
}

// namesOf extracts the "name" column of a list table, in row order.
func namesOf(t *testing.T, output string) []string {
	t.Helper()
	lines := (result{output: output}).lines()
	if len(lines) < 2 {
		return nil
	}
	var names []string
	for _, line := range lines[2:] {
		var name string
		for _, r := range line {
			if r == ' ' {
				break
			}
			name += string(r)
		}
		names = append(names, name)
	}
	return names
}

func TestListAlignsColumnsForWideCharacters(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "wide-title-test", "--title", "Fix 日本語 text")
	r.mustRun("new", "narrow-title-test", "--title", "Fix ascii text")

	lines := r.mustRun("list").lines()

	// Every row puts its state column at the same display column as the header.
	column := strings.Index(lines[0], "state")
	for _, line := range lines[2:] {
		at := strings.Index(line, "open")
		equal(t, column, runewidth.StringWidth(line[:at]))
	}
}

func setUpReadyIssues(t *testing.T) *testRepo {
	t.Helper()
	r := newTestRepo(t)
	r.mustRun("new", "auth-epic-test")
	r.mustRun("new", "login-sub-test", "--parent", "auth-epic-test")
	r.mustRun("new", "done-issue-test", "--state", "closed")
	r.mustRun("new", "unblocked-issue-test", "--depends-on", "done-issue-test", "--parent", "auth-epic-test")
	r.mustRun("new", "blocked-issue-test", "--depends-on", "login-sub-test", "--parent", "auth-epic-test")
	r.mustRun("new", "started-issue-test", "--state", "in-progress", "--bookmark", "ai/work", "--parent", "auth-epic-test")
	return r
}

func TestListReadyShowsTheIssuesThatCanStartNow(t *testing.T) {
	r := setUpReadyIssues(t)

	var rows []listRow
	if err := json.Unmarshal([]byte(r.mustRun("list", "--ready", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, row := range rows {
		names = append(names, row.Name)
	}

	// The epic waits on its open sub-issues, blocked-issue-test waits on login-sub-test, and a closed
	// issue is never ready.
	equalSlices(t, []string{"login-sub-test", "unblocked-issue-test", "started-issue-test"}, names)
}

func TestListReadyMatchesTheAvailableIssuesOfGraph(t *testing.T) {
	r := setUpReadyIssues(t)

	var graph issueGraph
	if err := json.Unmarshal([]byte(r.mustRun("graph", "auth-epic-test", "--json", "--full").output), &graph); err != nil {
		t.Fatal(err)
	}
	available := map[string]bool{}
	for _, node := range graph.Nodes {
		if node.Available {
			available[node.Name] = true
		}
	}
	var rows []listRow
	if err := json.Unmarshal([]byte(r.mustRun("list", "--ready", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}
	equal(t, len(available), len(rows))
	for _, row := range rows {
		isTrue(t, available[row.Name], row.Name+" is available in the graph")
	}
}

func TestListRefusesReadyTogetherWithBlocked(t *testing.T) {
	r := setUpReadyIssues(t)

	equal(t, 1, r.run("list", "--ready", "--blocked").exit)
}
