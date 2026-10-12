package referendum

import (
	"fmt"
	"strings"

	"github.com/runewake2/sisyphus/hephaestus/internal/airole"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
)

// BallotPrompt returns the prompt of an AI reviewer for its ballot on v.
func BallotPrompt(r *Referendum, reviewer Reviewer, role roles.Role, v Version) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are a %s reviewer in a hephaestus referendum, as the role %s. You cast one ballot on version v%d of the contents.\n\n", reviewer.Kind(), role.Name, v.Number)
	contents(&b, r, v)
	b.WriteString(`## What to do

1. Read the issue, and the change of the implementation from the base. Evaluate it only in the scope
   of your role, and against the current implementation (the base):
   - Does the change do what the issue asks, in your scope?
   - Does it break something in your scope that works now?
   - What is missing in your scope?
2. Cast one ballot:
   - yes: in your scope, the change is right as it is.
   - no: in your scope, the change must not go on. Give the reasons.
   - modify: in your scope, the change can go on with one modification. Make the modification in
     your working directory, which is a scratch copy at the implementation (commit it or not), or
     give a new text for one section of the issue in "section" and "text". Modify only what is in
     your scope.
`)
	if !reviewer.Required {
		b.WriteString("   - abstain: nothing of the change is in your scope. Only a non-required reviewer, as you are, can abstain.\n")
	}
	b.WriteString(`3. Do not change the issue, the tracker, or a bookmark, and do not push. Your ballot is your only
   output; hephaestus records it. You do not see the ballots of the other reviewers of this round.

End your output with your ballot, as one line of JSON, for example:

{"vote": "yes", "reasons": "The docs agree with the new flag."}
{"vote": "modify", "reasons": "The help text names the wrong flag.", "modification": "Correct the flag in the help text."}
{"vote": "modify", "reasons": "A criterion is missing.", "modification": "Add a criterion.", "section": "Acceptance criteria", "text": "<the new text of the section>"}
`)
	fmt.Fprintf(&b, "\n%s", airole.Instructions(role))
	return b.String()
}

// GatePrompt returns the prompt of a required AI reviewer at the gate of modification m of v.
func GatePrompt(r *Referendum, reviewer Reviewer, role roles.Role, v Version, m Modification) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are a required reviewer in a hephaestus referendum, as the role %s. A reviewer submitted a modification of version v%d. Accept it or reject it, in your scope. The modification goes on only when every required reviewer accepts it; then every reviewer votes again on the modified version.\n\n", role.Name, v.Number)
	contents(&b, r, v)
	fmt.Fprintf(&b, "## The modification\n\n- Author: %s\n- Summary: %s\n", m.Author, oneLine(m.Summary))
	if m.Commit != "" {
		fmt.Fprintf(&b, "- The modified implementation: commit %s. Your working directory is a scratch copy at it. See the modification with: git diff %s %s\n", m.Commit, v.Commit, m.Commit)
	}
	if m.Edit != nil {
		fmt.Fprintf(&b, "- The new text of the issue section %q:\n\n%s\n", m.Edit.Section, indent(m.Edit.Text))
	}
	b.WriteString(`
## What to do

Judge only the modification, in the scope of your role. Do not change anything: changes that you
make are not kept. End your output with your vote, as one line of JSON, for example:

{"gate": "accept", "reasons": "The modification is right in my scope."}
{"gate": "reject", "reasons": "The message is in the scope of another issue."}
`)
	fmt.Fprintf(&b, "\n%s", airole.Instructions(role))
	return b.String()
}

// contents writes the question, the issue, and the implementation of version v.
func contents(b *strings.Builder, r *Referendum, v Version) {
	fmt.Fprintf(b, "## The question\n\n%s\n\n", r.Question)
	fmt.Fprintf(b, "## The contents: v%d\n\n", v.Number)
	fmt.Fprintf(b, "- The issue: %s. Read it with: sisyphus show %s\n", r.Issue, r.Issue)
	fmt.Fprintf(b, "- The implementation: commit %s, from the bookmark %s. Your working directory is a scratch copy at it.\n", v.Commit, r.Bookmark)
	fmt.Fprintf(b, "- The current implementation (the base): %s, commit %s. See the change with: git diff %s %s\n", r.Base, r.BaseID, r.BaseID, v.Commit)
	if len(v.IssueEdits) > 0 {
		b.WriteString("- The issue text of this version has these sections changed; they replace what sisyphus shows:\n")
		for _, edit := range v.IssueEdits {
			fmt.Fprintf(b, "\n  ### %s\n\n%s\n", edit.Section, indent(edit.Text))
		}
	}
	if len(r.Versions) > 1 {
		b.WriteString("\n## The versions so far\n\n")
		for _, version := range r.Versions {
			if len(version.Modifications) == 0 {
				fmt.Fprintf(b, "- v%d: the bookmark as it was called.\n", version.Number)
				continue
			}
			fmt.Fprintf(b, "- v%d: the modifications of the round before it:\n", version.Number)
			for _, m := range version.Modifications {
				fmt.Fprintf(b, "  - %s: %s\n", m.Author, oneLine(m.Summary))
			}
		}
	}
	b.WriteString("\n")
}

func indent(text string) string {
	if strings.TrimSpace(text) == "" {
		return "    (no text)"
	}
	return "    " + strings.ReplaceAll(strings.TrimRight(text, "\n"), "\n", "\n    ")
}
