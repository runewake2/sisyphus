---
title: "Add a readability engineer that condenses sprawling changes"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
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

# Add a readability engineer that condenses sprawling changes

## Summary

Add the AI role readability-engineer: it measures the spread of each change — the distance in lines
between its related edits, the files it touches, and the directories it spans — ranks the sprawling
changes, and proposes options for condensing the code under them, so that one behavior can be edited
from a single screen. On request, it makes one small move with no change of behavior.

## Context

Requested by a human on 2026-10-20: a role that works to make changes condensed, by measuring the
distance of line numbers between related changes, and ranking changes that span multiple files or
directories, or that leave large gaps of unchanged lines between changes. It suggests options for
condensing sprawling code, and strives for code that can be edited from a single screen. The role
file is [[roles/readability-engineer]]. The roles may be removed after a review; a trial run shows
how useful each is.

## Acceptance criteria

- [x] `hephaestus/roles/readability-engineer.md` describes the role, its measure of spread (the
  directories, the files, and the gaps of unchanged lines, against a screen of about 40 lines), and
  its options for condensing: move the helper next to its caller, group what changes together, split
  a file that mixes behaviors, join files that always change together, keep the boundaries.
- [x] Its Claude Code subagent is in `.claude/agents/`.
- [x] A trial run on a recent bookmark reports its spread as numbers, and proposes at least one
  option to condense, or says why the spread has a reason.
- [x] A human checks the findings of the run and decides whether the role stays.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, trial run of the readability engineer, on the five changes that landed today. Measure: directories spanned, files touched, largest gap of unchanged lines between two edits in one file (from the @@ hunk headers; a screen is about 40 lines). Rank: directories above files, files above gaps.

| Rank | Change | Description | Dirs | Files | Largest gap (file) |
|---|---|---|---|---|---|
| 1 | uvtlrzlq | share the code the abstraction engineer found repeated | 12 | 24 | 605 (toolagent/toolagent.go) |
| 2 | xtqvxxsq | merge with jj in a scratch workspace | 7 | 15 | 48 (cmd/hephaestus/watch.go) |
| 3 | olxlywlx | add evict | 7 | 13 | 95 (mq/server.go) |
| 4 | nxkqtouq | add doctor and kick | 7 | 11 | 0 |
| 5 | yyxnlklr | fix the three confirmed defects | 5 | 9 | 141 (mergequeue/line.go) |

Judgments:
- uvtlrzlq: the spread has reasons, taken per sub-change: each deduplication edits the shared home plus every user of a copy, and six issue files move open to closed. But the commit bundles six separate changes (six issues) and one drive-by (it deletes the leftover open copy of draw-graph-nodes-in-linear-time.md, whose closed copy its parent change made); per this role, edits with different reasons are separate changes to report. The 605-line gap in toolagent.go also shows that file mixes behaviors: filed as [[hephaestus/core/split-toolagent-review]].
- xtqvxxsq: reasoned. A new interface (line) with two implementations, and a migration of every command from raw git to repo.Repo. It condenses more than it spreads: it removed isAncestor, which lived in referendum.go but served status.go and watch.go. Not a finding.
- olxlywlx: mostly reasoned. The evict-specific code already lives in cmd/hephaestus/evict.go and mq/evict.go; the residual edits in mq/server.go (6 hunks over about 300 lines: field, constructor, route, two seam checks, the Renewal struct) each sit at the behavior they gate. toolagent.go took 6 hunks with a gap of 87 because the eviction threads Run, Handle, and handle; recorded as supporting evidence in [[hephaestus/core/split-toolagent-review]]. One data point for the server-verb pattern; no issue filed for it.
- nxkqtouq: fully reasoned, and the model shape: the feature in its own new package (internal/hephaestus/doctor), its command, tests, docs, changelog, and issue. Largest gap 0.
- yyxnlklr: reasoned. Three separate fixes bundled (separate changes to report), each condensed on its own: the kick fixes in kick.go plus test; the rollback fix threads an error through the Rollback contract and both implementations (the 141-line gap is the distance between jjLine.Rollback and gitLine.Rollback, a contract and its implementations), and through every leave call inside run(), where the gaps stay under 26. Not a finding.

Filed: [[hephaestus/core/split-toolagent-review]] (one real finding). The first condensing move, if asked: option 1 of that issue, move Ballot, Review, and fileDiff to toolagent/review.go, same package, no behavior change. No code changed in this trial.

## Resolution

Completed: a human decided on 2026-10-20 that the role stays, after its trial run on the five changes of the day. The trial reported the spread of each as numbers (directories, files, largest gap against a screen of about 40 lines), ranked them, ruled four spreads reasoned with the reason named, and filed one finding (split-toolagent-review), whose smallest option landed the same day. The ranking table is in the Notes.

