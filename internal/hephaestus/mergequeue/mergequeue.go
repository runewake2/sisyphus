// Package mergequeue makes one linear history from a set of tasks ([[agent-delegation#Merge queue]]).
// It merges the bookmark of each task, one at a time, in the order of package mergeorder, onto a new
// line based on the base, gives the unreleased changelog entries of each bookmark their versions,
// renumbers decision records that collide, and runs the checks of the repo after each bookmark, or
// once for each batch of bookmarks. A
// bookmark that conflicts or fails the checks is left out, and does not change. The result is a new
// bookmark, and each merged bookmark moves onto it, so that it is in the history of the base when a
// human moves the base to the line. It never moves the base and never pushes.
//
// The merge happens in a scratch working copy of the repo: a jj workspace in a jj repo, where jj
// duplicate tries the commits on a line of copies and jj rebase then moves the commits themselves, or
// a git worktree of the git directory in a git repo, where git cherry-pick does the rebase.
package mergequeue

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/runewake2/sisyphus/internal/hephaestus/decisions"
	"github.com/runewake2/sisyphus/internal/hephaestus/mergeorder"
	"github.com/runewake2/sisyphus/internal/hephaestus/release"
	"github.com/runewake2/sisyphus/internal/hephaestus/repo"
	"github.com/runewake2/sisyphus/internal/hephaestus/shell"
)

// Options are the input of a run.
type Options struct {
	// Repo is the repo of the user, whose jj workspace or git worktree holds the commits.
	Repo repo.Repo
	// Tasks are the tasks to merge, with their bookmarks and dependencies.
	Tasks []mergeorder.Task
	// Base is the base of the new line, normally main.
	Base string
	// Name is the bookmark of the result, for example merge-queue/2026-10-12.
	Name string
	// Checks is a command line that must succeed after each bookmark, in the scratch working copy.
	// Empty: no checks.
	Checks string
	// Batch is how many bookmarks go on the line before one run of the checks. A batch that fails
	// splits, until the checks find the bookmark that fails. Less than 2: one at a time.
	Batch int
	// Out gets the output of the commands, for a person to follow the run.
	Out io.Writer
	// Now is the clock of the versions.
	Now func() time.Time
}

// Version is one unreleased entry that got a version.
type Version struct {
	Entry   string `json:"entry"`
	Version string `json:"version"`
}

// Merged is one bookmark that the run merged.
type Merged struct {
	Place      int                     `json:"place"`
	Bookmark   string                  `json:"bookmark"`
	Tasks      []string                `json:"tasks"`
	Commits    int                     `json:"commits"`
	Versions   []Version               `json:"versions"`
	Renumbered []decisions.Renumbering `json:"renumbered"`
	After      []mergeorder.Reason     `json:"after"`
}

// LeftOut is one bookmark that the run did not merge.
type LeftOut struct {
	Bookmark string   `json:"bookmark"`
	Tasks    []string `json:"tasks"`
	// Reason is conflict, checks, dependency, merges, immutable, versions, or nothing.
	Reason string `json:"reason"`
	Detail string `json:"detail"`
	// Files are the files of a conflict.
	Files []string `json:"files,omitempty"`
}

// Report is the result of a run.
type Report struct {
	Name     string    `json:"name"`
	Base     string    `json:"base"`
	BaseID   string    `json:"base-id"`
	Head     string    `json:"head"`
	Created  bool      `json:"created"`
	Merged   []Merged  `json:"merged"`
	LeftOut  []LeftOut `json:"left-out"`
	Warnings []string  `json:"warnings"`
	// Undo is the command that puts the merged bookmarks back. Empty when no bookmark moved.
	Undo string `json:"undo,omitempty"`
	// CheckRuns is how many times the checks ran.
	CheckRuns int `json:"check-runs"`
}

// A move is one merged bookmark, for the rebase at the end of a run.
type move struct {
	Bookmark string
	// From is the commit of the bookmark when the run tried it.
	From string
	// To is the copy of From on the line.
	To string
	// Version is the version commit of the bookmark on the line. Empty: it has none.
	Version string
}

// landed is the result of the rebase at the end of a run.
type landed struct {
	Head     string
	Undo     string
	Warnings []string
}

// Run merges the tasks. A task that cannot merge is left out, with the reason in the report; an
// error stops the run, and changes nothing outside the scratch working copy.
func Run(ctx context.Context, o Options) (Report, error) {
	report := Report{Name: o.Name, Base: o.Base, Merged: []Merged{}, LeftOut: []LeftOut{}, Warnings: []string{}}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Out == nil {
		o.Out = io.Discard
	}
	if o.Repo.HasBookmark(o.Name) {
		return report, fmt.Errorf("The bookmark %s already exists. Give another --name, or remove it.", o.Name)
	}
	baseID, err := o.Repo.CommitOf(o.Base)
	if err != nil || baseID == "" {
		return report, fmt.Errorf("The base %q does not exist.", o.Base)
	}
	report.BaseID, report.Head = baseID, baseID

	l, err := newLine(o.Repo, baseID)
	if err != nil {
		return report, err
	}
	defer l.Close()
	return run(ctx, o, l, report)
}

