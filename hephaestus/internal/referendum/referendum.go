// Package referendum runs a referendum ([[referendum]]): a series of reviewers, each an agent of one
// role, vote on an implementation in rounds. Required reviewers must all vote yes to pass it;
// non-required reviewers only give advice. A modification goes on only when every required reviewer
// accepts it at the gate; it then makes a new version, and every reviewer votes again on it.
//
// This file holds the procedure. Where the ballots come from (a model or a tool, in a scratch
// worktree) is the job of a Voter, and the record in the repo is the job of the caller.
package referendum

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Votes of a ballot.
const (
	Yes     = "yes"
	No      = "no"
	Modify  = "modify"
	Abstain = "abstain"
)

// Results of a referendum.
const (
	Passed   = "passed"
	Failed   = "failed"
	NoResult = "no-result"
)

// DefaultMaxRounds is the maximum number of rounds (R5 of [[referendum]]).
const DefaultMaxRounds = 3

// Reviewer is one reviewer of a referendum: a role, required or not.
type Reviewer struct {
	Role     string `json:"role"`
	Required bool   `json:"required"`
}

// Kind returns "required" or "non-required".
func (r Reviewer) Kind() string {
	if r.Required {
		return "required"
	}
	return "non-required"
}

// IssueEdit is a modification of the text of one section of the issue.
type IssueEdit struct {
	Section string `json:"section"`
	Text    string `json:"text"`
}

// Modification is a change of the contents that a reviewer submits with a modify ballot: new commits
// on the implementation, a new text for a section of the issue, or both.
type Modification struct {
	Author  string `json:"author"`
	Summary string `json:"summary"`
	// Commit is the head of the modified implementation. Empty: the implementation does not change.
	Commit string `json:"commit,omitempty"`
	// Diff is the change from the implementation of the version to Commit.
	Diff string     `json:"diff,omitempty"`
	Edit *IssueEdit `json:"issue-edit,omitempty"`
}

// Version is one version of the contents: the implementation, and the edits of the issue text that
// the gate accepted.
type Version struct {
	Number int `json:"number"`
	// Commit is the head of the implementation.
	Commit string `json:"commit"`
	// Ref is the git branch that keeps the commit, for example referendum/<id>/v2.
	Ref string `json:"ref,omitempty"`
	// Modifications are what made this version from the one before it, in the order that they were
	// applied. v1 has none.
	Modifications []Modification `json:"modifications,omitempty"`
	IssueEdits    []IssueEdit    `json:"issue-edits,omitempty"`
}

// Ballot is the vote of one reviewer in one round.
type Ballot struct {
	Reviewer
	Vote    string        `json:"vote"`
	Reasons string        `json:"reasons,omitempty"`
	Change  *Modification `json:"modification,omitempty"`
	// Error tells why the reviewer gave no valid ballot. The referendum then has no result.
	Error string `json:"error,omitempty"`
}

