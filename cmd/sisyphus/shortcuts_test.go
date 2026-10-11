package main

import (
	"strings"
	"testing"
)

func TestStartCloseAndAbandonChangeTheState(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("new", "old-idea-test")

	r.mustRun("start", "fix-login-bug", "--owner", "alice", "--bookmark", "ai/login")
	doc := r.frontmatter("issues/in-progress/fix-login-bug.md")
	equal(t, "alice", doc.get("owner"))
	equal(t, "ai/login", doc.get("bookmark"))

	res := r.mustRun("close", "fix-login-bug", "-m", "Fixed in 0.0.9.")
	equal(t, "", res.error)
	closed := r.frontmatter("issues/closed/fix-login-bug.md")
	equal(t, "completed", closed.get("resolution"))
	section, _ := sectionIn(closed.body, "Resolution")
	equal(t, "Fixed in 0.0.9.", strings.TrimSpace(strings.Join(section, "\n")))

	r.mustRun("abandon", "old-idea-test", "-m", "Replaced by [[fix-login-bug]].")
	equal(t, "abandoned", r.frontmatter("issues/closed/old-idea-test.md").get("resolution"))
}

func TestCloseWithoutAMessageWarnsAboutTheResolution(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")

	contains(t, r.mustRun("close", "fix-login-bug").error, "The Resolution section of 'fix-login-bug' holds no text")
}

func TestUpdateMessageNeedsTheStateClosed(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")

	res := r.run("update", "fix-login-bug", "in-progress", "-m", "x")

	equal(t, 1, res.exit)
	contains(t, res.error, "--message is the text of the Resolution section")
	equal(t, 0, r.run("update", "fix-login-bug", "closed", "-m", "Done.").exit)
}

func TestAMistakenCloseCanBeReopenedAndClosedAgain(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("close", "fix-login-bug")

	r.mustRun("update", "fix-login-bug", "open", "--reopen")
	r.mustRun("abandon", "fix-login-bug", "-m", "Not needed.")

	equal(t, "abandoned", r.frontmatter("issues/closed/fix-login-bug.md").get("resolution"))
}
