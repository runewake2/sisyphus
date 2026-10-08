package main

import (
	"slices"
	"strings"
	"testing"
)

func TestNewCreatesAnOpenIssueFromTheTemplate(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "retry-backoff")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "issues/open/retry-backoff.md", strings.TrimSpace(res.output))
	doc := r.frontmatter("issues/open/retry-backoff.md")
	equal(t, "Retry backoff", doc.get("title"))
	equal(t, "open", doc.get("state"))
	equal(t, "", doc.get("resolution"))
	equal(t, "medium", doc.get("priority"))
	equal(t, "medium", doc.get("effort"))
	equal(t, "[]", doc.get("tags"))
	equal(t, today(), doc.get("created"))
	equal(t, "", doc.get("closed"))
	equal(t, "", doc.get("bookmark"))
	equal(t, "", doc.get("deferred-from"))
	equal(t, "", doc.get("owner"))
	equal(t, "", doc.get("approver"))
	equal(t, "[]", doc.get("workspaces"))
	equal(t, "{}", doc.get("metadata"))
	isTrue(t, slices.ContainsFunc(doc.body, func(l string) bool { return l == "# Retry backoff" }), "the body has the title")
}

func TestNewRemovesTemplateInstructionsAndKeepsInlineComments(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "retry-backoff")

	doc := r.frontmatter("issues/open/retry-backoff.md")

	isTrue(t, !slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), "#") }), "no instruction lines")
	isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(l, "state: open ") && strings.Contains(l, " # ") }), "the state line keeps its comment")
}

func TestNewSetsTheGivenOptions(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "retry-backoff",
		"--title", "Add retry backoff",
		"--state", "in-progress",
		"--priority", "high",
		"--effort", "large",
		"--tags", "plan, widget-scheduler ,,repo",
		"--bookmark", "ai/work",
		"--owner", "alice",
		"--approver", "alice",
		"--workspace", "retry-backoff-work",
		"--metadata", "session-id=session-123")

	equal(t, 0, res.exit)
	doc := r.frontmatter("issues/in-progress/retry-backoff.md")
	equal(t, "Add retry backoff", doc.get("title"))
	equal(t, "high", doc.get("priority"))
	equal(t, "large", doc.get("effort"))
	equal(t, "[plan, widget-scheduler, repo]", doc.get("tags"))
	equal(t, "alice", doc.get("owner"))
	equal(t, "alice", doc.get("approver"))
	equal(t, "[retry-backoff-work]", doc.get("workspaces"))
	equal(t, `{session-id: "session-123"}`, doc.get("metadata"))
	isTrue(t, slices.ContainsFunc(doc.body, func(l string) bool { return l == "# Add retry backoff" }), "the body has the title")
}

func TestNewSetsMultipleMetadataEntries(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "open-issue-test", "--metadata", "session-id=session-123,note=hello")

	equal(t, 0, res.exit)
	doc := r.frontmatter("issues/open/open-issue-test.md")
	equal(t, `{note: "hello", session-id: "session-123"}`, doc.get("metadata"))
}

func TestNewQuotesTitlesWithSpecialCharacters(t *testing.T) {
	r := newTestRepo(t)
	const title = `Spike: try "quotes" # here`

	r.mustRun("new", "quoted-title-test", "--title", title)

	doc := r.frontmatter("issues/open/quoted-title-test.md")
	equal(t, title, doc.get("title"))
	isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return l == `title: "Spike: try \"quotes\" # here"` }), "the title is quoted")
}

func TestNewWritesDeferredFromAsAQuotedValue(t *testing.T) {
	cases := []struct{ input, expected string }{
		{"other-thing", "[[other-thing]]"},
		{"#other-thing", "[[other-thing]]"},
		{"[[other-thing]]", "[[other-thing]]"},
		{"ai/other-work", "ai/other-work"},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			r := newTestRepo(t)

			r.mustRun("new", "deferred-work-test", "--deferred-from", c.input)

			doc := r.frontmatter("issues/open/deferred-work-test.md")
			equal(t, c.expected, doc.get("deferred-from"))
			isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(l, `deferred-from: "`+c.expected+`"`) }), "the value is quoted")
		})
	}
}

func TestNewRejectsInvalidNames(t *testing.T) {
	for _, name := range []string{"Step-Deps-Test", "bad_name", "single", "one-two-three-four-five-six-seven", "trailing-", "-leading", "has space"} {
		t.Run(name, func(t *testing.T) {
			r := newTestRepo(t)

			res := r.run("new", "--", name)

			equal(t, 1, res.exit)
			contains(t, res.error, "Invalid issue name")
			isTrue(t, !r.exists("issues/open/"+name+".md"), "no issue is created")
		})
	}
}

