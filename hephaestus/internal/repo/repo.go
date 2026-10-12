// Package repo finds the root of a repo and asks its version control system (jj or git) for the
// facts that hephaestus needs: the files of a revision, the files that a line of work changes, and
// the order of revisions. It runs the jj or git command; it never changes the repo.
package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// VCS names the version control system of a repo.
type VCS string

const (
	JJ   VCS = "jj"
	Git  VCS = "git"
	None VCS = ""
)

// Repo is the root of a working copy and its version control system.
type Repo struct {
	Root string
	VCS  VCS
}

// Find returns the repo that holds dir: the nearest directory above it with .jj or .git. A
// directory in no repo gives a Repo with the VCS None and the root dir.
func Find(dir string) (Repo, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Repo{}, err
	}
	for d := abs; ; d = filepath.Dir(d) {
		if isDir(filepath.Join(d, ".jj")) {
			return Repo{Root: d, VCS: JJ}, nil
		}
		if exists(filepath.Join(d, ".git")) {
			return Repo{Root: d, VCS: Git}, nil
		}
		if filepath.Dir(d) == d {
			return Repo{Root: abs, VCS: None}, nil
		}
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// run runs the VCS command in the root, and returns its stdout. A failure gives the stderr of the
// command in the error.
func (r Repo) run(args ...string) (string, error) {
	if r.VCS == None {
		return "", errors.New("This directory is in no jj or git repo.")
	}
	cmd := exec.Command(string(r.VCS), args...)
	cmd.Dir = r.Root
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %v: %s", r.VCS, strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.String(), nil
}

// lines splits the output of a command into its lines that are not empty.
func lines(output string) []string {
	var result []string
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, filepath.ToSlash(line))
		}
	}
	return result
}

