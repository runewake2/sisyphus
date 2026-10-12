package main

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"testing"
)

func linksOf(t *testing.T, r *testRepo, document, content string) []linkRow {
	t.Helper()
	r.write(document, content)
	return decodeRows(t, r.run("links", document, "--json"))
}

func decodeRows(t *testing.T, res result) []linkRow {
	t.Helper()
	equal(t, 0, res.exit)
	var rows []linkRow
	if err := json.Unmarshal([]byte(res.output), &rows); err != nil {
		t.Fatalf("the output is not JSON: %v\n%s", err, res.output)
	}
	return rows
}

func rowOf(t *testing.T, rows []linkRow, link string) linkRow {
	t.Helper()
	var found []linkRow
	for _, row := range rows {
		if row.Link == link {
			found = append(found, row)
		}
	}
	if len(found) != 1 {
		t.Fatalf("want one row for %s, got %d", link, len(found))
	}
	return found[0]
}

func TestLinksFindsLinksInFrontmatterAndBodyWithLineNumbers(t *testing.T) {
	r := newTestRepo(t)

	rows := linksOf(t, r, "design/links-doc.md", "---\ndeferred-from: \"[[widget-scheduler]]\"\n---\n# Doc\n\nSee [[README]] and [[0.0.0]].\n")

	equalSlices(t, []linkRow{
		{2, "[[widget-scheduler]]", "ok", "design/widget-scheduler.md"},
		{6, "[[README]]", "ok", "README.md"},
		{6, "[[0.0.0]]", "ok", "changelog/0.0.0.md"},
	}, rows)
}

func TestLinksIgnoresLinksInCodeFencesAndInlineCode(t *testing.T) {
	r := newTestRepo(t)

	rows := linksOf(t, r, "design/links-doc.md",
		"# Doc\n```\n[[in-backtick-fence]]\n```\n~~~\n[[in-tilde-fence]]\n~~~\nText `[[in-inline-code]]` and [[README]].\n")

	equal(t, 1, len(rows))
	equal(t, "[[README]]", rows[0].Link)
}

func TestLinksChecksHeadingsInTheSameFile(t *testing.T) {
	r := newTestRepo(t)

	rows := linksOf(t, r, "design/links-doc.md",
		"# Doc\n## Real Heading\n```\n## Heading In Code\n```\n[[#Real Heading]] [[#real heading|text]] [[#Heading In Code]] [[#No Such Heading]] [[#^block-id]]\n")

	equal(t, "ok", rowOf(t, rows, "[[#Real Heading]]").Status)
	equal(t, "ok", rowOf(t, rows, "[[#real heading|text]]").Status)
	equal(t, "missing-heading", rowOf(t, rows, "[[#Heading In Code]]").Status)
	equal(t, "missing-heading", rowOf(t, rows, "[[#No Such Heading]]").Status)
	equal(t, "ok", rowOf(t, rows, "[[#^block-id]]").Status)
	for _, row := range rows {
		equal(t, "design/links-doc.md", row.Resolved)
	}
}

func TestLinksChecksHeadingsInOtherFiles(t *testing.T) {
	r := newTestRepo(t)

	rows := linksOf(t, r, "design/links-doc.md",
		"[[widget-scheduler#Status API]] [[widget-scheduler#status api]] [[widget-scheduler#No Such]] [[basic-plan.yaml#any]]\n")

	equal(t, "ok", rowOf(t, rows, "[[widget-scheduler#Status API]]").Status)
	equal(t, "ok", rowOf(t, rows, "[[widget-scheduler#status api]]").Status)
	equal(t, "missing-heading", rowOf(t, rows, "[[widget-scheduler#No Such]]").Status)
	equal(t, "ok", rowOf(t, rows, "[[basic-plan.yaml#any]]").Status)
}

func TestLinksReportsMissingAndAmbiguousLinks(t *testing.T) {
	r := newTestRepo(t)
	r.write("a/dup-doc.md", "# A\n")
	r.write("b/dup-doc.md", "# B\n")

	rows := linksOf(t, r, "design/links-doc.md", "[[nothing-here]] [[dup-doc]]\n")

	equal(t, linkRow{1, "[[nothing-here]]", "missing", ""}, rowOf(t, rows, "[[nothing-here]]"))
	equal(t, linkRow{1, "[[dup-doc]]", "ambiguous", "a/dup-doc.md, b/dup-doc.md"}, rowOf(t, rows, "[[dup-doc]]"))
}

