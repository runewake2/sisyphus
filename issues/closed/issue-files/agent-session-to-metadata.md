---
title: "Move agent-session to a generic metadata field"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli, schema]    # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/agent-session-metadata # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-agent-session-metadata] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Move agent-session to a generic metadata field

## Summary

Replace the explicit `agent-session` field with a generic `metadata` key-value field. Agents can
still record a session id in it, but it is no longer a named, hard-coded part of the schema, and it
is no longer auto-cleared on every state change: it persists so work can be resumed with context
later, including after an issue closes or reopens.

## Context

A real session id had been getting baked permanently into every closed issue's committed
frontmatter (ten of them, in this repo), with no way to opt out short of hand-editing. Raised as a
concern while preparing the repo for being public. A generic, opt-in `metadata` field fixes this at
the design level: nothing sets it unless an agent explicitly asks to.

- `metadata: {key: "value", ...}`, parsed/formatted by `parseMetadata`/`formatMetadata`
  (`issues.go`), the same comma-separated, one-value-no-commas shape as `tags`/`depends-on`.
- `sisyphus new --metadata key=value` and `sisyphus update --metadata key=value` (repeatable, or
  comma-separated, via cobra's `StringToStringVar`). `update` merges into the existing map (adds or
  overwrites a key); nothing clears a key automatically.
- Migrated every existing issue file's `agent-session:` line to `metadata: {}`, and the ten real
  session ids already scrubbed (blanked) in an earlier change stay blank under the new field.

## Acceptance criteria

- [x] `agent-session` is gone from the schema, the CLI flags, `issueView`, and `writeIssueText`.
- [x] `metadata` holds arbitrary key-value notes, settable via `--metadata` on `new` and `update`.
- [x] `update --metadata` merges into the existing map rather than replacing it.
- [x] Nothing clears `metadata` automatically on any state change.
- [x] Every existing issue file migrated from `agent-session:` to `metadata: {}`.
- [x] `CONTRIBUTING.md` (and the kit template) documents the field.

## Out of scope

- A way to remove a single metadata key (only add-or-overwrite is supported; revisit if needed).
- Validating metadata values in any way; they are free-form text.

## Notes

- 2026-10-07: Implemented as planned, superseding the previous fix (clear `agent-session` when not
  in-progress) from the same conversation: that fix is moot now that the field itself is gone.

## Resolution

Completed: `agent-session` replaced by a generic `metadata` field. See [[0.0.10]].

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
