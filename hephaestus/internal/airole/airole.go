// Package airole turns the role file of an AI role into what a model runs: a Claude Code subagent,
// and the prompt of one task. hephaestus runs AI roles first as Claude Code subagents, then through
// a runner command (Q4 of [[agent-delegation]]).
package airole

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
)

// sections are the sections of a role file that go into the instructions of the model, in order.
var sections = []string{"Charter", "Inputs", "Outputs", "Works with", "Tools", "Ask a human", "Prompt"}

// rules are the rules for every role of [[agent-delegation#Rules for every role]], as the model
// must follow them.
const rules = `## Rules for every role

- Work only on the task of your role. Work that you find for another role becomes a sisyphus issue
  for that role.
- Use sisyphus for every read and change of an issue. Never edit an issue file by hand.
- Work in your own workspace and bookmark. Never push, never open a pull request, and never move
  main.
- Record what you did in the repo: the Notes and the Resolution of the issue, decision records, and
  an unreleased changelog entry for a change.
- When the task needs a choice that it does not make: write the question in the Notes of the issue,
  stop the issue (sisyphus update <issue> open --add-tags needs-human), and end.
- Write in Simplified Technical English.`

// Instructions returns the instructions of the role: its title, sections, and the rules.
func Instructions(role roles.Role) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\nYou are a worker of the hephaestus role %s.\n", role.Title, role.Name)
	for _, name := range sections {
		if text := role.Section(name); text != "" {
			fmt.Fprintf(&b, "\n## %s\n\n%s\n", name, text)
		}
	}
	fmt.Fprintf(&b, "\n%s\n", rules)
	return b.String()
}

// ClaudeAgent returns the Claude Code subagent of an AI role: a file for .claude/agents/<role>.md.
func ClaudeAgent(role roles.Role) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", role.Flat())
	description := role.Charter + " A hephaestus role: use it for tasks delegated to " + role.Name + "."
	fmt.Fprintf(&b, "description: %s\n", quote(description))
	if len(role.AllowedTools) > 0 {
		fmt.Fprintf(&b, "tools: %s\n", strings.Join(role.AllowedTools, ", "))
	}
	if role.Model != "" {
		fmt.Fprintf(&b, "model: %s\n", role.Model)
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "<!-- Generated from hephaestus/roles/%s by hephaestus agent export claude. Do not edit: change the role file. -->\n\n", role.File())
	b.WriteString(Instructions(role))
	return b.String()
}

// TaskPrompt returns the prompt of one task: who the worker is, the task message, and what to do
// with the issue at the end.
//
// workspace is the directory where the worker works, and bookmark the bookmark of its work, if the
// task has a workspace of its own.
func TaskPrompt(role roles.Role, worker string, m mq.Message, workspace, bookmark string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are %s, a worker of the hephaestus role %s. Your owner name is %s.\n\n", worker, role.Name, worker)
	fmt.Fprintf(&b, "## Your task\n\nA message from %s, of the kind %s:\n\n%s\n", m.From, m.Kind, indent(m.Body))
	if len(m.Data) > 0 {
		var data any
		if json.Unmarshal(m.Data, &data) == nil {
			pretty, _ := json.MarshalIndent(data, "", "  ")
			fmt.Fprintf(&b, "\nIts data:\n\n```json\n%s\n```\n", pretty)
		}
	}
	if bookmark != "" {
		fmt.Fprintf(&b, "\nYour workspace for this task is %s, with the bookmark %s. hephaestus made it for you. Work and commit only there, and set the bookmark %s on your work. hephaestus removes the workspace when you finish the task; your commits and the bookmark stay.\n", workspace, bookmark, bookmark)
	} else {
		fmt.Fprintf(&b, "\nWork in %s.\n", workspace)
	}
	if m.Issue != "" && (m.To.Broadcast || m.To.All) {
		fmt.Fprintf(&b, "\nThe message is a notice about the issue %s, to every listener. Do not start, close, or reopen that issue: another worker owns it, or it is closed. Record your result in its Notes (sisyphus edit %s notes \"...\" --append), and file an issue for each piece of work that you find.\n", m.Issue, m.Issue)
	} else if m.Issue != "" {
		fmt.Fprintf(&b, "\nThe issue of the task is %s. You have it in progress. When you finish, close it with the result in its Resolution (sisyphus close %s -m \"...\"). When you must stop, stop it as the rules say. Do not end with the issue in progress.\n", m.Issue, m.Issue)
	}
	fmt.Fprintf(&b, "\n%s", Instructions(role))
	return b.String()
}

func indent(text string) string {
	if strings.TrimSpace(text) == "" {
		return "    (no text)"
	}
	return "    " + strings.ReplaceAll(strings.TrimRight(text, "\n"), "\n", "\n    ")
}

// quote returns a YAML double-quoted string.
func quote(text string) string {
	data, _ := json.Marshal(text)
	return string(data)
}
