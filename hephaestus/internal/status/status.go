// Package status gathers what needs a human in the government of hephaestus: the issues that wait
// for a human, the referendums that run or failed, the workers that stopped, the messages that
// wait, and the last line of the merge queue with the recommendation of the release manager.
package status

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
)

// Kinds of the messages that tell about a line of the merge queue.
const (
	LineKind           = "merge-queue-done"
	RecommendationKind = "release-recommendation"
)

// Input is what the report reads. The caller gathers it, so that a test can give it.
type Input struct {
	Now time.Time
	// Stale is the time after which a worker that has not spoken to the queue counts as stopped.
	Stale time.Duration
	// NeedsHuman are the open and in-progress issues tagged needs-human.
	NeedsHuman []sisyphus.Issue
	// WithReferendum are the issues that have the metadata key referendum, in any state.
	WithReferendum []sisyphus.Issue
	// Queue is the address of the message queue, or empty when it does not run.
	Queue    string
	QueueErr string
	Workers  []mq.WorkerStatus
	Pending  []mq.Pending
	// Log is the message log, oldest first: the last line and its recommendation come from it.
	Log []mq.Message
	// Merged tells whether the base already has a commit, for the last line.
	Merged func(commit, base string) bool
}

// Report is what needs a human.
type Report struct {
	Generated   time.Time    `json:"generated"`
	NeedsHuman  []Issue      `json:"needs-human"`
	Referendums []Referendum `json:"referendums"`
	Queue       Queue        `json:"queue"`
	Line        *Line        `json:"line,omitempty"`
	Attention   int          `json:"attention"`
	Notes       []string     `json:"notes,omitempty"`
}

// Issue is an issue that waits for a human.
type Issue struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	State    string `json:"state"`
	Owner    string `json:"owner,omitempty"`
	Priority string `json:"priority"`
}

// Referendum is a referendum that runs, failed, or has no result.
type Referendum struct {
	Issue  string `json:"issue"`
	State  string `json:"state"` // running, overdue, failed, or no-result
	Value  string `json:"value"` // the metadata value, for example round-2
	Record string `json:"record,omitempty"`
	// Deadline is when a running referendum times out.
	Deadline time.Time `json:"deadline,omitzero"`
	Action   string    `json:"action"`
}

// Queue is the state of the message queue: its workers and the deliveries that wait.
type Queue struct {
	Address string    `json:"address,omitempty"`
	Error   string    `json:"error,omitempty"`
	Live    []Worker  `json:"live"`
	Stopped []Worker  `json:"stopped"`
	Waiting []Waiting `json:"waiting"`
}

// Worker is a worker of the queue.
type Worker struct {
	Name     string    `json:"name"`
	Role     string    `json:"role,omitempty"`
	LastSeen time.Time `json:"last-seen,omitzero"`
	Holds    int       `json:"holds"`
}

// Waiting is the deliveries that wait for one receiver.
type Waiting struct {
	For        string    `json:"for"`
	Count      int       `json:"count"`
	Kinds      []string  `json:"kinds"`
	Oldest     time.Time `json:"oldest"`
	Leased     int       `json:"leased"`
	LiveWorker bool      `json:"live-worker"`
}

// Line is the last line of the merge queue.
type Line struct {
	Bookmark string    `json:"bookmark"`
	Base     string    `json:"base"`
	Head     string    `json:"head"`
	Merged   []string  `json:"merged"`
	LeftOut  []string  `json:"left-out"`
	Made     time.Time `json:"made"`
	// InBase is true when the base already has the head: a human moved it.
	InBase bool `json:"in-base"`
	// Recommendation is the last recommendation of the release manager for the line, if any.
	Recommendation *Recommendation `json:"recommendation,omitempty"`
}

// Recommendation is what the release manager recommends for a line.
type Recommendation struct {
	From      string    `json:"from"`
	Recommend string    `json:"recommend"` // move or hold
	Summary   string    `json:"summary"`
	Time      time.Time `json:"time"`
}

