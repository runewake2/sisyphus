package main

import (
	"encoding/json"
	"testing"
)

func setUpListIssues(t *testing.T) *testRepo {
	t.Helper()
	r := newTestRepo(t)
	r.mustRun("new", "open-low-test", "--priority", "low", "--tags", "plan")
	r.mustRun("new", "open-critical-test", "--priority", "critical", "--tags", "scheduler")
	r.mustRun("new", "in-progress-test", "--state", "in-progress", "--bookmark", "samw/ai/work", "--owner", "samw", "--priority", "high")
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

	res := r.run("list", "--owner", "samw")

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
