---
title: "Turn on shell completion for commands, flags, and issue names"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [ux, cli, feature] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Turn on shell completion for commands, flags, and issue names

## Summary

Shell completion is turned off (`CompletionOptions.DisableDefaultCmd = true`). Turn it on, and complete issue names, states, priorities, efforts, and tags, so that a user does not have to type or look up long names.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. Issue names are long kebab-case strings (`clarify-issue-name-descriptions`), and every command
that changes an issue needs one. cobra generates completion scripts for bash, zsh, fish, and PowerShell, and
`ValidArgsFunction` and `RegisterFlagCompletionFunc` can return the issue names from the index (only open and
in-progress ones for `update`, for example) and the valid enum values.

## Acceptance criteria

- [ ] `sisyphus completion bash|zsh|fish|powershell` prints a completion script.
- [ ] Commands that take an issue complete its name; `--state`, `--priority`, `--effort`, and `--resolution` complete their values; `--tags` completes the tags in use.
- [ ] README tells how to install the completion.
- [ ] A test checks the completion of an issue name.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
