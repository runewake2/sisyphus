# Unreleased: the merge queue can test a batch of bookmarks together

- Date: 2026-11-03
- Bookmark: samw/ai/merge-queue-batches
- Components: hephaestus
- Changelog: `hephaestus merge-queue --batch <n>` runs the checks once for up to n bookmarks with no conflicts. A batch that fails splits in two until the checks find each bookmark that fails. The report has a new field, `check-runs`.

## Summary

Long checks made the merge queue slow when much work was ready. A batch now uses one run of the
checks for several bookmarks.

## Changes

- `internal/hephaestus/mergequeue`: the run puts each batch on the line, then runs the checks once.
  A batch that fails goes back and splits.
- `cmd/hephaestus`: the flag `--batch`, the field `batch` of a message, and the line of the report
  that tells how many times the checks ran.

## Decisions

- [[0004-run-the-checks-once-for-a-batch]]

## Issues

- Closed [[merge-queue-batches-checks]].

## Notes

The default stays 1, and the role `merge-queue` does not use a batch until a human adds the flag to
its command or a request gives `batch`. The checks test only the tip of a batch.
