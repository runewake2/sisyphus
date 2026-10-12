---
title: "Renumber versions, changelogs, and decision records when work merges"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [merge-queue, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/merge-queue/merge-queue-agent]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Renumber versions, changelogs, and decision records when work merges

## Summary

When two lines of work pick the same next SHAME version or decision record number, the second one to merge must renumber them. Add a tool that does this in one step, so that the merge queue, and a human, can resolve these conflicts with no hand edits.

## Context

[[CONTRIBUTING#Changelog]] says: when you rebase onto `main`, renumber your changelog file, your `CHANGELOG.md` entry, `VERSION`, and the `Version:` lines of your decision records to the next free SHAME version. In practice, this also touches:

- every `[[0.0.N]]` link to the renumbered versions, in issues, changelogs, and docs,
- decision record files `NNNN-<title>.md` that have the same number as one on `main`, and the links to them,
- the order of entries in `CHANGELOG.md`.

These are the rules of the repo, not of issues, so this is hephaestus code, not sisyphus. `sisyphus rename` changes every link to a renamed issue in the same way, but only for issues. A possible form: `hephaestus renumber --onto main`, which finds the versions and decision numbers of the current line that `main` also has, and moves them to the next free numbers.

## Acceptance criteria

- [ ] One command renumbers the versions and decision records of a line of work that collide with `main`, and every link to them.
- [ ] `sisyphus lint` passes after it, and the tests cover a collision of each kind.
- [ ] It has a dry run that shows each change.

## Out of scope

- Conflicts in code.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.108]]. A human decided that versions come from the merge queue ([[0001-versions-come-from-the-merge-queue]]), so versions no longer collide: `hephaestus version assign` gives each unreleased changelog entry its version in the order of the merge, with every link to it. Decision record numbers still can collide; `hephaestus decisions renumber --onto main` renumbers the record that the base does not have.

