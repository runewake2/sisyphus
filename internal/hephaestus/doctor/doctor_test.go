package doctor

import (
	"strings"
	"testing"
	"time"

	"github.com/runewake2/sisyphus/internal/hephaestus/mq"
	"github.com/runewake2/sisyphus/internal/hephaestus/sisyphus"
)

var now = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

// base returns the input of a task that implementer-1 took: delegated, in progress, with its task
// message in the log.
func base() Input {
	return Input{
		Now:   now,
		Stale: 2 * time.Minute,
		Issue: sisyphus.Detail{
			Name: "fix-login-bug", Title: "Fix the login bug", State: "in-progress", Owner: "implementer-1",
			Metadata: map[string]string{"delegate": "implementer"},
		},
		Queue: "127.0.0.1:1",
		Workers: []mq.WorkerStatus{
			{Worker: "implementer-1", Role: "implementer", LastSeen: now.Add(-10 * time.Second), Holds: []mq.Held{{ID: "task-1", Until: now.Add(time.Minute)}}},
		},
		Log: []mq.Message{
			{ID: "task-1", From: "human", To: mq.Address{Role: "implementer"}, Kind: "task", Issue: "fix-login-bug", Time: now.Add(-time.Hour)},
		},
	}
}

func fixes(r Report) []string {
	var fixes []string
	for _, f := range r.Findings {
		fixes = append(fixes, f.Fix)
	}
	return fixes
}

func TestAWorkingTaskHasNoFindings(t *testing.T) {
	r := Diagnose(base())
	if len(r.Findings) != 0 {
		t.Fatalf("a worker that is live and holds the task gives findings: %+v", r.Findings)
	}
	if r.Role != "implementer" {
		t.Errorf("the role is %q, not implementer", r.Role)
	}
}

func TestAStoppedOwnerIsAFinding(t *testing.T) {
	in := base()
	in.Workers[0].LastSeen = now.Add(-time.Hour)
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Problem, "stopped") {
		t.Fatalf("a stopped owner gives %+v", r.Findings)
	}
	if r.Findings[0].Fix != "hephaestus kick fix-login-bug" {
		t.Errorf("the fix is %q, not the kick", r.Findings[0].Fix)
	}
}

func TestAnOwnerTheQueueDoesNotKnowIsAFinding(t *testing.T) {
	in := base()
	in.Workers = nil
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Problem, "does not know its owner") {
		t.Fatalf("an unknown owner gives %+v", r.Findings)
	}
}

func TestALiveOwnerWithoutTheTaskIsAFinding(t *testing.T) {
	in := base()
	in.Workers[0].Holds = nil
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Fix, "--force") {
		t.Fatalf("a live owner without the task gives %+v", r.Findings)
	}
}

func TestNeedsHumanIsAFinding(t *testing.T) {
	in := base()
	in.Issue.State, in.Issue.Owner = "open", ""
	in.Issue.Tags = []string{"needs-human"}
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Problem, "needs-human") {
		t.Fatalf("needs-human gives %+v", r.Findings)
	}
}

func TestADelegationWithoutATaskMessageIsAFinding(t *testing.T) {
	in := base()
	in.Issue.State, in.Issue.Owner = "open", ""
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Problem, "no task message waits") {
		t.Fatalf("a delegation without a message gives %+v", r.Findings)
	}
}

func TestAWaitingTaskWithoutALiveWorkerIsAFinding(t *testing.T) {
	in := base()
	in.Issue.State, in.Issue.Owner = "open", ""
	in.Workers = nil
	in.Pending = []mq.Pending{{ID: "task-1", Kind: "task", To: "role:implementer", For: "role:implementer", Issue: "fix-login-bug", Sent: now.Add(-time.Hour)}}
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Fix, "hephaestus agent run implementer") {
		t.Fatalf("a waiting task without a worker gives %+v", r.Findings)
	}
}

func TestNoQueueIsAFinding(t *testing.T) {
	in := base()
	in.Queue, in.Workers, in.Pending = "", nil, nil
	r := Diagnose(in)
	if len(fixes(r)) == 0 || !strings.Contains(fixes(r)[0], "hephaestus messages serve") {
		t.Fatalf("no queue gives %+v", r.Findings)
	}
}

