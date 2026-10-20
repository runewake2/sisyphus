package doctor

import (
	"fmt"
	"path"
	"slices"
)

// Kick is the plan that makes a stuck task run again. PlanKick makes it; the kick command acts on
// it: it reopens the issue, delegates it, and sends the task message, so that a worker of the role
// takes the task again.
type Kick struct {
	Issue string `json:"issue"`
	Role  string `json:"role"`
	// Reopen is true when the issue is in progress: the kick reopens it, which drops its owner.
	Reopen bool `json:"reopen"`
	// RemoveNeedsHuman is true when the issue has the tag needs-human: the kick removes it.
	RemoveNeedsHuman bool `json:"remove-needs-human"`
	// Send is true when no task message waits: the kick sends one, with the id MessageID. When one
	// waits already, the kick sends nothing: the reopened issue lets a worker take that message.
	Send      bool   `json:"send"`
	MessageID string `json:"message-id,omitempty"`
	// Bookmark goes into the data of the message, so that the task continues on its bookmark.
	Bookmark string `json:"bookmark,omitempty"`
	// Warnings are what the kick cannot fix, for example that no live worker of the role runs.
	Warnings []string `json:"warnings,omitempty"`
}

// PlanKick plans the kick of the task of the issue. role overrides the role of the task. Without
// force, it refuses a task that a live worker holds: that task is slow, not stuck.
func PlanKick(in Input, role string, force bool) (Kick, error) {
	k := Kick{Issue: in.Issue.Name, Role: role, Bookmark: in.Issue.Bookmark, Warnings: []string{}}
	if in.Issue.State == "closed" {
		return k, fmt.Errorf("The issue %s is closed: there is no task to kick.", k.Issue)
	}
	if k.Role == "" {
		k.Role = Role(in)
	}
	if k.Role == "" {
		return k, fmt.Errorf("No role owns the task of %s: the issue has no delegate metadata, and no task message names one. Give the role with --role, or delegate it: hephaestus delegate %s <role>", k.Issue, k.Issue)
	}
	if !force {
		if w := workerOf(in, in.Issue.Owner); in.Issue.State == "in-progress" && alive(in, w) && holdsTask(in, w) {
			return k, fmt.Errorf("The worker %s is live (last seen %s) and holds the task of %s: the task runs, slow or not. Kick it anyway with --force.", in.Issue.Owner, ago(in.Now, w.LastSeen), k.Issue)
		}
		for _, p := range pendingTasks(in) {
			if holder := workerOf(in, p.LeasedBy); p.LeasedBy != "" && alive(in, holder) {
				return k, fmt.Errorf("The live worker %s holds the task message %s of %s: the task runs, slow or not. Kick it anyway with --force.", p.LeasedBy, p.ID, k.Issue)
			}
		}
	}
	k.Reopen = in.Issue.State == "in-progress"
	k.RemoveNeedsHuman = slices.Contains(in.Issue.Tags, needsHuman)
	k.Send = len(pendingTasks(in)) == 0
	if k.Send {
		k.MessageID = fmt.Sprintf("kick-%s-%d", path.Base(k.Issue), in.Now.UnixNano())
	} else {
		p := pendingTasks(in)[0]
		k.Warnings = append(k.Warnings, fmt.Sprintf("The task message %s already waits for %s: no second message is sent.", p.ID, p.For))
	}
	if !liveRole(in, k.Role) {
		k.Warnings = append(k.Warnings, fmt.Sprintf("No live worker of %s runs: the task waits until one starts. Start one: hephaestus agent run %s", k.Role, k.Role))
	}
	return k, nil
}