func TestNewRejectsAFullNameThatAnIssueAlreadyHas(t *testing.T) {
	for _, existing := range []string{"issues/closed/existing-issue.md", "issues/in-progress/web/existing-issue.md"} {
		t.Run(existing, func(t *testing.T) {
			r := newTestRepo(t)
			r.write(existing, "---\nstate: closed\n---\n")
			name := strings.TrimSuffix(strings.SplitN(existing, "/", 3)[2], ".md")

			res := r.run("new", name)

			equal(t, 1, res.exit)
			contains(t, res.error, "already exists")
			contains(t, res.error, existing)
			isTrue(t, !r.exists("issues/open/"+name+".md"), "no issue is created")
		})
	}
}

func TestNewWarnsWhenOtherFilesHaveTheSameFileName(t *testing.T) {
	cases := []struct{ name, existing, advice string }{
		{"widget-scheduler", "design/widget-scheduler.md", "Put the issue in a subdirectory"},
		{"mixed-case-doc", "design/Mixed-Case-Doc.md", "Put the issue in a subdirectory"},
		{"web/existing-issue", "issues/closed/existing-issue.md", "[[web/existing-issue]]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := newTestRepo(t)
			if !r.exists(c.existing) {
				r.write(c.existing, "---\nstate: closed\n---\n")
			}

			res := r.run("new", c.name)

			equal(t, 0, res.exit)
			contains(t, res.error, "Warning: Other files also have the name")
			contains(t, res.error, c.existing)
			contains(t, res.error, c.advice)
			isTrue(t, r.exists("issues/open/"+c.name+".md"), "the issue is created")
		})
	}
}

func TestNewDefaultsAClosedIssueToCompleted(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "closed-issue-test", "--state", "closed")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "completed", r.frontmatter("issues/closed/closed-issue-test.md").get("resolution"))
}

func TestNewCreatesAClosedIssueWithAResolution(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "closed-issue-test", "--state", "closed", "--resolution", "abandoned")

	equal(t, 0, res.exit)
	equal(t, "issues/closed/closed-issue-test.md", strings.TrimSpace(res.output))
	doc := r.frontmatter("issues/closed/closed-issue-test.md")
	equal(t, "closed", doc.get("state"))
	equal(t, "abandoned", doc.get("resolution"))
	equal(t, today(), doc.get("closed"))
}

func TestNewRejectsAResolutionForAnIssueThatIsNotClosed(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "open-issue-test", "--resolution", "completed")

	equal(t, 1, res.exit)
	contains(t, res.error, "only when the state is closed")
	isTrue(t, !r.exists("issues/open/open-issue-test.md"), "no issue is created")
}

func TestNewWarnsWhenAnInProgressIssueHasNoBookmark(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "started-issue-test", "--state", "in-progress")

	equal(t, 0, res.exit)
	contains(t, res.error, "has no bookmark")
	isTrue(t, r.exists("issues/in-progress/started-issue-test.md"), "the issue is created")
}

func TestNewDoesNotWarnWhenAnInProgressIssueHasABookmark(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "started-issue-test", "--state", "in-progress", "--bookmark", "ai/work")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "ai/work", r.frontmatter("issues/in-progress/started-issue-test.md").get("bookmark"))
}

func TestNewSetsContextReplacingThePlaceholder(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "context-test", "--context", "Steps to reproduce: click the button twice.")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	doc := r.frontmatter("issues/open/context-test.md")
	body := strings.Join(doc.body, "\n")
	contains(t, body, "## Context\n\nSteps to reproduce: click the button twice.\n\n## Acceptance criteria")
	isTrue(t, !strings.Contains(body, "<Give the background"), "the placeholder is gone")
}

func TestNewRejectsValuesThatAreNotAllowed(t *testing.T) {
	cases := []struct{ option, value string }{
		{"--state", "blocked"},
		{"--priority", "urgent"},
		{"--effort", "huge"},
		{"--resolution", "done"},
	}
	for _, c := range cases {
		t.Run(c.option, func(t *testing.T) {
			r := newTestRepo(t)

			res := r.run("new", "bad-value-test", c.option, c.value)

			isTrue(t, res.exit != 0, "the command fails")
			contains(t, res.error, c.value)
			isTrue(t, !r.exists("issues/open/bad-value-test.md"), "no issue is created")
		})
	}
}
