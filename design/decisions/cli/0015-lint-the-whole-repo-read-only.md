# 0015: sisyphus lint checks the whole repo and never writes

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.45]]

## Context

Each command checks only the issue that it touches, and since
[[0012-warn-about-invalid-issues-on-read]] the read commands warn about the files that they read.
Nothing checked the whole repo at once, so CI could not fail on a broken issue file or a broken
wikilink ([[lint-issues-command]]).

## Decision

`sisyphus lint` checks every issue file and every Markdown file of the repo in one run, and never
writes:

- For each issue: the checks of `issueProblems` (frontmatter shape, state against the directory,
  enum values), one state directory per full name, no tag with a space, a parent that exists with
  no cycle above it, and depends-on entries that exist, are not the issue itself, and make no cycle.
- For each Markdown file: every wikilink is `ok`, with the same rules as `sisyphus links`.

Each problem is one line, `<file>[:<line>]: <problem>`, on stdout, and the exit code is 1. With no
problem, one summary line and exit code 0. `--json` gives the counts and the problems. The CI
workflow runs `sisyphus lint` after the tests. `sisyphus_lint` returns the JSON as a result, also
when there are problems, so that an agent can read them.

One walk of the repo resolves the links of all documents (`linkChecker`), so that the run stays
fast: about 20 ms for this repo.

## Alternatives considered

- **Repair what it can, like `update` does for a state field.** A linter that writes is not safe in
  CI, and a repair needs a choice that a person must make, for example which of two duplicate files
  to keep.
- **Report base-name clashes between an issue and another Markdown file as problems.** sisyphus
  allows them with a warning (see [[sisyphus-new]]); a clash is only a problem when a link becomes
  ambiguous, and the link check already reports that.
- **Run `sisyphus links` once for each file.** Each run walks the repo again, so the time grows with
  the square of the number of files.

## Consequences

- A broken issue file or wikilink fails CI.
- A broken frontmatter link, for example a parent that does not exist, is reported twice: once as
  an issue problem and once as a link problem. Both lines point to the same fix.
- [[migrate-command]] can use `lint` to check its result.
