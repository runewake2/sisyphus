package main

import "testing"

func TestStartRefusesAnIssueThatAnotherOwnerHas(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")

	res := r.run("start", "fix-login-bug", "--owner", "a2")

	equal(t, 1, res.exit)
	contains(t, res.error, "Issue 'fix-login-bug' is in progress, owned by a1")
	contains(t, res.error, "--take")
	equal(t, "a1", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestStartWithNoOwnerRefusesAnIssueThatAnOwnerHas(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")

	res := r.run("start", "fix-login-bug")

	equal(t, 1, res.exit)
	contains(t, res.error, "owned by a1")
}

func TestStartBySameOwnerIsSafeToRepeat(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")
	before := r.read("issues/in-progress/fix-login-bug.md")

	res := r.mustRun("start", "fix-login-bug", "--owner", "a1")

	equal(t, "", res.error)
	equal(t, before, r.read("issues/in-progress/fix-login-bug.md"))
}

func TestStartBySameOwnerFromTheEnvironmentIsSafeToRepeat(t *testing.T) {
	r := newTestRepo(t)
	t.Setenv(ownerVariable, "a1")
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug")

	equal(t, 0, r.run("start", "fix-login-bug").exit)
}

func TestStartWithTakeReplacesTheOwnerAndWarns(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")

	res := r.mustRun("start", "fix-login-bug", "--owner", "a2", "--take")

	contains(t, res.error, "Issue 'fix-login-bug' was owned by a1. Its owner is now a2.")
	equal(t, "a2", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestTakeNeedsANewOwner(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")

	res := r.run("start", "fix-login-bug", "--take")

	equal(t, 1, res.exit)
	contains(t, res.error, "--take needs the new owner")
}

func TestUpdateRefusesToChangeTheOwnerOfAnIssueInProgress(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")

	in := r.run("update", "fix-login-bug", "in-progress", "--owner", "a3")
	fields := r.run("update", "fix-login-bug", "--owner", "a3")

	equal(t, 1, in.exit)
	contains(t, in.error, "owned by a1")
	equal(t, 1, fields.exit)
	equal(t, 0, r.run("update", "fix-login-bug", "--owner", "a3", "--take").exit)
	equal(t, "a3", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestUpdateOfOtherFieldsDoesNotNeedTheOwner(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")

	equal(t, 0, r.run("update", "fix-login-bug", "in-progress", "--priority", "high").exit)
	equal(t, 0, r.run("update", "fix-login-bug", "--add-tags", "auth").exit)
}

func TestAnIssueInProgressWithNoOwnerCanGetOne(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug")

	equal(t, 0, r.run("start", "fix-login-bug", "--owner", "a2").exit)
	equal(t, "a2", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestAnotherOwnerCanStillCloseOrStopAnIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("new", "add-dark-mode")
	r.mustRun("start", "fix-login-bug", "--owner", "a1")
	r.mustRun("start", "add-dark-mode", "--owner", "a1")

	equal(t, 0, r.run("close", "fix-login-bug", "-m", "Done.").exit)
	equal(t, 0, r.run("update", "add-dark-mode", "open").exit)
}