func TestAnOverdueReferendumIsAFinding(t *testing.T) {
	in := base()
	in.Issue.Metadata["referendum"] = "round-2"
	in.Issue.Metadata["referendum-deadline"] = now.Add(-time.Hour).Format(time.RFC3339)
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Fix, "referendum decide") {
		t.Fatalf("an overdue referendum gives %+v", r.Findings)
	}
}

func TestAMissingBookmarkIsAFinding(t *testing.T) {
	in := base()
	in.Bookmark = &Bookmark{Name: "ai/fix-login-bug"}
	r := Diagnose(in)
	if len(r.Findings) != 1 || !strings.Contains(r.Findings[0].Problem, "does not exist") {
		t.Fatalf("a missing bookmark gives %+v", r.Findings)
	}
}

func TestTheRoleComesFromTheLastTaskMessage(t *testing.T) {
	in := base()
	delete(in.Issue.Metadata, "delegate")
	if role := Role(in); role != "implementer" {
		t.Fatalf("the role is %q, not implementer from the task message", role)
	}
}

func TestPlanKickOfAStoppedOwner(t *testing.T) {
	in := base()
	in.Workers[0].LastSeen = now.Add(-time.Hour)
	k, err := PlanKick(in, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !k.Reopen || k.Role != "implementer" || !k.Send || k.MessageID == "" {
		t.Fatalf("the plan is %+v", k)
	}
}

func TestPlanKickRefusesALiveOwnerWithoutTheTask(t *testing.T) {
	// The doctor sends this case to kick --force; a plain kick must refuse it, as its help says.
	in := base()
	in.Workers[0].Holds = nil
	if _, err := PlanKick(in, "", false); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("a live owner without the task is not refused: %v", err)
	}
	if _, err := PlanKick(in, "", true); err != nil {
		t.Fatalf("--force does not kick it: %v", err)
	}
}

func TestPlanKickRefusesWhenTheQueueDoesNotAnswer(t *testing.T) {
	// Without the workers and pending deliveries, the kick cannot tell a stuck task from a running
	// one: empty lists must not make the guards pass.
	in := base()
	in.QueueErr = "connection refused"
	in.Workers, in.Pending = nil, nil
	if _, err := PlanKick(in, "", false); err == nil || !strings.Contains(err.Error(), "does not answer") {
		t.Fatalf("a queue that does not answer is not refused: %v", err)
	}
	if _, err := PlanKick(in, "", true); err == nil {
		t.Fatal("--force kicks blind although the queue does not answer")
	}
}

func TestPlanKickRefusesALiveWorker(t *testing.T) {
	if _, err := PlanKick(base(), "", false); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("a live worker that holds the task is not refused: %v", err)
	}
	if _, err := PlanKick(base(), "", true); err != nil {
		t.Fatalf("--force does not kick it: %v", err)
	}
}

func TestPlanKickRefusesAClosedIssue(t *testing.T) {
	in := base()
	in.Issue.State = "closed"
	if _, err := PlanKick(in, "", false); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("a closed issue is not refused: %v", err)
	}
}

func TestPlanKickNeedsARole(t *testing.T) {
	in := base()
	in.Issue.State, in.Issue.Owner, in.Log = "open", "", nil
	delete(in.Issue.Metadata, "delegate")
	if _, err := PlanKick(in, "", false); err == nil || !strings.Contains(err.Error(), "--role") {
		t.Fatalf("a task without a role is not refused: %v", err)
	}
	k, err := PlanKick(in, "engineer", false)
	if err != nil || k.Role != "engineer" {
		t.Fatalf("--role does not name the role: %+v, %v", k, err)
	}
}

func TestPlanKickReusesAWaitingMessage(t *testing.T) {
	in := base()
	in.Issue.State, in.Issue.Owner, in.Workers = "open", "", nil
	in.Issue.Tags = []string{"needs-human"}
	in.Pending = []mq.Pending{{ID: "task-1", Kind: "task", To: "role:implementer", For: "role:implementer", Issue: "fix-login-bug", Sent: now.Add(-time.Hour)}}
	k, err := PlanKick(in, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if k.Send || !k.RemoveNeedsHuman || k.Reopen {
		t.Fatalf("the plan is %+v", k)
	}
	if len(k.Warnings) == 0 {
		t.Fatal("the plan does not warn that the message waits already")
	}
}