// run merges the tasks onto the line l, which starts at the base of the report.
func run(ctx context.Context, o Options, l line, report Report) (Report, error) {
	baseID := report.BaseID
	q := &queue{ctx: ctx, o: o, l: l, report: &report, head: baseID, tip: baseID, copies: map[string]string{}, leftOut: map[string]bool{}}
	done := false
	// An error must stop the run and change nothing outside the scratch working copy: without the
	// rollback, the commits picked since the last good head would stay, in jj as visible heads.
	defer func() {
		if !done {
			l.Rollback(q.head)
		}
	}()

	plan, err := mergeorder.Compute(l.Repo(), o.Tasks, baseID)
	if err != nil {
		return report, err
	}
	report.Warnings = append(report.Warnings, plan.Warnings...)
	for _, skipped := range plan.Skipped {
		report.LeftOut = append(report.LeftOut, LeftOut{Bookmark: skipped.Bookmark, Tasks: skipped.Tasks, Reason: "nothing", Detail: skipped.Why})
	}

	size := max(o.Batch, 1)
	for start := 0; start < len(plan.Order); start += size {
		if err := q.merge(plan.Order[start:min(start+size, len(plan.Order))], nil); err != nil {
			return report, err
		}
	}
	report.Head = q.head
	if q.head != baseID {
		result, err := l.Land(q.moves, q.head)
		report.Undo = result.Undo
		report.Warnings = append(report.Warnings, result.Warnings...)
		if err != nil {
			q.head = baseID // drop every copy
			return report, err
		}
		q.head, report.Head = result.Head, result.Head
		if err := l.Bookmark(o.Name, q.head); err != nil {
			return report, err
		}
		report.Created = true
	}
	done = true
	return report, nil
}

// A queue is the state of a run: the line, and what is on it.
type queue struct {
	ctx    context.Context
	o      Options
	l      line
	report *Report
	// head is the tip of the line that passed the checks; tip is the tip with the current batch.
	head, tip string
	merged    []string
	moves     []move
	copies    map[string]string
	leftOut   map[string]bool
}

// merge puts places on the line and runs the checks once, on the tip. A batch that fails goes back
// to the head, and each half tries again: without the failed work, a conflict in the batch can go
// away. failed is the output of checks that failed on the tree that places make, or nil.
func (q *queue) merge(places []mergeorder.Place, failed *string) error {
	merged, moves, picked, left := len(q.merged), len(q.moves), len(q.report.Merged), len(q.report.LeftOut)
	staged := false
	for _, place := range places {
		if q.ctx.Err() != nil {
			return q.ctx.Err()
		}
		ok, err := q.stage(place)
		if err != nil {
			return err
		}
		staged = staged || ok
	}
	if !staged || q.o.Checks == "" {
		q.head = q.tip
		return nil
	}
	var names []string
	for _, place := range places {
		names = append(names, place.Bookmark)
	}
	var output string
	if failed != nil {
		output = *failed
	} else {
		fmt.Fprintf(q.o.Out, "merge-queue: checks of %s: %s\n", strings.Join(names, ", "), q.o.Checks)
		q.report.CheckRuns++
		var ok bool
		if output, ok = runChecks(q.ctx, q.l.Dir(), q.o.Checks, q.o.Out); ok {
			q.head = q.tip
			return nil
		}
	}

	// Go back to the state before the batch.
	for _, out := range q.report.LeftOut[left:] {
		delete(q.leftOut, out.Bookmark)
	}
	q.merged, q.moves = q.merged[:merged], q.moves[:moves]
	q.report.Merged, q.report.LeftOut = q.report.Merged[:picked], q.report.LeftOut[:left]
	q.tip = q.head
	if err := q.l.Rollback(q.head); err != nil {
		return err
	}
	if len(places) == 1 {
		return q.leave(places[0], "checks", "the checks failed after it merged:\n"+tail(output, 30), nil)
	}
	fmt.Fprintf(q.o.Out, "merge-queue: the checks of %s failed; the batch splits\n", strings.Join(names, ", "))
	half := len(places) / 2
	if err := q.merge(places[:half], nil); err != nil {
		return err
	}
	// With all of the first half on the line, the second half makes the tree that failed: the run
	// does not spend the time of the checks to see that again.
	var known *string
	if len(q.report.LeftOut) == left {
		known = &output
	}
	return q.merge(places[half:], known)
}

// leave leaves a bookmark out, and puts the line back at its tip.
func (q *queue) leave(place mergeorder.Place, reason, detail string, files []string) error {
	q.leftOut[place.Bookmark] = true
	q.report.LeftOut = append(q.report.LeftOut, LeftOut{Bookmark: place.Bookmark, Tasks: place.Tasks, Reason: reason, Detail: detail, Files: files})
	fmt.Fprintf(q.o.Out, "merge-queue: left out %s: %s\n", place.Bookmark, detail)
	return q.l.Rollback(q.tip)
}

