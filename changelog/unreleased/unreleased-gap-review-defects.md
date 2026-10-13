# Unreleased: hephaestus gap review fixes, and role trials

- Date: 2026-10-13
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Bookmarks at one commit no longer stop the merge queue, `merge-queue` exits 2 when it leaves work out, duplicate role names are errors, and `hephaestus agent trial` tries a role on one issue in a scratch copy and records its cost.

## Summary

Three defects of the gap review are fixed, and a role can be tried on one issue with no change to
the repo.

## Changes

- `merge-order` and `merge-queue`: bookmarks at the same commit merge as one, under the first name,
  with a warning. Before, each was "based on" the other, and the run stopped with a cycle.
- `merge-queue` exits 2 when it leaves the work of a bookmark out: a conflict, the checks, or another
  reason in the report. The line of the other bookmarks is still made. Exit code 0 means that all
  work is in the line or was in the base already. As a tool agent, exit code 2 stops the request for
  a human.
- `hephaestus roles` reports a role with two role files (`x.md` and `x/role.md`), and roles with the
  same worker and subagent name (`a-b.md` and `a/b/role.md`). A command that finds a role with two
  role files fails.
- `hephaestus agent trial <role> <issue>` runs one task of a role, with the steps and limits of
  `agent run`, in a scratch copy of `--onto`. It claims, closes, or stops the issue only in the copy,
  and sends no message: the messages that it would send are in the record. `--notice <kind>` gives
  the issue as a broadcast, and `--keep` keeps the copy. Each trial goes into `trials.jsonl` in the
  shared hephaestus directory.
- `hephaestus agent trials [<role>...]` lists the trials: the result, the state of the issue before
  and after, the time, the tokens, the cost, and the number of changed files.

## Issues

- Closed [[gap-review-defects]].
