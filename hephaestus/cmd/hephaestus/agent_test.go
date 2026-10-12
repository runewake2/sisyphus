package main

import (
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/runewake2/sisyphus/hephaestus/internal/toolagent"
)

// toolRoleFile returns a role file of a tool role.
func toolRoleFile(name, command, onFailure, listens string) string {
	extra := ""
	if onFailure != "" {
		extra += "on-failure: " + onFailure + "\n"
	}
	if listens != "" {
		extra += "listens: " + listens + "\n"
	}
	return "---\nrole: " + name + "\nkind: tool\ncommand: " + command + "\ntrigger: request\n" + extra +
		"---\n\n# " + name + "\n\n## Charter\n\nA test role.\n\n## Inputs\n\nA task.\n\n## Outputs\n\nA result.\n\n## Tools\n\nsh.\n\n## Ask a human\n\nOn failure.\n"
}

// newAgentRepo makes a tracker repo with the tool roles of the tests, and a running message queue.
func newAgentRepo(t *testing.T) *testRepo {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the commands of the test roles are sh")
	}
	r := newTrackerRepo(t)
	t.Setenv(ownerVariable, "")
	r.write("hephaestus/roles/stamp.md", toolRoleFile("stamp", `echo "$HEPHAESTUS_MESSAGE_ISSUE by $SISYPHUS_OWNER" >> stamps.txt && echo "Stamped $HEPHAESTUS_MESSAGE_ISSUE."`, "", ""))
	r.write("hephaestus/roles/broken.md", toolRoleFile("broken", `echo "the build is broken"; exit 3`, "", ""))
	r.write("hephaestus/roles/handover.md", toolRoleFile("handover", `echo "cannot fix it"; exit 2`, "code-fixer", ""))
	r.write("hephaestus/roles/on-ready.md", toolRoleFile("on-ready", `echo "saw $HEPHAESTUS_MESSAGE_KIND" >> stamps.txt`, "", "[bookmark-ready]"))
	r.commit("roles")
	startQueue(t, r)
	return r
}

func (r *testRepo) runAgent(role string, args ...string) []toolagent.Result {
	r.t.Helper()
	res := r.mustRun(append([]string{"agent", "run", role, "--once", "--json"}, args...)...)
	var results []toolagent.Result
	if err := json.Unmarshal([]byte(res.output), &results); err != nil {
		r.t.Fatalf("%v\n%s", err, res.output)
	}
	return results
}

func (r *testRepo) showIssue(name string) map[string]any {
	r.t.Helper()
	out, err := exec.Command("sisyphus", "-C", r.root, "show", name, "--json", "--all").Output()
	if err != nil {
		r.t.Fatal(err)
	}
	var issue map[string]any
	if err := json.Unmarshal(out, &issue); err != nil {
		r.t.Fatal(err)
	}
	return issue
}

func TestAToolAgentDoesATaskAndClosesItsIssue(t *testing.T) {
	r := newAgentRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "stamp", "--kind", "task", "--issue", "fix-login-bug", "Stamp it.")

	results := r.runAgent("stamp")

	equal(t, 1, len(results))
	equal(t, "done", results[0].Status)
	equal(t, "fix-login-bug by stamp-1\n", r.read("stamps.txt"))
	issue := r.showIssue("fix-login-bug")
	equal(t, "closed", issue["state"])
	equal(t, "stamp-1", issue["owner"])
	contains(t, issue["body"].(string), "Done by the tool agent stamp-1 (role `stamp`)")
	contains(t, issue["body"].(string), "Stamped fix-login-bug.")
	// The sender gets the result, and the task does not come again.
	reply := receivedJSON(t, r, "--as", "alice")
	equal(t, 1, len(reply))
	equal(t, "result", reply[0].Kind)
	contains(t, string(reply[0].Data), `"status": "done"`)
	equal(t, 0, len(r.runAgent("stamp")))
}

func TestAFailedTaskGoesToAHuman(t *testing.T) {
	r := newAgentRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--to", "broken-1", "--kind", "task", "--issue", "fix-login-bug", "Build it.")

	results := r.runAgent("broken")

	equal(t, "failed", results[0].Status)
	equal(t, 3, results[0].Exit)
	equal(t, "needs-human", results[0].Next)
	issue := r.showIssue("fix-login-bug")
	equal(t, "open", issue["state"])
	contains(t, strings.Join(toStrings(issue["tags"]), ","), "needs-human")
	contains(t, issue["body"].(string), "the tool agent broken-1 ran `echo \"the build is broken\"; exit 3`: exit code 3.")
	contains(t, issue["body"].(string), "the build is broken")
}

