package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/referendum"
)

// fakeModel is the runner of the AI reviewers of the tests. In v1, docs documents the flag in its
// worktree and votes modify; ux votes no. Every required reviewer accepts at the gate.
const fakeModel = `#!/bin/sh
prompt=$(cat)
case "$prompt" in *"Accept it or reject it"*)
  echo "I read the modification."
  echo '{"gate": "accept", "reasons": "Right in my scope."}'
  exit 0;;
esac
case "$HEPHAESTUS_ROLE" in
docs)
  case "$prompt" in *"contents: v1"*)
    printf '# Login\n\nUse --reset to start again. See [[no-such-page]].\n' > docs.md
    echo '{"vote": "modify", "reasons": "The --reset flag has no docs.", "modification": "Document --reset in docs.md."}';;
  *) echo '{"vote": "yes", "reasons": "The docs agree."}';;
  esac;;
ux) echo '{"vote": "no", "reasons": "The error message is unclear."}';;
broken) echo "I forgot the ballot.";;
esac
`

func aiRoleFile(name, extra string) string {
	return strings.Replace(strings.Replace(reviewerRole, "role: reviewer", "role: "+name+"\n"+extra, 1), "# Reviewer", "# "+name, 1)
}

// newReferendumRepo makes a tracker repo with reviewers, and the bookmark ai/fix of fix-login-bug.
func newReferendumRepo(t *testing.T) (*testRepo, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake model is sh")
	}
	r := newTrackerRepo(t)
	t.Setenv(ownerVariable, "")
	t.Setenv(roleVariable, "")
	t.Setenv("HEPHAESTUS_AI_RUNNER", "")
	r.write("hephaestus/roles/check-ok.md", toolRoleFile("check-ok", "true", "", ""))
	r.write("hephaestus/roles/docs.md", aiRoleFile("docs", ""))
	r.write("hephaestus/roles/ux.md", aiRoleFile("ux", ""))
	r.write("hephaestus/roles/broken.md", aiRoleFile("broken", ""))
	r.write("hephaestus/roles/lead.md", aiRoleFile("lead", "can-call-referendum: true"))
	r.commit("tracker and roles")
	r.git("checkout", "-q", "-b", "ai/fix")
	r.write("login.go", "package login // with --reset\n")
	r.commit("fix the login")
	r.git("checkout", "-q", "main")
	r.mustRunSisyphus("update", "fix-login-bug", "in-progress", "--owner", "alice", "--bookmark", "ai/fix")

	model := filepath.Join(t.TempDir(), "model.sh")
	if err := os.WriteFile(model, []byte(fakeModel), 0o755); err != nil {
		t.Fatal(err)
	}
	return r, model
}

func (r *testRepo) mustRunSisyphus(args ...string) {
	r.t.Helper()
	cmdArgs := append([]string{"-C", r.root}, args...)
	if out, err := exec.Command("sisyphus", cmdArgs...).CombinedOutput(); err != nil {
		r.t.Fatalf("sisyphus %v: %v\n%s", args, err, out)
	}
}

