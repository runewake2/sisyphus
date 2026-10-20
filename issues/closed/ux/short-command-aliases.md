---
title: "Establish sis and forge as the short names of the two commands"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [ux]             # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed: 2026-10-20     # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Establish sis and forge as the short names of the two commands

## Summary

Typing sisyphus and hephaestus is slow. The short names are sis and forge, as symlinks (they work in scripts and agent runners, where shell aliases do not expand), with one completion line per alias. The docs and the role files keep the full names, so transcripts stay greppable. On a machine with Foundry, its forge collides; the README names hep as the fallback.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [x] The README of the repo has an Aliases section: the symlinks, the completion lines, the full-names rule, and the Foundry collision note.
- [x] The hephaestus README names the short names and points at the section.
- [x] sis and forge run on this machine (symlinks in ~/.local/bin).

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Done: a human chose sis and forge on 2026-10-20 (forge for the god of the forge, with hep as the documented fallback where Foundry owns the name). The Aliases section of README.md has the symlinks and the completion registration per alias; hephaestus/README.md points at it. Installed in ~/.local/bin of this machine.

