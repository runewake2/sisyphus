---
role: merge-queue         # The id of the role: the name of this file, and the value of metadata delegate.
kind: tool                # ai: a model does the work. tool: a program does it, with no model.
command: hephaestus merge-queue   # A tool role only: the program that does the work of a task.
trigger: request          # request, close, or schedule, or a list of them.
on-failure: needs-human   # needs-human, or an AI role that gets the work that the tool cannot do.
can-call-referendum: false
---

# Merge queue

## Charter

Make one linear history from a set of tasks.

The merge queue merges the bookmark of each task, one at a time, in the order of [[merge-order-planner]], onto one line based on `main`, and gives each
unreleased changelog entry its version as the work lands. It never moves `main`, never pushes, and
does nothing else ([[agent-delegation#Merge queue]]).

## Inputs

- A set of tasks: sisyphus issues, each with the bookmark of its work, given in a message or
  delegated to `merge-queue` (`metadata: {delegate: merge-queue}`).
- The base of the merge, normally `main`.

## Outputs

- A bookmark `merge-queue/<date>` with the merged work of every task that it could merge, in one
  line.
- The versions of the unreleased changelog entries (`hephaestus version assign`), and decision
  records with unique numbers (`hephaestus decisions renumber`).
- A report in the `Notes` of each task: its place in the order, what was renumbered, or why its
  work was left out.

## Tools

- The `hephaestus` command: `merge-order`, `version assign`, and `decisions renumber`.
- jj or git, in a scratch workspace.
- The checks of the repo: `gofmt`, `go vet`, the tests, and `sisyphus lint`.
- `sisyphus` to read the tasks and to record the report.

## Ask a human

- A conflict that is not mechanical, for example two changes of one function: stop that task,
  record the conflict, and send a message to the `engineer` role or a human.
- A check that fails after a rebase: leave the task out, with the output in its `Notes`.
- A cycle of dependencies between the tasks.
