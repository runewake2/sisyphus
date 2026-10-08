# Changelog

## [[0.0.23]] - 2026-10-08
- Call an issue that is ready to start "available" instead of "leaf" in `sisyphus graph`, its
  `--json` field, and its `--mermaid` class.

## [[0.0.22]] - 2026-10-08
- Place each issue in `sisyphus graph` one column after the farthest issue that points to it, so
  arrows no longer turn back or run side by side.

## [[0.0.21]] - 2026-10-08
- Draw each issue subdirectory in `sisyphus graph` as a box around its issues, show only file
  names in issue boxes, and mark the focus and leaves with box line styles instead of emoji.

## [[0.0.20]] - 2026-10-08
- `sisyphus slug` spells letters in ASCII (é becomes e), drops filler words before it keeps 6 words,
  and always prints ASCII.

## [[0.0.19]] - 2026-10-08
- Read the state directories once per command and parse each issue at most once; link-following
  commands run hundreds of times faster on a repo with thousands of issues.

## [[0.0.18]] - 2026-10-08
- Let issues sit in subdirectories below the state directories, found by their full name or by their
  file name; an ambiguous file name fails and lists the full names.

## [[0.0.17]] - 2026-10-08
- Fix `sisyphus update` in a colocated jj repo: move an issue that git does not track with a plain
  rename instead of `git mv`.

## [[0.0.16]] - 2026-10-08
- Draw `sisyphus graph` as a real dependency graph: everything below an issue and the path above
  it, with leaves marked, every linked issue with `--full`, plus `--mermaid` output and a new JSON format.

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
