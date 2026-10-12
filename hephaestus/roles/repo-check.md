---
role: repo-check
kind: tool
command: cd "$(git rev-parse --show-toplevel 2>/dev/null || jj root)" && test -z "$(gofmt -l cmd hephaestus)" && go vet ./... && (cd hephaestus && go vet ./...) && sisyphus lint && echo "gofmt, go vet, and sisyphus lint are clean in both modules and the repo."
trigger: request
on-failure: needs-human
listens: [bookmark-ready]
---

# Repo check

## Charter

Run the checks of this repo on the working copy of a task: gofmt, `go vet` of the sisyphus and the
hephaestus modules, and `sisyphus lint` of the repo.

It changes nothing. It is a tool role: it uses no model.

## Inputs

- A task message, with the issue to check and, in its data, the workspace to check.
- A broadcast of the kind `bookmark-ready`.

## Outputs

- The issue of the task, closed with the output of the checks, or stopped with the output in its
  `Notes` and tagged `needs-human`.
- A result message to the sender.

## Tools

jj, gofmt, go vet, and sisyphus.

## Ask a human

When a check fails: the output in `Notes` tells what to correct.