func TestLinksAcceptsAWikilinkToTheDocument(t *testing.T) {
	r := newTestRepo(t)
	r.write("design/links-doc.md", "[[README]]\n")

	rows := decodeRows(t, r.run("links", "[[links-doc]]", "--json"))

	equal(t, 1, len(rows))
}

func TestLinksReportsAMissingDocument(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("links", "missing-doc")

	equal(t, 1, res.exit)
	contains(t, res.error, "No file matches 'missing-doc'")
}

func TestLinksWritesJsonWithLowercaseKeys(t *testing.T) {
	r := newTestRepo(t)
	r.write("design/links-doc.md", "[[README]]\n")

	res := r.run("links", "design/links-doc.md", "--json")

	var rows []map[string]any
	if err := json.Unmarshal([]byte(res.output), &rows); err != nil {
		t.Fatal(err)
	}
	equal(t, float64(1), rows[0]["line"].(float64))
	equal(t, "[[README]]", rows[0]["link"].(string))
	equal(t, "ok", rows[0]["status"].(string))
	equal(t, "README.md", rows[0]["resolved"].(string))
}

func TestLinksWritesATableByDefault(t *testing.T) {
	r := newTestRepo(t)
	r.write("design/links-doc.md", "[[README]]\n")

	res := r.run("links", "design/links-doc.md")

	equal(t, 0, res.exit)
	lines := res.lines()
	equal(t, 3, len(lines))
	isTrue(t, regexp.MustCompile(`^line  link`).MatchString(lines[0]), "the first line is the header")
	isTrue(t, regexp.MustCompile(`^1\s+\[\[README\]\]\s+ok\s+README\.md$`).MatchString(lines[2]), "the third line is the row")
}

func TestLinksReportsADocumentWithoutLinks(t *testing.T) {
	r := newTestRepo(t)
	r.write("design/links-doc.md", "# No links\n")

	res := r.run("links", "design/links-doc.md")

	equal(t, 0, res.exit)
	equal(t, "", res.output)
	contains(t, res.error, "No wikilinks in design/links-doc.md")
}

func TestANestedTrackerResolvesLinksInTheWholeRepo(t *testing.T) {
	r := newTestRepo(t)
	r.write(".git/HEAD", "ref: refs/heads/main\n") // the working copy is the repo root
	r.write("CONTRIBUTING.md", "# Contributing\n\n## Issues\n")
	r.mustRun("new", "fix-login-bug")
	nested := &testRepo{t: t, root: filepath.Join(r.root, "tools")}
	nested.write("issues/TEMPLATE.md", kitTemplate(t))
	nested.mustRun("new", "build-tool")
	nested.write("notes.md", "See [[CONTRIBUTING#Issues]], [[fix-login-bug]], and [[build-tool]].\n")

	lint := nested.mustRun("lint")
	links := nested.mustRun("links", "notes.md", "--json").output
	list := nested.mustRun("list").output

	contains(t, lint.output, "1 issue and 3 Markdown files: no problems.")
	contains(t, links, `"resolved": "../CONTRIBUTING.md"`)
	contains(t, links, `"resolved": "../issues/open/fix-login-bug.md"`)
	contains(t, list, "build-tool")
	notContains(t, list, "fix-login-bug")
}

func TestTheRootTrackerChecksTheDocumentsOfANestedOne(t *testing.T) {
	r := newTestRepo(t)
	r.write(".git/HEAD", "ref: refs/heads/main\n")
	r.write("tools/issues/TEMPLATE.md", kitTemplate(t))
	r.write("tools/notes.md", "See [[missing-doc]].\n")

	res := r.run("lint")

	contains(t, res.output, "tools/notes.md:1: the link [[missing-doc]] is missing")
}

func TestAnAliasInATableRowIsReported(t *testing.T) {
	r := newTestRepo(t)

	rows := linksOf(t, r, "design/links-doc.md",
		"# Doc\n\n| Name | Link |\n|---|---|\n| one | [[README|the readme]] |\n| two | [[README]] |\n\nOutside a table: [[README|the readme]].\n")

	equal(t, 3, len(rows))
	equal(t, linkRow{5, "[[README|the readme]]", "alias-in-table", "README.md"}, rows[0])
	equal(t, "ok", rows[1].Status)
	equal(t, linkRow{8, "[[README|the readme]]", "ok", "README.md"}, rows[2])
}
