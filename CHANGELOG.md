# Changelog

## [[0.0.15]] - 2026-10-07
- Suggest `bug` and `feature` as built-in tags in the template and in `--tags` help.

## [[0.0.14]] - 2026-10-07
- Make sisyphus itself independent of any particular repo or issue: generic template, error
  examples, and tests.

## [[0.0.13]] - 2026-10-07
- `sisyphus --version` prints the version the binary was compiled from, in any directory.

## [[0.0.12]] - 2026-10-07
- Add `sisyphus-mcp`: an MCP server giving an agent every `sisyphus` command as a tool.

## [[0.0.11]] - 2026-10-07
- Add `sisyphus graph` to view a large task's family tree in the terminal.

## [[0.0.10]] - 2026-10-07
- Narrow `sisyphus init` to set up `issues/` only; replace `agent-session` with a generic
  `metadata` field.

## [[0.0.9]] - 2026-10-07
- Add the two GitHub Actions that finish github-source-of-truth: issue-to-pr and sync-to-github.
  Add `sisyphus slug` and `sisyphus new --context`; fix `sisyphus show` to include `remote` and
  `depends-on`.

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