// stage puts one bookmark on the line, with its versions, and without the checks. ok is false when
// it is left out.
func (q *queue) stage(place mergeorder.Place) (ok bool, err error) {
	l, o := q.l, q.o
	fmt.Fprintf(o.Out, "merge-queue: %d. %s (%s)\n", place.Place, place.Bookmark, strings.Join(place.Tasks, ", "))
	var missing []string
	for _, before := range place.After {
		if q.leftOut[before.Bookmark] {
			missing = append(missing, before.Bookmark)
		}
	}
	if len(missing) > 0 {
		return false, q.leave(place, "dependency", fmt.Sprintf("it must come after %s, which is left out", strings.Join(missing, ", ")), nil)
	}
	// The commits of the bookmark that the line does not have yet: not in the base, and not in a
	// bookmark that is merged already (a stack).
	commits, err := l.Commits(place.Bookmark, q.merged)
	if err != nil {
		return false, err
	}
	if merges, err := l.HasMerges(place.Bookmark, q.merged); err != nil {
		return false, err
	} else if merges {
		return false, q.leave(place, "merges", "its history has merge commits; the merge queue makes a linear history", nil)
	}
	if immutable, err := l.Immutable(place.Bookmark, q.merged); err != nil {
		return false, err
	} else if immutable {
		return false, q.leave(place, "immutable", "its history has immutable commits, which the merge queue cannot rebase", nil)
	}
	from, err := l.Repo().CommitOf(place.Bookmark)
	if err != nil {
		return false, err
	}
	for _, commit := range commits {
		files, ok, err := l.Pick(commit)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, q.leave(place, "conflict", fmt.Sprintf("its change %.12s conflicts with the line", commit), files)
		}
		if q.copies[commit], err = l.Head(); err != nil {
			return false, err
		}
	}
	merged := Merged{Place: len(q.report.Merged) + 1, Bookmark: place.Bookmark, Tasks: place.Tasks, Commits: len(commits), Versions: []Version{}, Renumbered: []decisions.Renumbering{}, After: place.After}

	// Decision records that collide with the line, then the versions of the new entries.
	renumber, err := decisions.PlanRenumber(l.Repo(), q.tip)
	if err != nil {
		return false, err
	}
	if err := decisions.Apply(l.Repo(), renumber); err != nil {
		return false, err
	}
	merged.Renumbered = renumber.Renumberings
	added, err := l.AddedSince(q.tip, release.UnreleasedDir)
	if err != nil {
		return false, err
	}
	var entries []string
	for _, file := range added {
		name := strings.TrimSuffix(path.Base(file), ".md")
		if _, err := os.Stat(l.Dir() + "/" + file); err == nil && !slices.Contains(entries, name) {
			entries = append(entries, name)
		}
	}
	if len(entries) > 0 {
		assign, err := release.PlanAssign(l.Repo(), entries, o.Now())
		if err != nil {
			return false, q.leave(place, "versions", "its unreleased entries cannot get versions: "+err.Error(), nil)
		}
		if err := release.Apply(l.Repo(), assign); err != nil {
			return false, err
		}
		for _, a := range assign.Assignments {
			merged.Versions = append(merged.Versions, Version{Entry: a.Entry.Name, Version: a.Version})
		}
	}
	versioned := len(entries) > 0 || len(renumber.Renumberings) > 0
	if versioned {
		var body []string
		for _, v := range merged.Versions {
			body = append(body, fmt.Sprintf("- %s is %s", v.Entry, v.Version))
		}
		for _, rn := range merged.Renumbered {
			body = append(body, fmt.Sprintf("- %s is now %s", rn.From, rn.To))
		}
		message := fmt.Sprintf("merge-queue: give versions to %s\n\n%s\n", place.Bookmark, strings.Join(body, "\n"))
		if err := l.CommitFiles(message); err != nil {
			return false, err
		}
	}
	if q.tip, err = l.Head(); err != nil {
		return false, err
	}
	q.merged = append(q.merged, place.Bookmark)
	q.report.Merged = append(q.report.Merged, merged)
	if to := q.copies[from]; to != "" {
		m := move{Bookmark: place.Bookmark, From: from, To: to}
		if versioned {
			m.Version = q.tip
		}
		q.moves = append(q.moves, m)
	}
	return true, nil
}

// runChecks runs the checks in dir, and returns their output and whether they passed. A stopped run
// stops the checks with every program that they started.
func runChecks(ctx context.Context, dir, checks string, out io.Writer) (string, bool) {
	cmd := shell.Command(ctx, checks)
	cmd.Dir = dir
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(&output, out)
	cmd.Stderr = io.MultiWriter(&output, out)
	err := cmd.Run()
	return output.String(), err == nil
}

func tail(text string, n int) string {
	all := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(all) > n {
		all = all[len(all)-n:]
	}
	return strings.Join(all, "\n")
}
