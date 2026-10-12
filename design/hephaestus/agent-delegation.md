# Agent delegation

- Date: 2026-10-11
- Issue: [[delegate-issues-to-agents]]
- Status: Draft. All open questions (Q1-Q14) are answered.

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
| Merge queue | `merge-queue` | Takes a set of tasks, done in parallel, and makes one linear history of their work, in the best order. It never moves `main`. | [[merge-queue-agent]] |
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
[[CONTRIBUTING#Changelog]]). These are roles with clear inputs
and outputs, so agents can do them, with a human at the points that need judgment.

## Rules for every role

These rules add to [[AGENTS]]. They apply to every role agent that hephaestus
runs.

1. **One role for each agent.** An agent works only on the issues delegated to its role, and only
   does the work of its role. Work that it finds for another role becomes an issue for that role.
2. **sisyphus is the interface to issues.** An agent finds, claims, and closes its work with
   sisyphus ([[agent-guide]]), never by editing issue files by hand.
   hephaestus gives each agent an owner name, its role id and an instance, for example
   `merge-queue-1`, in `SISYPHUS_OWNER`.
3. **Own workspace and bookmark.** hephaestus gives each agent its own jj workspace and a bookmark
   `samw/ai/<role>-<topic>`, and the agent claims each issue with `sisyphus start`
   ([[sisyphus-claims]]).
4. **Humans keep the code host and `main`.** No agent pushes, opens pull requests, or moves `main`
   unless a human tells it to. An agent prepares the result, and a human accepts it. A human can
   give one role more authority later, as a recorded decision, except the merge queue, which never
   moves `main` (Q3).
5. **Every result is in the repo.** An agent records what it did in the issue (`Notes`,
   `Resolution`), in decision records, and in the changelog, as
   [[CONTRIBUTING]] asks. A report that exists only in a chat is lost.
6. **Stop and ask.** When the work needs a choice that the issue does not make, the agent writes
   the question in `Notes`, stops the issue (`sisyphus update <name> open`), and tags it
   `needs-human`.

## AI agents and tool agents

hephaestus runs two kinds of agents (from the human, 2026-10-11):

| Kind | What does the work | Use for | Cost |
|---|---|---|---|
| **AI agent** | A model, for example a Claude Code subagent, then the Claude Agent SDK (Q4). | Work that needs judgment: a design, a review, a conflict of direction, a text. | Tokens for each task. |
| **Tool agent** | A program that does not use AI: a command with fixed rules. | Work that a program can do with no judgment: large refactors (a codemod, `gofmt -r`, a rename across the repo), lints, formatters, and renumbering. | No tokens: only the time of the program. |

A tool agent is a role like any other: it has a role file (`hephaestus/roles/<role>.md`, with
`kind: tool` and its command), an owner name, its own workspace and bookmark, and it claims,
records, and closes its tasks with sisyphus, as the [[#Rules for every role]] ask. Only the worker
differs: hephaestus runs the command of the role on the task, instead of a model.

- **Delegate to a tool when you can.** A large refactor that a codemod can do costs no tokens as a
  tool task, and many tokens as an AI task. Split a large task into the part that a tool can do and
  the part that needs judgment, and delegate each part to its kind of agent.
- **A tool does not decide.** A tool agent does what its command does, and nothing more. When the
  command fails, or cannot do the task, the tool agent stops the issue with the output of the
  command in `Notes`, and tags it `needs-human`, or hephaestus delegates the rest to an AI role
  (Q8).
- **Results are checked the same way.** The work of a tool agent goes through the same checks,
  referendums, and merge queue as the work of an AI agent.
- **Context is a message.** A tool agent gets the context of its task as a message (Q8, answered by
  the human on 2026-10-11): hephaestus, a human, or any other agent sends it, to the tool agent or
  as a public broadcast ([[#Messages between workers]]).
- **Reviewers too.** A tool agent can be a reviewer in a referendum (Q9, answered by the human on
  2026-10-11): a lint or the tests vote with their exit code, and a formatter can submit its changes
  as a modification.

`hephaestus agent run <role>` runs a worker of a tool role, and `hephaestus agent review <role>`
gives its ballot: [[0003-a-tool-agent-runs-its-command-for-each-message]] records how.

## Messages between workers

Agents must be able to send information to each other: one worker to another, and to every worker
at once, across kinds (AI and tool) and across workspaces (from the human, 2026-10-11). A tool agent
of its own, the **message queue** ([[message-queue]]), relays the messages:

- **Direct:** from one worker to another worker, by its owner name (for example
  `merge-queue-1`), or to any worker of a role (for example `engineer`) (Q13).
- **Broadcast:** from one worker to every worker that listens, as a public message. A tool agent
  can listen for broadcasts of one kind, for example a lint that runs on each "bookmark ready".
- **Any worker:** humans, AI agents, and tool agents send and receive the same messages. A message
  is how a tool agent gets the context of its task.
- **A message is information, not a decision.** Agents that must agree use a referendum
  ([[referendum]]); a message can call one, or tell its result.

A message has: an id, the sender, the receiver (a worker, a role, or `broadcast`), a kind (for
example `task`, `info`, or `referendum`), the issue that it is about, if any, the time, and a body
(text, and data as JSON for tool agents).

Answered by the human on 2026-10-11:

- **An append-only log** keeps every message (Q11). Nothing changes or removes a message after it is
  sent. A receiver still records in the `Notes` of an issue what a message changes for that issue,
  as [[#Rules for every role]] ask.
- **A service** passes the messages (Q12): the message queue runs as a service that the workers
  connect to, not as files in a shared directory.
- **At-least-once delivery** (Q14): each receiver gets each message at least once. A receiver can get
  a copy, so each message has an id, and a receiver ignores an id that it already handled.

- **Ordered** (Q14, answered by the human on 2026-10-12): the messages of one sender reach each
  receiver in the order that the sender sent them.
- **Addressing** (Q13, answered by the human on 2026-10-12): by owner name (one worker), by role
  (any one worker of the role, or all of them, by a flag of the message), or `broadcast` (every
  listener; a listener can filter by kind).

`hephaestus messages` implements it: [[0001-a-local-service-with-an-append-only-log]] records how.

## The form of hephaestus

hephaestus is a Go module of its own, in `hephaestus/` (Q7, answered by the human on 2026-10-12).
It uses sisyphus only through the `sisyphus` CLI and `sisyphus-mcp`, as any other client does. It
runs in two forms:

- **A CLI tool**, `hephaestus`, for people and agents: for example to delegate a task, to call a
  referendum, to give the merge queue a set of tasks, and to see what the workers do.
- **A background scheduler daemon**, which starts the workers on their triggers ([[#How an issue gets to an agent]],
  Q5) and keeps them running.

## How an issue gets to an agent

[[delegation-model]] decided it, and [[role-file]] describes the role files and the commands:

- An issue is delegated to a role with the `delegate` key of the sisyphus `metadata` field (Q1,
  answered by the human on 2026-10-11): `delegate` is the role assigned to own the task, and
  `owner` is the agent instance that works on it, for example `delegate: merge-queue` and
  `owner: merge-queue-1`. Set it with
  `sisyphus update <name> --metadata delegate=merge-queue`. sisyphus does not know what the key
  means.
- hephaestus finds the work of a role with a sisyphus filter on that key, for example
  `sisyphus list --ready --metadata delegate=merge-queue --json` (see
  [[#What hephaestus needs from sisyphus]]).
- Each role is described in a Markdown file of its own, `hephaestus/roles/<role>.md` (Q2, answered
  by the human on 2026-10-11): its charter, its inputs, its outputs, the
  tools it may use, the points where it must ask a human, and its prompt. hephaestus turns this
  file into an agent, for example a Claude Code subagent.
- Some roles also create their own work. For example, the technical docs agent looks at each issue
  that closes. The merge queue does not: it works only on the set of tasks that it is given.
- hephaestus starts the agents on a trigger (a schedule, a close, or a human): [[delegation-runner]].

## What hephaestus needs from sisyphus

Only generic issue data, which any client of sisyphus can use, and which has no meaning about
agents in sisyphus:

| Need | sisyphus today | sisyphus issue |
|---|---|---|
| Record the role of an issue | `metadata` (`--metadata key=value`) | none |
| Find the issues of a role | Not possible: no filter on `metadata` | [[filter-by-metadata]] |
| Know who works on an issue, in any workspace | `owner`, `SISYPHUS_OWNER`, claims | none |
| Read issues and their links as data | `--json` on `list`, `search`, `show`, `graph`, and `claims` | none |
| The issues of a bookmark | `bookmark` in the JSON rows of `list` | none (filter in hephaestus) |

A new need goes to sisyphus only if it is generic issue data. Everything else is hephaestus code.

## Merge queue

The first role. [[merge-queue-agent]] tracks it, and `hephaestus merge-queue` does it
([[0001-merge-with-git-in-a-scratch-worktree]]). It is a tool agent (Q10, answered by the human
on 2026-10-11): its order, rebases, renumbering, and checks follow fixed rules, so it uses no
model, and costs no tokens.

**Job:** the merge queue's only job is to make one linear history from a set of tasks. It never
moves `main` (Q3, answered by the human on 2026-10-11), and it does nothing else: it does not
choose the tasks, push, or change the work beyond the merge.

**Input:** a set of tasks: sisyphus issues, each with the bookmark of its work (the `bookmark`
field), whose tests pass and whose lint is clean. A human or an allowed agent gives the set, or
delegates the tasks to `merge-queue`. The set can hold the work of humans, and branches made beside
the queue: they land on the linear branch through the same supervised merge queue (Q6), whoever did
the work.

**Order:** the merge queue merges the bookmark of one task at a time, in an order that keeps the work correct
and the conflicts small ([[merge-order-planner]], `hephaestus merge-order`). Tasks with the same bookmark
merge together, and a bookmark with no change from the base is left out:

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
  stops at that bookmark, records the conflict in the notes of its issues, sends a message about it
  to the `engineer` role or a human, and goes on with the bookmarks that do not depend on it. A tool
  does not resolve a conflict that needs judgment.
- After each rebase: the checks of the repo must pass, for example `gofmt`, `go vet`, the tests,
  and `sisyphus lint`.

**Output:** a linear history: a bookmark `merge-queue/<date>` with all merged changes in one line
on `main`, and a report in the issues of the tasks: the order, what was renumbered, and what was
stopped and why. What happens to the bookmark next is not the merge queue's job: a human checks it
and moves `main`.

## Votes of several roles

The referendum is how hephaestus coordinates many concurrent agents. One role judges a change only in its own scope. A **referendum** lets several role agents vote on
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
  of the issues, and an order. It puts the proposal to the agents of the workstreams as a
  referendum ([[referendum]]), when a human allows it to call one. It does not merge; the merge
  queue does.
- **UX researcher** ([[ux-researcher-agent]]): uses what is being built as a new person and as an
  agent would (a first session, every `--help`, the docs, the MCP tools, and small models), and
  files one issue for each problem, with what it saw. The sisyphus reviews that filed the epics
  [[cli-and-docs-ux]] and [[agent-cooperation]] are examples, as a repeatable role.
- **Technical docs** ([[technical-docs-agent]]): for each issue that closes, checks that the
  command pages, the README, and the guides still agree with the behavior (run the examples,
  compare the flags), fixes the docs or files an issue, and writes the page for a new feature. It
  follows the writing style of [[CONTRIBUTING#Writing style]].

## Open questions

A human answers a question with a reply such as "Q1 yes", or another answer.

| Id | Question | Asked in | Proposed answer |
|---|---|---|---|
| Q1 | How is an issue delegated to a role: the `metadata` key `delegate`, a tag, or the `owner`? | [[delegation-model]] | Answered by the human on 2026-10-11: the `metadata` key `delegate`, the role assigned to own the task (`metadata: {delegate: merge-queue}`), with the sisyphus filter [[filter-by-metadata]]. Not a frontmatter field of its own. The `owner` stays the agent instance that works on it. |
| Q2 | Where are the roles described: `hephaestus/roles/<role>.md`, `.claude/agents/`, or in code? | [[delegation-model]] | Answered by the human on 2026-10-11: each role is a Markdown file of its own, `hephaestus/roles/<role>.md`, readable by people and by any runtime. hephaestus can generate `.claude/agents/` files from it. |
| Q3 | May the merge queue move `main`, or only prepare a bookmark? | [[merge-queue-agent]] | Answered by the human on 2026-10-11: the merge queue never moves `main`. Its only job is to make a linear history from a set of tasks. |
| Q4 | What runs the agents: Claude Code subagents, the Claude Agent SDK, Claude Managed Agents, or local models? | [[delegation-runner]] | Answered by the human on 2026-10-11: AI agents start as Claude Code subagents, started by a human, then hephaestus on the Agent SDK. Local models only for roles that a task suite shows they can do. hephaestus also runs tool agents, which use no AI ([[#AI agents and tool agents]]). |
| Q5 | Which roles start by themselves, and on what trigger? | [[delegation-runner]] | Answered by the human on 2026-10-11, for now: the merge queue on a human request, later on a schedule; technical docs on each close; the others on a human request. |
| Q6 | Does the merge queue merge the work of humans too, or only of agents? | [[merge-queue-agent]] | Answered by the human on 2026-10-12: yes. Human work, and branches made beside the queue, land on the linear branch through the supervised merge queue: every task in the set, whoever did its work. |
| Q7 | Is hephaestus a Go program in the sisyphus module (`cmd/hephaestus`) or a module of its own in `hephaestus/`? | [[delegate-issues-to-agents]] | Answered by the human on 2026-10-12: a Go module of its own in `hephaestus/`, which runs as a CLI tool and as a background scheduler daemon ([[#The form of hephaestus]]). |
| Q8 | How does a tool agent get its task and report its result? | [[tool-agents]] | Answered by the human: the context on 2026-10-11, the result on 2026-10-12. A tool agent gets its context as a message, which any agent can send or broadcast ([[#Messages between workers]]). hephaestus runs the command in the workspace of the task with the message; exit code 0 closes the issue with the summary that the command prints; another exit code stops it with the output in `Notes`, tagged `needs-human`, unless the role file names an AI role to delegate the rest to. |
| Q9 | Can a tool agent be a reviewer in a referendum ([[referendum]])? | [[tool-agents]] | Answered by the human on 2026-10-11: yes. A check, for example a lint or the tests, votes with its exit code, `yes` for 0 and `no` for another code. A formatter can submit its changes as a modification. A tool reviewer never abstains. |
| Q10 | Is the merge queue a tool agent? Its order, rebases, renumbering, and checks follow fixed rules. | [[merge-queue-agent]] | Answered by the human on 2026-10-11: yes. It stops at a conflict that is not mechanical, records it, and sends a message to the engineer or a human. |
| Q11 | Where are messages kept: only in the message queue, or also in the repo? | [[message-queue]] | Answered by the human on 2026-10-11: in an append-only log. A message that matters to an issue is also recorded in its `Notes` by the receiver. |
| Q12 | How do messages travel? | [[message-queue]] | Answered by the human on 2026-10-11: through a service, the message queue, that the workers connect to. |
| Q13 | How is a receiver addressed? | [[message-queue]] | Answered by the human on 2026-10-12: by owner name (one worker), by role (any one worker of the role, or all of them, by a flag of the message), or `broadcast` (every listener; a listener can filter by kind). |
| Q14 | What does the message queue promise? | [[message-queue]] | Answered by the human on 2026-10-11: at-least-once delivery; each message has an id, so that a receiver can ignore a copy. On 2026-10-12: the messages of one sender arrive in the order that it sent them. |