// Build makes the report.
func Build(in Input) Report {
	r := Report{Generated: in.Now, NeedsHuman: []Issue{}, Referendums: []Referendum{}}
	for _, issue := range in.NeedsHuman {
		r.NeedsHuman = append(r.NeedsHuman, Issue{Name: issue.Name, Title: issue.Title, State: issue.State, Owner: issue.Owner, Priority: issue.Priority})
	}
	r.Attention += len(r.NeedsHuman)

	for _, issue := range in.WithReferendum {
		value := issue.Metadata["referendum"]
		ref := Referendum{Issue: issue.Name, Value: value, Record: issue.Metadata["referendum-record"]}
		switch {
		case strings.HasPrefix(value, "round-"):
			ref.State, ref.Action = "running", "Wait for its result."
			if deadline, err := time.Parse(time.RFC3339, issue.Metadata["referendum-deadline"]); err == nil {
				ref.Deadline = deadline
				if in.Now.After(deadline) {
					ref.State, ref.Action = "overdue", "It stopped with no record. The next call records it as failed, or decide it: hephaestus referendum decide "+issue.Name+" passed|failed --reason <why>."
				}
			}
		case value == "no-result":
			ref.State, ref.Action = "no-result", "A human decides it: hephaestus referendum decide "+issue.Name+" passed|failed --reason <why>."
		case value == "failed" && issue.State != "closed":
			ref.State, ref.Action = "failed", "The owner of the issue corrects the work, and calls a new referendum, or a human overrules it."
		default:
			continue
		}
		r.Referendums = append(r.Referendums, ref)
		if ref.State != "running" {
			r.Attention++
		}
	}
	sort.Slice(r.Referendums, func(i, j int) bool { return r.Referendums[i].Issue < r.Referendums[j].Issue })

	r.Queue = Queue{Address: in.Queue, Error: in.QueueErr, Live: []Worker{}, Stopped: []Worker{}, Waiting: []Waiting{}}
	live := map[string]bool{} // worker names and roles with a live worker
	for _, w := range in.Workers {
		worker := Worker{Name: w.Worker, Role: w.Role, LastSeen: w.LastSeen, Holds: len(w.Holds)}
		if !w.LastSeen.IsZero() && in.Now.Sub(w.LastSeen) <= in.Stale {
			r.Queue.Live = append(r.Queue.Live, worker)
			live["worker:"+w.Worker] = true
			if w.Role != "" {
				live["role:"+w.Role] = true
			}
		} else {
			r.Queue.Stopped = append(r.Queue.Stopped, worker)
		}
	}
	groups := map[string]*Waiting{}
	var order []string
	for _, p := range in.Pending {
		g, ok := groups[p.For]
		if !ok {
			g = &Waiting{For: p.For, Oldest: p.Sent, Kinds: []string{}, LiveWorker: live[p.For]}
			groups[p.For] = g
			order = append(order, p.For)
		}
		g.Count++
		if p.LeasedBy != "" {
			g.Leased++
		}
		if !slices.Contains(g.Kinds, p.Kind) {
			g.Kinds = append(g.Kinds, p.Kind)
		}
		if p.Sent.Before(g.Oldest) {
			g.Oldest = p.Sent
		}
	}
	for _, key := range order {
		g := groups[key]
		r.Queue.Waiting = append(r.Queue.Waiting, *g)
		if !g.LiveWorker {
			r.Attention++
		}
	}
	if in.QueueErr != "" {
		r.Notes = append(r.Notes, "The message queue does not answer: "+in.QueueErr)
	}

	r.Line = lastLine(in)
	if r.Line != nil && !r.Line.InBase {
		r.Attention++
	}
	return r
}

// lastLine finds the last merge-queue-done in the log, and the last recommendation for its line.
func lastLine(in Input) *Line {
	var line *Line
	for _, m := range in.Log {
		if m.Kind != LineKind {
			continue
		}
		var data struct {
			Bookmark string   `json:"bookmark"`
			Base     string   `json:"base"`
			Head     string   `json:"head"`
			Merged   []string `json:"merged"`
			LeftOut  []string `json:"left-out"`
		}
		if json.Unmarshal(m.Data, &data) != nil || data.Bookmark == "" {
			continue
		}
		line = &Line{Bookmark: data.Bookmark, Base: data.Base, Head: data.Head, Merged: data.Merged, LeftOut: data.LeftOut, Made: m.Time}
	}
	if line == nil {
		return nil
	}
	if in.Merged != nil && line.Head != "" && line.Base != "" {
		line.InBase = in.Merged(line.Head, line.Base)
	}
	for _, m := range in.Log {
		if m.Kind != RecommendationKind || m.Time.Before(line.Made) {
			continue
		}
		var data struct {
			Bookmark  string `json:"bookmark"`
			Recommend string `json:"recommend"`
			Summary   string `json:"summary"`
		}
		if json.Unmarshal(m.Data, &data) != nil || data.Bookmark != line.Bookmark {
			continue
		}
		line.Recommendation = &Recommendation{From: m.From, Recommend: data.Recommend, Summary: data.Summary, Time: m.Time}
	}
	return line
}

