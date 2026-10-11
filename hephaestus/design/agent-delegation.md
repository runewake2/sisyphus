# Agent delegation

- Date: 2026-10-11
- Issue: [[delegate-issues-to-agents]]
- Status: Draft. The open questions at the end are not answered yet.

## Summary

hephaestus delegates sisyphus issues to agents that each have one role, and lets the agents
resolve the issues. This document describes the roles, the rules that all of them follow, and how
an issue gets to an agent. The first role to build is the **merge queue**, which merges many
parallel lines of work in the best order.

sisyphus does not manage agents: it is only a tool to interact with issues. Everything in this
document is part of hephaestus, except generic issue data that hephaestus asks sisyphus for (see
[[#What hephaestus needs from sisyphus]]).

| Role | Id | What it does | Issue |
|---|---|---|---|
| Merge queue | `merge-queue` | Takes the finished work of many parallel workspaces and merges it, one change line at a time, in the best order. | [[merge-queue-agent]] |
| Project manager | `project-manager` | Consolidates and focuses the issues: merges duplicates, splits large work, sets priorities, and keeps epics current. | [[project-manager-agent]] |
| Engineer | `engineer` | Takes two or more concurrent workstreams and organizes their work so that all of them go in the same direction. | [[engineer-agent]] |
| UX researcher | `ux-researcher` | Finds out how people and agents use what is being built, and files issues for problems and improvements. | [[ux-researcher-agent]] |
| Technical docs | `technical-docs` | Checks completed work against the documentation, and writes documentation for new features. | [[technical-docs-agent]] |

## Why

Today one agent at a time does each issue, and a human does the rest: the human merges the
bookmarks of the agents, notices that two workstreams go in different directions, decides which
issues matter, and checks that the docs still agree with the code. In the sisyphus repo, for
example, the work of several workspaces (`ux-fixes`, `tracker-fixes`, `edit-issue-sections`, and
`ollama-model-spike`) was merged by hand, and each merge had to renumber the SHAME versions that
the workspaces picked at the same time (see
[CONTRIBUTING.md, Changelog](../../CONTRIBUTING.md#changelog)). These are roles with clear inputs
and outputs, so agents can do them, with a human at the points that need judgment.

## Rules for every role

These rules add to [AGENTS.md](../../AGENTS.md). They apply to every role agent that hephaestus
runs.

1. **One role for each agent.** An agent works only on the issues delegated to its role, and only
   does the work of its role. Work that it finds for another role becomes an issue for that role.
2. **sisyphus is the interface to issues.** An agent finds, claims, and closes its work with
   sisyphus ([docs/agents.md](../../docs/agents.md)), never by editing issue files by hand.
   hephaestus gives each agent an owner name, its role id and an instance, for example
   `merge-queue-1`, in `SISYPHUS_OWNER`.
3. **Own workspace and bookmark.** hephaestus gives each agent its own jj workspace and a bookmark
   `samw/ai/<role>-<topic>`, and the agent claims each issue with `sisyphus start`
   ([docs/sisyphus-claims.md](../../docs/sisyphus-claims.md)).
4. **Humans keep the code host and `main`.** No agent pushes, opens pull requests, or moves `main`
   unless a human tells it to. An agent prepares the result, and a human accepts it. A human can
   give one role more authority later, as a recorded decision.
5. **Every result is in the repo.** An agent records what it did in the issue (`Notes`,
   `Resolution`), in decision records, and in the changelog, as
   [CONTRIBUTING.md](../../CONTRIBUTING.md) asks. A report that exists only in a chat is lost.
6. **Stop and ask.** When the work needs a choice that the issue does not make, the agent writes
   the question in `Notes`, stops the issue (`sisyphus update <name> open`), and tags it
   `needs-human`.

## How an issue gets to an agent

This is a draft; [[delegation-model]] decides it. The proposed model:

- An issue is delegated to a role with the generic `metadata` field of sisyphus:
  `sisyphus update <name> --metadata delegate=merge-queue`. sisyphus does not know what the key
  means.
- hephaestus finds the work of a role with a sisyphus filter on that key, for example
  `sisyphus list --ready --metadata delegate=merge-queue --json` (see
  [[#What hephaestus needs from sisyphus]]).
- A role is described in `hephaestus/roles/<role>.md`: its charter, its inputs, its outputs, the
  tools it may use, the points where it must ask a human, and its prompt. hephaestus turns this
  file into an agent, for example a Claude Code subagent.
- Some roles also create their own work. For example, the technical docs agent looks at each issue
  that closes, and the merge queue looks at each bookmark whose issues are all closed.
- hephaestus starts the agents on a trigger (a schedule, a close, or a human): [[delegation-runner]].

## What hephaestus needs from sisyphus

Only generic issue data, which any client of sisyphus can use, and which has no meaning about
agents in sisyphus:

| Need | sisyphus today | sisyphus issue |
|---|---|---|
| Record the role of an issue | `metadata` (`--metadata key=value`) | none |
| Find the issues of a role | Not possible: no filter on `metadata` | `filter-by-metadata` |
| Know who works on an issue, in any workspace | `owner`, `SISYPHUS_OWNER`, claims | none |
| Read issues and their links as data | `--json` on `list`, `search`, `show`, `graph`, and `claims` | none |
| The issues of a bookmark | `bookmark` in the JSON rows of `list` | none (filter in hephaestus) |

A new need goes to sisyphus only if it is generic issue data. Everything else is hephaestus code.

## Merge queue

The first role. [[merge-queue-agent]] tracks it.

**Input:** the bookmarks of finished work: bookmarks ahead of `main` whose issues are all closed,
whose tests pass, and whose lint is clean. A human can also give it a list of bookmarks.

**Order:** the merge queue merges one bookmark at a time, in an order that keeps the work correct
and the conflicts small ([[merge-order-planner]]):

1. Dependencies first: when an issue of one bookmark depends on an issue of another
   (`depends-on`), the bookmark of the dependency goes first.
2. Stacked bookmarks in their order: a bookmark based on another goes after it.
3. Then the bookmark that conflicts with the fewest others (by the files that both change), so
   that the large, risky merges come last and do not block the small ones.
4. Then the oldest first.

**Each merge:** rebase the bookmark onto the head of the queue in a scratch workspace, resolve the
mechanical conflicts, check, and go on:

- The mechanical conflicts are known: two workspaces pick the same next SHAME version, the same
  decision record number, and the same place at the top of `CHANGELOG.md`. The merge queue
  renumbers `VERSION`, the changelog file, the `CHANGELOG.md` entry, the `Version:` lines, the
  decision record files, and every link to them ([[renumber-versions-on-merge]]). These are the
  rules of the repo, not of sisyphus, so this is hephaestus code.
- Other conflicts, for example two changes of one function, are not mechanical. The merge queue
  stops at that bookmark, records the conflict in the notes of its issues, and goes on with the
  bookmarks that do not depend on it.
- After each rebase: the checks of the repo must pass, for example `gofmt`, `go vet`, the tests,
  and `sisyphus lint`.

**Output:** a bookmark `merge-queue/<date>` with all merged changes in one line on `main`, and a
report in the issue of the run: the order, what was renumbered, and what was stopped and why. A
human checks the report and moves `main`.

## Votes of several roles

One role judges a change only in its own scope. A **referendum** lets several role agents vote on
one implementation. Required reviewers must all vote `yes` to pass it; non-required reviewers can
only suggest modifications and raise concerns. A modification goes on only when every required
reviewer accepts it, and then every reviewer votes again on the modified version. [[referendum]] describes it, and [[referendum-voting]]
tracks it.

## The other roles

Each of these is a separate issue, with its own design before the work.

- **Project manager** ([[project-manager-agent]]): reads all open issues, and proposes, as one
  change for a human to accept, which issues to merge (duplicates), split (`x-large`), close (done
  or obsolete), and reprioritize, and which epics are complete. Uses `list`, `search`, `graph`,
  `claims`, and `lint`.
- **Engineer** ([[engineer-agent]]): takes two or more concurrent workstreams (bookmarks,
  workspaces, or epics), finds where they overlap or disagree (the same files, the same feature,
  decisions that contradict), and proposes one direction: a decision record, changes to the scope
  of the issues, and an order. It does not merge; the merge queue does.
- **UX researcher** ([[ux-researcher-agent]]): uses what is being built as a new person and as an
  agent would (a first session, every `--help`, the docs, the MCP tools, and small models), and
  files one issue for each problem, with what it saw. The sisyphus reviews that filed the epics
  `cli-and-docs-ux` and `agent-cooperation` are examples, as a repeatable role.
- **Technical docs** ([[technical-docs-agent]]): for each issue that closes, checks that the
  command pages, the README, and the guides still agree with the behavior (run the examples,
  compare the flags), fixes the docs or files an issue, and writes the page for a new feature. It
  follows the writing style of [CONTRIBUTING.md](../../CONTRIBUTING.md#writing-style).

## Open questions

A human answers a question with a reply such as "Q1 yes", or another answer.

| Id | Question | Asked in | Proposed answer |
|---|---|---|---|
| Q1 | How is an issue delegated to a role: the `metadata` key `delegate`, a tag, or the `owner`? | [[delegation-model]] | The `metadata` key `delegate`, with the sisyphus filter `filter-by-metadata`. It needs no field for agents in sisyphus. A tag mixes roles with components, and the owner is the agent instance that works on it. |
| Q2 | Where are the roles described: `hephaestus/roles/<role>.md`, `.claude/agents/`, or in code? | [[delegation-model]] | `hephaestus/roles/<role>.md`, readable by people and by any runtime. hephaestus can generate `.claude/agents/` files from it. |
| Q3 | May the merge queue move `main`, or only prepare a bookmark? | [[merge-queue-agent]] | Only prepare `merge-queue/<date>` and a report. A human moves `main`, as [AGENTS.md](../../AGENTS.md) asks now. |
| Q4 | What runs the agents: Claude Code subagents, the Claude Agent SDK, Claude Managed Agents, or local models? | [[delegation-runner]] | Start with Claude Code subagents, started by a human, then hephaestus on the Agent SDK. Local models only for roles that a task suite shows they can do. |
| Q5 | Which roles start by themselves, and on what trigger? | [[delegation-runner]] | Merge queue: on a human request, later on a schedule. Technical docs: on each close. The others: on a human request. |
| Q6 | Does the merge queue merge the work of humans too, or only of agents? | [[merge-queue-agent]] | Any bookmark that a human puts in its input list, and all `samw/ai/*` bookmarks of finished work. |
| Q7 | Is hephaestus a Go program in the sisyphus module (`cmd/hephaestus`) or a module of its own in `hephaestus/`? | [[delegate-issues-to-agents]] | A module of its own in `hephaestus/`, so that it can only use sisyphus through its CLI and MCP server, as any other client. |