// Files returns the paths, from the root and with "/", of the files of the working copy that the VCS
// tracks or would track: ignored files are left out. Without a VCS, every file below the root except
// in .git, .jj, and node_modules.
func (r Repo) Files() ([]string, error) {
	switch r.VCS {
	case JJ:
		out, err := r.run("file", "list")
		return lines(out), err
	case Git:
		out, err := r.run("ls-files", "--cached", "--others", "--exclude-standard")
		return lines(out), err
	}
	var files []string
	err := filepath.WalkDir(r.Root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() && slices.Contains([]string{".git", ".jj", "node_modules"}, entry.Name()) {
			return filepath.SkipDir
		}
		if !entry.IsDir() {
			rel, _ := filepath.Rel(r.Root, path)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	return files, err
}

// FilesAt returns the paths of the files of revision rev below dir (a path from the root).
func (r Repo) FilesAt(rev, dir string) ([]string, error) {
	switch r.VCS {
	case JJ:
		out, err := r.run("file", "list", "-r", rev, "--", dir)
		return lines(out), err
	case Git:
		out, err := r.run("ls-tree", "-r", "--name-only", rev, "--", dir)
		return lines(out), err
	}
	return nil, errors.New("Finding the files of a revision needs jj or git.")
}

// ChangedFiles returns the paths of the files that the changes from base to rev change.
func (r Repo) ChangedFiles(base, rev string) ([]string, error) {
	switch r.VCS {
	case JJ:
		out, err := r.run("diff", "--name-only", "--from", base, "--to", rev)
		return lines(out), err
	case Git:
		out, err := r.run("diff", "--name-only", base+"..."+rev)
		return lines(out), err
	}
	return nil, errors.New("Finding the changed files needs jj or git.")
}

// IsAncestor reports whether revision a is an ancestor of revision b, or the same revision.
func (r Repo) IsAncestor(a, b string) (bool, error) {
	switch r.VCS {
	case JJ:
		out, err := r.run("log", "--no-graph", "-r", fmt.Sprintf("(%s) & ::(%s)", a, b), "-T", `commit_id ++ "\n"`)
		return len(lines(out)) > 0, err
	case Git:
		cmd := exec.Command("git", "merge-base", "--is-ancestor", a, b)
		cmd.Dir = r.Root
		err := cmd.Run()
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 {
			return false, nil
		}
		return err == nil, err
	}
	return false, errors.New("Comparing revisions needs jj or git.")
}

// FirstChangeTime returns the author time of the oldest change from base to rev. With no change
// in that range, it returns the zero time.
func (r Repo) FirstChangeTime(base, rev string) (time.Time, error) {
	var out string
	var err error
	switch r.VCS {
	case JJ:
		out, err = r.run("log", "--no-graph", "-r", fmt.Sprintf("(%s)..(%s)", base, rev), "-T", `author.timestamp().format("%s") ++ "\n"`)
	case Git:
		out, err = r.run("log", "--format=%at", base+".."+rev)
	default:
		return time.Time{}, errors.New("Reading the time of changes needs jj or git.")
	}
	if err != nil {
		return time.Time{}, err
	}
	var oldest time.Time
	for _, line := range lines(out) {
		seconds, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			continue
		}
		if t := time.Unix(seconds, 0); oldest.IsZero() || t.Before(oldest) {
			oldest = t
		}
	}
	return oldest, nil
}

// Exists reports whether rev names a revision of the repo.
func (r Repo) Exists(rev string) bool {
	switch r.VCS {
	case JJ:
		_, err := r.run("log", "--no-graph", "-r", rev, "-T", `""`, "--limit", "1")
		return err == nil
	case Git:
		_, err := r.run("rev-parse", "--verify", "--quiet", rev+"^{commit}")
		return err == nil
	}
	return false
}

// SharedDir returns the directory of hephaestus that every workspace of the repo shares: hephaestus/
// in the jj repo directory (.jj/repo, which each jj workspace points to) or in the git common
// directory (which each git worktree points to). Without jj or git, it is .hephaestus/ in the root.
func (r Repo) SharedDir() string {
	switch r.VCS {
	case JJ:
		repoDir := filepath.Join(r.Root, ".jj", "repo")
		if info, err := os.Stat(repoDir); err == nil && !info.IsDir() {
			if target, err := os.ReadFile(repoDir); err == nil {
				repoDir = resolveFrom(filepath.Join(r.Root, ".jj"), strings.TrimSpace(string(target)))
			}
		}
		return filepath.Join(repoDir, "hephaestus")
	case Git:
		return filepath.Join(gitCommonDir(filepath.Join(r.Root, ".git")), "hephaestus")
	}
	return filepath.Join(r.Root, ".hephaestus")
}

// gitCommonDir returns the directory that every worktree of a git repo shares. In the main worktree,
// .git is that directory; in another worktree, .git is a file "gitdir: <dir>", and <dir>/commondir
// holds the path of the shared directory, relative to <dir>.
func gitCommonDir(dotGit string) string {
	info, err := os.Stat(dotGit)
	if err != nil || info.IsDir() {
		return dotGit
	}
	content, err := os.ReadFile(dotGit)
	if err != nil {
		return dotGit
	}
	gitDir, found := strings.CutPrefix(strings.TrimSpace(string(content)), "gitdir:")
	if !found {
		return dotGit
	}
	worktree := resolveFrom(filepath.Dir(dotGit), strings.TrimSpace(gitDir))
	common, err := os.ReadFile(filepath.Join(worktree, "commondir"))
	if err != nil {
		return worktree
	}
	return resolveFrom(worktree, strings.TrimSpace(string(common)))
}

func resolveFrom(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, filepath.FromSlash(path))
}

// WorkingDiff returns the changes of the working copy, as a git diff: in jj, the change of the
// working-copy commit; in git, the changes that are not committed, staged or not.
func (r Repo) WorkingDiff() (string, error) {
	switch r.VCS {
	case JJ:
		return r.run("diff", "--git")
	case Git:
		diff, err := r.run("diff", "HEAD")
		if err != nil {
			return "", err
		}
		// git diff leaves out new files that git does not track yet; jj diff has them.
		untracked, err := r.run("ls-files", "--others", "--exclude-standard")
		if err != nil {
			return "", err
		}
		for _, file := range lines(untracked) {
			cmd := exec.Command("git", "diff", "--no-index", "--", os.DevNull, file)
			cmd.Dir = r.Root
			out, _ := cmd.Output() // exit code 1 means that there is a difference
			diff += strings.ReplaceAll(string(out), "a/"+os.DevNull, "a/"+file)
		}
		return diff, nil
	}
	return "", errors.New("Finding the changes of the working copy needs jj or git.")
}

// DiffFiles returns the paths of the files that a git diff changes.
func DiffFiles(diff string) []string {
	var files []string
	for _, line := range strings.Split(diff, "\n") {
		if rest, ok := strings.CutPrefix(line, "diff --git a/"); ok {
			if i := strings.Index(rest, " b/"); i >= 0 {
				files = append(files, rest[:i])
			}
		}
	}
	return files
}
