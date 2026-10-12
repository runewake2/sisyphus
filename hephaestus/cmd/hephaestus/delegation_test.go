package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

// buildSisyphusOnce builds the sisyphus command of this repo, once, and puts it first on PATH:
// hephaestus uses sisyphus only through its command.
var buildSisyphusOnce = sync.OnceValue(func() error {
	dir, err := os.MkdirTemp("", "hephaestus-test-bin")
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "sisyphus"), "./cmd/sisyphus")
	cmd.Dir = filepath.Join("..", "..", "..") // the sisyphus module
	if out, err := cmd.CombinedOutput(); err != nil {
		return &buildError{string(out)}
	}
	return os.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
})

type buildError struct{ output string }

func (e *buildError) Error() string { return "building sisyphus: " + e.output }

const toolRole = `---
role: lint-runner
kind: tool
command: go vet ./...
trigger: [request, close]
on-failure: code-fixer
---

# Lint runner

## Charter

Run the linters on a task.

## Inputs

A task.

## Outputs

The result.

## Tools

go vet.

## Ask a human

When the linter cannot run.
`

const aiRole = `---
role: code-fixer
kind: ai
trigger: request
can-call-referendum: true
---

# Code fixer

## Charter

Fix what a linter finds.

## Inputs

A task.

## Outputs

A change.

## Tools

An editor.

## Ask a human

When the fix changes behavior.

## Prompt

Fix the problems, and change nothing else.
`

// newTrackerRepo makes a git repo with a sisyphus tracker, two roles, and three issues.
func newTrackerRepo(t *testing.T) *testRepo {
	t.Helper()
	if err := buildSisyphusOnce(); err != nil {
		t.Fatal(err)
	}
	r := newGitRepo(t)
	sisyphus := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("sisyphus", append([]string{"-C", r.root}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("sisyphus %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	sisyphus("init")
	sisyphus("new", "fix-login-bug")
	sisyphus("new", "add-dark-mode")
	sisyphus("new", "old-cleanup-task", "--state", "closed", "--metadata", "delegate=lint-runner")
	r.write("hephaestus/roles/lint-runner.md", toolRole)
	r.write("hephaestus/roles/code-fixer.md", aiRole)
	return r
}

func (r *testRepo) metadataOf(name string) map[string]string {
	r.t.Helper()
	cmd := exec.Command("sisyphus", "-C", r.root, "show", name, "--json")
	out, err := cmd.Output()
	if err != nil {
		r.t.Fatal(err)
	}
	var issue struct{ Metadata map[string]string }
	if err := json.Unmarshal(out, &issue); err != nil {
		r.t.Fatal(err)
	}
	return issue.Metadata
}

func TestRolesListsTheRoles(t *testing.T) {
	r := newTrackerRepo(t)

	res := r.mustRun("roles")

	equal(t, "role         kind  trigger        on-failure   charter\ncode-fixer   ai    request        needs-human  Fix what a linter finds.\nlint-runner  tool  request,close  code-fixer   Run the linters on a task.\n", res.output)
}

func TestRolesReportsEveryProblem(t *testing.T) {
	r := newTrackerRepo(t)
	r.write("hephaestus/roles/broken.md", "---\nrole: other\nkind: robot\ntrigger: never\n---\n\n# Broken\n\n## Charter\n\nText.\n")
	r.write("hephaestus/roles/no-command.md", "---\nrole: no-command\nkind: tool\ntrigger: request\non-failure: lint-runner\n---\n\n# No command\n\n## Charter\n\nA.\n\n## Inputs\n\nA.\n\n## Outputs\n\nA.\n\n## Tools\n\nA.\n\n## Ask a human\n\nA.\n")

	res := r.run("roles")

	equal(t, 1, res.exit)
	contains(t, res.errors, `broken.md: role is "other", but the file name is broken.md`)
	contains(t, res.errors, `broken.md: kind is "robot": use ai or tool`)
	contains(t, res.errors, `broken.md: trigger "never" is not one of: request, close, schedule`)
	contains(t, res.errors, `broken.md: the section "## Inputs" is missing or empty`)
	contains(t, res.errors, "no-command.md: a tool role needs a command")
	contains(t, res.output, "code-fixer") // the good roles are still listed
}

func TestOnFailureMustNameAnAIRole(t *testing.T) {
	r := newTrackerRepo(t)
	r.write("hephaestus/roles/code-fixer.md", toolRole[:0]+"---\nrole: code-fixer\nkind: tool\ncommand: fix\ntrigger: request\n---\n\n# Code fixer\n\n## Charter\n\nA.\n\n## Inputs\n\nA.\n\n## Outputs\n\nA.\n\n## Tools\n\nA.\n\n## Ask a human\n\nA.\n")

	res := r.run("roles")

	contains(t, res.errors, `lint-runner.md: on-failure names "code-fixer", which is not an AI role`)
}

func TestAnAIRoleNeedsAPrompt(t *testing.T) {
	r := newTrackerRepo(t)
	r.write("hephaestus/roles/code-fixer.md", aiRole[:len(aiRole)-len("## Prompt\n\nFix the problems, and change nothing else.\n")])

	res := r.run("roles")

	contains(t, res.errors, `code-fixer.md: the section "## Prompt" is missing or empty`)
}

func TestDelegateSetsTheMetadataKey(t *testing.T) {
	r := newTrackerRepo(t)

	res := r.mustRun("delegate", "fix-login-bug", "lint-runner")

	equal(t, "issues/open/fix-login-bug.md\n", res.output)
	equal(t, "lint-runner", r.metadataOf("fix-login-bug")["delegate"])
}

func TestDelegateRefusesARoleWithNoFile(t *testing.T) {
	r := newTrackerRepo(t)

	res := r.run("delegate", "fix-login-bug", "release-manager")

	equal(t, 1, res.exit)
	contains(t, res.errors, `No role "release-manager"`)
	contains(t, res.errors, "The roles are: code-fixer, lint-runner.")
	equal(t, "", r.metadataOf("fix-login-bug")["delegate"])
}

func TestDelegateGivesTheErrorOfSisyphus(t *testing.T) {
	r := newTrackerRepo(t)

	res := r.run("delegate", "no-such-issue", "lint-runner")

	equal(t, 1, res.exit)
	contains(t, res.errors, "Error: No issue named 'no-such-issue'")
}

func TestTasksListsTheReadyTasksOfARole(t *testing.T) {
	r := newTrackerRepo(t)
	r.mustRun("delegate", "fix-login-bug", "lint-runner")
	r.mustRun("delegate", "add-dark-mode", "code-fixer")

	ready := r.mustRun("tasks", "lint-runner")
	all := r.mustRun("tasks", "lint-runner", "--all", "--json")

	contains(t, ready.output, "fix-login-bug")
	notContains(t, ready.output, "add-dark-mode")
	notContains(t, ready.output, "old-cleanup-task")
	contains(t, all.output, `"name": "old-cleanup-task"`)
	contains(t, all.output, `"name": "fix-login-bug"`)
}

func TestClearRemovesTheDelegation(t *testing.T) {
	r := newTrackerRepo(t)
	r.mustRun("delegate", "fix-login-bug", "lint-runner")

	r.mustRun("delegate", "fix-login-bug", "--clear")
	res := r.mustRun("tasks", "lint-runner")

	equal(t, "", res.output)
	contains(t, res.errors, "No task delegated to lint-runner can start now.")
}
