# Unreleased: the review of a tool role lives in one file

- Date: 2026-10-20
- Bookmark: samw/ai/split-toolagent-review
- Components: hephaestus
- Changelog: The referendum review of a tool role (Ballot, Review) moved from toolagent.go to review.go in the same package: the task lifecycle and the voting are two behaviors, and each now edits from one place.

## Summary

The readability engineer's trial run measured the spread of the five changes of 2026-10-20: four
were reasoned (an interface and its implementations; a package with its command, tests, docs,
changelog, and issue), and the one finding was toolagent.go, which mixed the task lifecycle with the
referendum review — two of five changes had their largest in-file gap there (605 and 87 unchanged
lines between related edits, against a screen of about 40).

## Changes

- `internal/hephaestus/toolagent/review.go` holds `Ballot`, `Review`, and `fileDiff`; `toolagent.go`
  keeps the task lifecycle. No exported name, caller, or behavior changed.

## Issues

- Closed [[split-toolagent-review]], filed by the readability-engineer role in its trial run. The
  trial table is in the Notes of [[readability-engineer-agent]], whose trial criterion a human can
  now check.
