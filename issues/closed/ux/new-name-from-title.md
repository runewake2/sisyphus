---
title: "Let new make the issue name from the title"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [ux, cli, feature] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[ux/cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Let new make the issue name from the title

## Summary

`sisyphus new` needs a kebab-case name, so a user who types a title gets an error. Let `new` make the name from `--title` when no name is given, with the same rules as `sisyphus slug`.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo.

```
$ sisyphus new "Fix the login bug"
Error: Invalid issue name 'Fix the login bug'. Use 2-6 lowercase words in kebab-case, ...
```

`sisyphus slug` already turns a title into a valid, unique name, so today a user must run two commands, or type
the name and the title separately. Agents do the same (see `sisyphus slug` in the MCP tools).

## Acceptance criteria

- [ ] `sisyphus new --title "Fix the login bug"` without a name creates `fix-login-bug` (the name that `slug` gives), and prints its path.
- [ ] When the name argument is not a valid name but is valid title text, the error suggests the name that `slug` gives, for example `Did you mean: sisyphus new fix-login-bug --title "Fix the login bug"?`.
- [ ] `sisyphus_new` accepts a title without a name in the same way.
- [ ] [[sisyphus-new]] and tests cover both.

## Out of scope

- Asking the user questions interactively.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.71]]. `new` without a name makes it from `--title` as `slug` does, and a title given as the name gets the command to use instead.

