---
title: "GitHub Action: turn sisyphus issues into GitHub issues on main"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [github, actions] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/sync-to-github # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-sync-to-github] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "68fa9a09-fe4e-4ae8-9917-086c3313c5be"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[github-source-of-truth]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# GitHub Action: turn sisyphus issues into GitHub issues on main

## Summary

Add a GitHub Actions workflow that runs on `main` and creates or updates a GitHub issue for every
sisyphus issue, so sisyphus stays the source of truth while GitHub still shows current, readable
issues for anyone who looks there instead of at `issues/`.

## Context

Part of [[github-source-of-truth]]. This is the mirror direction, opposite of
[[github-action-issue-to-pr]]:

1. Trigger on push to `main` (or on PR merge to `main`).
2. For each sisyphus issue changed in the pushed commits (or, for a simple first version, every open
   and in-progress issue):
   - No `remote` field yet ([[pin-issue-to-remote-reference]]): create a GitHub issue from it, then
     write the new GitHub issue's URL back into the sisyphus issue's `remote` field in a follow-up
     commit to `main`.
   - Already has a `remote` field pointing at a GitHub issue: update that issue's title, body, and
     open/closed state from the sisyphus issue.
3. Sisyphus content always overwrites the GitHub issue, never the reverse: if someone edits the
   mirrored GitHub issue directly, the next sync overwrites their edit. Say so in the GitHub issue
   body so it is not a surprise.
4. Closing a sisyphus issue (state `closed`) closes its mirrored GitHub issue, with a note on the
   resolution (completed or abandoned).

Like [[github-action-issue-to-pr]], this probably wants a `sisyphus` subcommand or flag that lists
issues needing a sync (new or changed since a given point) rather than reimplementing frontmatter
parsing in the workflow.

## Acceptance criteria

- [ ] Pushing a new sisyphus issue to `main` creates a matching GitHub issue and records its URL in
      the sisyphus issue's `remote` field.
- [ ] Pushing a change to an already-mirrored sisyphus issue updates the matching GitHub issue
      (title, body, open/closed state).
- [ ] Closing a sisyphus issue closes its mirrored GitHub issue.
- [ ] The mirrored GitHub issue's body says it is generated from sisyphus and should not be edited
      directly.
- [ ] The workflow is idempotent: running it again with no new sisyphus changes makes no GitHub
      changes.

## Out of scope

- Reading edits back from GitHub into sisyphus (sisyphus always wins; see [[github-source-of-truth]]).
- Mirroring to Jira or any tracker other than GitHub issues.
- [[github-action-issue-to-pr]], the opposite direction.

## Notes

- 2026-10-07: Implemented without a new sisyphus subcommand: the workflow composes existing
  commands (`sisyphus list --state open,in-progress,closed --json` for names, `sisyphus show <name>
  --json` for title/body/state/resolution/remote, `sisyphus remote` to record a newly created
  issue's URL) plus `jq` and the `gh` CLI, rather than adding Go surface for something that is
  mostly "call gh, then call sisyphus remote". This did require fixing a real gap first: `sisyphus
  show --json` (and the human-readable form) was missing `remote` and `depends-on` entirely, added
  to `issueView`/`writeIssueText`.
- "Every open, in-progress, or closed issue, every run" rather than true changed-since-last-push
  diffing, per the issue's explicitly allowed simpler first version. `gh issue edit`/`close`/`reopen`
  are themselves idempotent, so re-running with no sisyphus changes makes no GitHub changes.
- Like [[github-action-issue-to-pr]], the YAML itself is reviewed by inspection; the `remote`/
  `depends-on` JSON fields it depends on are unit tested.

## Resolution

Completed: `.github/workflows/sync-to-github.yml` creates or updates a GitHub issue for every
sisyphus issue on push to `main`, recording new issues' URLs back in `remote`. See [[0.0.9]].
