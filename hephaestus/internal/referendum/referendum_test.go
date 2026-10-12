package referendum

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
)

// script is a fake Voter: the ballot of each reviewer in each version, and the gate vote of each
// required reviewer on the modification of each author.
type script struct {
	mu      sync.Mutex
	ballots map[string]Ballot   // "<role>@v<n>" or "<role>@v<n>#<call>"
	gates   map[string]GateVote // "<role>:<author>"
	calls   map[string]int
	// barrier, when set, makes each ballot wait until every reviewer of the round has been asked.
	barrier *sync.WaitGroup
	// conflicts are the authors whose commits do not apply on the ones before them.
	conflicts map[string]bool
}

func (s *script) Ballot(ctx context.Context, r *Referendum, reviewer Reviewer, v Version) Ballot {
	if s.barrier != nil {
		s.barrier.Done()
		done := make(chan struct{})
		go func() { s.barrier.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			return Ballot{Error: "the reviewers did not run at the same time"}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%s@v%d", reviewer.Role, v.Number)
	s.calls[key]++
	if b, ok := s.ballots[fmt.Sprintf("%s#%d", key, s.calls[key])]; ok {
		return b
	}
	if b, ok := s.ballots[key]; ok {
		return b
	}
	return Ballot{Vote: Yes}
}

func (s *script) Gate(ctx context.Context, r *Referendum, reviewer Reviewer, v Version, m Modification) GateVote {
	if g, ok := s.gates[reviewer.Role+":"+m.Author]; ok {
		return g
	}
	return GateVote{Accept: true}
}

func (s *script) Combine(r *Referendum, v Version, mods []Modification) (string, []int, error) {
	if len(mods) == 1 {
		return mods[0].Commit, nil, nil // as RoleVoter: one modification needs no work
	}
	head := v.Commit
	var conflicts []int
	for i, m := range mods {
		if s.conflicts[m.Author] {
			conflicts = append(conflicts, i)
			continue
		}
		head += "+" + m.Commit
	}
	return head, conflicts, nil
}

func (s *script) NewVersion(r *Referendum, v Version) (string, error) {
	return fmt.Sprintf("referendum/%s/v%d", r.ID, v.Number), nil
}

func newScript() *script {
	return &script{ballots: map[string]Ballot{}, gates: map[string]GateVote{}, calls: map[string]int{}}
}

func newReferendum(reviewers ...Reviewer) *Referendum {
	return &Referendum{ID: "2026-10-12-fix-login-bug", Issue: "fix-login-bug", Question: "Merge it?", Bookmark: "ai/fix", Reviewers: reviewers, Versions: []Version{{Number: 1, Commit: "c1"}}}
}

var (
	engineer = Reviewer{Role: "engineer", Required: true}
	docs     = Reviewer{Role: "technical-docs", Required: true}
	ux       = Reviewer{Role: "ux-researcher"}
)

func run(t *testing.T, r *Referendum, s *script) []string {
	t.Helper()
	var events []string
	err := Run(context.Background(), r, s, Options{On: func(e Event) {
		if e.Round != nil {
			events = append(events, fmt.Sprintf("round %d", e.Round.Number))
		} else {
			events = append(events, e.Kind)
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func equal[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func contains(t *testing.T, s, substring string) {
	t.Helper()
	if !strings.Contains(s, substring) {
		t.Errorf("want %q to contain %q", s, substring)
	}
}

// The example of [[referendum#Example]].
func TestTheExampleOfTheDesignPassesWithTwoConcerns(t *testing.T) {
	s := newScript()
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Reasons: "The flag is not documented.", Change: &Modification{Summary: "Add --reset to docs/login.md.", Commit: "c2"}}
	s.ballots["ux-researcher@v1"] = Ballot{Vote: No, Reasons: "The error message is unclear."}
	s.ballots["ux-researcher@v2"] = Ballot{Vote: Modify, Reasons: "Wrong flag.", Change: &Modification{Summary: "Name --reset in the message.", Commit: "c3"}}
	s.gates["engineer:ux-researcher"] = GateVote{Accept: false, Reasons: "The message is in the scope of another issue."}
	r := newReferendum(engineer, docs, ux)

	events := run(t, r, s)

	equal(t, "round 1,round 2,result", strings.Join(events, ","))
	equal(t, Passed, r.Result)
	equal(t, 2, len(r.Versions))
	equal(t, "c2", r.Current().Commit)
	equal(t, "technical-docs", r.Current().Modifications[0].Author)
	equal(t, "referendum/2026-10-12-fix-login-bug/v2", r.Current().Ref)
	equal(t, 1, s.calls["engineer@v2"]) // engineer votes again, because the contents changed
	equal(t, 2, len(r.Concerns))
	contains(t, r.Concerns[0], "ux-researcher (round 1): The error message is unclear.")
	contains(t, r.Concerns[1], "ux-researcher (round 2): the gate rejected its modification: Name --reset in the message.")
	equal(t, false, r.Rounds[1].Gates[0].Accepted)
}

func TestOneRequiredNoFailsTheReferendum(t *testing.T) {
	s := newScript()
	s.ballots["engineer@v1"] = Ballot{Vote: No, Reasons: "It undoes the session cache."}
	r := newReferendum(engineer, docs)

	run(t, r, s)

	equal(t, Failed, r.Result)
	contains(t, r.Reasons, "engineer: It undoes the session cache.")
}

func TestANonRequiredNoIsOnlyAConcern(t *testing.T) {
	s := newScript()
	s.ballots["ux-researcher@v1"] = Ballot{Vote: No, Reasons: "Unclear."}
	r := newReferendum(engineer, ux)

	run(t, r, s)

	equal(t, Passed, r.Result)
	equal(t, 1, len(r.Concerns))
}

func TestAFailedVoteOnAnEditedVersionFailsTheWholeReferendum(t *testing.T) {
	s := newScript()
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Docs.", Commit: "c2"}}
	s.ballots["engineer@v2"] = Ballot{Vote: No, Reasons: "The docs change breaks the build."}
	r := newReferendum(engineer, docs)

	run(t, r, s)

	equal(t, Failed, r.Result)
	equal(t, 2, len(r.Rounds))
	contains(t, r.Reasons, "on v2, the edited version, so the whole referendum fails")
}

func TestARequiredReviewerWhoseModificationIsRejectedVotesAgain(t *testing.T) {
	s := newScript()
	s.ballots["technical-docs@v1#1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Docs.", Commit: "c2"}}
	s.gates["engineer:technical-docs"] = GateVote{Accept: false, Reasons: "Out of scope."}
	r := newReferendum(engineer, docs)

	run(t, r, s)

	equal(t, Passed, r.Result)
	equal(t, 2, len(r.Rounds))
	equal(t, 1, r.Rounds[1].Version) // the same version (R8)
	equal(t, 2, s.calls["technical-docs@v1"])
	contains(t, r.Rounds[0].Outcome, "The gate rejected the modifications of technical-docs. They vote again on v1.")
}

func TestAfterTheMaximumRoundsAHumanDecides(t *testing.T) {
	s := newScript()
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Docs.", Commit: "c2"}}
	s.gates["engineer:technical-docs"] = GateVote{Accept: false}
	r := newReferendum(engineer, docs)

	run(t, r, s)

	equal(t, NoResult, r.Result)
	equal(t, DefaultMaxRounds, len(r.Rounds))
	contains(t, r.Reasons, "no result after 3 rounds. A human decides.")
}

func TestARequiredReviewerCannotAbstain(t *testing.T) {
	s := newScript()
	s.ballots["engineer@v1"] = Ballot{Vote: Abstain}
	s.ballots["ux-researcher@v1"] = Ballot{Vote: Abstain}
	r := newReferendum(engineer, ux)

	run(t, r, s)

	equal(t, Failed, r.Result)
	equal(t, No, r.Rounds[0].Ballots[0].Vote)
	equal(t, Abstain, r.Rounds[0].Ballots[1].Vote)
}

func TestAReviewerWithNoValidBallotGivesNoResult(t *testing.T) {
	s := newScript()
	s.ballots["technical-docs@v1"] = Ballot{Error: "its output does not end with a ballot"}
	r := newReferendum(engineer, docs)

	run(t, r, s)

	equal(t, NoResult, r.Result)
	contains(t, r.Reasons, "technical-docs: its output does not end with a ballot")
}

func TestEveryModificationOfARoundGoesToTheGateBeforeTheNextRound(t *testing.T) {
	s := newScript()
	s.ballots["ux-researcher@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Message.", Commit: "c-ux"}}
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Docs.", Commit: "c-docs"}}
	s.ballots["engineer@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Rename.", Commit: "c-eng"}}
	s.gates["technical-docs:engineer"] = GateVote{Accept: false, Reasons: "Out of scope."}
	r := newReferendum(engineer, docs, ux)

	run(t, r, s)

	// Each modification went to the gate alone, required reviewers first; the accepted ones together
	// made v2, and round 2 voted on it.
	first := r.Rounds[0]
	equal(t, 3, len(first.Gates))
	equal(t, "engineer", first.Gates[0].Modification.Author)
	equal(t, false, first.Gates[0].Accepted)
	equal(t, "technical-docs", first.Gates[1].Modification.Author)
	equal(t, "ux-researcher", first.Gates[2].Modification.Author)
	equal(t, 0, len(first.Returned))
	equal(t, 2, len(r.Versions[1].Modifications))
	equal(t, "c1+c-docs+c-ux", r.Versions[1].Commit)
	equal(t, 2, len(r.Rounds))
	equal(t, 2, r.Rounds[1].Version)
	equal(t, Passed, r.Result)
	contains(t, first.Outcome, "The gate accepted the modifications of technical-docs, ux-researcher: v2.")
}

func TestAnAcceptedModificationThatConflictsGoesBackToItsAuthor(t *testing.T) {
	s := newScript()
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Docs.", Commit: "c-docs"}}
	s.ballots["ux-researcher@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "Message.", Commit: "c-ux"}}
	s.conflicts = map[string]bool{"ux-researcher": true}
	r := newReferendum(engineer, docs, ux)

	run(t, r, s)

	equal(t, "ux-researcher", strings.Join(r.Rounds[0].Returned, ","))
	equal(t, "c1+c-docs", r.Versions[1].Commit) // two went to Combine; one conflicted
	equal(t, 1, len(r.Versions[1].Modifications))
	contains(t, strings.Join(r.Concerns, "\n"), "ux-researcher (round 1): the gate accepted its modification, but it conflicts with an earlier one")
	equal(t, Passed, r.Result)
}

func TestTwoEditsOfOneIssueSectionConflict(t *testing.T) {
	s := newScript()
	s.ballots["engineer@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "A.", Edit: &IssueEdit{Section: "Summary", Text: "A"}}}
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "B.", Edit: &IssueEdit{Section: "summary", Text: "B"}}}
	r := newReferendum(engineer, docs)

	run(t, r, s)

	equal(t, "technical-docs", strings.Join(r.Rounds[0].Returned, ","))
	equal(t, "A", r.Versions[1].IssueEdits[0].Text)
	equal(t, 1, len(r.Versions[1].IssueEdits))
}

func TestAModifyBallotWithNoModificationIsANo(t *testing.T) {
	s := newScript()
	s.ballots["engineer@v1"] = Ballot{Vote: Modify, Reasons: "Rename it."}
	r := newReferendum(engineer)

	run(t, r, s)

	equal(t, Failed, r.Result)
	contains(t, r.Rounds[0].Ballots[0].Reasons, "It voted modify, but gave no modification")
}

func TestAnIssueEditIsPartOfTheNextVersion(t *testing.T) {
	s := newScript()
	edit := &IssueEdit{Section: "Acceptance criteria", Text: "- [ ] The flag is documented."}
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Change: &Modification{Summary: "A criterion.", Edit: edit}}
	r := newReferendum(engineer, docs)

	run(t, r, s)

	equal(t, Passed, r.Result)
	equal(t, "c1", r.Current().Commit) // the implementation did not change
	equal(t, 1, len(r.Current().IssueEdits))
	contains(t, BallotPrompt(r, engineer, testRole(), r.Current()), "- [ ] The flag is documented.")
}

func TestTheReviewersOfARoundRunAtTheSameTime(t *testing.T) {
	s := newScript()
	s.barrier = &sync.WaitGroup{}
	s.barrier.Add(3)
	r := newReferendum(engineer, docs, ux)

	run(t, r, s)

	equal(t, Passed, r.Result)
}

func TestTheReviewersNeedARequiredOne(t *testing.T) {
	err := Check([]Reviewer{ux})
	contains(t, fmt.Sprint(err), "at least one required reviewer")
	err = Check([]Reviewer{engineer, {Role: "engineer"}})
	contains(t, fmt.Sprint(err), "The role engineer is a reviewer more than once.")
}

func TestTheRecordHasEveryRoundAndGate(t *testing.T) {
	s := newScript()
	s.ballots["technical-docs@v1"] = Ballot{Vote: Modify, Reasons: "No docs.", Change: &Modification{Summary: "Add the docs.", Commit: "c2", Diff: "+the docs"}}
	s.ballots["ux-researcher@v1"] = Ballot{Vote: No, Reasons: "Unclear | message."}
	r := newReferendum(engineer, docs, ux)
	r.Date, r.CalledBy, r.Base, r.BaseID = "2026-10-12", "a human", "main", "b1"

	run(t, r, s)
	record := Markdown(r)

	for _, part := range []string{
		"# Referendum: fix-login-bug",
		"- Issue: [[fix-login-bug]]",
		"- Result: **passed**",
		"| `ux-researcher` | non-required |",
		"### v2\n\n- Commit: `c2` (`referendum/2026-10-12-fix-login-bug/v2`)",
		"````diff\n+the docs\n````",
		"## Round 1: v1",
		"| `ux-researcher` | non-required | no | Unclear \\| message. |",
		"### Gate 1.1: the modification of `technical-docs`, accepted",
		"| `engineer` | accept |",
		"## Round 2: v2",
		"**passed**: Every required reviewer voted yes on v2.",
		"- ux-researcher (round 1): Unclear | message.",
	} {
		contains(t, record, part)
	}
	note := RoundNote(r, r.Rounds[0], r.ID)
	contains(t, note, "2026-10-12, referendum round 1 on v1: engineer (required) yes, technical-docs (required) modify, ux-researcher (non-required) no. The gate accepted the modification of technical-docs.")
	contains(t, note, "Record: [[2026-10-12-fix-login-bug]].")
	contains(t, ResultNote(r, r.ID), "The accepted implementation is v2, the bookmark `referendum/2026-10-12-fix-login-bug/v2`.")
}

func testRole() roles.Role {
	return roles.Role{Name: "engineer", Title: "Engineer", Kind: roles.AI}
}
