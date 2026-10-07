---
title: "GitHub Action: open a PR with a matching sisyphus issue for new GitHub issues"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [github, actions] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/issue-to-pr # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-issue-to-pr] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "68fa9a09-fe4e-4ae8-9917-086c3313c5be"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[github-source-of-truth]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# GitHub Action: open a PR with a matching sisyphus issue for new GitHub issues

## Summary

Add a GitHub Actions workflow that reacts to a newly filed GitHub issue by opening a PR that adds a
matching sisyphus issue, pinned to the GitHub issue that triggered it.

## Context

Part of [[github-source-of-truth]]. Humans who do not use sisyphus directly still need to be able to
file work by opening a GitHub issue. This workflow is the intake path:

1. Trigger on the `issues: opened` event.
2. Build a sisyphus issue from the GitHub issue: title from the GitHub issue title, `Summary` /
   `Context` from its body, `remote` set to the GitHub issue's URL (see
   [[pin-issue-to-remote-reference]]), and a generated issue name (slug of the title, made unique).
   This likely calls a new `sisyphus` subcommand rather than hand-rolling the frontmatter in the
   workflow (for example `sisyphus new <name> --remote <url> --title <title>`), so the logic for
   writing a valid issue lives in one place.
3. Commit the new file on a new branch and open a PR with `gh pr create`, so a human reviews and
   merges it like any other change to `issues/`.
4. Comment on the original GitHub issue with a link to the PR, so the filer can see their issue was
   picked up.

Only `issues: opened` needs to be handled for now; edits to the GitHub issue after the PR exists are
out of scope (see below).

## Acceptance criteria

- [ ] Opening a GitHub issue triggers a workflow run.
- [ ] The workflow opens a PR that adds one well-formed sisyphus issue file under `issues/open/`,
      pinned to the triggering GitHub issue via [[pin-issue-to-remote-reference]].
- [ ] The original GitHub issue gets a comment linking to the PR.
- [ ] The workflow does not run (or fails loudly rather than writing something broken) on a repo
      that was never initialized with `sisyphus init`.
- [ ] Covered by at least one test or dry-run that does not require actually filing a live GitHub
      issue (for example a script invoked with a recorded webhook payload).

## Out of scope

- Keeping the sisyphus issue in sync with further edits or comments on the GitHub issue after the
  PR is opened.
- Closing the GitHub issue automatically when the sisyphus issue closes.
- [[github-action-sync-issues-to-github]], the opposite direction.

## Notes

- 2026-10-07: Implemented as planned, plus two small CLI additions the workflow needed:
  `sisyphus slug <text>` (new command) turns a GitHub issue title into a unique, valid issue name
  (`slugify`/`uniqueSlug` in `issues.go`), and `sisyphus new --context <text>` (new flag, using the
  new `replaceSection` helper in `links.go`) fills the Context section from the GitHub issue body,
  instead of hand-rolling frontmatter in the workflow as the issue suggested avoiding.
- The workflow itself (`.github/workflows/issue-to-pr.yml`, added to the kit so every initialized
  repo gets it) is YAML I could not execute in this sandbox (no live GitHub webhook, no push
  access to test with); it is reviewed by inspection, not by a dry run, same as `ci.yml` earlier.
  The two sisyphus-side pieces it depends on (`slug`, `new --context`) are unit tested.

## Resolution

Completed: `.github/workflows/issue-to-pr.yml` opens a PR with a matching, pinned sisyphus issue
when a GitHub issue is filed, using the new `sisyphus slug` and `sisyphus new --context`. See
[[0.0.9]].
