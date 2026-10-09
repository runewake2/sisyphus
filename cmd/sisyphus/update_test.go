package main

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const updateTarget = "update-target-issue"

func createUpdateTarget(r *testRepo, options ...string) {
	r.t.Helper()
	r.mustRun(append([]string{"new", updateTarget}, options...)...)
}

func TestUpdateStartsAnIssue(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "issues/in-progress/"+updateTarget+".md", strings.TrimSpace(res.output))
	isTrue(t, !r.exists("issues/open/"+updateTarget+".md"), "the open file is gone")
	doc := r.frontmatter("issues/in-progress/" + updateTarget + ".md")
	equal(t, "in-progress", doc.get("state"))
	equal(t, "ai/work", doc.get("bookmark"))
	equal(t, "", doc.get("resolution"))
	equal(t, "", doc.get("closed"))
}

func TestUpdateClosesAnIssueAsCompletedAndKeepsTheBookmark(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "ai/work")

	res := r.run("update", updateTarget, "closed", "--resolution", "completed")

	equal(t, 0, res.exit)
	equal(t, "issues/closed/"+updateTarget+".md", strings.TrimSpace(res.output))
	isTrue(t, !r.exists("issues/in-progress/"+updateTarget+".md"), "the in-progress file is gone")
	doc := r.frontmatter("issues/closed/" + updateTarget + ".md")
	equal(t, "closed", doc.get("state"))
	equal(t, "completed", doc.get("resolution"))
	equal(t, today(), doc.get("closed"))
	equal(t, "ai/work", doc.get("bookmark"))
}

func TestUpdateClosesAnOpenIssueAsAbandoned(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "closed", "--resolution", "abandoned")

	equal(t, 0, res.exit)
	equal(t, "abandoned", r.frontmatter("issues/closed/"+updateTarget+".md").get("resolution"))
}

func TestUpdateStopsWorkAndClearsTheBookmark(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "ai/work")

	res := r.run("update", updateTarget, "open")

	equal(t, 0, res.exit)
	equal(t, "issues/open/"+updateTarget+".md", strings.TrimSpace(res.output))
	doc := r.frontmatter("issues/open/" + updateTarget + ".md")
	equal(t, "open", doc.get("state"))
	equal(t, "", doc.get("bookmark"))
	equal(t, "", doc.get("resolution"))
	equal(t, "", doc.get("closed"))
}

func TestUpdateKeepsAnExistingBookmarkWhenNoneIsGiven(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "ai/work")

	res := r.run("update", updateTarget, "in-progress")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "ai/work", r.frontmatter("issues/in-progress/"+updateTarget+".md").get("bookmark"))
}

func TestUpdateSetsOwnerApproverWorkspaceAndMetadata(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "in-progress",
		"--bookmark", "ai/work",
		"--owner", "alice",
		"--approver", "bob",
		"--workspace", "sisyphus-work",
		"--metadata", "session-id=session-123")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	doc := r.frontmatter("issues/in-progress/" + updateTarget + ".md")
	equal(t, "alice", doc.get("owner"))
	equal(t, "bob", doc.get("approver"))
	equal(t, "[sisyphus-work]", doc.get("workspaces"))
	equal(t, `{session-id: "session-123"}`, doc.get("metadata"))
}

func TestUpdateMergesMetadataAndKeepsItWhenClosed(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "ai/work",
		"--owner", "alice", "--metadata", "session-id=session-123")

	r.mustRun("update", updateTarget, "in-progress", "--metadata", "note=half-done")
	res := r.run("update", updateTarget, "closed", "--resolution", "completed")

	equal(t, 0, res.exit)
	doc := r.frontmatter("issues/closed/" + updateTarget + ".md")
	equal(t, `{note: "half-done", session-id: "session-123"}`, doc.get("metadata"))
	equal(t, "alice", doc.get("owner"))
	equal(t, "ai/work", doc.get("bookmark"))
}

