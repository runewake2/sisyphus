package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShowPrintsFieldsAndBody(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "explicit-step-dependencies", "--title", "Add explicit step dependencies", "--priority", "high", "--tags", "plan")

	res := r.run("show", "explicit-step-dependencies")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, res.output, "title:    Add explicit step dependencies")
	contains(t, res.output, "state:    open")
	contains(t, res.output, "priority: high")
	contains(t, res.output, "effort:   medium")
	contains(t, res.output, "tags:     [plan]")
	contains(t, res.output, "# Add explicit step dependencies")
	contains(t, res.output, "## Summary")
}

func TestShowAcceptsEveryFormOfTheName(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "explicit-step-dependencies")

	for _, reference := range []string{
		"explicit-step-dependencies",
		"[[explicit-step-dependencies]]",
		"#explicit-step-dependencies",
		"issues/open/explicit-step-dependencies.md",
	} {
		t.Run(reference, func(t *testing.T) {
			res := r.run("show", reference)

			equal(t, 0, res.exit)
			contains(t, res.output, "title:    Explicit step dependencies")
		})
	}
}

func TestShowPrintsJSON(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "explicit-step-dependencies", "--title", "Add explicit step dependencies", "--tags", "plan, repo")

	res := r.run("show", "--json", "explicit-step-dependencies")

	equal(t, 0, res.exit)
	var view issueView
	if err := json.Unmarshal([]byte(res.output), &view); err != nil {
		t.Fatalf("invalid JSON: %s\n%s", err, res.output)
	}
	equal(t, "explicit-step-dependencies", view.Name)
	equal(t, "Add explicit step dependencies", view.Title)
	equal(t, "open", view.State)
	equalSlices(t, []string{"plan", "repo"}, view.Tags)
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
	r.mustRun("new", "explicit-step-dependencies")
	r.write("issues/closed/explicit-step-dependencies.md", r.read("issues/open/explicit-step-dependencies.md"))

	res := r.run("show", "explicit-step-dependencies")

	equal(t, 1, res.exit)
	contains(t, res.error, "more than one state directory")
}