// GateVote is the vote of one required reviewer on one modification.
type GateVote struct {
	Role    string `json:"role"`
	Accept  bool   `json:"accept"`
	Reasons string `json:"reasons,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Gate is the vote of the required reviewers on one modification.
type Gate struct {
	Modification Modification `json:"modification"`
	Votes        []GateVote   `json:"votes"`
	Accepted     bool         `json:"accepted"`
}

// Round is one vote of every reviewer on one version.
type Round struct {
	Number  int      `json:"number"`
	Version int      `json:"version"`
	Ballots []Ballot `json:"ballots"`
	Gates   []Gate   `json:"gates,omitempty"`
	// Returned are the authors of modifications that the gate accepted, but that conflict with an
	// earlier accepted one of the round. They can submit them again against the new version.
	Returned []string `json:"returned,omitempty"`
	// Outcome tells what the round led to, in one sentence.
	Outcome string `json:"outcome"`
}

// Referendum is one referendum, from the call to the result.
type Referendum struct {
	ID        string     `json:"id"`
	Date      string     `json:"date"`
	Issue     string     `json:"issue"`
	Question  string     `json:"question"`
	CalledBy  string     `json:"called-by"`
	Bookmark  string     `json:"bookmark"`
	Base      string     `json:"base"`
	BaseID    string     `json:"base-id"`
	Reviewers []Reviewer `json:"reviewers"`
	MaxRounds int        `json:"max-rounds"`
	Versions  []Version  `json:"versions"`
	Rounds    []Round    `json:"rounds"`
	// Result is passed, failed, or no-result; empty while the referendum runs.
	Result string `json:"result,omitempty"`
	// Reasons tells why the referendum has its result.
	Reasons string `json:"reasons,omitempty"`
	// Concerns are the no ballots and the rejected modifications of non-required reviewers.
	Concerns []string `json:"concerns,omitempty"`
}

// Current returns the last version.
func (r *Referendum) Current() Version {
	return r.Versions[len(r.Versions)-1]
}

// Required returns the required reviewers.
func (r *Referendum) Required() []Reviewer {
	var required []Reviewer
	for _, reviewer := range r.Reviewers {
		if reviewer.Required {
			required = append(required, reviewer)
		}
	}
	return required
}

// Voter gets the ballots. A Voter must not show a reviewer the ballots of the other reviewers of the
// same round: Run asks for them all before it gives any to the caller.
type Voter interface {
	// Ballot asks one reviewer for its ballot on a version.
	Ballot(ctx context.Context, r *Referendum, reviewer Reviewer, v Version) Ballot
	// Gate asks one required reviewer to accept or reject a modification of a version.
	Gate(ctx context.Context, r *Referendum, reviewer Reviewer, v Version, m Modification) GateVote
	// Combine applies the implementation changes of modifications of v in order, each on the ones
	// before it, and returns the head and the indexes of the modifications that conflict and are left
	// out. Each modification has a Commit.
	Combine(r *Referendum, v Version, mods []Modification) (string, []int, error)
	// NewVersion keeps the commit of a new version, and returns its ref.
	NewVersion(r *Referendum, v Version) (string, error)
}

// Event tells the caller what happened, so that it can record it: "round" after each round, with
// its gates, and "result" at the end.
type Event struct {
	Kind  string
	Round *Round
}

// Options are the settings of a run.
type Options struct {
	// Parallel is the maximum number of reviewers that run at the same time. 0: all of them.
	Parallel int
	// On gets each event. It can be nil.
	On func(Event)
}

// Check returns an error when the reviewers of a referendum are not valid.
func Check(reviewers []Reviewer) error {
	seen := map[string]bool{}
	required := 0
	for _, reviewer := range reviewers {
		if seen[reviewer.Role] {
			return fmt.Errorf("The role %s is a reviewer more than once. A role is a required or a non-required reviewer, not both.", reviewer.Role)
		}
		seen[reviewer.Role] = true
		if reviewer.Required {
			required++
		}
	}
	if required == 0 {
		return fmt.Errorf("A referendum needs at least one required reviewer: give --required.")
	}
	return nil
}

// Run runs the rounds of the referendum until it has a result. r must have its reviewers and v1.
func Run(ctx context.Context, r *Referendum, voter Voter, o Options) error {
	if err := Check(r.Reviewers); err != nil {
		return err
	}
	if len(r.Versions) == 0 {
		return fmt.Errorf("The referendum has no contents to vote on.")
	}
	if r.MaxRounds <= 0 {
		r.MaxRounds = DefaultMaxRounds
	}
	emit := func(e Event) {
		if o.On != nil {
			o.On(e)
		}
	}
	finish := func(result, reasons string) error {
		r.Result, r.Reasons = result, reasons
		emit(Event{Kind: "result"})
		return nil
	}
	for number := 1; ; number++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if number > r.MaxRounds {
			return finish(NoResult, fmt.Sprintf("The referendum had no result after %d rounds. A human decides.", r.MaxRounds))
		}
		version := r.Current()
		round := Round{Number: number, Version: version.Number}
		round.Ballots = ballots(ctx, r, voter, version, o.Parallel)
		r.Rounds = append(r.Rounds, round)
		current := &r.Rounds[len(r.Rounds)-1]
		if ctx.Err() != nil {
			return ctx.Err()
		}

		var missing []string
		for _, b := range current.Ballots {
			if b.Error != "" {
				missing = append(missing, fmt.Sprintf("%s: %s", b.Role, b.Error))
			}
		}
		if len(missing) > 0 {
			current.Outcome = "Not every reviewer gave a valid ballot."
			emit(Event{Kind: "round", Round: current})
			return finish(NoResult, "Not every reviewer gave a valid ballot, so a human decides. "+strings.Join(missing, " "))
		}

		// Every modification of the round goes to the gate, each alone and against this version, in
		// the order of the reviewers (R2). The next round starts only after every one is accepted or
		// rejected.
		required := map[string]bool{}
		var accepted []Modification
		for _, b := range current.Ballots {
			required[b.Role] = b.Required
			if b.Vote != Modify || b.Change == nil {
				continue
			}
			gate := Gate{Modification: *b.Change, Votes: gateVotes(ctx, r, voter, version, *b.Change, o.Parallel)}
			gate.Accepted = true
			var gateErrors []string
			for _, v := range gate.Votes {
				if v.Error != "" {
					gateErrors = append(gateErrors, fmt.Sprintf("%s: %s", v.Role, v.Error))
				}
				gate.Accepted = gate.Accepted && v.Accept && v.Error == ""
			}
			current.Gates = append(current.Gates, gate)
			if len(gateErrors) > 0 {
				current.Outcome = "Not every required reviewer gave a valid vote at the gate."
				emit(Event{Kind: "round", Round: current})
				return finish(NoResult, "Not every required reviewer gave a valid vote at the gate, so a human decides. "+strings.Join(gateErrors, " "))
			}
			if gate.Accepted {
				accepted = append(accepted, gate.Modification)
			} else if !b.Required {
				r.Concerns = append(r.Concerns, fmt.Sprintf("%s (round %d): the gate rejected its modification: %s", b.Role, number, oneLine(b.Change.Summary)))
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Concerns of the non-required reviewers.
		for _, b := range current.Ballots {
			if !b.Required && b.Vote == No {
				r.Concerns = append(r.Concerns, fmt.Sprintf("%s (round %d): %s", b.Role, number, oneLine(b.Reasons)))
			}
		}

		// The accepted modifications make one new version together, and every reviewer votes on it.
		if len(accepted) > 0 {
			next, conflicts, err := combine(r, voter, version, accepted)
			if err != nil {
				return err
			}
			for _, m := range conflicts {
				current.Returned = append(current.Returned, m.Author)
				if !required[m.Author] {
					r.Concerns = append(r.Concerns, fmt.Sprintf("%s (round %d): the gate accepted its modification, but it conflicts with an earlier one: %s", m.Author, number, oneLine(m.Summary)))
				}
			}
			if next != nil {
				ref, err := voter.NewVersion(r, *next)
				if err != nil {
					return err
				}
				next.Ref = ref
				r.Versions = append(r.Versions, *next)
				current.Outcome = fmt.Sprintf("The gate accepted the %s of %s: v%d. Every reviewer votes again on it.", plural(len(next.Modifications), "modification", "modifications"), authors(next.Modifications), next.Number)
				emit(Event{Kind: "round", Round: current})
				continue
			}
		}

		// The count (R1): every required yes passes, one required no fails.
		var no, notYes []string
		for _, b := range current.Ballots {
			if !b.Required {
				continue
			}
			switch b.Vote {
			case Yes:
			case No:
				no = append(no, fmt.Sprintf("%s: %s", b.Role, oneLine(b.Reasons)))
			default:
				notYes = append(notYes, b.Role)
			}
		}
		switch {
		case len(no) > 0:
			current.Outcome = "A required reviewer voted no."
			emit(Event{Kind: "round", Round: current})
			reasons := "A required reviewer voted no. " + strings.Join(no, " ")
			if version.Number > 1 {
				// R7: a failed vote on an edited version fails the whole referendum.
				reasons = fmt.Sprintf("A required reviewer voted no on v%d, the edited version, so the whole referendum fails. ", version.Number) + strings.Join(no, " ")
			}
			return finish(Failed, reasons)
		case len(notYes) == 0:
			current.Outcome = "Every required reviewer voted yes."
			emit(Event{Kind: "round", Round: current})
			return finish(Passed, fmt.Sprintf("Every required reviewer voted yes on v%d.", version.Number))
		default:
			// R8: a required reviewer whose modification the gate rejected votes again.
			current.Outcome = fmt.Sprintf("The gate rejected the modifications of %s. They vote again on v%d.", strings.Join(notYes, ", "), version.Number)
			emit(Event{Kind: "round", Round: current})
		}
	}
}

// combine makes the next version from the accepted modifications of v, in order. A modification that
// conflicts with an earlier one, in the implementation or in the same section of the issue, is left
// out and returned. With no modification left, there is no next version.
func combine(r *Referendum, voter Voter, v Version, accepted []Modification) (*Version, []Modification, error) {
	var conflicts, kept []Modification
	sections := map[string]bool{}
	for _, m := range accepted {
		if m.Edit != nil {
			key := strings.ToLower(strings.TrimSpace(m.Edit.Section))
			if sections[key] {
				conflicts = append(conflicts, m)
				continue
			}
			sections[key] = true
		}
		kept = append(kept, m)
	}
	var withCommit []Modification
	var at []int
	for i, m := range kept {
		if m.Commit != "" {
			withCommit = append(withCommit, m)
			at = append(at, i)
		}
	}
	head := v.Commit
	left := map[int]bool{}
	if len(withCommit) > 0 {
		var failed []int
		var err error
		if head, failed, err = voter.Combine(r, v, withCommit); err != nil {
			return nil, nil, err
		}
		for _, i := range failed {
			left[at[i]] = true
		}
	}
	next := Version{Number: v.Number + 1, Commit: head, IssueEdits: append([]IssueEdit(nil), v.IssueEdits...)}
	for i, m := range kept {
		if left[i] {
			conflicts = append(conflicts, m)
			continue
		}
		next.Modifications = append(next.Modifications, m)
		if m.Edit != nil {
			next.IssueEdits = append(next.IssueEdits, *m.Edit)
		}
	}
	if len(next.Modifications) == 0 {
		return nil, conflicts, nil
	}
	return &next, conflicts, nil
}

func authors(mods []Modification) string {
	var names []string
	for _, m := range mods {
		names = append(names, m.Author)
	}
	return strings.Join(names, ", ")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ballots asks every reviewer for its ballot on v, at the same time, and returns them in the order
// of the reviewers. A reviewer sees no other ballot of the round.
func ballots(ctx context.Context, r *Referendum, voter Voter, v Version, parallel int) []Ballot {
	result := make([]Ballot, len(r.Reviewers))
	each(len(r.Reviewers), parallel, func(i int) {
		reviewer := r.Reviewers[i]
		b := voter.Ballot(ctx, r, reviewer, v)
		b.Reviewer = reviewer
		result[i] = normalize(b)
	})
	return result
}

// normalize applies the rules of a ballot: only a non-required reviewer abstains, and a modify
// ballot needs a modification.
func normalize(b Ballot) Ballot {
	if b.Error != "" {
		return b
	}
	switch b.Vote {
	case Yes, No:
		b.Change = nil
	case Abstain:
		b.Change = nil
		if b.Required {
			b.Vote = No
			b.Reasons = strings.TrimSpace("A required reviewer cannot abstain, so its abstain counts as no. " + b.Reasons)
		}
	case Modify:
		if b.Change == nil || (b.Change.Commit == "" && b.Change.Edit == nil) {
			b.Vote, b.Change = No, nil
			b.Reasons = strings.TrimSpace("It voted modify, but gave no modification: no change of the implementation and no issue edit. " + b.Reasons)
			break
		}
		b.Change.Author = b.Role
		if b.Change.Summary == "" {
			b.Change.Summary = b.Reasons
		}
	default:
		b.Error = fmt.Sprintf("the vote %q is not yes, no, modify, or abstain", b.Vote)
	}
	return b
}

// gateVotes asks every required reviewer to accept or reject m.
func gateVotes(ctx context.Context, r *Referendum, voter Voter, v Version, m Modification, parallel int) []GateVote {
	required := r.Required()
	result := make([]GateVote, len(required))
	each(len(required), parallel, func(i int) {
		vote := voter.Gate(ctx, r, required[i], v, m)
		vote.Role = required[i].Role
		result[i] = vote
	})
	return result
}

// each runs f for 0..n-1, at most parallel at the same time (0: all of them).
func each(n, parallel int, f func(int)) {
	if parallel <= 0 || parallel > n {
		parallel = n
	}
	slots := make(chan struct{}, max(parallel, 1))
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		slots <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-slots }()
			f(i)
		}(i)
	}
	wg.Wait()
}

// oneLine joins the lines of text, for a Notes line or a concern.
func oneLine(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return "(no reasons)"
	}
	return text
}