func TestAFailedTaskGoesToTheAIRoleOfOnFailure(t *testing.T) {
	r := newAgentRepo(t)
	r.mustRun("messages", "send", "--from", "alice", "--role", "handover", "--kind", "task", "--issue", "add-dark-mode", "Fix it.")

	results := r.runAgent("handover")

	equal(t, "code-fixer", results[0].Next)
	issue := r.showIssue("add-dark-mode")
	equal(t, "open", issue["state"])
	equal(t, "code-fixer", issue["metadata"].(map[string]any)["delegate"])
	task := receivedJSON(t, r, "--as", "code-fixer-1", "--role", "code-fixer")
	equal(t, 1, len(task))
	equal(t, "task", task[0].Kind)
	equal(t, "add-dark-mode", task[0].Issue)
	contains(t, task[0].Body, "could not do this task")
}

func TestAToolAgentDoesNotTakeAClaimedIssue(t *testing.T) {
	r := newAgentRepo(t)
	if out, err := exec.Command("sisyphus", "-C", r.root, "start", "fix-login-bug", "--owner", "bob").CombinedOutput(); err != nil {
		t.Fatal(string(out))
	}
	r.mustRun("messages", "send", "--from", "alice", "--role", "stamp", "--kind", "task", "--issue", "fix-login-bug", "Stamp it.")

	results := r.runAgent("stamp")

	equal(t, "refused", results[0].Status)
	contains(t, results[0].Detail, "owned by bob")
	isTrue(t, !r.exists("stamps.txt"), "the command did not run")
	equal(t, "bob", r.showIssue("fix-login-bug")["owner"])
}

func TestAToolAgentListensForItsBroadcasts(t *testing.T) {
	r := newAgentRepo(t)
	r.runAgent("on-ready") // the worker starts to listen
	r.mustRun("messages", "send", "--from", "merge-queue-1", "--broadcast", "--kind", "bookmark-ready", "ai/fix is ready")
	r.mustRun("messages", "send", "--from", "merge-queue-1", "--broadcast", "--kind", "other", "not for it")

	results := r.runAgent("on-ready")

	equal(t, 1, len(results))
	equal(t, "done", results[0].Status)
	equal(t, "saw bookmark-ready\n", r.read("stamps.txt"))
}

func TestAnAIRoleNeedsARunner(t *testing.T) {
	r := newAgentRepo(t)
	t.Setenv("HEPHAESTUS_AI_RUNNER", "")

	res := r.run("agent", "run", "code-fixer", "--once")

	equal(t, 1, res.exit)
	contains(t, res.errors, "The role code-fixer is an AI role: a model does its work, and a model costs tokens.")
	contains(t, res.errors, "claude -p --agent {role}")
}

func TestAToolReviewerVotes(t *testing.T) {
	r := newAgentRepo(t)
	r.write("hephaestus/roles/check-ok.md", toolRoleFile("check-ok", "true", "", ""))
	r.write("hephaestus/roles/check-no.md", toolRoleFile("check-no", `echo "2 problems"; exit 1`, "", ""))
	r.write("hephaestus/roles/formatter.md", toolRoleFile("formatter", `printf '# Readme, formatted\n' > README.md`, "", ""))
	r.commit("reviewers")

	ballot := func(role string) toolagent.Ballot {
		t.Helper()
		var b toolagent.Ballot
		if err := json.Unmarshal([]byte(r.mustRun("agent", "review", role, "--json").output), &b); err != nil {
			t.Fatal(err)
		}
		return b
	}
	yes, no, modify := ballot("check-ok"), ballot("check-no"), ballot("formatter")

	equal(t, "yes", yes.Vote)
	equal(t, "no", no.Vote)
	equal(t, 1, no.Exit)
	contains(t, no.Reasons, "2 problems")
	equal(t, "modify", modify.Vote)
	equal(t, "README.md", strings.Join(modify.Files, ","))
	contains(t, modify.Diff, "+# Readme, formatted")
}

func toStrings(value any) []string {
	var result []string
	if items, ok := value.([]any); ok {
		for _, item := range items {
			result = append(result, item.(string))
		}
	}
	return result
}
