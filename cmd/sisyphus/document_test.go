package main

import (
	"strings"
	"testing"
)

func TestCommandsReadAndKeepAByteOrderMark(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "bom-issue-test")
	r.write("issues/open/bom-issue-test.md", utf8BOM+r.read("issues/open/bom-issue-test.md"))

	contains(t, r.mustRun("show", "bom-issue-test").output, "Bom issue test")
	r.mustRun("update", "bom-issue-test", "in-progress", "--bookmark", "ai/bom")
	equal(t, "", r.mustRun("lint").error)

	content := r.read("issues/in-progress/bom-issue-test.md")
	isTrue(t, strings.HasPrefix(content, utf8BOM+"---\n"), "the file keeps its byte order mark")
	equal(t, "in-progress", r.frontmatter("issues/in-progress/bom-issue-test.md").get("state"))
}

func TestLinksFindAHeadingOnTheFirstLineAfterAByteOrderMark(t *testing.T) {
	r := newTestRepo(t)
	r.write("design/bom-doc.md", utf8BOM+"# First Heading\n")
	r.write("design/notes.md", "See [[bom-doc#First Heading]].\n")

	equal(t, 0, r.run("lint").exit)
}

func TestAnUpdateKeepsCRLFLineEndings(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "crlf-issue-test")
	r.write("issues/open/crlf-issue-test.md", strings.ReplaceAll(r.read("issues/open/crlf-issue-test.md"), "\n", "\r\n"))

	r.mustRun("update", "crlf-issue-test", "--priority", "high")

	content := r.read("issues/open/crlf-issue-test.md")
	equal(t, strings.Count(content, "\n"), strings.Count(content, "\r\n"))
	equal(t, "high", r.frontmatter("issues/open/crlf-issue-test.md").get("priority"))
}

func TestANewIssueTakesTheLineEndingsOfTheTemplate(t *testing.T) {
	r := newTestRepo(t)
	r.write("issues/TEMPLATE.md", strings.ReplaceAll(r.read("issues/TEMPLATE.md"), "\n", "\r\n"))

	r.mustRun("new", "crlf-issue-test")

	content := r.read("issues/open/crlf-issue-test.md")
	equal(t, strings.Count(content, "\n"), strings.Count(content, "\r\n"))
}

func TestLoadHeaderReadsTheSameFrontmatterAsLoadDocument(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug", "--tags", "bug")
	r.write("issues/open/crlf-bom-test.md", utf8BOM+strings.ReplaceAll(r.read("issues/open/fix-login-bug.md"), "\n", "\r\n"))

	for _, name := range []string{"fix-login-bug", "crlf-bom-test"} {
		path := r.pathOf("issues/open/" + name + ".md")
		full, err := loadDocument(path)
		if err != nil {
			t.Fatal(err)
		}
		header, err := loadHeader(path)
		if err != nil {
			t.Fatal(err)
		}
		equalSlices(t, full.front, header.front)
		equal(t, full.bom, header.bom)
		equal(t, full.crlf, header.crlf)
		equal(t, 0, len(header.body))
		isTrue(t, header.save(path) != nil, "a header cannot be saved")
	}
}

func TestLoadHeaderGivesTheSameErrorsAsLoadDocument(t *testing.T) {
	r := newTestRepo(t)
	r.write("no-frontmatter.md", "# Title\n")
	r.write("unclosed.md", "---\ntitle: x\n")
	for _, name := range []string{"no-frontmatter.md", "unclosed.md"} {
		_, fullErr := loadDocument(r.pathOf(name))
		_, headerErr := loadHeader(r.pathOf(name))
		isTrue(t, fullErr != nil && headerErr != nil, name+" fails")
		equal(t, fullErr.Error(), headerErr.Error())
	}
}
