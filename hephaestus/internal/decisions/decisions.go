// Package decisions renumbers decision records that two lines of work numbered the same. A decision
// record is a file NNNN-<title>.md in a decisions directory, for example
// design/decisions/cli/0021-completion-for-five-shells.md. Two lines of work can each take the next
// free number; when they merge, the record that the base does not have gets the next number after
// the others, and every link to it changes.
package decisions

import (
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/wikilink"
)

var recordPattern = regexp.MustCompile(`^(\d{4})-([a-z0-9-]+)\.md$`)

// Renumbering is one decision record that gets a new number.
type Renumbering struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Plan is what Renumber changes.
type Plan struct {
	Renumberings []Renumbering `json:"renumberings"`
	// Links counts the links that change, by the path of their file.
	Links map[string]int `json:"links"`
	// Problems are collisions that the base has itself, which Renumber does not change.
	Problems []string `json:"problems"`
	files    map[string]string
}

// PlanRenumber finds the decision records of the working copy that have the number of another
// record in the same directory, and gives each record that base does not have the next free number
// of its directory. It changes nothing.
func PlanRenumber(r repo.Repo, base string) (Plan, error) {
	files, err := r.Files()
	if err != nil {
		return Plan{}, err
	}
	byDir := map[string][]string{}
	for _, file := range files {
		if dir := path.Dir(file); isDecisionsDir(dir) && recordPattern.MatchString(path.Base(file)) {
			byDir[dir] = append(byDir[dir], file)
		}
	}
	plan := Plan{Renumberings: []Renumbering{}, Links: map[string]int{}, Problems: []string{}, files: map[string]string{}}
	renames := map[string]string{}
	for _, dir := range slices.Sorted(maps.Keys(byDir)) {
		records := byDir[dir]
		slices.Sort(records)
		numbers := map[string][]string{}
		highest := 0
		for _, record := range records {
			number := recordPattern.FindStringSubmatch(path.Base(record))[1]
			numbers[number] = append(numbers[number], record)
			if n, _ := strconv.Atoi(number); n > highest {
				highest = n
			}
		}
		var inBase []string
		for _, number := range slices.Sorted(maps.Keys(numbers)) {
			same := numbers[number]
			if len(same) < 2 {
				continue
			}
			if inBase == nil {
				if inBase, err = r.FilesAt(base, dir); err != nil {
					return Plan{}, err
				}
			}
			var kept []string
			var moved []string
			for _, record := range same {
				if slices.Contains(inBase, record) {
					kept = append(kept, record)
				} else {
					moved = append(moved, record)
				}
			}
			if len(kept) > 1 {
				plan.Problems = append(plan.Problems, fmt.Sprintf("%s already has %s with the same number. Renumber one of them by hand.", base, strings.Join(kept, " and ")))
				continue
			}
			if len(kept) == 0 {
				// None is in the base: the first keeps its number.
				moved = moved[1:]
			}
			for _, record := range moved {
				highest++
				title := recordPattern.FindStringSubmatch(path.Base(record))[2]
				to := path.Join(dir, fmt.Sprintf("%04d-%s.md", highest, title))
				plan.Renumberings = append(plan.Renumberings, Renumbering{From: record, To: to})
				renames[strings.TrimSuffix(path.Base(record), ".md")] = strings.TrimSuffix(path.Base(to), ".md")
			}
		}
	}
	if len(renames) == 0 {
		return plan, nil
	}
	for _, file := range files {
		if !strings.HasSuffix(strings.ToLower(file), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(r.Root, filepath.FromSlash(file)))
		if err != nil {
			continue
		}
		changed, n := wikilink.Rename(string(data), renames)
		if n > 0 {
			plan.Links[file] = n
			plan.files[file] = changed
		}
	}
	return plan, nil
}

// isDecisionsDir reports whether dir is a directory of decision records: a directory named
// decisions, or a directory below one.
func isDecisionsDir(dir string) bool {
	return slices.Contains(strings.Split(dir, "/"), "decisions")
}

// Apply writes the plan: the links first, then each record moves to its new name with the new
// number in its "# NNNN: Title" heading.
func Apply(r repo.Repo, plan Plan) error {
	for file, content := range plan.files {
		if err := os.WriteFile(filepath.Join(r.Root, filepath.FromSlash(file)), []byte(content), 0o644); err != nil {
			return err
		}
	}
	for _, rn := range plan.Renumberings {
		from := filepath.Join(r.Root, filepath.FromSlash(rn.From))
		data, err := os.ReadFile(from)
		if err != nil {
			return err
		}
		oldNumber := path.Base(rn.From)[:4]
		newNumber := path.Base(rn.To)[:4]
		text := strings.Replace(string(data), "# "+oldNumber+":", "# "+newNumber+":", 1)
		if err := os.WriteFile(filepath.Join(r.Root, filepath.FromSlash(rn.To)), []byte(text), 0o644); err != nil {
			return err
		}
		if err := os.Remove(from); err != nil {
			return err
		}
	}
	return nil
}
