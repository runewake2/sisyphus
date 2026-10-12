package referendum

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
	"github.com/runewake2/sisyphus/hephaestus/internal/toolagent"
)

// RoleVoter gets the ballots from the agents of the roles. Each reviewer works in a scratch git
// worktree of its own, at the commit of the version, so that it sees the implementation and no other
// ballot. A tool reviewer runs its command, and votes with its exit code
// ([[0003-a-tool-agent-runs-its-command-for-each-message]]). An AI reviewer runs the runner with the
// prompt of its ballot, and ends its output with its ballot as JSON. The changes that a reviewer
// makes in its worktree are its modification.
type RoleVoter struct {
	// GitDir is the git directory that holds the commits: a git repo, or the git store of a jj repo.
	GitDir string
	Roles  map[string]roles.Role
	// Runner runs the model of an AI reviewer, for example "claude -p --agent {role}".
	Runner string
	// Out gets the output of the reviewers, for a person to follow the run.
	Out io.Writer
	// worktrees serializes the changes to the worktrees of the git directory: reviewers run at the
	// same time, and git worktree add and remove of one must not see half of another.
	worktrees sync.Mutex
}

// Ballot asks one reviewer for its ballot on v.
func (rv *RoleVoter) Ballot(ctx context.Context, r *Referendum, reviewer Reviewer, v Version) Ballot {
	b := Ballot{Reviewer: reviewer}
	role, ok := rv.Roles[reviewer.Role]
	if !ok {
		b.Error = "it has no role file"
		return b
	}
	w, err := rv.worktree(v.Commit)
	if err != nil {
		b.Error = err.Error()
		return b
	}
	defer w.remove()
	if role.Kind == roles.Tool {
		ballot, err := toolagent.Review(ctx, role, w.dir, r.BaseID, rv.out())
		if err != nil {
			b.Error = err.Error()
			return b
		}
		b.Vote, b.Reasons = ballot.Vote, ballot.Reasons
		if ballot.Vote == Modify {
			b.Change = &Modification{Summary: fmt.Sprintf("The changes of the command of the tool role %s to %s.", role.Name, strings.Join(ballot.Files, ", "))}
			if b.Change.Commit, b.Change.Diff, err = w.commit(v.Commit, fmt.Sprintf("referendum %s: modification of %s", r.ID, role.Name)); err != nil {
				b.Error = err.Error()
			}
		}
		return b
	}

	output, err := rv.runModel(ctx, role, w.dir, r, BallotPrompt(r, reviewer, role, v))
	if err != nil {
		b.Error = err.Error()
		return b
	}
	var reply struct {
		Vote         string `json:"vote"`
		Reasons      string `json:"reasons"`
		Modification string `json:"modification"`
		Section      string `json:"section"`
		Text         string `json:"text"`
	}
	if !lastJSON(output, &reply, func() bool { return reply.Vote != "" }) {
		b.Error = `its output does not end with a ballot, for example {"vote": "yes", "reasons": "..."}`
		return b
	}
	b.Vote, b.Reasons = strings.ToLower(strings.TrimSpace(reply.Vote)), reply.Reasons
	if b.Vote != Modify {
		return b // a change that it made without a modify ballot is not a modification
	}
	b.Change = &Modification{Summary: reply.Modification}
	if reply.Section != "" {
		b.Change.Edit = &IssueEdit{Section: reply.Section, Text: reply.Text}
	}
	if b.Change.Commit, b.Change.Diff, err = w.commit(v.Commit, fmt.Sprintf("referendum %s: modification of %s\n\n%s", r.ID, role.Name, reply.Modification)); err != nil {
		b.Error = err.Error()
	}
	return b
}

