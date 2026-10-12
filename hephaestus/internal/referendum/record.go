package referendum

import (
	"fmt"
	"path"
	"strings"
)

// Dir is the directory of the records of referendums, from the root of the repo (R3 of
// [[referendum]]).
const Dir = "referendums"

// diffLines is the maximum number of lines of a diff in a record.
const diffLines = 200

// IssueLink returns the wikilink of the issue: its file name, which is unique in the tracker.
func IssueLink(issue string) string {
	return "[[" + path.Base(issue) + "]]"
}

// Markdown returns the full record of the referendum: the call, the reviewers, each version, each
// round with its ballots and gates, and the result.
func Markdown(r *Referendum) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Referendum: %s\n\n", path.Base(r.Issue))
	fmt.Fprintf(&b, "- Date: %s\n- Issue: %s\n- Called by: %s\n", r.Date, IssueLink(r.Issue), r.CalledBy)
	fmt.Fprintf(&b, "- Implementation: `%s`, based on `%s` (`%.12s`)\n", r.Bookmark, r.Base, r.BaseID)
	fmt.Fprintf(&b, "- Maximum rounds: %d\n", r.MaxRounds)
	result := r.Result
	if result == "" {
		result = "running"
	}
	fmt.Fprintf(&b, "- Result: **%s**\n\n", result)
	fmt.Fprintf(&b, "This is the historical record of one hephaestus referendum. hephaestus writes it while the referendum runs. Do not change it after the result.\n\n")
	fmt.Fprintf(&b, "## Question\n\n%s\n\n", r.Question)

	b.WriteString("## Reviewers\n\n| Role | Kind |\n|---|---|\n")
	for _, reviewer := range r.Reviewers {
		fmt.Fprintf(&b, "| `%s` | %s |\n", reviewer.Role, reviewer.Kind())
	}

	b.WriteString("\n## Versions\n")
	for _, v := range r.Versions {
		fmt.Fprintf(&b, "\n### v%d\n\n- Commit: `%s`", v.Number, v.Commit)
		if v.Ref != "" {
			fmt.Fprintf(&b, " (`%s`)", v.Ref)
		}
		b.WriteString("\n")
		if len(v.Modifications) == 0 {
			fmt.Fprintf(&b, "- The bookmark `%s` as it was called.\n", r.Bookmark)
			continue
		}
		for _, m := range v.Modifications {
			fmt.Fprintf(&b, "- The modification of `%s`: %s\n", m.Author, oneLine(m.Summary))
			modification(&b, m)
		}
	}

	for _, round := range r.Rounds {
		fmt.Fprintf(&b, "\n## Round %d: v%d\n\n| Reviewer | Kind | Vote | Reasons |\n|---|---|---|---|\n", round.Number, round.Version)
		for _, ballot := range round.Ballots {
			vote, reasons := ballot.Vote, ballot.Reasons
			if ballot.Error != "" {
				vote, reasons = "(no ballot)", ballot.Error
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", ballot.Role, ballot.Kind(), vote, cell(reasons))
		}
		for i, gate := range round.Gates {
			verdict := "rejected"
			if gate.Accepted {
				verdict = "accepted"
			}
			fmt.Fprintf(&b, "\n### Gate %d.%d: the modification of `%s`, %s\n\n%s\n", round.Number, i+1, gate.Modification.Author, verdict, oneLine(gate.Modification.Summary))
			if !gate.Accepted {
				modification(&b, gate.Modification)
			}
			b.WriteString("\n| Required reviewer | Vote | Reasons |\n|---|---|---|\n")
			for _, vote := range gate.Votes {
				verdict, reasons := "reject", vote.Reasons
				if vote.Accept {
					verdict = "accept"
				}
				if vote.Error != "" {
					verdict, reasons = "(no vote)", vote.Error
				}
				fmt.Fprintf(&b, "| `%s` | %s | %s |\n", vote.Role, verdict, cell(reasons))
			}
		}
		if len(round.Returned) > 0 {
			fmt.Fprintf(&b, "\nAccepted modifications that conflict with an earlier one, and went back to their authors: %s.\n", strings.Join(round.Returned, ", "))
		}
		fmt.Fprintf(&b, "\nOutcome: %s\n", round.Outcome)
	}

	if r.Result != "" {
		fmt.Fprintf(&b, "\n## Result\n\n**%s**: %s\n", r.Result, r.Reasons)
		if len(r.Concerns) > 0 {
			b.WriteString("\nConcerns of the non-required reviewers:\n\n")
			for _, concern := range r.Concerns {
				fmt.Fprintf(&b, "- %s\n", concern)
			}
		}
	}
	return b.String()
}

// modification writes the issue edit and the diff of a modification.
func modification(b *strings.Builder, m Modification) {
	if m.Edit != nil {
		fmt.Fprintf(b, "\nThe new text of the issue section \"%s\":\n\n%s\n", m.Edit.Section, quoteBlock(m.Edit.Text))
	}
	if m.Diff != "" {
		lines := strings.Split(strings.TrimRight(m.Diff, "\n"), "\n")
		if len(lines) > diffLines {
			lines = append(lines[:diffLines], fmt.Sprintf("... (%d more lines; see the commit)", len(lines)-diffLines))
		}
		fmt.Fprintf(b, "\n````diff\n%s\n````\n", strings.Join(lines, "\n"))
	}
}

// RoundNote returns the Notes line of a round, for the issue.
func RoundNote(r *Referendum, round Round, record string) string {
	var votes []string
	for _, ballot := range round.Ballots {
		vote := ballot.Vote
		if ballot.Error != "" {
			vote = "no ballot"
		}
		votes = append(votes, fmt.Sprintf("%s (%s) %s", ballot.Role, ballot.Kind(), vote))
	}
	var gates []string
	for _, gate := range round.Gates {
		verdict := "rejected"
		if gate.Accepted {
			verdict = "accepted"
		}
		gates = append(gates, fmt.Sprintf("the gate %s the modification of %s", verdict, gate.Modification.Author))
	}
	note := fmt.Sprintf("%s, referendum round %d on v%d: %s.", r.Date, round.Number, round.Version, strings.Join(votes, ", "))
	if len(gates) > 0 {
		note += " " + capitalize(strings.Join(gates, "; ")) + "."
	}
	return note + " " + round.Outcome + " Record: [[" + record + "]]."
}

// ResultNote returns the Notes line of the result, for the issue.
func ResultNote(r *Referendum, record string) string {
	note := fmt.Sprintf("%s, referendum result: %s. %s", r.Date, r.Result, r.Reasons)
	if len(r.Concerns) > 0 {
		note += fmt.Sprintf(" %d concerns of non-required reviewers are in the record.", len(r.Concerns))
	}
	final := r.Current()
	if r.Result == Passed && final.Number > 1 {
		note += fmt.Sprintf(" The accepted implementation is v%d, the bookmark `%s`.", final.Number, final.Ref)
	}
	return note + " Record: [[" + record + "]]."
}

func cell(text string) string {
	return strings.ReplaceAll(oneLine(text), "|", "\\|")
}

func quoteBlock(text string) string {
	return "> " + strings.ReplaceAll(strings.TrimRight(text, "\n"), "\n", "\n> ")
}

func capitalize(text string) string {
	if text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}
