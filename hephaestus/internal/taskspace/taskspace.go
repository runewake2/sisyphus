// Package taskspace makes the workspace of one task: a jj workspace or a git worktree of its own,
// with its own bookmark, so that workers that run at the same time never edit the same files
// ([[AGENTS#Workspaces and bookmarks]]). It is in <workspaces>/<repo>-<name>, as the rules of the
// repo ask, and hephaestus removes it when the task is done.
package taskspace

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/scratch"
)

// Space is the workspace of one task.
type Space struct {
	// Name is the name of the workspace, for example implementer-fix-login-bug.
	Name string `json:"name"`
	Dir  string `json:"dir"`
	// Bookmark is the bookmark of the work, for example ai/implementer-fix-login-bug.
	Bookmark string `json:"bookmark"`
	// Reused is true when the workspace was there already, for example from a task that stopped.
	Reused bool `json:"reused,omitempty"`
}

var unsafe = regexp.MustCompile(`[^a-z0-9-]+`)

// Name returns a workspace name from its parts: lowercase words with "-".
func Name(parts ...string) string {
	name := strings.ToLower(strings.Join(parts, "-"))
	name = unsafe.ReplaceAllString(name, "-")
	return strings.Trim(regexp.MustCompile(`-+`).ReplaceAllString(name, "-"), "-")
}

// Dir returns the directory of the workspace name: <dir of the main root>/workspaces/<repo>-<name>,
// or <workspaces>/<repo>-<name> when workspaces is given.
func Dir(r repo.Repo, workspaces, name string) string {
	main := r.MainRoot()
	if workspaces == "" {
		workspaces = filepath.Join(filepath.Dir(main), "workspaces")
	}
	return filepath.Join(workspaces, filepath.Base(main)+"-"+name)
}

// Create makes the workspace name, based on base, with the bookmark: a jj workspace in a jj repo, and
// a git worktree on a new branch in a git repo. In jj, the worker sets the bookmark on its work. A
// workspace that is there already is used again.
func Create(r repo.Repo, workspaces, name, base, bookmark string) (Space, error) {
	s := Space{Name: name, Dir: Dir(r, workspaces, name), Bookmark: bookmark}
	if info, err := os.Stat(s.Dir); err == nil && info.IsDir() {
		s.Reused = true
		return s, nil
	}
	if err := os.MkdirAll(filepath.Dir(s.Dir), 0o755); err != nil {
		return s, err
	}
	switch r.VCS {
	case repo.JJ:
		if _, err := run(r.Root, "jj", "workspace", "add", s.Dir, "--name", name, "-r", base); err != nil {
			return s, err
		}
	case repo.Git:
		gitDir, err := r.GitDir()
		if err != nil {
			return s, err
		}
		if _, err := scratch.Git(gitDir, "", "rev-parse", "--verify", "--quiet", "refs/heads/"+bookmark); err == nil {
			_, err = scratch.Git(gitDir, "", "worktree", "add", "--quiet", s.Dir, bookmark)
			if err != nil {
				return s, err
			}
		} else if _, err := scratch.Git(gitDir, "", "worktree", "add", "--quiet", "-b", bookmark, s.Dir, base); err != nil {
			return s, err
		}
	default:
		return s, fmt.Errorf("A workspace for a task needs jj or git.")
	}
	return s, nil
}

// Dirty reports whether the workspace has changes that no commit of a bookmark keeps: in git,
// changes that are not committed; in jj, a working-copy commit with changes. Removing it would lose
// them.
func Dirty(r repo.Repo, s Space) bool {
	switch r.VCS {
	case repo.JJ:
		out, err := run(s.Dir, "jj", "diff", "--summary", "-r", "@")
		return err != nil || out != ""
	case repo.Git:
		out, err := run(s.Dir, "git", "status", "--porcelain")
		return err != nil || out != ""
	}
	return false
}

// Remove removes the workspace: jj forgets it, or git removes the worktree, and its directory goes.
// The commits and the bookmark of its work stay.
func Remove(r repo.Repo, s Space) error {
	switch r.VCS {
	case repo.JJ:
		run(s.Dir, "jj", "status") // record the last changes of the working copy in its commit
		if _, err := run(r.Root, "jj", "workspace", "forget", s.Name); err != nil {
			return err
		}
	case repo.Git:
		gitDir, err := r.GitDir()
		if err != nil {
			return err
		}
		if _, err := scratch.Git(gitDir, "", "worktree", "remove", "--force", s.Dir); err != nil {
			return err
		}
	}
	return os.RemoveAll(s.Dir)
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}
