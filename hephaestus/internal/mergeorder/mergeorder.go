// Package mergeorder computes the order in which the merge queue merges the work of a set of tasks.
// The work of each task is a bookmark (or branch); tasks with the same bookmark merge together. The
// rules, in order of importance ([[agent-delegation#Merge queue]]):
//
//  1. Dependencies: when a task depends on a task of another bookmark, that bookmark goes first.
//  2. Stacks: a bookmark based on another bookmark of the set goes after it.
//  3. Overlap: then the bookmark whose changed files the fewest other bookmarks also change.
//  4. Age: then the bookmark whose oldest change is the oldest.
//  5. Then the name of the bookmark.
//
// Rules 1 and 2 are constraints; rules 3-5 only order the bookmarks that the constraints allow.
package mergeorder

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

// Task is one sisyphus issue of the set.
type Task struct {
	Name      string
	State     string
	Bookmark  string
	DependsOn []Dependency
}

// Dependency is an issue that a task depends on, with its state.
type Dependency struct {
	Name  string
	State string
}

// VCS is what the planner needs from the repo.
type VCS interface {
	Exists(rev string) bool
	IsAncestor(a, b string) (bool, error)
	ChangedFiles(base, rev string) ([]string, error)
	FirstChangeTime(base, rev string) (time.Time, error)
}

// Reason is why one bookmark comes after another.
type Reason struct {
	Bookmark string `json:"bookmark"`
	Why      string `json:"why"`
}

// Overlap is a bookmark that changes some of the same files.
type Overlap struct {
	Bookmark string   `json:"bookmark"`
	Files    []string `json:"files"`
}

// Place is one bookmark in the order.
type Place struct {
	Place       int       `json:"place"`
	Bookmark    string    `json:"bookmark"`
	Tasks       []string  `json:"tasks"`
	After       []Reason  `json:"after"`
	Overlaps    []Overlap `json:"overlaps"`
	FirstChange time.Time `json:"first-change"`
}

// Skipped is a bookmark that the order leaves out.
type Skipped struct {
	Bookmark string   `json:"bookmark"`
	Tasks    []string `json:"tasks"`
	Why      string   `json:"why"`
}

// Plan is the merge order of a set of tasks.
type Plan struct {
	Onto     string    `json:"onto"`
	Order    []Place   `json:"order"`
	Skipped  []Skipped `json:"skipped"`
	Warnings []string  `json:"warnings"`
}

type unit struct {
	bookmark string
	tasks    []string
	files    []string
	first    time.Time
	after    map[string]string // bookmark -> why
}

