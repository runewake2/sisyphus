---
title: "Rewrite the GitHub issue intake workflow on the GitHub plugin"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [github, actions] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[integration-plugins]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[github-plugin]]", "[[plugin-sync-commands]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Rewrite the GitHub issue intake workflow on the GitHub plugin

## Summary

Rewrite the removed `issue-to-pr` workflow on top of the GitHub plugin: when an owner or org member
opens a GitHub issue, open a PR that adds the matching, pinned sisyphus issue.

## Context

Part of [[integration-plugins]]. The workflow was removed until the plugin exists (decision P4).
The old version ([[github-action-issue-to-pr]]) built the issue itself with `sisyphus slug` and
`sisyphus new`; the new one calls `sisyphus pull "$ISSUE_URL"` and lets [[github-plugin]] do the
translation.

Keep the protections of the old version:

- Run only for issues whose author is the repo owner or an org member (`author_association`).
- Pass every untrusted value (title, body, URL) through `env:`, never as `${{ }}` inside `run:`.

It lives in this repo's `.github/workflows/`; `sisyphus init` does not ship it.

## Acceptance criteria

- [ ] An issue opened by an owner or member results in a PR that adds the pinned sisyphus issue,
      and a comment on the issue that links to the PR.
- [ ] An issue opened by anyone else does nothing.
- [ ] No untrusted value is interpolated into a shell script.

## Out of scope

- Mirroring sisyphus issues back to GitHub ([[github-sync-workflow]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
