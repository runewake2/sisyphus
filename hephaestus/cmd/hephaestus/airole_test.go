package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const reviewerRole = `---
role: reviewer
kind: ai
trigger: request
on-failure: needs-human
allowed-tools: [Bash, Read]
model: sonnet
---

# Reviewer

## Charter

Review the change of a task.

## Inputs

A task.

## Outputs

A review.

## Tools

sisyphus.

## Ask a human

When the change is unclear.

## Prompt

Review the change, and close the issue with your review.
`

// newAIRepo makes an agent repo with the AI role reviewer, whose fake model is runner.
func newAIRepo(t *testing.T) *testRepo {
	t.Helper()
	r := newAgentRepo(t)
	r.write("hephaestus/roles/reviewer.md", reviewerRole)
	r.commit("reviewer")
	return r
}

func TestAnAIWorkerRunsTheModelWithThePromptOfTheTask(t *testing.T) {
	r := newAIRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "--data", `{"bookmark": "ai/fix"}`, "Review the fix.")
	runner := `cat > prompt.txt; echo {role} > role.txt; sisyphus close "$HEPHAESTUS_MESSAGE_ISSUE" -m "Reviewed: no problems."`

	results := r.runAgent("reviewer", "--runner", runner)

	equal(t, "done", results[0].Status)
	equal(t, "reviewer\n", r.read("role.txt"))
	prompt := r.read("prompt.txt")
	for _, part := range []string{
		"You are reviewer-1, a worker of the hephaestus role reviewer.",
		"A message from alice, of the kind task:\n\n    Review the fix.",
		`"bookmark": "ai/fix"`,
		"The issue of the task is fix-login-bug. You have it in progress.",
		"## Charter\n\nReview the change of a task.",
		"## Prompt\n\nReview the change, and close the issue with your review.",
		"## Rules for every role",
	} {
		contains(t, prompt, part)
	}
	issue := r.showIssue("fix-login-bug")
	equal(t, "closed", issue["state"])
	equal(t, "reviewer-1", issue["owner"])
}

func TestAnAIWorkerCanStopItsIssue(t *testing.T) {
	r := newAIRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "Review it.")

	results := r.runAgent("reviewer", "--runner", `sisyphus update "$HEPHAESTUS_MESSAGE_ISSUE" open --add-tags needs-human`)

	equal(t, "stopped", results[0].Status)
	equal(t, "open", r.showIssue("fix-login-bug")["state"])
}

func TestAnIssueThatTheModelLeavesInProgressGoesToAHuman(t *testing.T) {
	r := newAIRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "Review it.")

	results := r.runAgent("reviewer", "--runner", `echo "I looked at it."`)

	equal(t, "failed", results[0].Status)
	equal(t, "needs-human", results[0].Next)
	issue := r.showIssue("fix-login-bug")
	equal(t, "open", issue["state"])
	contains(t, issue["body"].(string), "the AI worker reviewer-1 ran")
	contains(t, issue["body"].(string), "The model ended with the issue still in progress.")
}

func TestAFailedModelGoesToAHuman(t *testing.T) {
	r := newAIRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "Review it.")
	t.Setenv("HEPHAESTUS_AI_RUNNER", `echo "rate limited"; exit 1`)

	results := r.runAgent("reviewer")

	equal(t, "failed", results[0].Status)
	equal(t, "needs-human", results[0].Next)
	contains(t, toStrings(r.showIssue("fix-login-bug")["tags"])[len(toStrings(r.showIssue("fix-login-bug")["tags"]))-1], "needs-human")
}

func TestExportWritesAClaudeSubagentForEachAIRole(t *testing.T) {
	r := newAIRepo(t)

	r.mustRun("agent", "export", "claude")

	agent := r.read(".claude/agents/reviewer.md")
	isTrue(t, bytes.HasPrefix([]byte(agent), []byte("---\nname: reviewer\ndescription: \"Review the change of a task. A hephaestus role: use it for tasks delegated to reviewer.\"\ntools: Bash, Read\nmodel: sonnet\n---\n")), agent)
	contains(t, agent, "Generated from hephaestus/roles/reviewer.md")
	contains(t, agent, "## Prompt\n\nReview the change, and close the issue with your review.")
	isTrue(t, !r.exists(".claude/agents/stamp.md"), "a tool role has no subagent")
	equal(t, 0, r.run("agent", "export", "claude", "--check").exit)

	r.write("hephaestus/roles/reviewer.md", reviewerRole+"\nMore text.\n")
	stale := r.run("agent", "export", "claude", "--check")

	equal(t, 1, stale.exit)
	contains(t, stale.errors, "reviewer.md")
}

// TestTheSubagentsOfThisRepoAreInSync checks the subagents in .claude/agents of this repo against
// the role files, so that a change of a role file is exported in the same change.
func TestTheSubagentsOfThisRepoAreInSync(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "hephaestus", "roles")); err != nil {
		t.Skip("not in the sisyphus repo")
	}
	var out, errOut bytes.Buffer

	exit := run([]string{"-C", root, "agent", "export", "claude", "--check"}, &out, &errOut)

	if exit != 0 {
		t.Errorf("the subagents are not in sync: %s", errOut.String())
	}
}
