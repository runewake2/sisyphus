package main

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestParentNewIssuesHaveNoParentByDefault(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "parent-issue-test")

	equal(t, "", r.frontmatter("issues/open/parent-issue-test.md").get("parent"))
}

func TestParentNewWritesTheParentAsAQuotedWikilink(t *testing.T) {
	for _, reference := range []string{"parent-issue-test", "[[parent-issue-test]]", "#parent-issue-test"} {
		t.Run(reference, func(t *testing.T) {
			r := newTestRepo(t)
			r.mustRun("new", "parent-issue-test")

			res := r.run("new", "child-issue-test", "--parent", reference)

			equal(t, 0, res.exit)
			equal(t, "", res.error)
			doc := r.frontmatter("issues/open/child-issue-test.md")
			equal(t, "[[parent-issue-test]]", doc.get("parent"))
			isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(l, `parent: "[[parent-issue-test]]"`) }), "the parent is quoted")
		})
	}
}

func TestParentNewRefusesAParentThatDoesNotExist(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "child-issue-test", "--parent", "no-such-issue")

	equal(t, 1, res.exit)
	contains(t, res.error, "The parent issue 'no-such-issue' does not exist")
	isTrue(t, !r.exists("issues/open/child-issue-test.md"), "no issue is created")
}

func TestParentNewWarnsWhenTheParentIsClosed(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "parent-issue-test", "--state", "closed", "--resolution", "completed")

	res := r.run("new", "child-issue-test", "--parent", "parent-issue-test")

	equal(t, 0, res.exit)
	contains(t, res.error, "The parent issue 'parent-issue-test' is closed")
}

func TestParentSetsAndClearsTheParent(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "parent-issue-test")
	r.mustRun("new", "child-issue-test", "--state", "in-progress", "--bookmark", "ai/work")

	set := r.run("parent", "child-issue-test", "[[parent-issue-test]]")

	equal(t, 0, set.exit)
	equal(t, "issues/in-progress/child-issue-test.md", strings.TrimSpace(set.output))
	doc := r.frontmatter("issues/in-progress/child-issue-test.md")
	equal(t, "[[parent-issue-test]]", doc.get("parent"))
	isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(l, "parent: ") && strings.Contains(l, " # ") }), "the parent line keeps its comment")

	clear := r.run("parent", "child-issue-test", "--clear")

	equal(t, 0, clear.exit)
	equal(t, "", r.frontmatter("issues/in-progress/child-issue-test.md").get("parent"))
}

func TestParentNeedsEitherAParentOrClear(t *testing.T) {
	cases := map[string][]string{
		"neither": {},
		"both":    {"parent-issue-test", "--clear"},
	}
	for name, extra := range cases {
		t.Run(name, func(t *testing.T) {
			r := newTestRepo(t)
			r.mustRun("new", "parent-issue-test")
			r.mustRun("new", "child-issue-test")

			res := r.run(append([]string{"parent", "child-issue-test"}, extra...)...)

			equal(t, 1, res.exit)
			contains(t, res.error, "Give a parent issue or --clear, but not both")
		})
	}
}

func TestParentRefusesTheIssueItself(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "child-issue-test")

	res := r.run("parent", "child-issue-test", "child-issue-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "cannot be its own parent")
}

func TestParentRefusesAParentThatDoesNotExist(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "child-issue-test")

	res := r.run("parent", "child-issue-test", "no-such-issue")

	equal(t, 1, res.exit)
	contains(t, res.error, "does not exist")
	equal(t, "", r.frontmatter("issues/open/child-issue-test.md").get("parent"))
}

func TestParentReportsAMissingIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "parent-issue-test")

	res := r.run("parent", "no-such-issue", "parent-issue-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "No issue named 'no-such-issue'")
}

func TestParentRefusesADirectCycle(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "first-issue-test")
	r.mustRun("new", "second-issue-test", "--parent", "first-issue-test")

	res := r.run("parent", "first-issue-test", "second-issue-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "cannot be one of its sub-issues")
	equal(t, "", r.frontmatter("issues/open/first-issue-test.md").get("parent"))
}

func TestParentRefusesAnIndirectCycle(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "first-issue-test")
	r.mustRun("new", "second-issue-test", "--parent", "first-issue-test")
	r.mustRun("new", "third-issue-test", "--parent", "second-issue-test")

	res := r.run("parent", "first-issue-test", "third-issue-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "cannot be one of its sub-issues")
}

func TestParentClosingAnIssueWarnsAboutSubIssuesThatAreNotClosed(t *testing.T) {
	for _, resolution := range []string{"completed", "abandoned"} {
		t.Run(resolution, func(t *testing.T) {
			r := newTestRepo(t)
			r.mustRun("new", "parent-issue-test")
			r.mustRun("new", "open-child-test", "--parent", "parent-issue-test")
			r.mustRun("new", "started-child-test", "--parent", "parent-issue-test", "--state", "in-progress", "--bookmark", "ai/work")
			r.mustRun("new", "closed-child-test", "--parent", "parent-issue-test", "--state", "closed", "--resolution", "completed")

			res := r.run("update", "parent-issue-test", "closed", "--resolution", resolution)

			equal(t, 0, res.exit)
			contains(t, res.error, "has sub-issues that are not closed: open-child-test, started-child-test.")
			isTrue(t, r.exists("issues/closed/parent-issue-test.md"), "the issue is closed")
		})
	}
}

func TestParentClosingAnIssueDoesNotWarnWhenAllSubIssuesAreClosed(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "parent-issue-test")
	r.mustRun("new", "closed-child-test", "--parent", "parent-issue-test", "--state", "closed", "--resolution", "completed")
	r.mustRun("edit", "parent-issue-test", "resolution", "Completed.")

	res := r.run("update", "parent-issue-test", "closed", "--resolution", "completed")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
}

func TestParentLinkResolves(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "parent-issue-test")
	r.mustRun("new", "child-issue-test", "--parent", "parent-issue-test")

	res := r.run("links", "[[child-issue-test]]")

	equal(t, 0, res.exit)
	isTrue(t, regexp.MustCompile(`\[\[parent-issue-test\]\]\s+ok\s+issues/open/parent-issue-test\.md`).MatchString(res.output), "the parent link resolves")
}