// Gate asks one required reviewer to accept or reject m. It works in a worktree at the modified
// implementation; the changes that it makes there are not kept.
func (rv *RoleVoter) Gate(ctx context.Context, r *Referendum, reviewer Reviewer, v Version, m Modification) GateVote {
	vote := GateVote{Role: reviewer.Role}
	role, ok := rv.Roles[reviewer.Role]
	if !ok {
		vote.Error = "it has no role file"
		return vote
	}
	commit := v.Commit
	if m.Commit != "" {
		commit = m.Commit
	}
	w, err := rv.worktree(commit)
	if err != nil {
		vote.Error = err.Error()
		return vote
	}
	defer w.remove()
	if role.Kind == roles.Tool {
		// A tool judges only what the modification makes worse in its scope: what it finds on the
		// version already is not caused by the modification.
		after, err := toolagent.Review(ctx, role, w.dir, r.BaseID, rv.out())
		if err != nil {
			vote.Error = err.Error()
			return vote
		}
		before := toolagent.Ballot{Vote: Yes}
		if after.Vote != Yes {
			wv, err := rv.worktree(v.Commit)
			if err != nil {
				vote.Error = err.Error()
				return vote
			}
			before, err = toolagent.Review(ctx, role, wv.dir, r.BaseID, rv.out())
			wv.remove()
			if err != nil {
				vote.Error = err.Error()
				return vote
			}
		}
		vote.Accept, vote.Reasons = toolGate(role, before, after)
		return vote
	}
	output, err := rv.runModel(ctx, role, w.dir, r, GatePrompt(r, reviewer, role, v, m))
	if err != nil {
		vote.Error = err.Error()
		return vote
	}
	var reply struct {
		Gate    string `json:"gate"`
		Reasons string `json:"reasons"`
	}
	if !lastJSON(output, &reply, func() bool { return reply.Gate == "accept" || reply.Gate == "reject" }) {
		vote.Error = `its output does not end with a gate vote, for example {"gate": "accept", "reasons": "..."}`
		return vote
	}
	vote.Accept, vote.Reasons = reply.Gate == "accept", reply.Reasons
	return vote
}

// toolGate decides the gate vote of a tool from its ballots on the version (before) and on the
// modified version (after): it rejects only when the modification makes its result worse. A no is
// worse than a modify, and a modify is worse than a yes; a modify is also worse when it would change
// a file that it would not change before.
func toolGate(role roles.Role, before, after toolagent.Ballot) (bool, string) {
	rank := map[string]int{Yes: 0, Modify: 1, No: 2}
	switch {
	case after.Vote == Yes:
		return true, after.Reasons
	case rank[after.Vote] > rank[before.Vote]:
		return false, strings.TrimSpace(fmt.Sprintf("The modification makes the result of %s worse: %s before, %s after. %s", role.Name, before.Vote, after.Vote, after.Reasons))
	case after.Vote == Modify:
		var added []string
		for _, file := range after.Files {
			if !slices.Contains(before.Files, file) {
				added = append(added, file)
			}
		}
		if len(added) > 0 {
			return false, fmt.Sprintf("%s would change %s, which the modification changed.", role.Name, strings.Join(added, ", "))
		}
	}
	return true, fmt.Sprintf("The modification does not make the result of %s worse: %s before and after.", role.Name, after.Vote)
}

// Combine applies the commits of each modification on the ones before it, in a scratch worktree of
// v. A modification whose commits do not apply is left out. One modification needs no work.
func (rv *RoleVoter) Combine(r *Referendum, v Version, mods []Modification) (string, []int, error) {
	if len(mods) == 1 {
		return mods[0].Commit, nil, nil
	}
	w, err := rv.worktree(v.Commit)
	if err != nil {
		return "", nil, err
	}
	defer w.remove()
	var conflicts []int
	for i, m := range mods {
		head, err := rv.git(w.dir, "rev-parse", "HEAD")
		if err != nil {
			return "", nil, err
		}
		if _, err := rv.git(w.dir, "cherry-pick", "--allow-empty", "--keep-redundant-commits", v.Commit+".."+m.Commit); err != nil {
			rv.git(w.dir, "cherry-pick", "--abort")
			if _, err := rv.git(w.dir, "reset", "--quiet", "--hard", head); err != nil {
				return "", nil, err
			}
			conflicts = append(conflicts, i)
		}
	}
	head, err := rv.git(w.dir, "rev-parse", "HEAD")
	return head, conflicts, err
}

// NewVersion keeps the commit of v on the branch referendum/<id>/v<number>.
func (rv *RoleVoter) NewVersion(r *Referendum, v Version) (string, error) {
	ref := fmt.Sprintf("referendum/%s/v%d", r.ID, v.Number)
	if _, err := rv.git("", "branch", "--force", ref, v.Commit); err != nil {
		return "", err
	}
	return ref, nil
}

func (rv *RoleVoter) out() io.Writer {
	if rv.Out == nil {
		return io.Discard
	}
	return rv.Out
}

