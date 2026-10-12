// Package mergequeue makes one linear history from a set of tasks ([[agent-delegation#Merge queue]]).
// It merges the bookmark of each task, one at a time, in the order of package mergeorder, onto a new
// line based on the base, gives the unreleased changelog entries of each bookmark their versions,
// renumbers decision records that collide, and runs the checks of the repo after each bookmark. A
// bookmark that conflicts or fails the checks is left out. The result is a new bookmark; the base,
// the bookmarks of the tasks, and the working copies of the users do not change. It never moves the
// base and never pushes.
//
// The merge happens with git, in a scratch worktree of the git directory of the repo: a git repo, or
// the git store of a jj repo, where every jj bookmark is a git branch.
package mergequeue

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/decisions"
	"github.com/runewake2/sisyphus/hephaestus/internal/mergeorder"
	"github.com/runewake2/sisyphus/hephaestus/internal/release"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
)

// Options are the input of a run.
type Options struct {
	// Repo is the repo of the user; its git directory holds the commits.
	Repo repo.Repo
	// Tasks are the tasks to merge, with their bookmarks and dependencies.
	Tasks []mergeorder.Task
	// Base is the base of the new line, normally main.
	Base string
	// Name is the bookmark of the result, for example merge-queue/2026-10-12.
	Name string
	// Checks is a command line that must succeed after each bookmark, in the scratch worktree. Empty:
	// no checks.
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

// worktree is the scratch worktree of a run.
type worktree struct {
	gitDir   string
	dir      string
	out      io.Writer
	identity []string
}

// git runs git in the worktree, and returns its stdout.
func (w worktree) git(args ...string) (string, error) {
	cmd := exec.Command("git", append(slices.Clone(w.identity), args...)...)
	cmd.Dir = w.dir
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}

// identity gives git a committer when the repo has none, so that a run never fails for it. A
// configured name and email win.
func identity(dir string) []string {
	var args []string
	for key, fallback := range map[string]string{"user.name": "hephaestus merge queue", "user.email": "hephaestus@localhost"} {
		cmd := exec.Command("git", "config", key)
		cmd.Dir = dir
		if out, err := cmd.Output(); err != nil || strings.TrimSpace(string(out)) == "" {
			args = append(args, "-c", key+"="+fallback)
		}
	}
	slices.Sort(args)
	return args
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
// error stops the run, and changes nothing outside the scratch worktree.
func Run(ctx context.Context, o Options) (Report, error) {
	report := Report{Name: o.Name, Base: o.Base, Merged: []Merged{}, LeftOut: []LeftOut{}, Warnings: []string{}}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Out == nil {
		o.Out = io.Discard
	}
	gitDir, err := o.Repo.GitDir()
	if err != nil {
		return report, err
	}
	git := func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"--git-dir=" + gitDir}, args...)...)
		var out, errOut bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errOut
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
		}
		return strings.TrimSpace(out.String()), nil
	}
	if out, _ := git("rev-parse", "--verify", "--quiet", "refs/heads/"+o.Name); out != "" {
		return report, fmt.Errorf("The bookmark %s already exists. Give another --name, or remove it.", o.Name)
	}
	baseID, err := git("rev-parse", "--verify", o.Base+"^{commit}")
	if err != nil {
		return report, fmt.Errorf("The base %q does not exist.", o.Base)
	}
	report.BaseID, report.Head = baseID, baseID

	scratch, err := os.MkdirTemp("", "hephaestus-merge-queue-")
	if err != nil {
		return report, err
	}
	os.Remove(scratch) // git worktree add makes it
	if _, err := git("worktree", "add", "--detach", "--quiet", scratch, baseID); err != nil {
		return report, err
	}
	defer func() {
		git("worktree", "remove", "--force", scratch)
		os.RemoveAll(scratch)
		git("worktree", "prune")
	}()
	w := worktree{gitDir: gitDir, dir: scratch, out: o.Out, identity: identity(scratch)}
	scratchRepo := repo.Repo{Root: scratch, VCS: repo.Git}

	plan, err := mergeorder.Compute(scratchRepo, o.Tasks, baseID)
	if err != nil {
		return report, err
	}
	report.Warnings = append(report.Warnings, plan.Warnings...)
	for _, skipped := range plan.Skipped {
		report.LeftOut = append(report.LeftOut, LeftOut{Bookmark: skipped.Bookmark, Tasks: skipped.Tasks, Reason: "nothing", Detail: skipped.Why})
	}

	head := baseID
	var mergedBookmarks []string
	leftOut := map[string]bool{}
	for _, place := range plan.Order {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		fmt.Fprintf(o.Out, "merge-queue: %d. %s (%s)\n", place.Place, place.Bookmark, strings.Join(place.Tasks, ", "))
		leave := func(reason, detail string, files []string) {
			leftOut[place.Bookmark] = true
			report.LeftOut = append(report.LeftOut, LeftOut{Bookmark: place.Bookmark, Tasks: place.Tasks, Reason: reason, Detail: detail, Files: files})
			w.git("cherry-pick", "--abort")
			w.git("reset", "--quiet", "--hard", head)
			w.git("clean", "-fdq")
			fmt.Fprintf(o.Out, "merge-queue: left out %s: %s\n", place.Bookmark, detail)
		}
		var missing []string
		for _, before := range place.After {
			if leftOut[before.Bookmark] {
				missing = append(missing, before.Bookmark)
			}
		}
		if len(missing) > 0 {
			leave("dependency", fmt.Sprintf("it must come after %s, which is left out", strings.Join(missing, ", ")), nil)
			continue
		}
		// The commits of the bookmark that the line does not have yet: not in the base, and not in a
		// bookmark that is merged already (a stack).
		args := []string{"rev-list", "--reverse", "--topo-order", place.Bookmark, "^" + baseID}
		for _, done := range mergedBookmarks {
			args = append(args, "^"+done)
		}
		list, err := w.git(args...)
		if err != nil {
			return report, err
		}
		commits := lines(list)
		if merges, _ := w.git(append([]string{"rev-list", "--merges"}, args[3:]...)...); len(lines(merges)) > 0 {
			leave("merges", "its history has merge commits; the merge queue makes a linear history", nil)
			continue
		}
		conflict := false
		for _, commit := range commits {
			if _, err := w.git("cherry-pick", "--allow-empty", "--keep-redundant-commits", commit); err != nil {
				files, _ := w.git("diff", "--name-only", "--diff-filter=U")
				leave("conflict", fmt.Sprintf("its change %.12s conflicts with the line", commit), lines(files))
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}
		merged := Merged{Place: len(report.Merged) + 1, Bookmark: place.Bookmark, Tasks: place.Tasks, Commits: len(commits), Versions: []Version{}, Renumbered: []decisions.Renumbering{}, After: place.After}

		// Decision records that collide with the line, then the versions of the new entries.
		renumber, err := decisions.PlanRenumber(scratchRepo, head)
		if err != nil {
			return report, err
		}
		if err := decisions.Apply(scratchRepo, renumber); err != nil {
			return report, err
		}
		merged.Renumbered = renumber.Renumberings
		added, err := w.git("log", "--reverse", "--diff-filter=A", "--name-only", "--format=", head+"..HEAD", "--", release.UnreleasedDir)
		if err != nil {
			return report, err
		}
		var entries []string
		for _, file := range lines(added) {
			name := strings.TrimSuffix(path.Base(file), ".md")
			if _, err := os.Stat(scratch + "/" + file); err == nil && !slices.Contains(entries, name) {
				entries = append(entries, name)
			}
		}
		if len(entries) > 0 {
			assign, err := release.PlanAssign(scratchRepo, entries, o.Now())
			if err != nil {
				leave("versions", "its unreleased entries cannot get versions: "+err.Error(), nil)
				continue
			}
			if err := release.Apply(scratchRepo, assign); err != nil {
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
			if _, err := w.git("add", "-A"); err != nil {
				return report, err
			}
			message := fmt.Sprintf("merge-queue: give versions to %s\n\n%s\n", place.Bookmark, strings.Join(body, "\n"))
			if _, err := w.git("commit", "--quiet", "-m", message); err != nil {
				return report, err
			}
		}
		if o.Checks != "" {
			fmt.Fprintf(o.Out, "merge-queue: checks of %s: %s\n", place.Bookmark, o.Checks)
			if output, ok := runChecks(ctx, scratch, o.Checks, o.Out); !ok {
				leave("checks", "the checks failed after it merged:\n"+tail(output, 30), nil)
				continue
			}
		}
		newHead, err := w.git("rev-parse", "HEAD")
		if err != nil {
			return report, err
		}
		head = strings.TrimSpace(newHead)
		mergedBookmarks = append(mergedBookmarks, place.Bookmark)
		report.Merged = append(report.Merged, merged)
	}
	report.Head = head
	if head != baseID {
		if _, err := git("branch", o.Name, head); err != nil {
			return report, err
		}
		report.Created = true
	}
	return report, nil
}

// runChecks runs the checks in dir, and returns their output and whether they passed.
func runChecks(ctx context.Context, dir, checks string, out io.Writer) (string, bool) {
	cmd := exec.CommandContext(ctx, "sh", "-c", checks)
	if isWindows() {
		cmd = exec.CommandContext(ctx, "cmd", "/C", checks)
	}
	cmd.Dir = dir
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(&output, out)
	cmd.Stderr = io.MultiWriter(&output, out)
	err := cmd.Run()
	return output.String(), err == nil
}

func isWindows() bool { return os.PathSeparator == '\\' }

func tail(text string, n int) string {
	all := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(all) > n {
		all = all[len(all)-n:]
	}
	return strings.Join(all, "\n")
}
