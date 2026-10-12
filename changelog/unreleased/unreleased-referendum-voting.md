# Unreleased: Referendums

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: hephaestus runs referendums: role agents vote on the implementation of an issue in rounds, with required and non-required reviewers and a modification gate, and the result is recorded in the issue and in `referendums/`.

## Summary

hephaestus can call a referendum ([[referendum]]). Several role agents vote on the implementation of
an issue, each in its own scope, and agree or fail through one recorded procedure.

## Changes

- `hephaestus referendum call <issue>` takes these inputs:
  - `--question`;
  - `--required` and `--optional` reviewers;
  - `--runner` for AI reviewers;
  - `--bookmark`, `--onto`, `--max-rounds` (default 3), and `--parallel`.

  It runs the rounds to the result:
  - Every reviewer of a round votes at the same time, each in a scratch git worktree at the commit
    of the version, so that no reviewer sees another ballot.
  - Every modification of a round goes to the gate, each alone. Every required reviewer must accept
    it. The accepted ones together make one new version on the branch `referendum/<id>/v<n>`, and
    every reviewer votes again. One that conflicts with an earlier one goes back to its author.
  - Every required `yes` passes. One required `no` fails, also on an edited version (R7). A
    non-required `no` is a concern.
  - A required reviewer whose modification the gate rejected votes again (R8).
  - After the maximum rounds, or when a reviewer gives no valid ballot, the result is `no-result`,
    and the issue gets the tag `needs-human`.
- Tool reviewers vote with their exit code, and modify with the changes of their command. AI
  reviewers get a ballot prompt and end their output with a line of JSON.
- The record:
  - `metadata` `referendum` and `referendum-record`;
  - a Notes line for each round and for the result;
  - the full record in `referendums/<date>-<issue>.md`;
  - a `referendum` message to the owner when the message queue runs.

  When the referendum passes on a modified version, its issue edits are applied, and the bookmark of
  the issue becomes the branch of that version.
- Only a human, or a worker whose role file has `can-call-referendum: true`, can call a referendum.
- `hephaestus referendum decide <issue> passed|failed --reason` lets a human decide after
  `no-result`, or overrule a result. The record shows the decision.
- The engineer role calls its referendums with `hephaestus referendum call`.

## Decisions

- [[0001-a-referendum-runs-each-reviewer-in-a-scratch-worktree]]

## Issues

- Implements [[referendum-voting]]. A test run with three role agents on a real bookmark, checked by
  a human, is still to do.
