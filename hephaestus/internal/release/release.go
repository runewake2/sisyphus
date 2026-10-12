// Package release gives versions to unreleased changelog entries. A line of work does not pick a
// version: it writes an unreleased entry, changelog/unreleased/unreleased-<topic>.md, and links to
// it. When the merge queue lands the work, Assign gives each entry the next SHAME version, in the
// order of the merge, so that two lines of work never pick the same version.
package release

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/wikilink"
)

// UnreleasedDir is the directory of the unreleased entries, from the root of the repo.
const UnreleasedDir = "changelog/unreleased"

// Prefix starts the name of every unreleased entry, so that a link to it is unique and clear.
const Prefix = "unreleased-"

var (
	titlePattern     = regexp.MustCompile(`^# Unreleased: (.+)$`)
	changelogPattern = regexp.MustCompile(`^- Changelog: (.+)$`)
	datePattern      = regexp.MustCompile(`^- Date: (\d{4}-\d{2}-\d{2})$`)
	versionPattern   = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
)

// Entry is one unreleased changelog entry.
type Entry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Title     string `json:"title"`
	Changelog string `json:"changelog"`
	Date      string `json:"date"`
	content   string
}

// Entries reads every unreleased entry of the repo, sorted by date, then name.
func Entries(root string) ([]Entry, error) {
	files, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(UnreleasedDir), "*.md"))
	if err != nil {
		return nil, err
	}
	var entries []Entry
	var problems []string
	for _, file := range files {
		entry, err := readEntry(root, file)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		entries = append(entries, entry)
	}
	if len(problems) > 0 {
		return nil, errors.New(strings.Join(problems, "\n"))
	}
	slices.SortFunc(entries, func(a, b Entry) int {
		if c := strings.Compare(a.Date, b.Date); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})
	return entries, nil
}

func readEntry(root, file string) (Entry, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Entry{}, err
	}
	rel, _ := filepath.Rel(root, file)
	entry := Entry{Name: strings.TrimSuffix(filepath.Base(file), ".md"), Path: filepath.ToSlash(rel), content: string(data)}
	if !strings.HasPrefix(entry.Name, Prefix) {
		return Entry{}, fmt.Errorf("%s: the name of an unreleased entry starts with %q, for example %s%s.md.", entry.Path, Prefix, Prefix, "fix-login-bug")
	}
	for _, line := range strings.Split(strings.ReplaceAll(entry.content, "\r\n", "\n"), "\n") {
		if m := titlePattern.FindStringSubmatch(line); m != nil && entry.Title == "" {
			entry.Title = strings.TrimSpace(m[1])
		}
		if m := changelogPattern.FindStringSubmatch(line); m != nil && entry.Changelog == "" {
			entry.Changelog = strings.TrimSpace(m[1])
		}
		if m := datePattern.FindStringSubmatch(line); m != nil && entry.Date == "" {
			entry.Date = m[1]
		}
	}
	var missing []string
	if entry.Title == "" {
		missing = append(missing, `a first line "# Unreleased: <title>"`)
	}
	if entry.Changelog == "" {
		missing = append(missing, `a line "- Changelog: <one line for CHANGELOG.md>"`)
	}
	if len(missing) > 0 {
		return Entry{}, fmt.Errorf("%s needs %s.", entry.Path, strings.Join(missing, ", and "))
	}
	return entry, nil
}

// Assignment is one entry and the version that it gets.
type Assignment struct {
	Entry   Entry  `json:"entry"`
	Version string `json:"version"`
	To      string `json:"to"`
}

// Plan is what Assign changes.
type Plan struct {
	From        string       `json:"from"`
	To          string       `json:"to"`
	Assignments []Assignment `json:"assignments"`
	// Links counts the links that change, by the path of their file.
	Links map[string]int `json:"links"`
	date  string
	files map[string]string
}