func TestUpdateAppendsToWorkspacesWithoutDuplicating(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "ai/work", "--workspace", "sisyphus-work")

	r.mustRun("update", updateTarget, "in-progress", "--workspace", "sisyphus-work-2")
	r.mustRun("update", updateTarget, "in-progress", "--workspace", "sisyphus-work")

	doc := r.frontmatter("issues/in-progress/" + updateTarget + ".md")
	equal(t, "[sisyphus-work, sisyphus-work-2]", doc.get("workspaces"))
}

func TestUpdateClearsOwnerButKeepsWorkspacesApproverAndMetadataWhenReopened(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "ai/work",
		"--owner", "alice", "--approver", "bob", "--workspace", "sisyphus-work", "--metadata", "session-id=session-123")

	res := r.run("update", updateTarget, "open")

	equal(t, 0, res.exit)
	doc := r.frontmatter("issues/open/" + updateTarget + ".md")
	equal(t, "", doc.get("owner"))
	equal(t, "", doc.get("bookmark"))
	equal(t, "bob", doc.get("approver"))
	equal(t, "[sisyphus-work]", doc.get("workspaces"))
	equal(t, `{session-id: "session-123"}`, doc.get("metadata"))
}

func TestUpdateMovesTheIssueWithGitMvInAGitRepo(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)
	r.git("init", "-q")
	r.git("add", "-A")
	r.git("-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "initial")

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	isTrue(t, !r.exists("issues/open/"+updateTarget+".md"), "the open file is gone")
	isTrue(t, r.exists("issues/in-progress/"+updateTarget+".md"), "the issue moved")

	cmd := exec.Command("git", "status", "--porcelain=v1")
	cmd.Dir = r.root
	raw, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	output := string(raw)
	isTrue(t, !strings.Contains(output, "??"), "nothing is untracked: "+output)
	isTrue(t, !strings.Contains(output, " D "), "nothing is an unstaged delete: "+output)
	movedAsRename := strings.Contains(output, "issues/open/"+updateTarget+".md -> issues/in-progress/"+updateTarget+".md")
	movedAsStagedDeleteAndAdd := strings.Contains(output, "D  issues/open/"+updateTarget+".md") &&
		strings.Contains(output, "A  issues/in-progress/"+updateTarget+".md")
	isTrue(t, movedAsRename || movedAsStagedDeleteAndAdd, "the move is staged in the git index: "+output)
}

func TestUpdateMovesAnIssueThatGitDoesNotTrackInAGitRepo(t *testing.T) {
	r := newTestRepo(t)
	r.git("init", "-q")
	r.git("add", "-A")
	r.git("-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "initial")
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	isTrue(t, !r.exists("issues/open/"+updateTarget+".md"), "the open file is gone")
	isTrue(t, r.exists("issues/in-progress/"+updateTarget+".md"), "the issue moved")
}

func TestUpdateMovesTheIssueWithPlainRenameWithoutGit(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	equal(t, 0, res.exit)
	isTrue(t, !r.exists("issues/open/"+updateTarget+".md"), "the open file is gone")
	isTrue(t, r.exists("issues/in-progress/"+updateTarget+".md"), "the issue moved")
}

func TestUpdateChangesPriorityEffortAndTags(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "open", "--priority", "critical", "--effort", "large", "--tags", "plan, widget-scheduler ,,repo")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	doc := r.frontmatter("issues/open/" + updateTarget + ".md")
	equal(t, "critical", doc.get("priority"))
	equal(t, "large", doc.get("effort"))
	equal(t, "[plan, widget-scheduler, repo]", doc.get("tags"))
}

func TestUpdateKeepsPriorityEffortAndTagsWhenNotGiven(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--priority", "high", "--effort", "large", "--tags", "plan")

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	equal(t, 0, res.exit)
	doc := r.frontmatter("issues/in-progress/" + updateTarget + ".md")
	equal(t, "high", doc.get("priority"))
	equal(t, "large", doc.get("effort"))
	equal(t, "[plan]", doc.get("tags"))
}

