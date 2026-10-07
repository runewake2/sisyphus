package main

import (
	"slices"
	"strings"
	"testing"
)

func TestNewCreatesAnOpenIssueFromTheTemplate(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "explicit-step-dependencies")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "issues/open/explicit-step-dependencies.md", strings.TrimSpace(res.output))
	doc := r.frontmatter("issues/open/explicit-step-dependencies.md")
	equal(t, "Explicit step dependencies", doc.get("title"))
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
	equal(t, "", doc.get("agent-session"))
	isTrue(t, slices.ContainsFunc(doc.body, func(l string) bool { return l == "# Explicit step dependencies" }), "the body has the title")
}

func TestNewRemovesTemplateInstructionsAndKeepsInlineComments(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "explicit-step-dependencies")

	doc := r.frontmatter("issues/open/explicit-step-dependencies.md")

	isTrue(t, !slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(strings.TrimSpace(l), "#") }), "no instruction lines")
	isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(l, "state: open ") && strings.Contains(l, " # ") }), "the state line keeps its comment")
}

func TestNewSetsTheGivenOptions(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "explicit-step-dependencies",
		"--title", "Add explicit step dependencies",
		"--priority", "high",
		"--effort", "large",
		"--tags", "plan, widget-scheduler ,,repo",
		"--owner", "samw",
		"--approver", "samw",
		"--workspace", "sisyphus-explicit-steps",
		"--agent-session", "session-123")

	equal(t, 0, res.exit)
	doc := r.frontmatter("issues/open/explicit-step-dependencies.md")
	equal(t, "Add explicit step dependencies", doc.get("title"))
	equal(t, "high", doc.get("priority"))
	equal(t, "large", doc.get("effort"))
	equal(t, "[plan, widget-scheduler, repo]", doc.get("tags"))
	equal(t, "samw", doc.get("owner"))
	equal(t, "samw", doc.get("approver"))
	equal(t, "[sisyphus-explicit-steps]", doc.get("workspaces"))
	equal(t, "session-123", doc.get("agent-session"))
	isTrue(t, slices.ContainsFunc(doc.body, func(l string) bool { return l == "# Add explicit step dependencies" }), "the body has the title")
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
		{"samw/ai/other-work", "samw/ai/other-work"},
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

func TestNewRejectsNamesThatAreNotUniqueAcrossAllMarkdownFiles(t *testing.T) {
	cases := []struct{ name, existing string }{
		{"widget-scheduler", "design/widget-scheduler.md"},
		{"mixed-case-doc", "design/Mixed-Case-Doc.md"},
		{"existing-issue", "issues/closed/existing-issue.md"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := newTestRepo(t)
			if !r.exists(c.existing) {
				r.write(c.existing, "---\nstate: closed\n---\n")
			}

			res := r.run("new", c.name)

			equal(t, 1, res.exit)
			contains(t, res.error, "is not unique")
			contains(t, res.error, c.existing)
			isTrue(t, !r.exists("issues/open/"+c.name+".md"), "no issue is created")
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

	res := r.run("new", "started-issue-test", "--state", "in-progress", "--bookmark", "samw/ai/work")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "samw/ai/work", r.frontmatter("issues/in-progress/started-issue-test.md").get("bookmark"))
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
