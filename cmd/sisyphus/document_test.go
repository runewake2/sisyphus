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