// Text writes the report for a person.
func (r Report) Text(now time.Time) string {
	var b strings.Builder
	ago := func(t time.Time) string {
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

	fmt.Fprintf(&b, "Issues that wait for a human (%d)\n", len(r.NeedsHuman))
	for _, i := range r.NeedsHuman {
		owner := ""
		if i.Owner != "" {
			owner = ", owner " + i.Owner
		}
		fmt.Fprintf(&b, "  %s  %s, %s%s: %s\n", i.Name, i.State, i.Priority, owner, i.Title)
	}

	fmt.Fprintf(&b, "\nReferendums that run or need a human (%d)\n", len(r.Referendums))
	for _, ref := range r.Referendums {
		detail := ref.State
		if ref.State == "running" || ref.State == "overdue" {
			detail = fmt.Sprintf("%s (%s, deadline %s)", ref.State, ref.Value, ref.Deadline.Local().Format("2006-01-02 15:04"))
		}
		fmt.Fprintf(&b, "  %s  %s. %s\n", ref.Issue, detail, ref.Action)
	}

	q := r.Queue
	if q.Address == "" {
		fmt.Fprintf(&b, "\nMessage queue: not running. Start it with: hephaestus messages serve\n")
	} else {
		fmt.Fprintf(&b, "\nWorkers at %s (%d live, %d stopped)\n", q.Address, len(q.Live), len(q.Stopped))
		for _, w := range q.Live {
			fmt.Fprintf(&b, "  live     %s%s, last seen %s, holds %d\n", w.Name, roleSuffix(w.Role), ago(w.LastSeen), w.Holds)
		}
		for _, w := range q.Stopped {
			fmt.Fprintf(&b, "  stopped  %s%s, last seen %s\n", w.Name, roleSuffix(w.Role), ago(w.LastSeen))
		}
		total := 0
		for _, g := range q.Waiting {
			total += g.Count
		}
		fmt.Fprintf(&b, "\nMessages that wait (%d)\n", total)
		for _, g := range q.Waiting {
			state := "a live worker has them"
			if !g.LiveWorker {
				state = "no live worker: start one"
			}
			fmt.Fprintf(&b, "  %s  %d (%s), oldest %s, %d leased: %s\n", g.For, g.Count, strings.Join(g.Kinds, ", "), ago(g.Oldest), g.Leased, state)
		}
	}
	for _, note := range r.Notes {
		fmt.Fprintf(&b, "  %s\n", note)
	}

	fmt.Fprintf(&b, "\nLast line of the merge queue\n")
	if r.Line == nil {
		fmt.Fprintf(&b, "  none in the message log\n")
	} else {
		l := r.Line
		moved := fmt.Sprintf("%s does not have it yet: a human checks it, and moves %s", l.Base, l.Base)
		if l.InBase {
			moved = fmt.Sprintf("%s has it already", l.Base)
		}
		fmt.Fprintf(&b, "  %s on %s, made %s: %d merged, %d left out. %s.\n", l.Bookmark, l.Base, ago(l.Made), len(l.Merged), len(l.LeftOut), moved)
		if len(l.LeftOut) > 0 {
			fmt.Fprintf(&b, "  left out: %s\n", strings.Join(l.LeftOut, ", "))
		}
		if rec := l.Recommendation; rec != nil {
			fmt.Fprintf(&b, "  release manager (%s, %s): %s. %s\n", rec.From, ago(rec.Time), rec.Recommend, rec.Summary)
		} else if !l.InBase {
			fmt.Fprintf(&b, "  release manager: no recommendation yet\n")
		}
	}

	if r.Attention == 0 {
		fmt.Fprintf(&b, "\nNothing needs a human now.\n")
	} else {
		fmt.Fprintf(&b, "\n%d %s need a human.\n", r.Attention, map[bool]string{true: "item", false: "items"}[r.Attention == 1])
	}
	return b.String()
}

func roleSuffix(role string) string {
	if role == "" {
		return ""
	}
	return " (" + role + ")"
}