func TestAReferendumRunsItsRoundsAndRecordsThem(t *testing.T) {
	r, model := newReferendumRepo(t)

	res := r.run("referendum", "call", "fix-login-bug", "--question", "Merge ai/fix as the fix of fix-login-bug?",
		"--required", "check-ok,docs", "--optional", "ux", "--runner", model, "--json")

	if res.exit != 0 {
		t.Fatalf("the referendum did not pass: %s", res.errors)
	}
	var ref referendum.Referendum
	if err := json.Unmarshal([]byte(res.output), &ref); err != nil {
		t.Fatalf("%v\n%s", err, res.output)
	}
	id := time.Now().Format("2006-01-02") + "-fix-login-bug"
	equal(t, id, ref.ID)
	equal(t, referendum.Passed, ref.Result)
	equal(t, 2, len(ref.Rounds))
	equal(t, "a human", ref.CalledBy)
	equal(t, "Document --reset in docs.md.", ref.Versions[1].Modification.Summary)
	contains(t, ref.Versions[1].Modification.Diff, "+Use --reset to start again.")
	equal(t, 2, len(ref.Concerns))

	// The modification is the new version, on its own branch, and the issue now names it.
	v2 := "referendum/" + id + "/v2"
	equal(t, "# Login\n\nUse --reset to start again. See [[no-such-page]].\n", r.git("show", v2+":docs.md"))
	contains(t, r.git("log", "-1", "--format=%s", v2), "referendum "+id+": modification of docs")
	issue := r.showIssue("fix-login-bug")
	equal[any](t, v2, issue["bookmark"])
	metadata := r.metadataOf("fix-login-bug")
	equal(t, "passed", metadata["referendum"])
	equal(t, id, metadata["referendum-record"])
	notes := issue["body"].(string)
	contains(t, notes, "referendum round 1 on v1: check-ok (required) yes, docs (required) modify, ux (non-required) no. The gate accepted the modification of docs.")
	contains(t, notes, "referendum round 2 on v2:")
	contains(t, notes, "referendum result: passed. Every required reviewer voted yes on v2.")
	contains(t, notes, "Record: [["+id+"]].")

	// The reviewers worked in scratch worktrees: the working copy of the user did not change.
	isTrue(t, !r.exists("docs.md"), "a reviewer changed the working copy of the user")
	notContains(t, r.git("worktree", "list"), "hephaestus-referendum-")
	record := r.read("referendums/" + id + ".md")
	contains(t, record, "## Round 2: v2")
	contains(t, record, "| `ux` | non-required | no | The error message is unclear. |")
	contains(t, record, "### Gate 1.1: the modification of `docs`, accepted")
	// The record and the Notes link only to what exists, also with a link in the diff of a modification.
	r.mustRunSisyphus("lint")
}

func TestAReferendumFailsOnARequiredNo(t *testing.T) {
	r, model := newReferendumRepo(t)

	res := r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "ux", "--runner", model)

	equal(t, 1, res.exit)
	contains(t, res.output, "Referendum "+time.Now().Format("2006-01-02")+"-fix-login-bug on fix-login-bug: failed")
	contains(t, res.errors, "The referendum on fix-login-bug has the result failed. A required reviewer voted no. ux: The error message is unclear.")
	equal(t, "failed", r.metadataOf("fix-login-bug")["referendum"])
	equal[any](t, "ai/fix", r.showIssue("fix-login-bug")["bookmark"])
}

func TestAReviewerWithNoBallotGoesToAHumanWhoDecides(t *testing.T) {
	r, model := newReferendumRepo(t)

	res := r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "broken", "--runner", model)

	equal(t, 1, res.exit)
	equal(t, "no-result", r.metadataOf("fix-login-bug")["referendum"])
	contains(t, strings.Join(toStrings(r.showIssue("fix-login-bug")["tags"]), ","), "needs-human")

	r.mustRun("referendum", "decide", "fix-login-bug", "passed", "--reason", "I checked the fix myself.")

	equal(t, "passed", r.metadataOf("fix-login-bug")["referendum"])
	id := time.Now().Format("2006-01-02") + "-fix-login-bug"
	contains(t, r.read("referendums/"+id+".md"), "## Human decision\n\n")
	contains(t, r.read("referendums/"+id+".md"), "a human decided **passed** (the result was no-result). I checked the fix myself.")
	notes := r.showIssue("fix-login-bug")["body"].(string)
	contains(t, notes, "a human decided the referendum: passed (the result was no-result).")
}

func TestOnlyAnAllowedRoleCallsAReferendum(t *testing.T) {
	r, model := newReferendumRepo(t)

	t.Setenv(roleVariable, "docs")
	res := r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "check-ok", "--runner", model)
	equal(t, 1, res.exit)
	contains(t, res.errors, "The role docs cannot call a referendum: its role file does not have can-call-referendum: true.")

	t.Setenv(roleVariable, "lead")
	t.Setenv(ownerVariable, "lead-1")
	res = r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "check-ok", "--json")
	equal(t, 0, res.exit)
	contains(t, res.output, `"called-by": "the role lead (lead-1)"`)

	res = r.run("referendum", "decide", "fix-login-bug", "failed", "--reason", "no")
	equal(t, 1, res.exit)
	contains(t, res.errors, "Only a human decides a referendum")
}