func TestUpdateRejectsAnInvalidPriorityOrEffort(t *testing.T) {
	cases := []struct{ flag, value string }{
		{"--priority", "urgent"},
		{"--effort", "huge"},
	}
	for _, c := range cases {
		t.Run(c.flag, func(t *testing.T) {
			r := newTestRepo(t)
			createUpdateTarget(r)

			res := r.run("update", updateTarget, "open", c.flag, c.value)

			equal(t, 1, res.exit)
			contains(t, res.error, "Invalid "+strings.TrimPrefix(c.flag, "--"))
		})
	}
}

func TestUpdateWarnsWhenAnInProgressIssueHasNoBookmark(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "in-progress")

	equal(t, 0, res.exit)
	contains(t, res.error, "has no bookmark")
	isTrue(t, r.exists("issues/in-progress/"+updateTarget+".md"), "the issue moved")
}

func TestUpdateAcceptsEveryFormOfTheName(t *testing.T) {
	for _, reference := range []string{updateTarget, "[[" + updateTarget + "]]", "#" + updateTarget, "issues/open/" + updateTarget + ".md"} {
		t.Run(reference, func(t *testing.T) {
			r := newTestRepo(t)
			createUpdateTarget(r)

			res := r.run("update", reference, "in-progress", "--bookmark", "ai/work")

			equal(t, 0, res.exit)
			isTrue(t, r.exists("issues/in-progress/"+updateTarget+".md"), "the issue moved")
		})
	}
}

func TestUpdateDefaultsToCompletedWhenClosingWithoutAResolution(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)
	r.mustRun("edit", updateTarget, "resolution", "Completed.")

	res := r.run("update", updateTarget, "closed")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "completed", r.frontmatter("issues/closed/"+updateTarget+".md").get("resolution"))
}

func TestUpdateRejectsAResolutionWhenTheNewStateIsNotClosed(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "in-progress", "--resolution", "completed")

	equal(t, 1, res.exit)
	contains(t, res.error, "only when the new state is closed")
	isTrue(t, r.exists("issues/open/"+updateTarget+".md"), "the issue did not move")
}

func TestUpdateRefusesToReopenAClosedIssue(t *testing.T) {
	for _, state := range []string{"open", "in-progress"} {
		t.Run(state, func(t *testing.T) {
			r := newTestRepo(t)
			createUpdateTarget(r, "--state", "closed", "--resolution", "completed")

			res := r.run("update", updateTarget, state)

			equal(t, 1, res.exit)
			contains(t, res.error, "Do not reopen a closed issue")
			contains(t, res.error, "[["+updateTarget+"]]")
			isTrue(t, r.exists("issues/closed/"+updateTarget+".md"), "the issue did not move")
		})
	}
}

func TestUpdateReportsAMissingIssue(t *testing.T) {
	r := newTestRepo(t)

	res := r.run("update", "no-such-issue", "open")

	equal(t, 1, res.exit)
	contains(t, res.error, "No issue named 'no-such-issue'")
}

func TestUpdateReportsAnIssueInMoreThanOneDirectory(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)
	r.write("issues/closed/"+updateTarget+".md", r.read("issues/open/"+updateTarget+".md"))

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	equal(t, 1, res.exit)
	contains(t, res.error, "more than one state directory")
}

func TestUpdateCorrectsAStateFieldThatDoesNotMatchTheDirectory(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)
	path := "issues/open/" + updateTarget + ".md"
	r.write(path, strings.Replace(r.read(path), "state: open ", "state: closed ", 1))

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	equal(t, 0, res.exit)
	contains(t, res.error, "its state field was 'closed'")
	equal(t, "in-progress", r.frontmatter("issues/in-progress/"+updateTarget+".md").get("state"))
}

func TestUpdateKeepsInlineCommentsAndTheBody(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)
	before := r.frontmatter("issues/open/" + updateTarget + ".md")

	r.mustRun("update", updateTarget, "in-progress", "--bookmark", "ai/work")

	after := r.frontmatter("issues/in-progress/" + updateTarget + ".md")
	isTrue(t, slices.ContainsFunc(after.front, func(l string) bool { return strings.HasPrefix(l, "state: in-progress ") && strings.Contains(l, " # ") }), "the state line keeps its comment")
	equalSlices(t, before.body, after.body)
	equal(t, len(before.front), len(after.front))
}
