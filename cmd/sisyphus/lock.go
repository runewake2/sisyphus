package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// sharedStateDir returns the directory where sisyphus keeps state that every workspace of the repo
// shares: <jj repo>/sisyphus in a jj repo, and <git common dir>/sisyphus in a git repo. Every jj
// workspace and every git worktree of one repo finds the same directory. ok is false when root is in
// no jj or git repo; then nothing is shared.
func sharedStateDir(root string) (dir string, ok bool) {
	for d := root; ; d = filepath.Dir(d) {
		if info, err := os.Stat(filepath.Join(d, ".jj")); err == nil && info.IsDir() {
			if repo, ok := jjRepoDir(filepath.Join(d, ".jj")); ok {
				return filepath.Join(repo, "sisyphus"), true
			}
		}
		if common, ok := gitCommonDir(filepath.Join(d, ".git")); ok {
			return filepath.Join(common, "sisyphus"), true
		}
		if filepath.Dir(d) == d {
			return "", false
		}
	}
}

// jjRepoDir returns the repo directory of the jj workspace whose .jj directory is dotJJ. In the first
// workspace, .jj/repo is the directory. In each other workspace, .jj/repo is a file that holds the path
// of that directory, relative to .jj or absolute.
func jjRepoDir(dotJJ string) (string, bool) {
	repo := filepath.Join(dotJJ, "repo")
	info, err := os.Stat(repo)
	if err != nil {
		return "", false
	}
	if info.IsDir() {
		return repo, true
	}
	target, err := os.ReadFile(repo)
	if err != nil {
		return "", false
	}
	return resolveFrom(dotJJ, strings.TrimSpace(string(target))), true
}

// gitCommonDir returns the directory that every worktree of the git repo shares. In the main
// worktree, .git is that directory. In each other worktree, .git is a file "gitdir: <dir>", and
// <dir>/commondir holds the path of the shared directory, relative to <dir>.
func gitCommonDir(dotGit string) (string, bool) {
	info, err := os.Stat(dotGit)
	if err != nil {
		return "", false
	}
	if info.IsDir() {
		return dotGit, true
	}
	content, err := os.ReadFile(dotGit)
	if err != nil {
		return "", false
	}
	gitDir, found := strings.CutPrefix(strings.TrimSpace(string(content)), "gitdir:")
	if !found {
		return "", false
	}
	worktree := resolveFrom(filepath.Dir(dotGit), strings.TrimSpace(gitDir))
	common, err := os.ReadFile(filepath.Join(worktree, "commondir"))
	if err != nil {
		return worktree, true
	}
	return resolveFrom(worktree, strings.TrimSpace(string(common))), true
}

// resolveFrom returns path, made absolute from base when it is relative.
func resolveFrom(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, filepath.FromSlash(path))
}

// lockFile returns the path of the lock that a command holds while it changes issues: in the shared
// state directory, so that it is outside every working copy and the same for every workspace, or in
// issues/ when the repo uses no jj and no git.
func lockFile(root string) string {
	if dir, ok := sharedStateDir(root); ok {
		return filepath.Join(dir, "lock")
	}
	return filepath.Join(root, "issues", ".lock")
}

// lockWait is how long a command waits for another command to release the lock. Tests make it shorter.
var lockWait = 10 * time.Second

const (
	// lockStale is the age of a lock that no command holds any more, for example after a crash. A
	// command that changes issues holds the lock for much less time.
	lockStale = 30 * time.Second
	lockRetry = 20 * time.Millisecond
)

// lockIssues takes the lock of the repo of root, and waits while another command holds it. A lock
// older than lockStale is removed. The returned function releases the lock.
func lockIssues(root string) (func(), error) {
	path := lockFile(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(lockWait)
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(file, "pid %d since %s\n", os.Getpid(), time.Now().Format(time.RFC3339))
			file.Close()
			return func() { os.Remove(path) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("Cannot take the lock %s: %w", path, err)
		}
		if info, err := os.Stat(path); err == nil && time.Since(info.ModTime()) > lockStale {
			os.Remove(path)
			continue
		}
		if time.Now().After(deadline) {
			holder, _ := os.ReadFile(path)
			return nil, fmt.Errorf("Another sisyphus command holds the lock %s (%s). Try again. If no sisyphus command runs, remove the file.", path, strings.TrimSpace(string(holder)))
		}
		time.Sleep(lockRetry)
	}
}

// writeFileAtomically writes data to a temporary file in the directory of path, then renames it to
// path, so that a command that reads path at the same time sees the old file or the new file, never a
// part of one.
func writeFileAtomically(path string, data []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name()) // nothing to remove after the rename
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.Chmod(temp.Name(), mode); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}
