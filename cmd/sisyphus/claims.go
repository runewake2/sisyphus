package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// A claim records that an issue is in progress in one workspace, in the shared state directory (see
// sharedStateDir), so that every jj workspace and git worktree of the repo sees it before the work
// merges. The issue file in the working copy that started the work is the truth: a claim whose
// workspace no longer has the issue in progress is stale.
type claim struct {
	Name  string `json:"name"`
	Owner string `json:"owner,omitempty"`
	// Workspace is the absolute path of the repo root (the directory of issues/) in the working copy
	// that claimed the issue.
	Workspace string `json:"workspace"`
	Bookmark  string `json:"bookmark,omitempty"`
	Since     string `json:"since"`
}

// claimStore is the directory of the claims of one tracker. ok is false when the repo uses no jj and
// no git: then each working copy is alone, and the issue files are enough.
type claimStore struct {
	dir string
	// here is the workspace of this command, in the form of claim.Workspace.
	here string
	ok   bool
}

func openClaims(root string) claimStore {
	shared, workingCopy, ok := findSharedState(root)
	if !ok {
		return claimStore{}
	}
	here := workspacePath(root)
	// A repo can have more than one tracker, in different directories. Each has its own claims.
	tracker := "root"
	if rel, err := filepath.Rel(workspacePath(workingCopy), here); err == nil && rel != "." {
		tracker = "nested-" + strings.ReplaceAll(filepath.ToSlash(rel), "/", "--")
	}
	return claimStore{dir: filepath.Join(shared, "claims", tracker), here: here, ok: true}
}

// workspacePath returns the absolute path of dir with symbolic links resolved, so that two paths of
// one directory compare equal.
func workspacePath(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	return dir
}

func (s claimStore) path(name string) string {
	return filepath.Join(s.dir, filepath.FromSlash(name)+".json")
}

// get returns the claim of the issue name, if there is one.
func (s claimStore) get(name string) (claim, bool) {
	if !s.ok {
		return claim{}, false
	}
	data, err := os.ReadFile(s.path(name))
	if err != nil {
		return claim{}, false
	}
	var c claim
	if json.Unmarshal(data, &c) != nil {
		return claim{}, false
	}
	return c, true
}

func (s claimStore) put(c claim) error {
	if !s.ok {
		return nil
	}
	path := s.path(c.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomically(path, append(data, '\n'))
}

func (s claimStore) remove(name string) {
	if s.ok {
		os.Remove(s.path(name))
		removeEmptyClaimDirectories(s.dir, filepath.Dir(s.path(name)))
	}
}

// all returns every claim of the tracker, sorted by name.
func (s claimStore) all() []claim {
	var claims []claim
	if !s.ok {
		return claims
	}
	filepath.WalkDir(s.dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		rel, err := filepath.Rel(s.dir, strings.TrimSuffix(path, ".json"))
		if err != nil {
			return nil
		}
		if c, ok := s.get(filepath.ToSlash(rel)); ok {
			claims = append(claims, c)
		}
		return nil
	})
	slices.SortFunc(claims, func(a, b claim) int { return strings.Compare(a.Name, b.Name) })
	return claims
}

// isHere reports whether c is a claim of the workspace of this command.
func (s claimStore) isHere(c claim) bool {
	return c.Workspace == s.here
}

// isStale reports whether the workspace of c no longer has the issue in progress, for example
// because the workspace was removed, or its change was abandoned.
func isStale(c claim) bool {
	return !fileExists(filepath.Join(c.Workspace, "issues", "in-progress", filepath.FromSlash(c.Name)+".md"))
}

// claimedElsewhere returns the claim of the issue name by another workspace, when it is not stale.
func (s claimStore) claimedElsewhere(name string) (claim, bool) {
	c, ok := s.get(name)
	if !ok || s.isHere(c) || isStale(c) {
		return claim{}, false
	}
	return c, true
}

