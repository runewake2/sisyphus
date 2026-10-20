// Package doctor diagnoses the task of one issue that hephaestus manages: whether a worker works on
// it, waits for it, stopped on it, or lost it, and what makes it run again. The doctor command
// reports its findings; the kick command acts on its plan ([[agent-delegation]]).
package doctor

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/runewake2/sisyphus/internal/hephaestus/mq"
	"github.com/runewake2/sisyphus/internal/hephaestus/sisyphus"
)

// Input is what the doctor reads about one task. The caller gathers it, so that a test can give it.
type Input struct {
	Now time.Time
	// Stale is the time after which a worker that has not spoken to the queue counts as stopped.
	Stale time.Duration
	// Issue is the issue of the task.
	Issue sisyphus.Detail
	// Queue is the address of the message queue, or empty when it does not run.
	Queue    string
	QueueErr string
	Workers  []mq.WorkerStatus
	Pending  []mq.Pending
	// Log is the message log, oldest first. It works also when the queue does not run.
	Log []mq.Message
	// Workspace is the task workspace of the issue, when one exists.
	Workspace *Workspace
	// Bookmark is the bookmark of the issue, when the issue has one.
	Bookmark *Bookmark
}

// Workspace is the task workspace of an issue.
type Workspace struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
	// Dirty is true when the workspace has changes that no commit keeps.
	Dirty bool `json:"dirty"`
}

// Bookmark is the bookmark of an issue, as the repo sees it.
type Bookmark struct {
	Name   string `json:"name"`
	Exists bool   `json:"exists"`
	Head   string `json:"head,omitempty"`
	// Merged is true when the base already has the head.
	Merged bool `json:"merged"`
}

// Finding is one problem of the task, with what fixes it.
type Finding struct {
	Problem string `json:"problem"`
	Fix     string `json:"fix"`
}

// Report is the diagnosis of one task.
type Report struct {
	Issue string `json:"issue"`
	Title string `json:"title"`
	State string `json:"state"`
	Owner string `json:"owner,omitempty"`
	// Role is the role of the task: the metadata delegate, or the receiver of its last task message.
	Role string `json:"role,omitempty"`
	// Facts are the observations that are not problems.
	Facts    []string  `json:"facts"`
	Findings []Finding `json:"findings"`
}

// delegateKey is the sisyphus metadata key of the role that owns a task.
const delegateKey = "delegate"

// needsHuman is the tag of an issue that a worker stopped for a human.
const needsHuman = "needs-human"

// Role returns the role of the task of the issue: the metadata delegate, or the role that its last
// task message went to. Empty when neither names one.
func Role(in Input) string {
	if role := in.Issue.Metadata[delegateKey]; role != "" {
		return role
	}
	tasks := taskMessages(in)
	if len(tasks) == 0 {
		return ""
	}
	last := tasks[len(tasks)-1]
	if last.To.Role != "" {
		return last.To.Role
	}
	for _, w := range in.Workers {
		if w.Worker == last.To.Worker {
			return w.Role
		}
	}
	return ""
}

// taskMessages returns the messages of the log that are tasks of the issue: to one worker, or to one
// worker of a role, with the issue; oldest first. A broadcast or a message to every worker of a role
// is a notice, not a task.
func taskMessages(in Input) []mq.Message {
	var tasks []mq.Message
	for _, m := range in.Log {
		if m.Issue == in.Issue.Name && !m.To.Broadcast && !m.To.All && m.Kind == "task" {
			tasks = append(tasks, m)
		}
	}
	return tasks
}

// pendingTasks returns the deliveries of task messages of the issue that still wait.
func pendingTasks(in Input) []mq.Pending {
	var waits []mq.Pending
	for _, p := range in.Pending {
		if p.Issue == in.Issue.Name && p.Kind == "task" && p.To != "broadcast" && !strings.HasSuffix(p.To, ":all") {
			waits = append(waits, p)
		}
	}
	return waits
}

// workerOf returns the worker of the queue with the name, or nil.
func workerOf(in Input, name string) *mq.WorkerStatus {
	for i := range in.Workers {
		if in.Workers[i].Worker == name {
			return &in.Workers[i]
		}
	}
	return nil
}

// alive reports whether the worker spoke to the queue within the stale time.
func alive(in Input, w *mq.WorkerStatus) bool {
	return w != nil && !w.LastSeen.IsZero() && in.Now.Sub(w.LastSeen) <= in.Stale
}

// holdsTask reports whether the worker holds a lease on a task message of the issue.
func holdsTask(in Input, w *mq.WorkerStatus) bool {
	if w == nil {
		return false
	}
	tasks := taskMessages(in)
	for _, held := range w.Holds {
		for _, m := range tasks {
			if m.ID == held.ID {
				return true
			}
		}
	}
	return false
}

// liveRole reports whether a live worker of the role runs.
func liveRole(in Input, role string) bool {
	for i := range in.Workers {
		if in.Workers[i].Role == role && alive(in, &in.Workers[i]) {
			return true
		}
	}
	return false
}

