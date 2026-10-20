---
title: "Pin the GitHub Actions of the CI workflow to commit SHAs"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [ci, security]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Pin the GitHub Actions of the CI workflow to commit SHAs

## Summary

Pin each action of .github/workflows/ci.yml to the full commit SHA of its release, with the version in a comment, so that dependency-check/github-actions passes.

## Context

Found on 2026-10-12 by the first run of dependency-check/github-actions ([[dependency-check-roles]]) on this repo: .github/workflows/ci.yml:13 uses actions/checkout@v4, and line 15 uses actions/setup-go@v5. A tag can move to other code, so the workflow can run code that nobody reviewed.

## Acceptance criteria

- [ ] actions/checkout and actions/setup-go are pinned to the commit SHAs of their releases, with the version in a comment.
- [ ] hephaestus agent review dependency-check/github-actions passes on the repo.

- [ ] (proposed by triage) Each pinned SHA is verified against the published release of the action, so that the pin points at reviewed code.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, triage: checked .github/workflows/ci.yml; lines 13 and 15 still use actions/checkout@v4 and actions/setup-go@v5, so the issue is valid and not stale. Moved it from issue-files/ to repo/: issue-files/ is about the issue file format, and this issue is about the CI of the repo itself, like the closed repo/one-go-module. Not a duplicate; [[dependency-check-roles]] is the source of the finding, and [[github-intake-workflow]] is about another workflow. The criteria are checkable as written. Proposed extra criterion: the pinned SHAs are verified against the published releases of each action (compare with the tag on github.com), so that the pin does not point at unreviewed code. Proposal: priority medium (keep; security), role implementer.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
