package main

import (
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

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "samw/ai/work")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "issues/in-progress/"+updateTarget+".md", strings.TrimSpace(res.output))
	isTrue(t, !r.exists("issues/open/"+updateTarget+".md"), "the open file is gone")
	doc := r.frontmatter("issues/in-progress/" + updateTarget + ".md")
	equal(t, "in-progress", doc.get("state"))
	equal(t, "samw/ai/work", doc.get("bookmark"))
	equal(t, "", doc.get("resolution"))
	equal(t, "", doc.get("closed"))
}

func TestUpdateClosesAnIssueAsCompletedAndKeepsTheBookmark(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "samw/ai/work")

	res := r.run("update", updateTarget, "closed", "--resolution", "completed")

	equal(t, 0, res.exit)
	equal(t, "issues/closed/"+updateTarget+".md", strings.TrimSpace(res.output))
	isTrue(t, !r.exists("issues/in-progress/"+updateTarget+".md"), "the in-progress file is gone")
	doc := r.frontmatter("issues/closed/" + updateTarget + ".md")
	equal(t, "closed", doc.get("state"))
	equal(t, "completed", doc.get("resolution"))
	equal(t, today(), doc.get("closed"))
	equal(t, "samw/ai/work", doc.get("bookmark"))
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
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "samw/ai/work")

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
	createUpdateTarget(r, "--state", "in-progress", "--bookmark", "samw/ai/work")

	res := r.run("update", updateTarget, "in-progress")

	equal(t, 0, res.exit)
	equal(t, "", res.error)
	equal(t, "samw/ai/work", r.frontmatter("issues/in-progress/"+updateTarget+".md").get("bookmark"))
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

			res := r.run("update", reference, "in-progress", "--bookmark", "samw/ai/work")

			equal(t, 0, res.exit)
			isTrue(t, r.exists("issues/in-progress/"+updateTarget+".md"), "the issue moved")
		})
	}
}

func TestUpdateRequiresAResolutionToClose(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)

	res := r.run("update", updateTarget, "closed")

	equal(t, 1, res.exit)
	contains(t, res.error, "give --resolution")
	isTrue(t, r.exists("issues/open/"+updateTarget+".md"), "the issue did not move")
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

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "samw/ai/work")

	equal(t, 1, res.exit)
	contains(t, res.error, "more than one state directory")
}

func TestUpdateCorrectsAStateFieldThatDoesNotMatchTheDirectory(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)
	path := "issues/open/" + updateTarget + ".md"
	r.write(path, strings.Replace(r.read(path), "state: open ", "state: closed ", 1))

	res := r.run("update", updateTarget, "in-progress", "--bookmark", "samw/ai/work")

	equal(t, 0, res.exit)
	contains(t, res.error, "its state field was 'closed'")
	equal(t, "in-progress", r.frontmatter("issues/in-progress/"+updateTarget+".md").get("state"))
}

func TestUpdateKeepsInlineCommentsAndTheBody(t *testing.T) {
	r := newTestRepo(t)
	createUpdateTarget(r)
	before := r.frontmatter("issues/open/" + updateTarget + ".md")

	r.mustRun("update", updateTarget, "in-progress", "--bookmark", "samw/ai/work")

	after := r.frontmatter("issues/in-progress/" + updateTarget + ".md")
	isTrue(t, slices.ContainsFunc(after.front, func(l string) bool { return strings.HasPrefix(l, "state: in-progress ") && strings.Contains(l, " # ") }), "the state line keeps its comment")
	equalSlices(t, before.body, after.body)
	equal(t, len(before.front), len(after.front))
}
