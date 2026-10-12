# Unreleased: hephaestus status

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: `hephaestus status` shows in one view what needs a human: issues, referendums, workers, waiting messages, and the last line of the merge queue with the recommendation of the release manager.

## Summary

A human no longer has to read four sources to know what waits for them.

## Changes

- `hephaestus status` (and `--json`) shows:
  - the open and in-progress issues tagged `needs-human`;
  - the referendums that run (with their deadline), that are past their deadline, that failed on an
    open issue, or that have no result, each with the next action;
  - the workers of the message queue, live or stopped (no keep-alive for `--stale`, 2 minutes by
    default);
  - the messages that wait, by receiver, and whether a live worker can take them;
  - the last line of the merge queue, whether the base has it, and the last recommendation of the
    release manager for it.
- The message queue answers `GET /pending`: each delivery that waits, also for a worker or a role
  that has not started.
- The release manager sends its recommendation as a `release-recommendation` broadcast.
- `hephaestus messages log` and `status` share one reader of the message log.

## Issues

- Closed [[status-command]].
