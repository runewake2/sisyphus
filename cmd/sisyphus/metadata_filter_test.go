package main

import (
	"encoding/json"
	"testing"
)

func setUpMetadataIssues(t *testing.T) *testRepo {
	t.Helper()
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug", "--metadata", "delegate=merge-queue", "--metadata", "session-id=abc")
	r.mustRun("new", "add-dark-mode", "--metadata", "delegate=technical-docs")
	r.mustRun("new", "plain-issue-test")
	return r
}

func jsonNames(t *testing.T, output string) []string {
	t.Helper()
	var rows []listRow
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, row := range rows {
		names = append(names, row.Name)
	}
	return names
}

func TestListFiltersOnAMetadataValue(t *testing.T) {
	r := setUpMetadataIssues(t)

	names := jsonNames(t, r.mustRun("list", "--metadata", "delegate=merge-queue", "--json").output)

	equalSlices(t, []string{"fix-login-bug"}, names)
}

func TestAMetadataKeyAloneMatchesAnyValue(t *testing.T) {
	r := setUpMetadataIssues(t)

	names := jsonNames(t, r.mustRun("list", "--metadata", "delegate", "--json").output)

	equalSlices(t, []string{"add-dark-mode", "fix-login-bug"}, names)
}

func TestEveryMetadataPairMustMatch(t *testing.T) {
	r := setUpMetadataIssues(t)

	both := jsonNames(t, r.mustRun("list", "--metadata", "delegate=merge-queue", "--metadata", "session-id=abc", "--json").output)
	wrong := jsonNames(t, r.mustRun("list", "--metadata", "delegate=merge-queue", "--metadata", "session-id=xyz", "--json").output)

	equalSlices(t, []string{"fix-login-bug"}, both)
	equal(t, 0, len(wrong))
}

func TestSearchFiltersOnMetadataToo(t *testing.T) {
	r := setUpMetadataIssues(t)

	res := r.mustRun("search", "--metadata", "delegate=technical-docs", "--json")

	var rows []searchRow
	if err := json.Unmarshal([]byte(res.output), &rows); err != nil {
		t.Fatal(err)
	}
	equal(t, 1, len(rows))
	equal(t, "add-dark-mode", rows[0].Name)
	equal(t, "technical-docs", rows[0].Metadata["delegate"])
}

func TestListRowsHaveTheMetadata(t *testing.T) {
	r := setUpMetadataIssues(t)

	var rows []listRow
	if err := json.Unmarshal([]byte(r.mustRun("list", "--json").output), &rows); err != nil {
		t.Fatal(err)
	}

	for _, row := range rows {
		switch row.Name {
		case "fix-login-bug":
			equal(t, "abc", row.Metadata["session-id"])
		case "plain-issue-test":
			isTrue(t, row.Metadata == nil, "an issue with no metadata has none in its row")
		}
	}
	notContains(t, r.mustRun("list", "--json").output, `"metadata": {}`)
}

func TestAMetadataFilterWithNoKeyIsAnError(t *testing.T) {
	r := setUpMetadataIssues(t)

	res := r.run("list", "--metadata", "=merge-queue")

	equal(t, 1, res.exit)
	contains(t, res.error, "has no key")
}