// PlanAssign gives the next versions, after the version in VERSION, to the entries with the given
// names, in their order, or to every unreleased entry when names is empty. It changes nothing.
func PlanAssign(r repo.Repo, names []string, today time.Time) (Plan, error) {
	all, err := Entries(r.Root)
	if err != nil {
		return Plan{}, err
	}
	entries := all
	if len(names) > 0 {
		entries = nil
		for _, name := range names {
			name = strings.TrimSuffix(filepath.Base(name), ".md")
			index := slices.IndexFunc(all, func(e Entry) bool { return e.Name == name || e.Name == Prefix+name })
			if index < 0 {
				return Plan{}, fmt.Errorf("No unreleased entry is named %q in %s.", name, UnreleasedDir)
			}
			if slices.ContainsFunc(entries, func(e Entry) bool { return e.Name == all[index].Name }) {
				return Plan{}, fmt.Errorf("The entry %s is given twice.", all[index].Name)
			}
			entries = append(entries, all[index])
		}
	}
	if len(entries) == 0 {
		return Plan{}, fmt.Errorf("No unreleased entry in %s: there is nothing to give a version to.", UnreleasedDir)
	}
	current, err := os.ReadFile(filepath.Join(r.Root, "VERSION"))
	if err != nil {
		return Plan{}, fmt.Errorf("Cannot read VERSION: %w", err)
	}
	from := strings.TrimSpace(string(current))
	m := versionPattern.FindStringSubmatch(from)
	if m == nil {
		return Plan{}, fmt.Errorf("VERSION holds %q, not a version PROUD.DEFAULT.SHAME.", from)
	}
	shame, _ := strconv.Atoi(m[3])
	plan := Plan{From: from, Links: map[string]int{}, date: today.Format("2006-01-02"), files: map[string]string{}}
	renames := map[string]string{}
	for i, entry := range entries {
		version := fmt.Sprintf("%s.%s.%d", m[1], m[2], shame+i+1)
		to := "changelog/" + version + ".md"
		if _, err := os.Stat(filepath.Join(r.Root, filepath.FromSlash(to))); err == nil {
			return Plan{}, fmt.Errorf("%s already exists. VERSION (%s) is behind the changelog.", to, from)
		}
		plan.Assignments = append(plan.Assignments, Assignment{Entry: entry, Version: version, To: to})
		renames[entry.Name] = version
		plan.To = version
	}
	files, err := r.Files()
	if err != nil {
		return Plan{}, err
	}
	for _, file := range files {
		if !strings.HasSuffix(strings.ToLower(file), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(r.Root, filepath.FromSlash(file)))
		if err != nil {
			continue // a file that the VCS knows but that is gone from the working copy
		}
		changed, n := wikilink.Rename(string(data), renames)
		if n > 0 {
			plan.Links[file] = n
			plan.files[file] = changed
		}
	}
	return plan, nil
}

// Apply writes the plan: each entry moves to changelog/<version>.md with the heading "# <version>"
// and no Changelog line, CHANGELOG.md gets one entry for each version, newest first, VERSION gets the
// last version, and every link to an entry gets its version.
func Apply(r repo.Repo, plan Plan) error {
	for file, content := range plan.files {
		if err := os.WriteFile(filepath.Join(r.Root, filepath.FromSlash(file)), []byte(content), 0o644); err != nil {
			return err
		}
	}
	changelogPath := filepath.Join(r.Root, "CHANGELOG.md")
	changelog, err := os.ReadFile(changelogPath)
	if err != nil {
		return fmt.Errorf("Cannot read CHANGELOG.md: %w", err)
	}
	text := string(changelog)
	for _, a := range plan.Assignments {
		source := filepath.Join(r.Root, filepath.FromSlash(a.Entry.Path))
		data, err := os.ReadFile(source) // with its own links already changed
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(r.Root, filepath.FromSlash(a.To)), []byte(released(string(data), a.Version)), 0o644); err != nil {
			return err
		}
		if err := os.Remove(source); err != nil {
			return err
		}
		block := fmt.Sprintf("## [[%s]] - %s\n- %s\n\n", a.Version, plan.date, strings.TrimPrefix(a.Entry.Changelog, "- "))
		if i := strings.Index(text, "## [["); i >= 0 {
			text = text[:i] + block + text[i:]
		} else {
			text = strings.TrimRight(text, "\n") + "\n\n" + block
		}
	}
	if err := os.WriteFile(changelogPath, []byte(text), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.Root, "VERSION"), []byte(plan.To+"\n"), 0o644)
}

// released returns the text of an unreleased entry as the changelog file of version: the heading
// "# <version>", and no Changelog line.
func released(content, version string) string {
	lines := strings.SplitAfter(content, "\n")
	var out []string
	headed := false
	for _, line := range lines {
		bare := strings.TrimRight(line, "\r\n")
		if !headed && titlePattern.MatchString(bare) {
			out = append(out, "# "+version+line[len(bare):])
			headed = true
			continue
		}
		if changelogPattern.MatchString(bare) {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "")
}