// ago writes how long ago t was, or that it never happened.
func ago(now, t time.Time) string {
	if t.IsZero() {
		return "never since the queue started"
	}
	d := now.Sub(t).Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// Diagnose reads the task of the issue and reports what it sees: the facts, and a finding with its
// fix for each problem.
func Diagnose(in Input) Report {
	r := Report{Issue: in.Issue.Name, Title: in.Issue.Title, State: in.Issue.State, Owner: in.Issue.Owner, Role: Role(in), Facts: []string{}, Findings: []Finding{}}
	kick := "hephaestus kick " + r.Issue

	switch {
	case in.QueueErr != "":
		r.Findings = append(r.Findings, Finding{
			Problem: "The message queue does not answer: " + in.QueueErr,
			Fix:     "Start it again: hephaestus messages serve",
		})
	case in.Queue == "":
		r.Findings = append(r.Findings, Finding{
			Problem: "No message queue runs for the repo, so no worker gets a task.",
			Fix:     "Start it: hephaestus messages serve",
		})
	}

	if r.Role == "" {
		r.Facts = append(r.Facts, "No role owns the task: the issue has no delegate metadata, and no task message names it.")
		if r.State == "open" && r.Owner == "" {
			r.Findings = append(r.Findings, Finding{
				Problem: "The issue is open, but no role owns it, so no worker will take it.",
				Fix:     fmt.Sprintf("Delegate it: hephaestus delegate %s <role>, then %s", r.Issue, kick),
			})
		}
	} else {
		r.Facts = append(r.Facts, fmt.Sprintf("The role of the task is %s.", r.Role))
	}

	tasks := taskMessages(in)
	if len(tasks) > 0 {
		last := tasks[len(tasks)-1]
		r.Facts = append(r.Facts, fmt.Sprintf("The last task message is %s, from %s to %s, sent %s.", last.ID, last.From, last.To.String(), ago(in.Now, last.Time)))
	} else {
		r.Facts = append(r.Facts, "The log has no task message for the issue.")
	}

	waits := pendingTasks(in)
	queueKnown := in.Queue != "" && in.QueueErr == ""

	switch r.State {
	case "in-progress":
		diagnoseInProgress(in, &r, waits, queueKnown, kick)
	case "open":
		diagnoseOpen(in, &r, waits, queueKnown, kick)
	case "closed":
		r.Facts = append(r.Facts, "The issue is closed: the task is done.")
	}

	if b := in.Bookmark; b != nil {
		switch {
		case !b.Exists:
			r.Findings = append(r.Findings, Finding{
				Problem: fmt.Sprintf("The bookmark %s of the issue does not exist in the repo.", b.Name),
				Fix:     "Read the Notes of the issue: the work was never committed, or the bookmark name is wrong.",
			})
		case b.Merged:
			r.Facts = append(r.Facts, fmt.Sprintf("The bookmark %s is at %.12s, and the base has it already.", b.Name, b.Head))
		default:
			r.Facts = append(r.Facts, fmt.Sprintf("The bookmark %s is at %.12s; the base does not have it yet.", b.Name, b.Head))
		}
	}
	if ws := in.Workspace; ws != nil {
		dirty := ""
		if ws.Dirty {
			dirty = ", with changes that no commit keeps"
		}
		r.Facts = append(r.Facts, fmt.Sprintf("The task workspace %s exists%s; a kicked task continues there.", ws.Dir, dirty))
	}

	diagnoseReferendum(in, &r)
	return r
}

// diagnoseInProgress finds what is wrong with a task whose issue is in progress: a worker should
// work on it.
func diagnoseInProgress(in Input, r *Report, waits []mq.Pending, queueKnown bool, kick string) {
	if r.Owner == "" {
		r.Findings = append(r.Findings, Finding{
			Problem: "The issue is in progress with no owner: no worker works on it, and no worker can start it.",
			Fix:     kick,
		})
		return
	}
	if !queueKnown {
		r.Facts = append(r.Facts, fmt.Sprintf("The owner is %s; without the queue, the doctor cannot tell whether it works.", r.Owner))
		return
	}
	w := workerOf(in, r.Owner)
	switch {
	case w == nil:
		r.Findings = append(r.Findings, Finding{
			Problem: fmt.Sprintf("The issue is in progress, but the queue does not know its owner %s. When no one works on it, the task is stuck.", r.Owner),
			Fix:     kick,
		})
	case !alive(in, w):
		r.Findings = append(r.Findings, Finding{
			Problem: fmt.Sprintf("The owner %s stopped: last seen %s. The task does not run, and its message will not come back by itself, because the worker acknowledged it.", r.Owner, ago(in.Now, w.LastSeen)),
			Fix:     kick,
		})
	case holdsTask(in, w):
		r.Facts = append(r.Facts, fmt.Sprintf("The worker %s works on the task: last seen %s, and it holds the lease of the task message.", r.Owner, ago(in.Now, w.LastSeen)))
	default:
		r.Findings = append(r.Findings, Finding{
			Problem: fmt.Sprintf("The owner %s is live (last seen %s), but holds no message of this task: it acknowledged the task and left the issue in progress.", r.Owner, ago(in.Now, w.LastSeen)),
			Fix:     kick + " --force (the worker is live)",
		})
	}
	_ = waits
}

// diagnoseOpen finds what is wrong with a task whose issue is open: a worker should take it.
func diagnoseOpen(in Input, r *Report, waits []mq.Pending, queueKnown bool, kick string) {
	if slices.Contains(in.Issue.Tags, needsHuman) {
		r.Findings = append(r.Findings, Finding{
			Problem: "A worker stopped the task for a human (the tag needs-human). The reason is in the Notes of the issue.",
			Fix:     "Correct the cause (for example new API keys, or the question in the Notes), then: " + kick,
		})
	}
	if !queueKnown {
		return
	}
	if len(waits) == 0 {
		if r.Role != "" && !slices.Contains(in.Issue.Tags, needsHuman) {
			r.Findings = append(r.Findings, Finding{
				Problem: fmt.Sprintf("The issue is delegated to %s, but no task message waits: nothing will start the task.", r.Role),
				Fix:     kick,
			})
		}
		return
	}
	for _, p := range waits {
		switch {
		case p.LeasedBy != "":
			holder := workerOf(in, p.LeasedBy)
			if alive(in, holder) {
				r.Facts = append(r.Facts, fmt.Sprintf("The task message %s is leased by %s, live: it works on the task.", p.ID, p.LeasedBy))
			} else {
				r.Findings = append(r.Findings, Finding{
					Problem: fmt.Sprintf("The task message %s is leased by %s, which stopped. The queue gives the message to another worker when the lease ends, at %s.", p.ID, p.LeasedBy, p.LeaseUntil.Local().Format("15:04:05")),
					Fix:     fmt.Sprintf("Wait for the lease, and have a live worker of %s: hephaestus agent run %s", r.Role, r.Role),
				})
			}
		case liveRole(in, roleOf(p.For)):
			r.Facts = append(r.Facts, fmt.Sprintf("The task message %s waits for %s; a live worker of the role runs and will take it.", p.ID, p.For))
		default:
			r.Findings = append(r.Findings, Finding{
				Problem: fmt.Sprintf("The task message %s waits for %s, but no live worker of the role runs.", p.ID, p.For),
				Fix:     fmt.Sprintf("Start one: hephaestus agent run %s", roleOf(p.For)),
			})
		}
	}
}

// roleOf returns the role of a pending receiver "role:<role>", or the whole receiver.
func roleOf(waitsFor string) string {
	return strings.TrimPrefix(waitsFor, "role:")
}

// diagnoseReferendum reports the referendum of the issue, as status does for the repo.
func diagnoseReferendum(in Input, r *Report) {
	value := in.Issue.Metadata["referendum"]
	decide := fmt.Sprintf("hephaestus referendum decide %s passed|failed --reason <why>", r.Issue)
	switch {
	case strings.HasPrefix(value, "round-"):
		deadline, err := time.Parse(time.RFC3339, in.Issue.Metadata["referendum-deadline"])
		if err == nil && in.Now.After(deadline) {
			r.Findings = append(r.Findings, Finding{
				Problem: fmt.Sprintf("The referendum of the issue (%s) is past its deadline, %s.", value, deadline.Local().Format("2006-01-02 15:04")),
				Fix:     "The next referendum call records it as failed, or decide it: " + decide,
			})
			return
		}
		r.Facts = append(r.Facts, fmt.Sprintf("A referendum runs on the issue (%s).", value))
	case value == "no-result":
		r.Findings = append(r.Findings, Finding{
			Problem: "The referendum of the issue ended with no result.",
			Fix:     "A human decides it: " + decide,
		})
	case value == "failed" && in.Issue.State != "closed":
		r.Facts = append(r.Facts, "The last referendum of the issue failed: the owner corrects the work and calls a new one, or a human overrules it.")
	}
}

// Text writes the report for a person.
func (r Report) Text() string {
	var b strings.Builder
	owner := ""
	if r.Owner != "" {
		owner = ", owner " + r.Owner
	}
	role := "no role"
	if r.Role != "" {
		role = "role " + r.Role
	}
	fmt.Fprintf(&b, "%s: %s\n  %s%s, %s\n", r.Issue, r.Title, r.State, owner, role)
	for _, fact := range r.Facts {
		fmt.Fprintf(&b, "  %s\n", fact)
	}
	fmt.Fprintf(&b, "\nFindings (%d)\n", len(r.Findings))
	for _, f := range r.Findings {
		fmt.Fprintf(&b, "  - %s\n    Fix: %s\n", f.Problem, f.Fix)
	}
	if len(r.Findings) == 0 {
		fmt.Fprintf(&b, "  none: the task is not stuck\n")
	}
	return b.String()
}
