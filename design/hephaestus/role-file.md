# Role files

- Date: 2026-10-12
- Issue: [[delegation-model]]

Each role of hephaestus is a Markdown file of its own, `hephaestus/roles/<role>.md`
([[0002-each-role-is-a-markdown-file]]). People read it to know what the role does; hephaestus reads
its frontmatter to run the role; for an AI role, the model gets its sections as its instructions.
[[merge-queue|roles/merge-queue.md]] and [[repo-check|roles/repo-check.md]] are examples of tool roles. [[engineer|roles/engineer.md]], [[project-manager|roles/project-manager.md]],
[[ux-researcher|roles/ux-researcher.md]], and [[technical-docs|roles/technical-docs.md]] are AI roles.

## Frontmatter

| Field | Values | Meaning |
|---|---|---|
| `role` | kebab-case, the name of the file | The id of the role. A task is delegated to it with the sisyphus metadata key `delegate` ([[0001-delegate-is-a-metadata-key]]). |
| `kind` | `ai` or `tool` | Who does the work: a model, or a program with no model ([[agent-delegation#AI agents and tool agents]]). |
| `command` | A command line | A tool role only: the program that does the work of a task. An AI role has none. |
| `trigger` | `request`, `close`, `schedule`, or a list such as `[request, schedule]` | What starts the role (Q5 of [[agent-delegation]]). |
| `on-failure` | `needs-human` (the default), or an AI role | Where the work goes that the role cannot do. A tool can give it to an AI role; nothing gives work back to a tool. |
| `listens` | A list of kinds, for example `[bookmark-ready]` | Optional: the kinds of broadcast that the workers of the role react to ([[message-queue]]). They also get every message to them and to the role. |
| `can-call-referendum` | `true` or `false` (the default) | Whether a human allows the role to call a referendum ([[referendum]]). |

A comment after `#` is allowed on each line.

## Sections

The body has a title, `# <Role name>`, and these sections, none of them empty:

| Section | What it holds |
|---|---|
| `## Charter` | What the role does, in one paragraph, and what it does not do. The first paragraph is the summary of `hephaestus roles`. |
| `## Inputs` | What the role gets: the tasks, the messages, and the repo state that it reads. |
| `## Outputs` | What the role makes: changes, bookmarks, issues, notes, and messages. |
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
