# 0004: Run the checks once for a batch

- Status: Accepted
- Date: 2026-11-03
- Version: [[unreleased-merge-queue-batches-checks]]

## Context

The merge queue ran the checks after each bookmark. With long tests, the queue becomes slow when
much work is ready. A human asked that work with no conflicts is tested together, so that the queue
can catch up.

## Decision

- `hephaestus merge-queue --batch <n>` puts up to n bookmarks on the line, and runs the checks once,
  on the tip. The default is 1, the behavior before this decision. A message can give `batch`.
- A bookmark that conflicts with the line stays out of the batch, and the batch goes on.
- A batch that fails goes back to the last line that passed, with its versions and its left-out
  records. Each half then tries again, until the checks find each bookmark that fails.
- A conflict with a member of a failed batch is tried again, because that member can fail.
- When the first half merges completely, the second half makes the tree that failed. The run does
  not run the checks on that tree again.
- The report has the number of runs of the checks, in the field `check-runs`.

## Alternatives considered

- **Run the checks of each bookmark in parallel, each in a working copy of its own.** It costs the
  same CPU in less time, and it does not test the bookmarks together. The line still needs a check
  of the combined result.
- **Keep bookmarks that change the same file in different batches.** The conflict check of the VCS
  already finds a real conflict. This adds runs when the work does not conflict.
- **Make batches the default.** A batch checks less than one run for each bookmark. A person must
  choose that.

## Consequences

- When a batch passes, n bookmarks cost one run of the checks. When one bookmark fails, the search
  costs more runs, about two for each halving.
- The checks test the tip of each batch, not each bookmark in it. A commit inside a batch can fail
  the tests if a later bookmark of the batch fixes it, so a bisect of `main` can stop on it.
- Checks that fail at random can blame the wrong bookmark, because the run trusts the failure of a
  tree and does not run the checks on that tree again.
