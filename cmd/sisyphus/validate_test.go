package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestListWarnsAboutAnIssueWithInvalidYAML(t *testing.T) {
	r := newTestRepo(t)
	r.write("issues/open/broken-yaml-test.md", "---\ntitle: [broken\nstate: open\npriority: medium\neffort: medium\n---\n# Broken\n")

	res := r.run("list", "--json")

	equal(t, 0, res.exit)
	contains(t, res.error, "Warning: issues/open/broken-yaml-test.md: the value of 'title' opens a list with '[' and does not close it with ']'")
	var rows []listRow
	isTrue(t, json.Unmarshal([]byte(res.output), &rows) == nil, "stdout stays valid JSON")
}

func TestListWarnsAboutAStateThatDoesNotMatchTheDirectory(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "moved-by-hand-test")
	// A person moves the file without changing its state field.
	r.write("issues/in-progress/moved-by-hand-test.md", r.read("issues/open/moved-by-hand-test.md"))
	if err := os.Remove(r.pathOf("issues/open/moved-by-hand-test.md")); err != nil {
		t.Fatal(err)
	}

	res := r.run("list", "--state", "in-progress")

	contains(t, res.error, "issues/in-progress/moved-by-hand-test.md: its state field is 'open', but it is in issues/in-progress/")
}

func TestListWarnsAboutAnInvalidPriorityEvenWhenAFilterHidesTheIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "urgent-issue-test")
	r.write("issues/open/urgent-issue-test.md", strings.Replace(r.read("issues/open/urgent-issue-test.md"), "priority: medium", "priority: urgent", 1))

	res := r.run("list", "--priority", "high")

	contains(t, res.error, "its priority 'urgent' is not one of: critical, high, medium, low")
	isTrue(t, !strings.Contains(res.output, "urgent-issue-test"), "the filter still hides the issue")
}

func TestSearchAndShowWarnAboutAnInvalidIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "bad-effort-test")
	r.write("issues/open/bad-effort-test.md", strings.Replace(r.read("issues/open/bad-effort-test.md"), "effort: medium", "effort: huge", 1))

	contains(t, r.run("search", "bad").error, "its effort 'huge' is not one of")
	show := r.run("show", "bad-effort-test", "--json")
	contains(t, show.error, "its effort 'huge' is not one of")
	var view issueView
	isTrue(t, json.Unmarshal([]byte(show.output), &view) == nil, "stdout stays valid JSON")
}

func TestListWarnsAboutAFileWithNoFrontmatter(t *testing.T) {
	r := newTestRepo(t)
	r.write("issues/open/no-frontmatter-test.md", "# Just a heading\n")

	contains(t, r.run("list").error, "issues/open/no-frontmatter-test.md cannot be read")
}

func TestValidIssuesGiveNoWarnings(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "valid-issue-test", "--title", `A "quoted": title #1`, "--tags", "bug,ui", "--metadata", "note=a=b")
	r.mustRun("update", "valid-issue-test", "closed", "--resolution", "abandoned")

	equal(t, "", r.run("list", "--state", "closed").error)
}
