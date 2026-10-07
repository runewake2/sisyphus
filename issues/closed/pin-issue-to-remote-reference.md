---
title: "Add a field to pin a sisyphus issue to a remote reference (GitHub issue, Jira ticket)"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [schema, cli, github] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/pin-remote-reference # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-pin-remote-reference] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
agent-session:          # AI agent session id of the current agent working on the issue, if available.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[github-source-of-truth]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Add a field to pin a sisyphus issue to a remote reference (GitHub issue, Jira ticket)

## Summary

Add a frontmatter field that pins a sisyphus issue to the remote item that tracks it for people
outside the repo, for example a GitHub issue or a Jira ticket. This is the field that
[[github-action-issue-to-pr]] and [[github-action-sync-issues-to-github]] both read and write.

## Context

Part of [[github-source-of-truth]]. The issue template (`cmd/sisyphus/kit/issues/TEMPLATE.md.tmpl`)
and the frontmatter schema (`document.go` / `issues.go`) need a new optional field, for example
`remote:`, that holds a single URL (a GitHub issue URL or a Jira ticket URL). Keep it to one field
and one value for now: an issue has at most one remote reference.

Needs at least:

- A new frontmatter key, validated loosely (non-empty URL) the way other optional fields
  (`deferred-from`, `parent`) are validated.
- A CLI way to set it: a `--remote` flag on `sisyphus new`, and a way to set or clear it on an
  existing issue (either a new `sisyphus remote <name> <url>` command, mirroring `sisyphus parent`,
  or a `--remote`/`--clear-remote` pair of flags on `sisyphus update`). Pick whichever matches the
  existing command shape better.
- `sisyphus resolve` / `sisyphus links` do not need to follow this field; it points outside the repo.

## Acceptance criteria

- [ ] An issue's frontmatter can hold a `remote` field with a URL.
- [ ] There is a CLI command or flag to set and clear it, consistent with how `parent` is set
      (see `sisyphus parent`).
- [ ] `issues/TEMPLATE.md` documents the field with an example, the way the other fields are
      documented.
- [ ] Tests cover setting, clearing, and round-tripping the field.

## Out of scope

- Validating that the URL actually points at a real GitHub issue or Jira ticket.
- Following or rendering the remote reference anywhere other than the issue's own frontmatter.

## Notes

- 2026-10-07: Implemented as planned, choosing the dedicated-command option: `sisyphus remote <name>
  [<url>] [--clear]`, mirroring `sisyphus parent` exactly (including the "give one or the other, not
  both" error). Also added `--remote` to `sisyphus new`. Validation is loose, per the issue: the
  value must parse as a URL with a scheme and a host (`net/url`), nothing more.

## Resolution

Completed: issues can hold a `remote` URL, set via `sisyphus new --remote` or `sisyphus remote`.
See [[0.0.4]].
