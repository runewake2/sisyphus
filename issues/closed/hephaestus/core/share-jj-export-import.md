---
title: "One place exports and imports the jj bookmarks for git"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed: 2026-10-20     # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[hephaestus/core/export-repo-command-plumbing]]", "[[hephaestus/core/evict-command]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# One place exports and imports the jj bookmarks for git

## Summary

Three places run jj git export, and two run jj git import, so that git and jj see the same bookmarks in a colocated repo. The copies:

- internal/hephaestus/toolagent/toolagent.go:467-474 (exportJJ) and 476-480 (importJJ): no VCS guard; the callers at 348-352, 417-421, and 457-459 each check a.Repo.VCS == repo.JJ first.
- cmd/hephaestus/referendum.go:350-370 (exportJJ, importJJ): the same code, with the guard r.VCS != repo.JJ inside.
- cmd/hephaestus/mergequeue.go:97-102: an inline jj git export after a merge, with the guard at the caller and the error dropped.

The proposal: two functions on repo.Repo in internal/hephaestus/repo, for example ExportBookmarks() error and ImportBookmarks(): in a jj repo, run jj git export (or import) in the root; in a git repo, do nothing and return nil. Export returns the error with the combined output, as both copies format it today ("jj git export: %v: %s"); Import stays best-effort, as both copies are today. The three copies and their guards move to the methods.

Boundary: repo is below toolagent and is imported by cmd/hephaestus already; nothing is shared with the sisyphus CLI ([[design/decisions/cli/0029-sisyphus-does-not-manage-agents|0029]] stays respected). What it removes: ~30 lines and three copies of a contract that must stay one — when the guard or the error handling of the export changes, every caller must change with it (the queue copy already dropped the error silently). What it costs: the best-effort import and the strict export become a fixed contract; mergequeue.go's inline export today ignores the error, and the caller decides again whether to keep ignoring it (keep the behavior: ignore it, with a comment).

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] repo.Repo has ExportBookmarks and ImportBookmarks, with tests in a jj repo and a git repo (the git repo is a no-op).
- [ ] toolagent, cmd/hephaestus/referendum.go, and cmd/hephaestus/mergequeue.go have no exec of jj git export or jj git import of their own.
- [ ] Behavior stays: the same guards, the same error text, and the queue still ignores the export error after a merge.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: this issue now depends on [[export-repo-command-plumbing]] and [[evict-command]]. Reasons: it removes exportJJ and importJJ from toolagent.go, and hephaestus evict is in work in toolagent now; and it changes cmd/hephaestus/referendum.go next to revParse, which the plumbing issue also moves. The new methods build on the exported repo.Run.

## Resolution

Done: repo.Repo.ExportBookmarks and ImportBookmarks are the one place that exports and imports the jj bookmarks for git, guarded inside (a no-op outside a jj repo), with tests in a jj and a git repo. toolagent, referendum.go, trial.go, and mergequeue.go have no jj git export or import of their own; the merge queue still ignores the export error after a merge, with the reason in a comment. The methods sit on the private runner of the repo package, so the dependency on export-repo-command-plumbing did not block this direction; the plumbing move can rebase on them.

