package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestRenameMovesTheIssueToADirectory(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug", "--title", "Fix the login bug")

	res := r.mustRun("rename", "fix-login-bug", "web/auth/fix-login-bug")

	equal(t, "issues/open/web/auth/fix-login-bug.md\n", res.output)
	isTrue(t, !r.exists("issues/open/fix-login-bug.md"), "the old file is gone")
	equal(t, "Fix the login bug", r.frontmatter("issues/open/web/auth/fix-login-bug.md").get("title"))
}

func TestRenameKeepsTheState(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug", "--state", "in-progress", "--owner", "alice")

	r.mustRun("rename", "fix-login-bug", "web/fix-login-bug")

	isTrue(t, r.exists("issues/in-progress/web/fix-login-bug.md"), "the issue stays in progress")
}

func TestRenameRemovesTheDirectoriesThatItEmpties(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "web/auth/fix-login-bug")

	r.mustRun("rename", "web/auth/fix-login-bug", "fix-login-bug")

	isTrue(t, !r.exists("issues/open/web"), "the empty directories are gone")
	isTrue(t, r.exists("issues/open/fix-login-bug.md"), "the issue moved")
}

func TestRenameUpdatesTheLinksToTheIssue(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("new", "split-login-form", "--parent", "fix-login-bug", "--depends-on", "fix-login-bug")
	r.write("design/login.md", "# Login\n\nSee [[fix-login-bug]], [[fix-login-bug#Summary|the summary]], and ![[fix-login-bug]].\n\n`[[fix-login-bug]]` in code.\n\n```\n[[fix-login-bug]]\n```\n")

	r.mustRun("rename", "fix-login-bug", "fix-login-loop")

	equal(t, "# Login\n\nSee [[fix-login-loop]], [[fix-login-loop#Summary|the summary]], and ![[fix-login-loop]].\n\n`[[fix-login-bug]]` in code.\n\n```\n[[fix-login-bug]]\n```\n", r.read("design/login.md"))
	sub := r.read("issues/open/split-login-form.md")
	contains(t, sub, "\nparent: \"[[fix-login-loop]]\"")
	contains(t, sub, "\ndepends-on: [\"[[fix-login-loop]]\"]")
	equal(t, 0, r.run("lint").exit)
}

func TestAMoveToADirectoryKeepsLinksWithTheFileName(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("new", "split-login-form", "--parent", "fix-login-bug")
	r.write("design/login.md", "See [[fix-login-bug]] and [[issues/open/fix-login-bug]].\n")

	r.mustRun("rename", "fix-login-bug", "web/fix-login-bug")

	equal(t, "See [[fix-login-bug]] and [[web/fix-login-bug]].\n", r.read("design/login.md"))
	// sisyphus writes the full name in the frontmatter.
	contains(t, r.read("issues/open/split-login-form.md"), "\nparent: \"[[web/fix-login-bug]]\"")
	equal(t, 0, r.run("lint").exit)
}

func TestARenameToANameThatAnotherFileHasUsesTheFullName(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.write("design/login-notes.md", "# Login notes\n")
	r.write("design/login.md", "See [[fix-login-bug]].\n")

	res := r.mustRun("rename", "fix-login-bug", "web/login-notes")

	contains(t, res.error, "Other files also have the name 'login-notes'")
	equal(t, "See [[web/login-notes]].\n", r.read("design/login.md"))
}

func TestRenameUpdatesLinksInTheIssueItself(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug", "--context", "See [[fix-login-bug#Notes]].")

	r.mustRun("rename", "fix-login-bug", "fix-login-loop")

	contains(t, r.read("issues/open/fix-login-loop.md"), "See [[fix-login-loop#Notes]].")
}

func TestRenameKeepsCRLFLineEndings(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.write("design/login.md", "# Login\r\n\r\nSee [[fix-login-bug]].\r\n")

	r.mustRun("rename", "fix-login-bug", "fix-login-loop")

	equal(t, "# Login\r\n\r\nSee [[fix-login-loop]].\r\n", r.read("design/login.md"))
}

func TestRenameRefusesAnInvalidOrUsedName(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.mustRun("new", "add-dark-mode", "--state", "closed")

	invalid := r.run("rename", "fix-login-bug", "Fix Login")
	used := r.run("rename", "fix-login-bug", "add-dark-mode")
	same := r.run("rename", "fix-login-bug", "fix-login-bug")

	contains(t, invalid.error, "Invalid issue name 'Fix Login'")
	contains(t, used.error, "Issue 'add-dark-mode' already exists: issues/closed/add-dark-mode.md.")
	contains(t, same.error, "already has the name")
	isTrue(t, r.exists("issues/open/fix-login-bug.md"), "nothing moved")
}

func TestRenameDryRunChangesNothing(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.write("design/login.md", "See [[fix-login-bug]].\n")

	text := r.mustRun("rename", "fix-login-bug", "fix-login-loop", "--dry-run")
	asJSON := r.mustRun("rename", "fix-login-bug", "fix-login-loop", "--dry-run", "--json")

	equal(t, "issues/open/fix-login-bug.md -> issues/open/fix-login-loop.md\ndesign/login.md:1: [[fix-login-bug]] -> [[fix-login-loop]]\n", text.output)
	var plan renamePlan
	if err := json.Unmarshal([]byte(asJSON.output), &plan); err != nil {
		t.Fatal(err)
	}
	equal(t, 1, len(plan.Links))
	equal(t, "design/login.md", plan.Links[0].File)
	isTrue(t, r.exists("issues/open/fix-login-bug.md"), "nothing moved")
	equal(t, "See [[fix-login-bug]].\n", r.read("design/login.md"))
}

func TestRenameMovesTheClaim(t *testing.T) {
	first, _ := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")

	first.mustRun("rename", "fix-login-bug", "web/fix-login-bug")

	claims := claimsOf(t, first)
	equal(t, 1, len(claims))
	equal(t, "web/fix-login-bug", claims[0].Name)
	equal(t, "here", claims[0].Status)
}

func TestRenameRefusesAnIssueThatAnotherWorkspaceHasStarted(t *testing.T) {
	first, second := newWorkspaces(t)
	first.mustRun("start", "fix-login-bug", "--owner", "a1")

	res := second.run("rename", "fix-login-bug", "web/fix-login-bug")

	equal(t, 1, res.exit)
	contains(t, res.error, "rename the issue after that work merges")
}

func TestRenameWithGitRecordsAMove(t *testing.T) {
	r := newTestRepo(t)
	r.mustRun("new", "fix-login-bug")
	r.git("init", "-q")
	r.git("add", "-A")
	r.git("-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "initial")

	r.mustRun("rename", "fix-login-bug", "web/fix-login-bug")

	cmd := exec.Command("git", "status", "--porcelain=v1")
	cmd.Dir = r.root
	raw, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	status := string(raw)
	contains(t, status, "R  issues/open/fix-login-bug.md -> issues/open/web/fix-login-bug.md")
	isTrue(t, !strings.Contains(status, "??"), "nothing is untracked: "+status)
}
