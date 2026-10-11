package main

import "testing"

func TestStartTakesTheOwnerFromTheEnvironment(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	t.Setenv(ownerVariable, "agent-7")

	res := r.mustRun("start", "fix-login-bug")

	equal(t, "", res.error)
	equal(t, "agent-7", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestOwnerFlagWinsOverTheEnvironment(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	t.Setenv(ownerVariable, "agent-7")

	r.mustRun("update", "fix-login-bug", "in-progress", "--owner", "alice")

	equal(t, "alice", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestNewInProgressTakesTheOwnerFromTheEnvironment(t *testing.T) {
	r := newTestRepo(t)
	t.Setenv(ownerVariable, "agent-7")

	r.mustRun("new", "fix-login-bug", "--state", "in-progress")

	equal(t, "agent-7", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestTheEnvironmentDoesNotChangeTheOwnerOfAnIssueInProgress(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("start", "fix-login-bug", "--owner", "alice")
	t.Setenv(ownerVariable, "agent-7")

	r.mustRun("update", "fix-login-bug", "--priority", "high")

	equal(t, "alice", r.frontmatter("issues/in-progress/fix-login-bug.md").get("owner"))
}

func TestTheEnvironmentDoesNotGiveAnOpenIssueAnOwner(t *testing.T) {
	r := newTestRepo(t)
	t.Setenv(ownerVariable, "agent-7")

	r.mustRun("new", "fix-login-bug")
	r.mustRun("update", "fix-login-bug", "--priority", "high")

	equal(t, "", r.frontmatter("issues/open/fix-login-bug.md").get("owner"))
}

func TestStartWithNoOwnerWarns(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")

	res := r.mustRun("start", "fix-login-bug")

	contains(t, res.error, "in progress with no owner")
	contains(t, res.error, ownerVariable)
}

func TestListMineFiltersOnTheEnvironmentOwner(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug", "--state", "in-progress", "--owner", "agent-7")
	r.mustRun("new", "add-dark-mode", "--state", "in-progress", "--owner", "alice")
	t.Setenv(ownerVariable, "agent-7")

	list := r.mustRun("list", "--mine").output
	search := r.mustRun("search", "--mine").output

	contains(t, list, "fix-login-bug")
	notContains(t, list, "add-dark-mode")
	contains(t, search, "fix-login-bug")
	notContains(t, search, "add-dark-mode")
}

func TestListMineWithoutTheEnvironmentIsAnError(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("list", "--mine")

	equal(t, 1, res.exit)
	contains(t, res.error, ownerVariable+" is not set")
}
