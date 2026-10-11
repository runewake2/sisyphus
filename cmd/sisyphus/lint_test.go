package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintPassesACleanRepo(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "auth-epic-test")
	r.mustRun("new", "login-sub-test", "--parent", "auth-epic-test", "--tags", "bug")
	r.write("design/notes.md", "See [[login-sub-test]] and [[widget-scheduler#Status API]].\n")

	res := r.run("lint")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	contains(t, res.output, "2 issues and ")
	contains(t, res.output, "Markdown files: no problems.")
}

// replaceIn edits a file of the test repo by hand, the way a person would.
func (r *testRepo) replaceIn(relative, old, replacement string) {
	r.t.Helper()
	content := r.read(relative)
	if !strings.Contains(content, old) {
		r.t.Fatalf("%s does not contain %q", relative, old)
	}
	r.write(relative, strings.Replace(content, old, replacement, 1))
}

func TestLintReportsEachKindOfProblem(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "first-issue-test")
	r.mustRun("new", "second-issue-test", "--depends-on", "first-issue-test", "--parent", "first-issue-test")
	r.mustRun("new", "loose-issue-test")

	// Hand edits that the commands would refuse.
	r.write("issues/open/no-frontmatter-test.md", "# No frontmatter\n")
	r.replaceIn("issues/open/loose-issue-test.md", "priority: medium", "priority: urgent")
	r.replaceIn("issues/open/loose-issue-test.md", "tags: []", "tags: [needs review]")
	r.replaceIn("issues/open/loose-issue-test.md", "title: \"Loose issue test\"", "title: [broken")
	r.replaceIn("issues/open/loose-issue-test.md", "remote:", "remote: x\n  - stray line")
	r.replaceIn("issues/open/first-issue-test.md", "parent:", `parent: "[[second-issue-test]]"`)
	r.replaceIn("issues/open/first-issue-test.md", "depends-on: []", `depends-on: ["[[second-issue-test]]", "[[gone-issue-test]]"]`)
	r.write("issues/closed/loose-issue-test.md", strings.Replace(r.read("issues/open/loose-issue-test.md"), "state: open", "state: closed", 1))
	r.write("design/notes.md", "# Notes\n\nSee [[nowhere-at-all]] and [[widget-scheduler#No Such Heading]].\n")

	res := r.run("lint")

	equal(t, 1, res.exit)
	for _, want := range []string{
		"issues/open/no-frontmatter-test.md: the file cannot be read",
		"issues/open/loose-issue-test.md: its priority 'urgent' is not one of",
		"issues/open/loose-issue-test.md: the tag 'needs review' has a space",
		"issues/open/loose-issue-test.md: the value of 'title' opens a list",
		`issues/open/loose-issue-test.md: the frontmatter line "  - stray line`,
		"issues/open/loose-issue-test.md: the issue 'loose-issue-test' is in more than one state directory: issues/open/loose-issue-test.md, issues/closed/loose-issue-test.md",
		"issues/open/first-issue-test.md: its parent chain has a cycle",
		"issues/open/second-issue-test.md: its parent chain has a cycle",
		"issues/open/first-issue-test.md: its depends-on entry 'gone-issue-test' does not exist",
		"issues/open/first-issue-test.md: its dependencies form a cycle: 'second-issue-test' depends back on it",
		"design/notes.md:3: the link [[nowhere-at-all]] is missing",
		"design/notes.md:3: the link [[widget-scheduler#No Such Heading]] is missing-heading",
	} {
		contains(t, res.output, want)
	}
	contains(t, res.error, "problems in")
}

func TestLintReportsAParentThatDoesNotExistAndADependencyOnItself(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "lonely-issue-test")
	r.replaceIn("issues/open/lonely-issue-test.md", "parent:", `parent: "[[gone-epic-test]]"`)
	r.replaceIn("issues/open/lonely-issue-test.md", "depends-on: []", `depends-on: ["[[lonely-issue-test]]"]`)

	res := r.run("lint")

	contains(t, res.output, "issues/open/lonely-issue-test.md: its parent [[gone-epic-test]] does not exist")
	contains(t, res.output, "issues/open/lonely-issue-test.md: it depends on itself")
}

func TestLintWritesNothing(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "first-issue-test")
	r.replaceIn("issues/open/first-issue-test.md", "state: open", "state: closed")
	before := snapshotFiles(t, r.root)

	r.run("lint")

	equal(t, before, snapshotFiles(t, r.root))
}

func TestLintPrintsJSON(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "first-issue-test")
	r.replaceIn("issues/open/first-issue-test.md", "effort: medium", "effort: huge")

	res := r.run("lint", "--json")

	equal(t, 1, res.exit)
	var result lintResult
	if err := json.Unmarshal([]byte(res.output), &result); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, res.output)
	}
	equal(t, 1, result.Issues)
	equal(t, 1, len(result.Problems))
	equal(t, "issues/open/first-issue-test.md", result.Problems[0].File)
	contains(t, result.Problems[0].Problem, "its effort 'huge'")
}

// snapshotFiles returns every file below root with its content, as one string, so that a test can
// see any change.
func snapshotFiles(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		content, err := os.ReadFile(path)
		b.WriteString(path + "\n" + string(content) + "\n")
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}
