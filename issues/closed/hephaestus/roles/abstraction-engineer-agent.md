---
title: "Add an abstraction engineer that promotes code reuse and clear abstractions"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed: 2026-10-20     # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add an abstraction engineer that promotes code reuse and clear abstractions

## Summary

Add the AI role abstraction-engineer: it reviews each new change for code that the repo already has, scans the repo for code that repeats, and proposes abstractions as issues, with their boundary, contract, and cost. On request, it makes one small abstraction with no change of behavior.

## Context

Requested by a human on 2026-10-12: an agent whose entire objective is clear abstractions and code reuse. It promotes reuse of code and shared implementation where the boundary makes sense, and builds sensible abstractions when code reuse occurs often. The role file is [[roles/abstraction-engineer]]. The roles may be removed after a review; a trial run shows how useful each is.

## Acceptance criteria

- [x] `hephaestus/roles/abstraction-engineer.md` describes the role and its rules: reuse first, three or more copies, one reason to change, respect the boundaries, clearer and not only shorter, behavior stays.
- [x] Its Claude Code subagent is in `.claude/agents/`.
- [x] A trial run on `internal/hephaestus` finds the known copies below, and files issues with no duplicates.
- [x] A human checks the issues of the run and decides whether the role stays.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-12, from claude: known copies in hephaestus, for the trial run to find: the scratch git worktree add and remove (mergequeue.Run and referendum.RoleVoter; only the second got the fix for the race of two reviewers), the git committer identity fallback (mergequeue.identity and referendum.identity), tail (toolagent and mergequeue), lines (repo and mergequeue), indent (airole and referendum), the shell of a check (mergequeue.runChecks repeats toolagent.Shell, without the stop of the process group), and toolagent.contains (slices.Contains).

2026-10-20, from the abstraction engineer (trial run on internal/hephaestus and cmd/hephaestus): I verified the known copies of 2026-10-12 against the code of today, found new ones, and filed seven issues, each for one abstraction.

The known copies, today: (1) The scratch worktree add/remove is now package scratch, and referendum.RoleVoter uses it — but the merge queue was rewritten (mergequeue/line.go), and gitLine (line.go:217-233, 332-336) repeats the lifecycle again, without the lock against the prune race: [[hephaestus/core/gitline-reuses-scratch]]. (2) The identity fallback has three copies now: scratch.Identity (scratch.go:120), line.go identity (263), and line.go jjIdentity (102), with different fallback names and lookup scopes: [[hephaestus/core/one-vcs-identity-fallback]]. (3) runChecks (mergequeue.go:260) still repeats toolagent.Shell (toolagent.go:670) without the process-group stop: [[hephaestus/core/share-platform-shell]]. (4) lines has two copies (repo.go:82 with ToSlash, mergequeue.go:91 without): part of [[hephaestus/core/export-repo-command-plumbing]]. (5) indent (airole.go:95, referendum/prompt.go:101) is identical, and referendum already imports airole: [[hephaestus/core/export-airole-indent]]. (6) toolagent.contains (toolagent.go:759) still repeats slices.Contains: [[hephaestus/core/use-slices-contains]].

New copies the list did not have: (7) exportJJ/importJJ in three places (toolagent.go:467-480, cmd/hephaestus/referendum.go:350-370, inline in cmd/hephaestus/mergequeue.go:97-102, where the error is dropped): [[hephaestus/core/share-jj-export-import]]. (8) Six command runners with the same error format (repo.run, scratch.Git, taskspace.run, jjLine.jjIn, gitLine.git, gitLine.gitDirCmd, plus revParse in cmd/hephaestus/referendum.go:372): [[hephaestus/core/export-repo-command-plumbing]]. (9) ago() twice, identical (doctor.go:166, the closure in status.go:255): [[hephaestus/core/share-ago-format]].

Copies that I do not propose to share, and why: tail (toolagent.go:682 with the "(N earlier lines are not shown)" marker, mergequeue.go:275 without) — two copies with different outputs; one function with a flag would be two functions, and a silent unification changes an output; unify only when the queue wants the marker too. jjLine workspace add (line.go:68-79) against taskspace.Create (taskspace.go:53-84) — the shared part is three lines around jj workspace add; the contracts differ (a named task workspace that is reused, against a temp scratch workspace that rides the tip), and jj workspace forget has no prune race; two reasons to change. The test helper contains(t, s, substring) in three _test.go files (cmd/hephaestus/helpers_test.go:115, status_test.go:13, referendum/referendum_test.go:115) — three copies of six lines, but sharing needs an exported test-support package that couples the test code of three packages; the cost is above the gain. The seconds-to-duration one-liner (toolagent.go:496, cmd/hephaestus/trial.go:320) — two one-liners, below the bar. Boundary check: every proposed home (scratch, repo, airole, mq, a new shell package) is inside internal/hephaestus; nothing is shared with the sisyphus CLI, so [[design/decisions/cli/0029-sisyphus-does-not-manage-agents|0029]] stays respected.

The first move I would make: [[hephaestus/core/gitline-reuses-scratch]] — it is small, it removes the largest exact copy, and it closes a real race (the lock is the fix that one copy got and the other did not). I changed no code in this run.

## Resolution

Completed: a human decided on 2026-10-20 that the role stays, after its trial run on internal/hephaestus and cmd/hephaestus. The trial found the known copies of the Notes (those still live) and new ones, filed seven issues with no duplicates, and ruled four copy groups not worth sharing, with reasons. Five of its proposals landed the same day (gitline-reuses-scratch, one-vcs-identity-fallback — which exposed a latent flag-order defect — share-platform-shell, use-slices-contains, share-ago-format, export-airole-indent); export-repo-command-plumbing and share-jj-export-import stay open with their dependencies recorded.

