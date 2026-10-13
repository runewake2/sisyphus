package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func entry(title, changelog, date string) string {
	return "# Unreleased: " + title + "\n\n- Date: " + date + "\n- Bookmark: ai/work\n- Components: cli\n- Changelog: " + changelog + "\n\n## Summary\n\n" + title + ".\n"
}

// newReleaseRepo makes a repo at version 0.0.6 with two unreleased entries and documents that
// link to them.
func newReleaseRepo(t *testing.T) *testRepo {
	t.Helper()
	r := newGitRepo(t)
	r.write("VERSION", "0.0.6\n")
	r.write("CHANGELOG.md", "# Changelog\n\n## [[0.0.6]] - 2026-10-01\n- An older change.\n")
	r.write("changelog/0.0.6.md", "# 0.0.6\n")
	r.write("changelog/unreleased/unreleased-fix-login.md", entry("Fix the login loop", "Fix the login loop after a password reset.", "2026-10-10"))
	r.write("changelog/unreleased/unreleased-dark-mode.md", entry("Add a dark mode", "Add a dark mode.", "2026-10-11"))
	r.write("issues/closed/fix-login-bug.md", "## Resolution\n\nCompleted in [[unreleased-fix-login]], see [[unreleased-fix-login#Summary|the summary]].\n")
	r.write("design/notes.md", "See [[changelog/unreleased/unreleased-dark-mode]].\n\n`[[unreleased-dark-mode]]` in code.\n")
	r.commit("work")
	return r
}

func TestAssignGivesTheNextVersionsInDateOrder(t *testing.T) {
	r := newReleaseRepo(t)

	res := r.mustRun("version", "assign")

	contains(t, res.output, "changelog/unreleased/unreleased-fix-login.md -> changelog/0.0.7.md\n")
	contains(t, res.output, "changelog/unreleased/unreleased-dark-mode.md -> changelog/0.0.8.md\n")
	contains(t, res.output, "VERSION: 0.0.6 -> 0.0.8\n")
	equal(t, "0.0.8\n", r.read("VERSION"))
	isTrue(t, !r.exists("changelog/unreleased/unreleased-fix-login.md"), "the entry moved")
	released := r.read("changelog/0.0.7.md")
	isTrue(t, strings.HasPrefix(released, "# 0.0.7\n\n- Date: 2026-10-10\n"), "the heading is the version: "+released)
	notContains(t, released, "Changelog:")
	today := time.Now().Format("2006-01-02")
	equal(t, "# Changelog\n\n## [[0.0.8]] - "+today+"\n- Add a dark mode.\n\n## [[0.0.7]] - "+today+"\n- Fix the login loop after a password reset.\n\n## [[0.0.6]] - 2026-10-01\n- An older change.\n", r.read("CHANGELOG.md"))
}

func TestAssignChangesEveryLinkToAnEntry(t *testing.T) {
	r := newReleaseRepo(t)

	r.mustRun("version", "assign")

	equal(t, "## Resolution\n\nCompleted in [[0.0.7]], see [[0.0.7#Summary|the summary]].\n", r.read("issues/closed/fix-login-bug.md"))
	equal(t, "See [[0.0.8]].\n\n`[[unreleased-dark-mode]]` in code.\n", r.read("design/notes.md"))
}

func TestAssignTakesTheOrderOfTheMerge(t *testing.T) {
	r := newReleaseRepo(t)

	r.mustRun("version", "assign", "unreleased-dark-mode", "fix-login")

	contains(t, r.read("changelog/0.0.7.md"), "Add a dark mode")
	contains(t, r.read("changelog/0.0.8.md"), "Fix the login loop")
}

func TestAssignOneEntryLeavesTheOthers(t *testing.T) {
	r := newReleaseRepo(t)

	r.mustRun("version", "assign", "unreleased-dark-mode")

	equal(t, "0.0.7\n", r.read("VERSION"))
	isTrue(t, r.exists("changelog/unreleased/unreleased-fix-login.md"), "the other entry stays")
}

func TestAssignDryRunChangesNothing(t *testing.T) {
	r := newReleaseRepo(t)

	res := r.mustRun("version", "assign", "--dry-run", "--json")

	var plan struct {
		From, To    string
		Assignments []struct{ Version, To string }
		Links       map[string]int
	}
	if err := json.Unmarshal([]byte(res.output), &plan); err != nil {
		t.Fatal(err)
	}
	equal(t, "0.0.8", plan.To)
	equal(t, 2, plan.Links["issues/closed/fix-login-bug.md"])
	equal(t, "0.0.6\n", r.read("VERSION"))
	isTrue(t, r.exists("changelog/unreleased/unreleased-fix-login.md"), "nothing moved")
}

func TestAssignRefusesAnIncompleteEntry(t *testing.T) {
	r := newReleaseRepo(t)
	r.write("changelog/unreleased/unreleased-broken.md", "# Broken\n")
	r.write("changelog/unreleased/no-prefix.md", entry("No prefix", "No prefix.", "2026-10-10"))

	res := r.run("version", "assign")

	equal(t, 1, res.exit)
	contains(t, res.errors, `changelog/unreleased/unreleased-broken.md needs a first line "# Unreleased: <title>", and a line "- Changelog: <one line for CHANGELOG.md>".`)
	contains(t, res.errors, `changelog/unreleased/no-prefix.md: the name of an unreleased entry starts with "unreleased-"`)
	equal(t, "0.0.6\n", r.read("VERSION"))
}

func TestAssignRefusesAVersionThatExists(t *testing.T) {
	r := newReleaseRepo(t)
	r.write("changelog/0.0.7.md", "# 0.0.7\n")

	res := r.run("version", "assign")

	equal(t, 1, res.exit)
	contains(t, res.errors, "changelog/0.0.7.md already exists")
}

func TestAssignWithNoEntryIsAnError(t *testing.T) {
	r := newGitRepo(t)
	r.write("VERSION", "0.0.6\n")

	res := r.run("version", "assign")

	equal(t, 1, res.exit)
	contains(t, res.errors, "No unreleased entry")
}

func TestVersionListShowsTheEntries(t *testing.T) {
	r := newReleaseRepo(t)

	res := r.mustRun("version", "list")

	equal(t, "2026-10-10  unreleased-fix-login  Fix the login loop\n2026-10-11  unreleased-dark-mode  Add a dark mode\n", res.output)
}
