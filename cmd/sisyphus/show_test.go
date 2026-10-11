package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShowPrintsFieldsAndBody(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff", "--title", "Add retry backoff", "--priority", "high", "--tags", "plan")

	res := r.run("show", "retry-backoff")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, res.output, "title:      Add retry backoff")
	contains(t, res.output, "state:      open")
	contains(t, res.output, "priority:   high")
	contains(t, res.output, "effort:     medium")
	contains(t, res.output, "tags:       [plan]")
	contains(t, res.output, "# Add retry backoff")
	contains(t, r.mustRun("show", "retry-backoff", "--all").output, "## Summary")
}

func TestShowPrintsTheResolutionOfAClosedIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")
	r.mustRun("update", "retry-backoff", "closed", "--resolution", "abandoned")

	res := r.run("show", "retry-backoff")

	equal(t, 0, res.exit)
	contains(t, res.output, "state:      closed\nresolution: abandoned\n")
	contains(t, res.output, "closed:     "+today()+"\n")
	contains(t, res.output, "created:    "+today()+"\n")
}

func TestShowAcceptsEveryFormOfTheName(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	for _, reference := range []string{
		"retry-backoff",
		"[[retry-backoff]]",
		"#retry-backoff",
		"issues/open/retry-backoff.md",
	} {
		t.Run(reference, func(t *testing.T) {
			res := r.run("show", reference)

			equal(t, 0, res.exit)
			contains(t, res.output, "title:      Retry backoff")
		})
	}
}

func TestShowPrintsJSON(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "blocking-issue-test")
	r.mustRun("new", "retry-backoff",
		"--title", "Add retry backoff",
		"--tags", "plan, repo",
		"--remote", "https://github.com/acme/widgets/issues/42",
		"--depends-on", "blocking-issue-test")

	res := r.run("show", "--json", "retry-backoff")

	equal(t, 0, res.exit)
	var view issueView
	if err := json.Unmarshal([]byte(res.output), &view); err != nil {
		t.Fatalf("invalid JSON: %s\n%s", err, res.output)
	}
	equal(t, "retry-backoff", view.Name)
	equal(t, "Add retry backoff", view.Title)
	equal(t, "open", view.State)
	equalSlices(t, []string{"plan", "repo"}, view.Tags)
	equal(t, "https://github.com/acme/widgets/issues/42", view.Remote)
	equalSlices(t, []string{"blocking-issue-test"}, view.DependsOn)
	isTrue(t, strings.Contains(view.Body, "## Summary"), "the body is included")
}

func TestShowReportsAMissingIssue(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("show", "no-such-issue")

	equal(t, 1, res.exit)
	contains(t, res.error, "No issue named 'no-such-issue'")
}

func TestShowReportsAnIssueInMoreThanOneDirectory(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")
	r.write("issues/closed/retry-backoff.md", r.read("issues/open/retry-backoff.md"))

	res := r.run("show", "retry-backoff")

	equal(t, 1, res.exit)
	contains(t, res.error, "more than one state directory")
}

func TestCommandsFindAnIssueWithoutRegardToCase(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "first-issue-test")
	r.mustRun("new", "web/fix-login-bug")

	contains(t, r.mustRun("show", "FIRST-ISSUE-TEST").output, "First issue test")
	contains(t, r.mustRun("show", "[[Fix-Login-Bug]]").output, "Fix login bug")
	r.mustRun("update", "First-Issue-Test", "--priority", "high")
	r.mustRun("depends-on", "fix-login-bug", "[[FIRST-ISSUE-TEST]]")

	equal(t, "high", r.frontmatter("issues/open/first-issue-test.md").get("priority"))
	equalSlices(t, []string{"first-issue-test"}, parseDependsOn(r.frontmatter("issues/open/web/fix-login-bug.md").get("depends-on")))
}

func TestShowHidesEmptyFieldsAndTemplateSections(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug", "--tags", "bug")
	r.mustRun("edit", "fix-login-bug", "Summary", "The login fails after a reset.")

	short := r.mustRun("show", "fix-login-bug").output
	full := r.mustRun("show", "fix-login-bug", "--all").output

	contains(t, short, "tags:       [bug]")
	contains(t, short, "## Summary\n\nThe login fails after a reset.")
	for _, hidden := range []string{"owner:", "remote:", "metadata:", "## Context", "## Acceptance criteria", "<Condition>", "## Resolution"} {
		isTrue(t, !strings.Contains(short, hidden), hidden+" is hidden")
		contains(t, full, hidden)
	}
}

func TestShowKeepsASectionWithRealText(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("edit", "fix-login-bug", "Acceptance criteria", "--", "- [ ] The login works after a reset.")

	contains(t, r.mustRun("show", "fix-login-bug").output, "- [ ] The login works after a reset.")
}