func removeEmptyClaimDirectories(top, dir string) {
	for dir != top && strings.HasPrefix(dir, top) {
		if os.Remove(dir) != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

// checkClaims fails when another workspace has a claim on the issue name that is not stale, unless
// take is true. It tells about a stale claim, which recordClaim then replaces. Call it before any
// change, and recordClaim after the checks. The caller holds the lock (see lockIssues).
func checkClaims(s claimStore, name string, take bool, warnings io.Writer) error {
	c, ok := s.get(name)
	if !ok || s.isHere(c) {
		return nil
	}
	where := s.display(c.Workspace)
	switch {
	case isStale(c):
		warn(warnings, fmt.Sprintf("Removed the claim of '%s' by %s: %s no longer has the issue in progress.", name, describeOwner(c.Owner), where))
	case !take:
		return fmt.Errorf("Issue '%s' is in progress in another workspace, %s, owned by %s, since %s. Select another issue, or to take it, add --take.", name, where, describeOwner(c.Owner), displayTime(c.Since))
	default:
		warn(warnings, fmt.Sprintf("Issue '%s' was in progress in another workspace, %s, owned by %s. This workspace now has the claim.", name, where, describeOwner(c.Owner)))
	}
	return nil
}

// recordClaim records that this workspace has the issue name in progress, owned by owner. A claim of
// the same owner in this workspace keeps its start time.
func recordClaim(s claimStore, name, owner, bookmark string) error {
	if !s.ok {
		return nil
	}
	since := time.Now().UTC().Format(time.RFC3339)
	if c, ok := s.get(name); ok && s.isHere(c) && c.Owner == owner {
		since = c.Since
	}
	return s.put(claim{Name: name, Owner: owner, Workspace: s.here, Bookmark: bookmark, Since: since})
}

// display returns the path of workspace from this workspace, for example ../workspaces/fix-login.
func (s claimStore) display(workspace string) string {
	if rel, err := filepath.Rel(s.here, workspace); err == nil {
		return filepath.ToSlash(rel)
	}
	return workspace
}

// displayTime returns an RFC 3339 time in local time, to the minute.
func displayTime(value string) string {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.Local().Format("2006-01-02 15:04")
	}
	return value
}

// releaseIssue removes the claim of this workspace on the issue name, when the issue leaves
// in-progress. The claim of another workspace stays, with a warning: that workspace still works on the
// issue.
func releaseIssue(s claimStore, name string, warnings io.Writer) {
	c, ok := s.get(name)
	if !ok {
		return
	}
	if s.isHere(c) || isStale(c) {
		s.remove(name)
		return
	}
	warn(warnings, fmt.Sprintf("Issue '%s' is also in progress in another workspace, %s, owned by %s. Its claim stays until that workspace stops the work.", name, s.display(c.Workspace), describeOwner(c.Owner)))
}

func describeOwner(owner string) string {
	if owner == "" {
		return "no owner"
	}
	return owner
}

// claimView is one row of sisyphus claims.
type claimView struct {
	claim
	// Status is "here" for a claim of this workspace, "active" for a claim of another workspace, and
	// "stale" for a claim whose workspace no longer has the issue in progress.
	Status string `json:"status"`
}

// listClaims returns every claim of the tracker of root. With prune, it removes the stale claims, and
// returns only the others.
func listClaims(root string, prune bool) ([]claimView, error) {
	s := openClaims(root)
	if !s.ok {
		return nil, errors.New("Claims need a jj or git repo: without one, there are no other workspaces. sisyphus list --state in-progress shows the issues in progress.")
	}
	views := []claimView{}
	for _, c := range s.all() {
		status := "active"
		switch {
		case isStale(c):
			status = "stale"
		case s.isHere(c):
			status = "here"
		}
		if prune && status == "stale" {
			s.remove(c.Name)
			continue
		}
		views = append(views, claimView{claim: c, Status: status})
	}
	return views, nil
}

func claimsCommand(findRoot func() (string, error)) *cobra.Command {
	var asJSON, prune bool
	cmd := &cobra.Command{
		Use:   "claims",
		Short: "List the issues that each workspace has in progress.",
		Long: "List the claims of every jj workspace and git worktree of the repo: the issues that a " +
			"workspace has started, before its work merges. sisyphus start records a claim, and refuses an " +
			"issue that another workspace has claimed. A claim goes away when the issue leaves in-progress " +
			"in its workspace.\n\n" +
			"The status of a claim is here (this workspace), active (another workspace), or stale (the " +
			"workspace no longer has the issue in progress, for example because the workspace was removed). " +
			"sisyphus ignores stale claims, and --prune removes them.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			var claims []claimView
			if prune {
				unlock, err := lockIssues(root)
				if err != nil {
					return err
				}
				claims, err = listClaims(root, true)
				unlock()
				if err != nil {
					return err
				}
			} else if claims, err = listClaims(root, false); err != nil {
				return err
			}
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(claims)
			}
			if len(claims) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "No workspace has claimed an issue.")
				return nil
			}
			store := openClaims(root)
			cells := make([][]string, len(claims))
			for i, c := range claims {
				cells[i] = []string{c.Name, c.Owner, c.Status, store.display(c.Workspace), c.Bookmark, displayTime(c.Since)}
			}
			writeTable(cmd.OutOrStdout(), []string{"name", "owner", "status", "workspace", "bookmark", "since"}, cells)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print JSON instead of a table.")
	cmd.Flags().BoolVar(&prune, "prune", false, "Remove the stale claims, and list the others.")
	return cmd
}