// Compute returns the merge order of tasks onto base. It fails for a task with no bookmark, a
// bookmark that does not exist, and a cycle of constraints.
func Compute(vcs VCS, tasks []Task, base string) (Plan, error) {
	plan := Plan{Onto: base, Order: []Place{}, Skipped: []Skipped{}, Warnings: []string{}}
	if !vcs.Exists(base) {
		return plan, fmt.Errorf("The base %q does not exist.", base)
	}
	units := map[string]*unit{}
	unitOf := map[string]string{} // task -> bookmark
	var problems []string
	for _, task := range tasks {
		switch {
		case task.Bookmark == "":
			problems = append(problems, fmt.Sprintf("task %s has no bookmark, so its work cannot be found. Set it with: sisyphus update %s --bookmark <bookmark>", task.Name, task.Name))
			continue
		case !vcs.Exists(task.Bookmark):
			problems = append(problems, fmt.Sprintf("the bookmark %s of task %s does not exist", task.Bookmark, task.Name))
			continue
		}
		u, ok := units[task.Bookmark]
		if !ok {
			u = &unit{bookmark: task.Bookmark, after: map[string]string{}}
			units[task.Bookmark] = u
		}
		u.tasks = append(u.tasks, task.Name)
		unitOf[task.Name] = task.Bookmark
	}
	if len(problems) > 0 {
		return plan, fmt.Errorf("%s.", strings.Join(problems, ".\n"))
	}

	// A bookmark with no change from the base has nothing to merge.
	for _, name := range slices.Sorted(maps.Keys(units)) {
		u := units[name]
		files, err := vcs.ChangedFiles(base, u.bookmark)
		if err != nil {
			return plan, err
		}
		if len(files) == 0 {
			plan.Skipped = append(plan.Skipped, Skipped{Bookmark: u.bookmark, Tasks: u.tasks, Why: fmt.Sprintf("it changes nothing from %s: its work is already there", base)})
			delete(units, name)
			continue
		}
		u.files = files
		if u.first, err = vcs.FirstChangeTime(base, u.bookmark); err != nil {
			return plan, err
		}
	}

	// Rule 1: dependencies between the tasks of different bookmarks.
	for _, task := range tasks {
		u, ok := units[unitOf[task.Name]]
		if !ok {
			continue
		}
		for _, dependency := range task.DependsOn {
			other, inSet := unitOf[dependency.Name]
			switch {
			case !inSet:
				if dependency.State != "closed" {
					plan.Warnings = append(plan.Warnings, fmt.Sprintf("task %s depends on %s, which is not in the set and not closed", task.Name, dependency.Name))
				}
			case other != u.bookmark:
				if _, merged := units[other]; merged {
					addReason(u, other, fmt.Sprintf("%s depends on %s", task.Name, dependency.Name))
				}
			}
		}
	}
	// Rule 2: stacks.
	names := slices.Sorted(maps.Keys(units))
	for _, a := range names {
		for _, b := range names {
			if a == b {
				continue
			}
			below, err := vcs.IsAncestor(a, b)
			if err != nil {
				return plan, err
			}
			if below {
				addReason(units[b], a, fmt.Sprintf("%s is based on %s", b, a))
			}
		}
	}
	// Rule 3: overlaps.
	overlaps := map[string][]Overlap{}
	for _, a := range names {
		for _, b := range names {
			if a == b {
				continue
			}
			if shared := intersect(units[a].files, units[b].files); len(shared) > 0 {
				overlaps[a] = append(overlaps[a], Overlap{Bookmark: b, Files: shared})
			}
		}
	}

	waiting := map[string]int{}
	for _, name := range names {
		waiting[name] = len(units[name].after)
	}
	placed := map[string]bool{}
	for len(placed) < len(names) {
		var ready []string
		for _, name := range names {
			if !placed[name] && waiting[name] == 0 {
				ready = append(ready, name)
			}
		}
		if len(ready) == 0 {
			var rest []string
			for _, name := range names {
				if !placed[name] {
					rest = append(rest, fmt.Sprintf("%s (%s)", name, strings.Join(units[name].tasks, ", ")))
				}
			}
			return plan, fmt.Errorf("The dependencies and stacks of these bookmarks form a cycle, so no order is possible: %s.", strings.Join(rest, "; "))
		}
		slices.SortFunc(ready, func(a, b string) int {
			if d := len(overlaps[a]) - len(overlaps[b]); d != 0 {
				return d
			}
			if c := units[a].first.Compare(units[b].first); c != 0 {
				return c
			}
			return strings.Compare(a, b)
		})
		next := units[ready[0]]
		placed[next.bookmark] = true
		place := Place{Place: len(plan.Order) + 1, Bookmark: next.bookmark, Tasks: next.tasks, After: []Reason{}, Overlaps: overlaps[next.bookmark], FirstChange: next.first}
		for _, other := range slices.Sorted(maps.Keys(next.after)) {
			place.After = append(place.After, Reason{Bookmark: other, Why: next.after[other]})
		}
		if place.Overlaps == nil {
			place.Overlaps = []Overlap{}
		}
		plan.Order = append(plan.Order, place)
		for _, name := range names {
			if _, ok := units[name].after[next.bookmark]; ok {
				waiting[name]--
			}
		}
	}
	return plan, nil
}

func addReason(u *unit, before, why string) {
	if existing, ok := u.after[before]; ok {
		if !strings.Contains(existing, why) {
			u.after[before] = existing + "; " + why
		}
		return
	}
	u.after[before] = why
}

func intersect(a, b []string) []string {
	var shared []string
	for _, file := range a {
		if slices.Contains(b, file) {
			shared = append(shared, file)
		}
	}
	return shared
}
