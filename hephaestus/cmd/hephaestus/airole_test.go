package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "--data", `{"note": "ai/fix"}`, "Review the fix.")
	out := t.TempDir()
	runner := `cat > ` + out + `/prompt.txt; echo {role} > ` + out + `/role.txt; pwd > ` + out + `/pwd.txt; sisyphus close "$HEPHAESTUS_MESSAGE_ISSUE" -m "Reviewed: no problems." && git add -A && git -c user.name=t -c user.email=t@t commit -qm "Review fix-login-bug"`

	results := r.runAgent("reviewer", "--runner", runner)

	if results[0].Status != "done" {
		t.Fatalf("%+v", results[0])
	}
	read := func(name string) string {
		data, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	equal(t, "reviewer\n", read("role.txt"))
	// The task ran in a workspace of its own, which is gone now; its commit and bookmark stay.
	equal(t, results[0].Workspace+"\n", read("pwd.txt"))
	isTrue(t, results[0].Removed, "the workspace of the done task was not removed")
	_, err := os.Stat(results[0].Workspace)
	isTrue(t, os.IsNotExist(err), "the workspace directory is still there")
	contains(t, r.git("show", "ai/reviewer-fix-login-bug:issues/closed/fix-login-bug.md"), "owner: reviewer-1")
	prompt := read("prompt.txt")
	for _, part := range []string{
		"You are reviewer-1, a worker of the hephaestus role reviewer.",
		"Your workspace for this task is " + results[0].Workspace + ", with the bookmark ai/reviewer-fix-login-bug.",
		"A message from alice, of the kind task:\n\n    Review the fix.",
		`"note": "ai/fix"`,
		"The issue of the task is fix-login-bug. You have it in progress.",
		"## Charter\n\nReview the change of a task.",
		"## Prompt\n\nReview the change, and close the issue with your review.",
		"## Rules for every role",
	} {
		contains(t, prompt, part)
	}
}

func TestAnAIWorkerCanStopItsIssue(t *testing.T) {
	r := newAIRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "Review it.")

	results := r.runAgent("reviewer", "--runner", `sisyphus update "$HEPHAESTUS_MESSAGE_ISSUE" open --add-tags needs-human`)

	equal(t, "stopped", results[0].Status)
	equal[any](t, "open", showIssueIn(t, results[0].Workspace, "fix-login-bug")["state"])
}

func TestAnIssueThatTheModelLeavesInProgressGoesToAHuman(t *testing.T) {
	r := newAIRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "Review it.")

	results := r.runAgent("reviewer", "--runner", `echo "I looked at it."`)

	equal(t, "failed", results[0].Status)
	equal(t, "needs-human", results[0].Next)
	isTrue(t, !results[0].Removed, "the workspace of a failed task was removed")
	issue := showIssueIn(t, results[0].Workspace, "fix-login-bug") // the issue changed in the task workspace
	equal(t, "open", issue["state"])
	contains(t, issue["body"].(string), "the AI worker reviewer-1 ran")
	contains(t, issue["body"].(string), "The model ended with the issue still in progress.")
	contains(t, issue["body"].(string), "the workspace of the task is kept for a human: "+results[0].Workspace)
}

func TestAFailedModelGoesToAHuman(t *testing.T) {
	r := newAIRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "reviewer", "--kind", "task", "--issue", "fix-login-bug", "Review it.")
	t.Setenv("HEPHAESTUS_AI_RUNNER", `echo "rate limited"; exit 1`)

	results := r.runAgent("reviewer")

	equal(t, "failed", results[0].Status)
	equal(t, "needs-human", results[0].Next)
	contains(t, strings.Join(toStrings(showIssueIn(t, results[0].Workspace, "fix-login-bug")["tags"]), ","), "needs-human")
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

// showIssueIn returns sisyphus show --json of an issue in the working copy dir.
func showIssueIn(t *testing.T, dir, name string) map[string]any {
	t.Helper()
	out, err := exec.Command("sisyphus", "-C", dir, "show", name, "--json", "--all").Output()
	if err != nil {
		t.Fatalf("sisyphus show %s in %s: %v", name, dir, err)
	}
	var issue map[string]any
	if err := json.Unmarshal(out, &issue); err != nil {
		t.Fatal(err)
	}
	return issue
}
