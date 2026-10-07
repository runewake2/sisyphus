---
title: "Add sisyphus migrate to upgrade every issue to the current schema"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli, schema]    # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add sisyphus migrate to upgrade every issue to the current schema

## Summary

Add `sisyphus migrate` to crawl every issue in the repo and bring its frontmatter up to date with
the schema of the installed sisyphus version: add missing fields, carry renamed fields over, and
record which version last migrated the issue in its `metadata`. Today a schema change means a
hand-written `sed` over every issue file.

## Context

Part of [[issue-manager-gaps]]. Motivating case: replacing `agent-session` with `metadata` (see
[[agent-session-to-metadata]]) was done by hand, with `sed` over every file in `issues/`. Every
repo that uses sisyphus would have to repeat that, and the next schema change will need it again.

Suggested design:

- **Target schema**: the `issues/TEMPLATE.md` embedded in the running binary (the `kit`), not the
  repo's own `issues/TEMPLATE.md`, which is exactly what may be out of date. `migrate` also brings
  the repo's `issues/TEMPLATE.md` up to date.
- **Missing fields**: add every frontmatter field the template has and the issue lacks, with the
  template's default value and inline comment, at the template's position. `document.set` only
  appends a new key at the end of the frontmatter, so this needs a small insert-at-position helper.
- **Renamed or replaced fields**: a table of known migrations, applied in order, for example
  `agent-session: <id>` -> `metadata: {session-id: "<id>"}`. A value is moved, never dropped.
- **Unknown fields**: fields neither the template nor a migration knows about are kept as they
  are, and reported, so custom fields are never lost.
- **Version stamp**: record `sisyphus-version` in each migrated issue's `metadata` (from
  `sisyphus.Version()`), so a later `migrate` can tell which migrations an issue still needs. Open
  question: is a per-issue stamp worth the noise, compared with one stamp for the whole repo?
- **Safety**: `--dry-run` prints what would change per file and writes nothing. The body of an
  issue is never touched, inline comments stay (as `document.set` already keeps them), and
  running `migrate` twice changes nothing the second time.

## Acceptance criteria

- [ ] `sisyphus migrate` adds every missing template field to every issue, at the template's
      position, with the template's default value and comment.
- [ ] A known renamed field (at least `agent-session` -> `metadata.session-id`) is migrated with
      its value kept.
- [ ] Unknown fields are kept and reported, never removed.
- [ ] The repo's `issues/TEMPLATE.md` is updated to the binary's template.
- [ ] `--dry-run` reports the changes and writes nothing.
- [ ] Running `migrate` twice in a row makes no changes the second time.
- [ ] Bodies and inline comments are unchanged.
- [ ] Also available as an MCP tool in `sisyphus-mcp` ([[sisyphus-mcp-server]]).
- [ ] Covered by tests for each item above.

## Out of scope

- Migrating anything outside `issues/`.
- Downgrading to an older schema.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
