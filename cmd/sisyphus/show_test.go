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
	contains(t, res.output, "## Summary")
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