// runModel runs the runner of an AI reviewer in dir, with the prompt on its standard input, and
// returns its standard output.
func (rv *RoleVoter) runModel(ctx context.Context, role roles.Role, dir string, r *Referendum, prompt string) (string, error) {
	if err := toolagent.Validate(role, rv.Runner); err != nil {
		return "", err
	}
	line := strings.ReplaceAll(rv.Runner, "{role}", role.Flat())
	cmd := toolagent.Shell(ctx, line)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(prompt)
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(&output, rv.out())
	cmd.Stderr = rv.out()
	cmd.Env = append(os.Environ(),
		"SISYPHUS_OWNER="+role.Flat()+"-reviewer",
		"HEPHAESTUS_ROLE="+role.Name,
		"HEPHAESTUS_REVIEW=1",
		"HEPHAESTUS_REFERENDUM="+r.ID,
	)
	err := cmd.Run()
	var exit *exec.ExitError
	switch {
	case errors.As(err, &exit):
		return output.String(), fmt.Errorf("`%s` ended with exit code %d", line, exit.ExitCode())
	case err != nil:
		return output.String(), fmt.Errorf("`%s` could not start: %v", line, err)
	}
	return output.String(), nil
}

// lastJSON reads the last line of output that is a JSON object for which ok returns true. A model
// can write other text before its ballot.
func lastJSON(output string, into any, ok func() bool) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.Trim(strings.TrimSpace(lines[i]), "`")
		if !strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}") {
			continue
		}
		if json.Unmarshal([]byte(line), into) == nil && ok() {
			return true
		}
	}
	return false
}

// git runs git on the git directory, in the worktree dir if it is not empty.
func (rv *RoleVoter) git(dir string, args ...string) (string, error) {
	full := []string{"--git-dir=" + rv.GitDir}
	if dir != "" {
		full = []string{"-C", dir}
	}
	full = append(full, identity()...)
	cmd := exec.Command("git", append(full, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// identity gives git a committer when the user has none, so that a modification never fails for it.
func identity() []string {
	var args []string
	for _, key := range []string{"user.name", "user.email"} {
		if out, err := exec.Command("git", "config", key).Output(); err != nil || strings.TrimSpace(string(out)) == "" {
			args = append(args, "-c", key+"="+map[string]string{"user.name": "hephaestus referendum", "user.email": "hephaestus@localhost"}[key])
		}
	}
	return args
}

// scratch is the scratch worktree of one reviewer.
type scratch struct {
	rv  *RoleVoter
	dir string
}

func (rv *RoleVoter) worktree(commit string) (*scratch, error) {
	dir, err := os.MkdirTemp("", "hephaestus-referendum-")
	if err != nil {
		return nil, err
	}
	os.Remove(dir) // git worktree add makes it
	rv.worktrees.Lock()
	defer rv.worktrees.Unlock()
	if _, err := rv.git("", "worktree", "add", "--detach", "--quiet", dir, commit); err != nil {
		return nil, err
	}
	return &scratch{rv: rv, dir: dir}, nil
}

func (s *scratch) remove() {
	s.rv.worktrees.Lock()
	defer s.rv.worktrees.Unlock()
	if _, err := s.rv.git("", "worktree", "remove", "--force", s.dir); err != nil {
		os.RemoveAll(s.dir)
		s.rv.git("", "worktree", "prune")
	}
}

// commit commits the changes of the worktree, if any, and returns the head and the diff from base.
// With no commit and no change, it returns no head.
func (s *scratch) commit(base, message string) (string, string, error) {
	status, err := s.rv.git(s.dir, "status", "--porcelain")
	if err != nil {
		return "", "", err
	}
	if status != "" {
		if _, err := s.rv.git(s.dir, "add", "-A"); err != nil {
			return "", "", err
		}
		if _, err := s.rv.git(s.dir, "commit", "--quiet", "-m", message); err != nil {
			return "", "", err
		}
	}
	head, err := s.rv.git(s.dir, "rev-parse", "HEAD")
	if err != nil || head == base {
		return "", "", err
	}
	if merges, _ := s.rv.git(s.dir, "rev-list", "--merges", base+"..HEAD"); merges != "" {
		return "", "", errors.New("its modification has merge commits; a modification is a linear line of commits on the version")
	}
	if ancestors, _ := s.rv.git(s.dir, "rev-list", "HEAD.."+base); ancestors != "" {
		return "", "", errors.New("its modification does not start at the commit of the version")
	}
	diff, err := s.rv.git(s.dir, "diff", base, "HEAD")
	return head, diff, err
}

// RoleMap returns the roles of the reviewers by name.
func RoleMap(dir string, reviewers []Reviewer) (map[string]roles.Role, error) {
	result := map[string]roles.Role{}
	var problems []string
	for _, reviewer := range reviewers {
		role, err := roles.Find(dir, reviewer.Role)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		result[reviewer.Role] = role
	}
	if len(problems) > 0 {
		slices.Sort(problems)
		return nil, errors.New(strings.Join(problems, "\n"))
	}
	return result, nil
}
