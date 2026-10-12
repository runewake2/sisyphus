# hephaestus

hephaestus is a worker manager: it provisions and manages agents, each with one role, and gives
them sisyphus issues to resolve. It depends on sisyphus, and uses sisyphus issues as its record of
work. It is a separate project that lives beside sisyphus, in this directory of the sisyphus repo.

> **Status:** early. Delegation, the message queue, tool and AI agents, referendums, the merge
> queue, and the versioning tools work; the scheduler daemon does not exist yet. [[agent-delegation]]
> is the design, and [[delegate-issues-to-agents]] is the first epic.

hephaestus is a Go module of its own, in this directory. It runs as a CLI tool, `hephaestus`, and as
a background scheduler daemon that starts the workers on their triggers.

## hephaestus and sisyphus

The two tools must not overlap:

| sisyphus | hephaestus |
|---|---|
| A tool to interact with issues: create, change, find, link, and check them. | A worker manager: provisions agents, starts them, gives them work, and records their runs. |
| Knows nothing about roles, agents, or runners. | Knows the roles, and runs the agents. |
| Gives hephaestus data: fields, `metadata`, claims, filters, and JSON. | Uses only the sisyphus CLI or `sisyphus-mcp` to read and change issues. |

All agent-specific work is in hephaestus. The only sisyphus work that hephaestus can ask for is
generic issue data that any client can use, for example a filter on a `metadata` key. See
[[0029-sisyphus-does-not-manage-agents]].

## Commands

Build it with `go build ./cmd/hephaestus` in this directory. Every command takes `-C <directory>`,
as `git -C` does.

| Command | What it does |
|---|---|
| `hephaestus roles` | List the roles of `hephaestus/roles/<role>.md`, and report every problem of a role file ([[role-file]]). |
| `hephaestus delegate <issue> <role>` | Delegate an issue to a role: `sisyphus update <issue> --metadata delegate=<role>`. `--clear` removes it. |
| `hephaestus tasks <role>` | List the tasks of a role that can start now; `--all` in every state. |
| `hephaestus merge-order <task>...` | Print the order in which the merge queue merges the work of a set of tasks, with the reason for each place. `--onto` gives the base (default `main`); `--json` prints it as JSON. |
| `hephaestus messages serve` | Run the message queue of the repo: a service on 127.0.0.1 that every workspace of the repo finds, with an append-only log ([[0001-a-local-service-with-an-append-only-log]]). |
| `hephaestus messages send` | Send a message `--to <worker>`, `--role <role>` (`--all` for every worker of it), or `--broadcast`, with `--kind`, `--issue`, and `--data`. |
| `hephaestus messages receive`, `ack` | Print the messages that wait for a worker (`--as`, default `SISYPHUS_OWNER`), and acknowledge them, so that they do not come again. |
| `hephaestus messages listen -- <command>` | Run a command for each message of a worker, as a tool agent does; exit code 0 acknowledges the message. |
| `hephaestus messages log` | Print every message of the log, also when the service does not run. |
| `hephaestus agent run <role>` | Run a worker of a tool role: for each message, claim its issue, run the command of the role, and close or stop the issue from the exit code. No model call ([[0003-a-tool-agent-runs-its-command-for-each-message]]). |
| `hephaestus agent run <ai-role> --runner <command>` | Run a worker of an AI role: the runner (for example `claude -p --agent {role}`) gets the prompt of each task. Never without a runner. |
| `hephaestus agent export claude` | Write a Claude Code subagent in `.claude/agents/` for each AI role; `--check` tells whether they are in sync with the role files. |
| `hephaestus agent review <role>` | Run a tool role as a referendum reviewer, and print its ballot: `yes`, `no`, or `modify` with the changes that it made. |
| `hephaestus referendum call <issue>` | Let role agents vote on the implementation of an issue: `--question`, `--required` and `--optional` reviewers, and `--runner` for AI reviewers. Each reviewer votes in a scratch worktree. It runs the rounds and the modification gates to the result, and records each round in the issue and in `referendums/<id>.md` ([[referendum]]). Exit code 0 only when it passes. |
| `hephaestus referendum decide <issue> passed\|failed` | Decide a referendum as a human: after `no-result`, or to overrule its result, with `--reason`. |
| `hephaestus merge-queue <task>...` | Make one linear history from the work of a set of tasks, onto a new bookmark `merge-queue/<date>`: merge each bookmark in the order of `merge-order`, renumber decision records, give the versions, and run `--checks`. It never moves `main` ([[0001-merge-with-git-in-a-scratch-worktree]]). |
| `hephaestus version list` | List the unreleased changelog entries, in the order that `assign` uses by default. |
| `hephaestus version assign [<entry>...]` | Give the next SHAME versions to unreleased changelog entries, in the given order: move each to `changelog/<version>.md`, add the `CHANGELOG.md` entry, set `VERSION`, and change every link to it. `--dry-run` and `--json` show the plan. |
| `hephaestus decisions renumber [--onto main]` | Give a new number to each decision record that has the number of another one in its directory and that the base does not have, and change every link to it. `--dry-run` and `--json` show the plan. |

The merge queue runs `version assign` and `decisions renumber` as it lands work: versions come
from the merge queue ([[0001-versions-come-from-the-merge-queue]]).
A human who lands work by hand runs them too.

## Layout

```
hephaestus/
  README.md     This page
  go.mod        The Go module of hephaestus
  cmd/          The hephaestus command
  internal/     Its packages
  roles/        One role file for each role
```

hephaestus shares the issues and the design log of the repo with sisyphus
([[0002-one-tracker-and-one-design-log-for-both-projects]]): its issues are in
`issues/<state>/hephaestus/`, its design documents in `design/hephaestus/`, and its decision
records in `design/decisions/hephaestus/`. The rules of the repo apply ([[AGENTS]] and
[[CONTRIBUTING#hephaestus]]).
