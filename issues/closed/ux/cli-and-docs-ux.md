---
title: "Make the sisyphus CLI and its documentation easier to learn and use"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [ux, cli, docs]  # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
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

# Make the sisyphus CLI and its documentation easier to learn and use

## Summary

sisyphus works, but a new user has to learn it from the source: the help has almost no examples, common actions need long commands, write commands give little output, and the README is written for this repo, not for people who use sisyphus in theirs. This epic tracks the improvements to the CLI and the documentation.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo.

The review tried a first session as a new user (`list` before `init`, `init`, `new` with a title, start and close
an issue, `show`, `--state all`, `--json` on `new`, `version`), compared the flags in `--help` with every page in
`docs/`, and read the README and the docs as a reader on GitHub would.

What works well: the flags in the help and in `docs/` agree, error messages say what is wrong and what is valid,
and every read command has `--json`.

The sub-issues are in two groups:

- CLI: [[new-name-from-title]], [[workflow-shortcut-commands]], [[state-all-and-value-suggestions]],
  [[shell-completion]], [[json-output-for-write-commands]], [[help-examples-and-groups]],
  [[readable-show-output]], [[list-fits-terminal]], [[open-issue-in-editor]], [[vcs-neutral-defaults]],
  [[guide-new-users]], [[global-directory-flag]].
- Documentation: [[readme-for-users]], [[issue-format-guide]], [[output-conventions-doc]],
  [[consistent-command-pages]], [[mcp-setup-and-examples]].

Related open work: [[suggest-names-on-missing-issue]] and [[clarify-issue-name-descriptions]] (error messages and
tool descriptions), and [[ollama-task-manager-guide]] (a guide for local models).

## Acceptance criteria

- [x] All sub-issues are closed or deferred with a reason.

## Out of scope

- The plugin work ([[integration-plugins]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.68]] to [[0.0.84]] and [[0.0.103]]. All 18 sub-issues are closed: the CLI has workflow shortcuts, completion for five shells, JSON for write commands, grouped help with examples, readable show and list output, an editor form of edit, a global -C flag, and consistent command pages, and the docs have a user README, an issue format guide, MCP setup, and the output conventions.

