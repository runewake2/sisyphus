package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// newWorkspaces returns two working copies of one jj repo, as jj workspace add makes them: the first
// has the repo directory in .jj/repo, and the second has a .jj/repo file that points to it. Both have
// the issue fix-login-bug, open, as two workspaces based on one change have.
func newWorkspaces(t *testing.T) (first, second *testRepo) {
	t.Helper()
	t.Setenv(ownerVariable, "")
	base := t.TempDir()
	first = &testRepo{t: t, root: filepath.Join(base, "repo")}
	second = &testRepo{t: t, root: filepath.Join(base, "workspaces", "second")}
	first.write(".jj/repo/store/type", "git")
	second.write(".jj/repo", "../../../repo/.jj/repo")
	for _, r := range []*testRepo{first, second} {
		r.write("issues/TEMPLATE.md", kitTemplate(t))
		r.mustRun("new", "fix-login-bug", "--title", "Fix the login bug")
		r.mustRun("new", "add-dark-mode", "--title", "Add a dark mode")
	}
	return first, second
}

func TestStartRefusesAnIssueThatAnotherWorkspaceHasStarted(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1", "--bookmark", "ai/one")

	res := second.run("start", "fix-login-bug", "--owner", "a2")

	equal(t, 1, res.exit)
	contains(t, res.error, "Issue 'fix-login-bug' is in progress in another workspace, ../../repo, owned by a1")
	contains(t, res.error, "--take")
	isTrue(t, second.exists("issues/open/fix-login-bug.md"), "the issue stays open in the second workspace")
}

func TestTheSameOwnerInAnotherWorkspaceIsRefusedToo(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "claude")

	res := second.run("start", "fix-login-bug", "--owner", "claude")

	equal(t, 1, res.exit)
	contains(t, res.error, "in another workspace")
}

func TestTakeMovesTheClaimToThisWorkspace(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")

	res := second.mustRun("start", "fix-login-bug", "--owner", "a2", "--take")

	contains(t, res.error, "was in progress in another workspace")
	claims := claimsOf(t, first)
	equal(t, 1, len(claims))
	equal(t, "a2", claims[0].Owner)
	equal(t, "active", claims[0].Status)
}

func TestAClaimEndsWhenTheIssueLeavesInProgress(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")
	first.mustRun("close", "fix-login-bug", "-m", "Done.")

	equal(t, 0, len(claimsOf(t, first)))
	equal(t, 0, second.run("start", "fix-login-bug", "--owner", "a2").exit)
}

func TestAStopEndsTheClaim(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")
	first.mustRun("update", "fix-login-bug", "open")

	equal(t, 0, second.run("start", "fix-login-bug", "--owner", "a2").exit)
}

func TestAStaleClaimIsRemovedWithAWarning(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")
	// The first workspace loses the work, for example with jj abandon.
	if err := os.Rename(first.pathOf("issues/in-progress/fix-login-bug.md"), first.pathOf("issues/open/fix-login-bug.md")); err != nil {
		t.Fatal(err)
	}

	equal(t, "stale", claimsOf(t, second)[0].Status)
	res := second.mustRun("start", "fix-login-bug", "--owner", "a2")

	contains(t, res.error, "Removed the claim of 'fix-login-bug' by a1")
	equal(t, "here", claimsOf(t, second)[0].Status)
}

func TestPruneRemovesStaleClaims(t *testing.T) {
	first, _ := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")
	first.mustRun("start", "add-dark-mode", "--owner", "a1")
	if err := os.Remove(first.pathOf("issues/in-progress/add-dark-mode.md")); err != nil {
		t.Fatal(err)
	}

	res := first.mustRun("claims", "--prune", "--json")

	var claims []claimView
	if err := json.Unmarshal([]byte(res.output), &claims); err != nil {
		t.Fatal(err)
	}
	equal(t, 1, len(claims))
	equal(t, "fix-login-bug", claims[0].Name)
	equal(t, 1, len(claimsOf(t, first)))
}

func TestReadyLeavesOutAnIssueThatAnotherWorkspaceHasClaimed(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")

	ready := second.mustRun("list", "--ready", "--state", "open").output

	notContains(t, ready, "fix-login-bug")
	contains(t, ready, "add-dark-mode")
}

func TestClaimsTableShowsEachWorkspace(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1", "--bookmark", "ai/one")
	second.mustRun("start", "add-dark-mode", "--owner", "a2")

	table := second.mustRun("claims").output

	contains(t, table, "add-dark-mode  a2     here")
	contains(t, table, "fix-login-bug  a1     active  ../../repo")
	contains(t, table, "ai/one")
}

func TestClaimsWithoutAVCSIsAnError(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("claims")

	equal(t, 1, res.exit)
	contains(t, res.error, "Claims need a jj or git repo")
}

func TestNewInProgressClaimsTheIssue(t *testing.T) {
	first, _ := newWorkspaces(t)
	first.mustRun("new", "retry-backoff", "--state", "in-progress", "--owner", "a1")

	claims := claimsOf(t, first)

	equal(t, 1, len(claims))
	equal(t, "retry-backoff", claims[0].Name)
}

func TestNestedIssueNamesHaveClaims(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("new", "web/fix-session-expiry")
	second.mustRun("new", "web/fix-session-expiry")
	first.mustRun("start", "web/fix-session-expiry", "--owner", "a1")

	res := second.run("start", "web/fix-session-expiry", "--owner", "a2")

	equal(t, 1, res.exit)
	contains(t, res.error, "in another workspace")
}

func claimsOf(t *testing.T, r *testRepo) []claimView {
	t.Helper()
	var claims []claimView
	if err := json.Unmarshal([]byte(r.mustRun("claims", "--json").output), &claims); err != nil {
		t.Fatal(err)
	}
	return claims
}
