---
title: "Add \"sisyphus show\" to print a single issue's metadata and body"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/show-command # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-show-command] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
agent-session:          # AI agent session id of the current agent working on the issue, if available.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Add "sisyphus show" to print a single issue's metadata and body

## Summary

Add `sisyphus show <name>` to print one issue: its frontmatter fields and its Markdown body,
without the caller first having to find which of `issues/open`, `issues/in-progress`, or
`issues/closed` holds it.

## Context

Part of [[issue-manager-gaps]]. `loadIssue` (`cmd/sisyphus/issues.go`) already resolves a reference
(name, `[[name]]`, `#name`, or a path) to a file and parses it; `show` is mostly a thin command
around that plus a renderer, much like `resolve` and `links` are thin commands around existing
lookup logic.

Suggested shape:

```bash
sisyphus show explicit-step-dependencies     # human-readable: fields, then the body
sisyphus show "[[explicit-step-dependencies]]"
sisyphus show --json explicit-step-dependencies   # frontmatter as a JSON object, body as a string
```

The human-readable form should be easy to scan: print the key fields (title, state, priority,
effort, tags, owner, approver, bookmark, parent) and then the body, rather than dumping the raw
frontmatter block.

## Acceptance criteria

- [ ] `sisyphus show <name>` accepts every reference form `update`/`parent` already accept (name,
      `[[name]]`, `#name`, path).
- [ ] Prints the issue's fields and body in a readable form.
- [ ] `--json` prints the frontmatter fields and body as JSON.
- [ ] A missing or ambiguous issue produces the same errors as `update` does today (reuses
      `findIssue`/`loadIssue`, not a new lookup).
- [ ] Covered by tests, including the missing-issue and ambiguous-issue error cases.

## Out of scope

- Rendering wikilinks or Markdown to HTML; this prints the raw Markdown body.
- Editing the issue; this is read-only (see [[update-editable-metadata]] for edits).

## Notes

- 2026-10-07: Implemented as planned. `issueView` (`issues.go`) holds every frontmatter field plus
  the body; `newIssueView` builds it from a loaded `document`, reusing `parseList` for `tags` and
  `workspaces`. `showCommand` (`cli.go`) is a thin wrapper around `loadIssue`, matching `resolve` and
  `links`. Text output (`writeIssueText`) prints the nine key fields the issue specified, aligned by
  label width, then the body with its leading blank lines trimmed.

## Resolution

Completed: added `sisyphus show <name>` (and `--json`). See [[0.0.1]].
