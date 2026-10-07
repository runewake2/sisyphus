# Changelog

## [[0.0.8]] - 2026-10-07
- Default a closed issue's resolution to `completed` instead of requiring `--resolution`.

## [[0.0.7]] - 2026-10-07
- Add `depends-on` and `sisyphus depends-on`, with cycle detection and a close-time warning, plus
  `--blocked` on `list`/`search`.

## [[0.0.6]] - 2026-10-07
- Add `sisyphus search` to find issues by frontmatter filters and content, with section-scoped
  results.

## [[0.0.5]] - 2026-10-07
- Add `sisyphus list` to list and filter issues; update AGENTS.md's claimed-issue check to use it.

## [[0.0.4]] - 2026-10-07
- Add `remote` to pin an issue to a GitHub issue or Jira ticket, via `sisyphus new --remote` or
  `sisyphus remote`.

## [[0.0.3]] - 2026-10-07
- Let `sisyphus update` change `priority`, `effort`, and `tags`.

## [[0.0.2]] - 2026-10-07
- Move issue files through `git mv` when the repo has a `.git` directory.

## [[0.0.1]] - 2026-10-07
- Add `sisyphus show` to print one issue's key fields and body.

## [[0.0.0]] - 2026-10-07
- Start the workflow: contributing guide, agent rules, issues, and the changelog.
