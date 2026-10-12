# Role files

- Date: 2026-10-12
- Issue: [[delegation-model]]

Each role of hephaestus is a Markdown file of its own, `hephaestus/roles/<role>.md`
([[0002-each-role-is-a-markdown-file]]). A role of a family, for example one formatter for each
language, is `hephaestus/roles/<family>/<variant>/role.md`. Its name is the path, for example
`formatter/golang`, and its directory holds its implementation
([[0005-roles-of-a-family-live-in-a-directory]]). People read it to know what the role does; hephaestus reads
its frontmatter to run the role; for an AI role, the model gets its sections as its instructions.
[[roles/merge-queue|roles/merge-queue.md]] and [[roles/repo-check|roles/repo-check.md]] are examples of tool roles. [[roles/engineer|roles/engineer.md]], [[roles/project-manager|roles/project-manager.md]],
[[roles/ux-researcher|roles/ux-researcher.md]], and [[roles/technical-docs|roles/technical-docs.md]] are AI roles.

Link to a role file as `[[roles/<role>]]`: the subagents in `.claude/agents/` have the same file
names ([[#Running an AI role]]).

## Frontmatter

| Field | Values | Meaning |
|---|---|---|
| `role` | kebab-case, the name of the file; for a role of a family, its path, for example `formatter/golang` | The id of the role. A task is delegated to it with the sisyphus metadata key `delegate` ([[0001-delegate-is-a-metadata-key]]). |
| `kind` | `ai` or `tool` | Who does the work: a model, or a program with no model ([[agent-delegation#AI agents and tool agents]]). |
| `command` | A command line | A tool role only: the program that does the work of a task. An AI role has none. The command gets `HEPHAESTUS_ROLE` (the role), `HEPHAESTUS_ROLE_DIR` (the directory of the role file, for a script there), and, in a referendum, `HEPHAESTUS_REVIEW=1` and `HEPHAESTUS_BASE`. |
| `trigger` | `request`, `close`, `schedule`, or a list such as `[request, schedule]` | What starts the role (Q5 of [[agent-delegation]]). |
| `on-failure` | `needs-human` (the default), or an AI role | Where the work goes that the role cannot do. A tool can give it to an AI role; nothing gives work back to a tool. |
| `listens` | A list of kinds, for example `[bookmark-ready]` | Optional: the kinds of broadcast that the workers of the role react to ([[message-queue]]). They also get every message to them and to the role. |
| `allowed-tools` | A list of Claude Code tools, for example `[Bash, Read, Grep, Glob]` | An AI role only: the tools of its subagent. Without it, the tools of the session. |
| `model` | A model, for example `sonnet` | An AI role only: the model of its subagent. Without it, the model of the session. |
| `can-call-referendum` | `true` or `false` (the default) | Whether a human allows the role to call a referendum ([[referendum]]). |

A comment after `#` is allowed on each line.

## Sections

The body has a title, `# <Role name>`, and these sections, none of them empty:

| Section | What it holds |
|---|---|
| `## Charter` | What the role does, in one paragraph, and what it does not do. The first paragraph is the summary of `hephaestus roles`. |
| `## Inputs` | What the role gets: the tasks, the messages, and the repo state that it reads. |
| `## Outputs` | What the role makes: changes, bookmarks, issues, notes, and messages. |
| `## Works with` | Optional: how the role works with the other roles, one item for each, which starts with the role or family in backticks: who gives it work, who gets its work, and where the scope of each ends. `hephaestus roles` checks that each names a role or family that exists, and an AI role gets the section with its instructions. |
| `## Tools` | The commands and tools that the role may use. |
| `## Ask a human` | The points where the role must stop and ask, as the rules of [[agent-delegation#Rules for every role]] ask. |
| `## Prompt` | An AI role only: the instructions of the model, in addition to the sections above. |

`hephaestus roles` lists the roles and reports every problem of every role file, for example a
missing section, a tool role with no command, or an `on-failure` role that is not an AI role.

## Delegation

- `hephaestus delegate <issue> <role>` runs `sisyphus update <issue> --metadata delegate=<role>`,
  after it checks that the role has a role file. `--clear` removes the delegation.
- `hephaestus tasks <role>` lists the tasks of the role that can start now:
  `sisyphus list --metadata delegate=<role> --state open --ready`. `--all` lists them in every
  state.
- `delegate` is the role that owns the task; the sisyphus `owner` is the worker that works on it,
  for example `merge-queue-1`, which hephaestus gives in `SISYPHUS_OWNER`.

## Running an AI role

hephaestus runs AI roles in two ways ([[0004-ai-roles-run-as-claude-code-subagents]]):

- **As Claude Code subagents.** `hephaestus agent export claude` writes `.claude/agents/<role>.md`
  (with `-` for each `/` of a role of a family)
  for each AI role: its description, `allowed-tools`, `model`, sections, and the rules of every
  role. A person in Claude Code gives a task to the subagent of a role. The files are generated: a
  test fails when they differ from the role files, so run the export in the same change as a change
  of a role file.
- **As a worker of hephaestus.** `hephaestus agent run <role> --runner "claude -p --agent {role}"`
  (or the environment variable `HEPHAESTUS_AI_RUNNER`) gets the messages of the role, claims the
  issue of each, and runs the runner with the prompt of the task on its standard input: the role
  file, the rules, and the message. The model closes or stops the issue itself; an issue that it
  leaves in progress, or a runner that fails, goes to a human. hephaestus never starts a model
  without a runner, because a model costs tokens.

