package main

import (
	"slices"
	"strings"
	"testing"
)

func TestDependsOnNewIssuesHaveNoDependenciesByDefault(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "depends-on-test")

	equal(t, "[]", r.frontmatter("issues/open/depends-on-test.md").get("depends-on"))
}

func TestNewSetsDependsOnAsAQuotedWikilink(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "blocking-issue-test")

	res := r.run("new", "depends-on-test", "--depends-on", "blocking-issue-test")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	doc := r.frontmatter("issues/open/depends-on-test.md")
	equal(t, `["[[blocking-issue-test]]"]`, doc.get("depends-on"))
}

func TestNewRefusesADependsOnThatDoesNotExist(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("new", "depends-on-test", "--depends-on", "no-such-issue")

	equal(t, 1, res.exit)
	contains(t, res.error, "The issue 'no-such-issue' does not exist")
	isTrue(t, !r.exists("issues/open/depends-on-test.md"), "no issue is created")
}

func TestDependsOnAddsAndRemovesOneDependency(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "blocking-issue-test")
	r.mustRun("new", "other-blocking-issue-test")
	r.mustRun("new", "depends-on-test", "--state", "in-progress", "--bookmark", "ai/work")

	add := r.run("depends-on", "depends-on-test", "blocking-issue-test")
	equal(t, 0, add.exit)
	equal(t, "issues/in-progress/depends-on-test.md", strings.TrimSpace(add.output))

	addAnother := r.run("depends-on", "depends-on-test", "other-blocking-issue-test")
	equal(t, 0, addAnother.exit)
	doc := r.frontmatter("issues/in-progress/depends-on-test.md")
	equal(t, `["[[blocking-issue-test]]", "[[other-blocking-issue-test]]"]`, doc.get("depends-on"))
	isTrue(t, slices.ContainsFunc(doc.front, func(l string) bool { return strings.HasPrefix(l, "depends-on: ") && strings.Contains(l, " # ") }), "the line keeps its comment")

	removeOne := r.run("depends-on", "depends-on-test", "blocking-issue-test", "--clear")
	equal(t, 0, removeOne.exit)
	equal(t, `["[[other-blocking-issue-test]]"]`, r.frontmatter("issues/in-progress/depends-on-test.md").get("depends-on"))

	clearAll := r.run("depends-on", "depends-on-test", "--clear")
	equal(t, 0, clearAll.exit)
	equal(t, "[]", r.frontmatter("issues/in-progress/depends-on-test.md").get("depends-on"))
}

func TestDependsOnIsIdempotent(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "blocking-issue-test")
	r.mustRun("new", "depends-on-test")

	r.mustRun("depends-on", "depends-on-test", "blocking-issue-test")
	r.mustRun("depends-on", "depends-on-test", "blocking-issue-test")

	equal(t, `["[[blocking-issue-test]]"]`, r.frontmatter("issues/open/depends-on-test.md").get("depends-on"))
}

func TestDependsOnRejectsSelfDependency(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "depends-on-test")

	res := r.run("depends-on", "depends-on-test", "depends-on-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "cannot depend on itself")
}

func TestDependsOnRejectsACycle(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "issue-a-test")
	r.mustRun("new", "issue-b-test")
	r.mustRun("depends-on", "issue-a-test", "issue-b-test")

	res := r.run("depends-on", "issue-b-test", "issue-a-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "would create a cycle")
	equal(t, "[]", r.frontmatter("issues/open/issue-b-test.md").get("depends-on"))
}

func TestDependsOnRejectsATransitiveCycle(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "issue-a-test")
	r.mustRun("new", "issue-b-test")
	r.mustRun("new", "issue-c-test")
	r.mustRun("depends-on", "issue-a-test", "issue-b-test")
	r.mustRun("depends-on", "issue-b-test", "issue-c-test")

	res := r.run("depends-on", "issue-c-test", "issue-a-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "would create a cycle")
}

func TestDependsOnRejectsAMissingIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "depends-on-test")

	res := r.run("depends-on", "depends-on-test", "no-such-issue")

	equal(t, 1, res.exit)
	contains(t, res.error, "The issue 'no-such-issue' does not exist")
}

func TestDependsOnRequiresAnIssueOrClear(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "depends-on-test")

	res := r.run("depends-on", "depends-on-test")

	equal(t, 1, res.exit)
	contains(t, res.error, "Give a blocking issue to add, or --clear")
}

func TestUpdateWarnsWhenClosingAnIssueThatOthersDependOn(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "blocking-issue-test")
	r.mustRun("new", "depends-on-test")
	r.mustRun("depends-on", "depends-on-test", "blocking-issue-test")

	res := r.run("update", "blocking-issue-test", "closed", "--resolution", "completed")

	equal(t, 0, res.exit)
	contains(t, res.error, "these issues still depend on it: depends-on-test")
}

func TestUpdateDoesNotWarnWhenNoOpenIssueDependsOnIt(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "blocking-issue-test")
	r.mustRun("edit", "blocking-issue-test", "resolution", "Completed.")

	res := r.run("update", "blocking-issue-test", "closed", "--resolution", "completed")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
}

func TestListFiltersByBlocked(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "blocking-issue-test")
	r.mustRun("new", "depends-on-test")
	r.mustRun("depends-on", "depends-on-test", "blocking-issue-test")

	blocked := r.run("list", "--blocked")
	equal(t, 0, blocked.exit)
	contains(t, blocked.output, "depends-on-test")
	isTrue(t, !strings.Contains(blocked.output, "blocking-issue-test"), "the blocking issue itself is not blocked")

	r.mustRun("update", "blocking-issue-test", "closed", "--resolution", "completed")

	afterClose := r.run("list", "--blocked")
	equal(t, 0, afterClose.exit)
	equal(t, "", strings.TrimSpace(namesOnly(afterClose.output)))
}

// namesOnly strips the header and dashes from a list table, leaving only data rows (or "" if none).
func namesOnly(output string) string {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) <= 2 {
		return ""
	}
	return strings.Join(lines[2:], "\n")
}
