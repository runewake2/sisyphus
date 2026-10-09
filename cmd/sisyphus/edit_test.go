package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestEditReplacesASection(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	res := r.run("edit", "retry-backoff", "summary", "Retry with backoff.\nCap the delay at 30 seconds.")

	equal(t, 0, res.exit)
	equal(t, "issues/open/retry-backoff.md\n", res.output)
	content := r.read("issues/open/retry-backoff.md")
	contains(t, content, "## Summary\n\nRetry with backoff.\nCap the delay at 30 seconds.\n\n## Context\n")
	isTrue(t, !strings.Contains(content, "<One or two sentences"), "the placeholder is gone")
	contains(t, content, "## Context\n\n<Give the background")
}

func TestEditKeepsTheFrontmatter(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff", "--priority", "high", "--tags", "plan")
	before := r.frontmatter("issues/open/retry-backoff.md").front

	r.mustRun("edit", "retry-backoff", "Summary", "Retry with backoff.")

	equalSlices(t, before, r.frontmatter("issues/open/retry-backoff.md").front)
}

func TestEditFillsTheResolutionOfAClosedIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")
	r.mustRun("update", "retry-backoff", "closed", "--resolution", "abandoned")

	r.mustRun("edit", "[[retry-backoff]]", "resolution", "Abandoned. The client already retries.")

	content := r.read("issues/closed/retry-backoff.md")
	isTrue(t, strings.HasSuffix(content, "## Resolution\n\nAbandoned. The client already retries.\n\n"), content)
}

func TestEditAppendReplacesThePlaceholder(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	r.mustRun("edit", "retry-backoff", "notes", "2026-10-09: first note.", "--append")
	r.mustRun("edit", "retry-backoff", "notes", "2026-10-10: second note.", "--append")

	contains(t, r.read("issues/open/retry-backoff.md"),
		"## Notes\n\n2026-10-09: first note.\n\n2026-10-10: second note.\n\n## Resolution\n")
}

func TestEditReadsTextFromStandardInput(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	root := newRootCommand(func() (string, error) { return r.root, nil })
	root.SetArgs([]string{"edit", "retry-backoff", "summary", "-"})
	root.SetIn(strings.NewReader("Line one.\n\nLine two.\n"))
	root.SetOut(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	contains(t, r.read("issues/open/retry-backoff.md"), "## Summary\n\nLine one.\n\nLine two.\n\n## Context\n")
}

func TestEditReportsAnUnknownSection(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")
	before := r.read("issues/open/retry-backoff.md")

	res := r.run("edit", "retry-backoff", "design", "Text.")

	equal(t, 1, res.exit)
	contains(t, res.error, "Issue 'retry-backoff' has no section 'design'. Its sections are: Summary, Context, Acceptance criteria, Out of scope, Notes, Resolution.")
	equal(t, before, r.read("issues/open/retry-backoff.md"))
}

func TestEditDoesNotReplaceTheTitle(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	res := r.run("edit", "retry-backoff", "Retry backoff", "Text.")

	equal(t, 1, res.exit)
	contains(t, res.error, "has no section 'Retry backoff'")
}

func TestEditRejectsEmptyText(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	res := r.run("edit", "retry-backoff", "summary", "  \n")

	equal(t, 1, res.exit)
	contains(t, res.error, "The text is empty.")
}

func TestUpdateWarnsAboutAnEmptyResolution(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	res := r.run("update", "retry-backoff", "closed")

	equal(t, 0, res.exit)
	contains(t, res.error, `Warning: The Resolution section of 'retry-backoff' holds no text. Complete it with: sisyphus edit retry-backoff resolution "<text>"`)
}

func TestUpdateDoesNotWarnAboutACompleteResolution(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")
	r.mustRun("edit", "retry-backoff", "resolution", "Completed.")

	res := r.run("update", "retry-backoff", "closed")

	equal(t, 0, res.exit)
	isTrue(t, !strings.Contains(res.error, "Resolution section"), res.error)
}
