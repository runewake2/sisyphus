# 0019: Skip the files that .gitignore ignores in the repo walk

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.66]]

## Context

`resolve`, `links`, `lint`, `new`, and `slug` walk the whole repo to find link targets and Markdown
names. Only `.git`, `.jj`, `bin`, `obj`, and `node_modules` were skipped, so a large `vendor/`,
`target/`, `dist/`, or `.venv/` made them slow: with 100,000 vendored files, `resolve` took 137 ms
instead of 4 ms, and `lint` took 9.8 s, because every vendored file was a link target
([[skip-ignored-files-in-repo-walk]]).

## Decision

The walk skips what `.gitignore` ignores:

- In a git repo (a `.git` at the root, and `git` on PATH), one call of
  `git ls-files --others --ignored --exclude-standard --directory` gives the ignored paths, with an
  ignored directory as one entry. The walk skips them. Every rule of git applies: nested `.gitignore`
  files, negation, and `.git/info/exclude`, and a tracked file in an ignored directory still counts.
- Without git, the walk applies the simple rules of the root `.gitignore`: names and globs that match
  any path part, patterns with `/` from the root, and a trailing `/` for a directory. Negation is not
  supported.
- `.git`, `.jj`, `bin`, `obj`, and `node_modules` are always skipped, as before.

`.gitignore` is the per-repo setting; sisyphus adds no setting of its own.

## Alternatives considered

- **List the files with `git ls-files --cached --others`.** It needs a second call, `--deleted`, to
  drop tracked files that were deleted, which checks every tracked file: 92 ms for `resolve` at
  20,000 issues, against 38 ms before.
- **Only the simple root rules, also in a git repo.** About 20 ms faster in a repo with nothing to
  ignore, but nested `.gitignore` files and negation would not apply, and a tracked file in an
  ignored directory would be lost as a link target.
- **A longer built-in skip list** (`vendor`, `target`, `dist`, ...). It would hide directories that
  a repo really uses, and could not be changed per repo.

## Consequences

- With ignored directories, the commands that walk the repo are 10 to 100 times faster.
- In a git repo with nothing to ignore, the git call adds about 20 ms at 20,000 issues (`resolve`:
  38 → 60 ms). An ignored file is no longer a link target: a link to it is `missing`.
