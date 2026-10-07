---
title: "Add \"sisyphus lint\" to validate the whole issues/ tree at once"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Add "sisyphus lint" to validate the whole issues/ tree at once

## Summary

Add `sisyphus lint` to validate every issue in `issues/` at once: well-formed frontmatter, state
matching its directory, valid enum values, resolving wikilinks, and no broken or cyclic
relationships. Today each of these is only checked against the one issue a command happens to touch.

## Context

Part of [[issue-manager-gaps]]. The checks already exist piecemeal and just need to run across every
issue instead of one:

- State matches directory: `updateIssue` already detects and *corrects* this for the one issue it
  touches (`doc.get("state") != current.state`); `lint` should *report* it for every issue without
  changing anything (lint never writes).
- Valid `state`/`priority`/`effort`/`resolution` values: `oneOf` against `states`/`priorities`/
  `efforts`/`resolutions` (`cmd/sisyphus/issues.go`), applied to every issue's recorded field, not
  just a value a flag just set.
- Wikilinks resolve: `listLinks` (`links.go`) already reports `ok`/`missing`/`ambiguous`/
  `missing-heading` for one document; `lint` runs it over every Markdown file in the repo and fails
  on anything other than `ok`.
- `parent` points at a real, non-cyclic issue: the same logic `parentValueFor` uses to validate a
  new parent, applied to every issue's existing `parent` field.
- Issue name uniqueness: the same check `newIssue` runs before creating a file, applied across all
  existing files instead of one candidate name.

This also gives the [[github-action-sync-issues-to-github]] work and `.github/workflows/ci.yml` a
natural CI step: `sisyphus lint` can run alongside `go test` so a bad issue file fails CI the same
way a broken build does.

## Acceptance criteria

- [ ] `sisyphus lint` checks every issue for: frontmatter parses, state matches directory, enum
      fields hold valid values, `parent` resolves and has no cycle, and the issue name is unique.
- [ ] `sisyphus lint` checks every Markdown file in the repo for wikilinks that are not `ok`.
- [ ] Exit code is non-zero and problems are listed (one line per problem, with the file) if
      anything fails; zero and silent (or a short "N issues, 0 problems" line) otherwise.
- [ ] `sisyphus lint` never writes to any file.
- [ ] `.github/workflows/ci.yml` runs `sisyphus lint` as a step.
- [ ] Covered by tests for at least one failure of each kind above.

## Out of scope

- Auto-fixing problems; `lint` only reports. (`update` already auto-corrects a mismatched `state`
  field for the one issue it touches; that behavior is unchanged.)
- Linting the `design/decisions/` or `changelog/` content beyond their wikilinks.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example [[0.0.5]].
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
