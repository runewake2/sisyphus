// Package mergequeue makes one linear history from a set of tasks ([[agent-delegation#Merge queue]]).
// It merges the bookmark of each task, one at a time, in the order of package mergeorder, onto a new
// line based on the base, gives the unreleased changelog entries of each bookmark their versions,
// renumbers decision records that collide, and runs the checks of the repo after each bookmark. A
// bookmark that conflicts or fails the checks is left out. The result is a new bookmark; the base,
// the bookmarks of the tasks, and the working copies of the users do not change. It never moves the
// base and never pushes.
//
// The merge happens in a scratch working copy of the repo: a jj workspace in a jj repo, where jj
// duplicate puts the commits onto the line, or a git worktree of the git directory in a git repo,
// where git cherry-pick does.
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
	// Reason is conflict, checks, dependency, merges, or nothing.
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
}

func lines(text string) []string {
	var result []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}
	return result
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
	head := baseID
	done := false
	// An error must stop the run and change nothing outside the scratch working copy: without the
	// rollback, the commits picked since the last good head would stay, in jj as visible heads.
	defer func() {
		if !done {
			l.Rollback(head)
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

	var mergedBookmarks []string
	leftOut := map[string]bool{}
	for _, place := range plan.Order {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		fmt.Fprintf(o.Out, "merge-queue: %d. %s (%s)\n", place.Place, place.Bookmark, strings.Join(place.Tasks, ", "))
		leave := func(reason, detail string, files []string) error {
			leftOut[place.Bookmark] = true
			report.LeftOut = append(report.LeftOut, LeftOut{Bookmark: place.Bookmark, Tasks: place.Tasks, Reason: reason, Detail: detail, Files: files})
			fmt.Fprintf(o.Out, "merge-queue: left out %s: %s\n", place.Bookmark, detail)
			return l.Rollback(head)
		}
		var missing []string
		for _, before := range place.After {
			if leftOut[before.Bookmark] {
				missing = append(missing, before.Bookmark)
			}
		}
		if len(missing) > 0 {
			if err := leave("dependency", fmt.Sprintf("it must come after %s, which is left out", strings.Join(missing, ", ")), nil); err != nil {
				return report, err
			}
			continue
		}
		// The commits of the bookmark that the line does not have yet: not in the base, and not in a
		// bookmark that is merged already (a stack).
		commits, err := l.Commits(place.Bookmark, mergedBookmarks)
		if err != nil {
			return report, err
		}
		if merges, err := l.HasMerges(place.Bookmark, mergedBookmarks); err != nil {
			return report, err
		} else if merges {
			if err := leave("merges", "its history has merge commits; the merge queue makes a linear history", nil); err != nil {
				return report, err
			}
			continue
		}
		conflict := false
		for _, commit := range commits {
			files, ok, err := l.Pick(commit)
			if err != nil {
				return report, err
			}
			if !ok {
				if err := leave("conflict", fmt.Sprintf("its change %.12s conflicts with the line", commit), files); err != nil {
					return report, err
				}
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}
		merged := Merged{Place: len(report.Merged) + 1, Bookmark: place.Bookmark, Tasks: place.Tasks, Commits: len(commits), Versions: []Version{}, Renumbered: []decisions.Renumbering{}, After: place.After}

		// Decision records that collide with the line, then the versions of the new entries.
		renumber, err := decisions.PlanRenumber(l.Repo(), head)
		if err != nil {
			return report, err
		}
		if err := decisions.Apply(l.Repo(), renumber); err != nil {
			return report, err
		}
		merged.Renumbered = renumber.Renumberings
		added, err := l.AddedSince(head, release.UnreleasedDir)
		if err != nil {
			return report, err
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
				if err := leave("versions", "its unreleased entries cannot get versions: "+err.Error(), nil); err != nil {
					return report, err
				}
				continue
			}
			if err := release.Apply(l.Repo(), assign); err != nil {
				return report, err
			}
			for _, a := range assign.Assignments {
				merged.Versions = append(merged.Versions, Version{Entry: a.Entry.Name, Version: a.Version})
			}
		}
		if len(entries) > 0 || len(renumber.Renumberings) > 0 {
			var body []string
			for _, v := range merged.Versions {
				body = append(body, fmt.Sprintf("- %s is %s", v.Entry, v.Version))
			}
			for _, rn := range merged.Renumbered {
				body = append(body, fmt.Sprintf("- %s is now %s", rn.From, rn.To))
			}
			message := fmt.Sprintf("merge-queue: give versions to %s\n\n%s\n", place.Bookmark, strings.Join(body, "\n"))
			if err := l.CommitFiles(message); err != nil {
				return report, err
			}
		}
		if o.Checks != "" {
			fmt.Fprintf(o.Out, "merge-queue: checks of %s: %s\n", place.Bookmark, o.Checks)
			if output, ok := runChecks(ctx, l.Dir(), o.Checks, o.Out); !ok {
				if err := leave("checks", "the checks failed after it merged:\n"+tail(output, 30), nil); err != nil {
					return report, err
				}
				continue
			}
		}
		newHead, err := l.Head()
		if err != nil {
			return report, err
		}
		head = newHead
		mergedBookmarks = append(mergedBookmarks, place.Bookmark)
		report.Merged = append(report.Merged, merged)
	}
	report.Head = head
	if head != baseID {
		if err := l.Bookmark(o.Name, head); err != nil {
			return report, err
		}
		report.Created = true
	}
	done = true
	return report, nil
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