func TestAReferendumChecksItsCallFirst(t *testing.T) {
	r, _ := newReferendumRepo(t)

	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--required", "check-ok"}, "Give the question of the referendum with --question"},
		{[]string{"-q", "?", "--optional", "check-ok"}, "A referendum needs at least one required reviewer"},
		{[]string{"-q", "?", "--required", "docs"}, "The role docs is an AI role: a model does its work, and a model costs tokens."},
		{[]string{"-q", "?", "--required", "nobody"}, `No role "nobody"`},
		{[]string{"-q", "?", "--required", "check-ok", "--bookmark", "ai/none"}, `The bookmark "ai/none" does not exist.`},
	} {
		res := r.run(append([]string{"referendum", "call", "fix-login-bug"}, c.args...)...)
		equal(t, 1, res.exit)
		contains(t, res.errors, c.want)
	}
	res := r.run("referendum", "call", "add-dark-mode", "-q", "?", "--required", "check-ok")
	contains(t, res.errors, "The issue add-dark-mode has no bookmark. Give the implementation with --bookmark.")

	r.mustRunSisyphus("update", "fix-login-bug", "--metadata", "referendum=round-2")
	res = r.run("referendum", "call", "fix-login-bug", "-q", "?", "--required", "check-ok")
	contains(t, res.errors, "The issue fix-login-bug has a referendum that runs (referendum=round-2")
}

func TestAReferendumThatTimesOutFailsAndStopsItsReviewers(t *testing.T) {
	r, _ := newReferendumRepo(t)
	marker := filepath.Join(t.TempDir(), "still-running")
	slow := filepath.Join(t.TempDir(), "slow.sh")
	if err := os.WriteFile(slow, []byte("#!/bin/sh\ncat >/dev/null\nsleep 3\ntouch "+marker+"\necho '{\"vote\": \"yes\"}'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	started := time.Now()

	res := r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "docs", "--runner", slow, "--timeout", "500ms")

	equal(t, 1, res.exit)
	isTrue(t, time.Since(started) < 2500*time.Millisecond, "the referendum waited for its reviewer after the timeout")
	contains(t, res.errors, "has the result failed. The referendum timed out: it had no result within 500ms.")
	equal(t, "failed", r.metadataOf("fix-login-bug")["referendum"])
	issue := r.showIssue("fix-login-bug")
	equal[any](t, "in-progress", issue["state"]) // the issue stays open for its owner
	notContains(t, strings.Join(toStrings(issue["tags"]), ","), "needs-human")
	contains(t, issue["body"].(string), "referendum result: failed. The referendum timed out")
	time.Sleep(3 * time.Second)
	_, err := os.Stat(marker)
	isTrue(t, os.IsNotExist(err), "the reviewer went on after the timeout")
	notContains(t, r.git("worktree", "list"), "hephaestus-referendum-")
}

func TestAReferendumPastItsDeadlineIsRecordedAsFailed(t *testing.T) {
	r, model := newReferendumRepo(t)
	r.write("referendums/2026-10-01-fix-login-bug.md", "# Referendum: fix-login-bug\n\n- Result: **running**\n")
	r.mustRunSisyphus("update", "fix-login-bug", "--metadata", "referendum=round-2", "--metadata", "referendum-record=2026-10-01-fix-login-bug",
		"--metadata", "referendum-deadline="+time.Now().Add(time.Hour).UTC().Format(time.RFC3339))

	res := r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "check-ok", "--runner", model)
	equal(t, 1, res.exit)
	contains(t, res.errors, "has a referendum that runs (referendum=round-2, record 2026-10-01-fix-login-bug), until ")

	r.mustRunSisyphus("update", "fix-login-bug", "--metadata", "referendum-deadline="+time.Now().Add(-time.Minute).UTC().Format(time.RFC3339))
	res = r.run("referendum", "call", "fix-login-bug", "-q", "Merge it?", "--required", "check-ok", "--runner", model)

	equal(t, 0, res.exit)
	contains(t, res.errors, "The referendum 2026-10-01-fix-login-bug on fix-login-bug had no result by its deadline: it is recorded as failed.")
	contains(t, r.read("referendums/2026-10-01-fix-login-bug.md"), "## Timed out\n\n")
	body := r.showIssue("fix-login-bug")["body"].(string)
	contains(t, body, "referendum result: failed. The referendum timed out: it had no result by its deadline")
	contains(t, body, "referendum result: passed.")
}
