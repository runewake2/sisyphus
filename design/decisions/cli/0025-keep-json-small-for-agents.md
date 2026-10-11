# 0025: Keep the JSON of show and search small for agents

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.94]]

## Context

Agents read the JSON of `show` and `search` into their context, and the MCP tools return JSON by
default. `show --json` returned the body with the template text of each section that nobody had
filled, about 970 bytes for a new issue. `search --json` returned the whole body of each match: a
search for `mcp` in this repo returned 45,673 bytes for 17 issues
([[compact-json-for-agents]]). Search also matched the template text, so a word of a placeholder
matched every issue.

## Decision

- `show --json`, and the `--json` output of the commands that change an issue, leave out the
  sections that hold only template text, as the text of `show` already does
  ([[readable-show-output]]). `show --json --all` gives the body as it is in the file.
- `search` does not match template text.
- `search` with a query returns the matched lines, each with the heading of its section, in
  `matches`, at most 5 for each issue, with a count of the others in `more-matches`. A line is cut
  to 200 characters around the match. `content` is now returned only with `--section` or with the
  new `--content`.
- `list` and `search` take `--limit`. The JSON stays an array; stderr tells how many issues were
  left out.

## Alternatives considered

- **Keep `content` by default and add a flag for the lines.** The large output is the default that
  every agent gets, so the default must be the small one.
- **An object with `results` and `total` for `--limit`.** It would break every script that reads
  the array.
- **Leave out the body of `show --json`.** An agent calls `show` to read the issue.

## Consequences

- The search for `mcp` in this repo returns 13,284 bytes (29%).
- A script that read `content` from `search --json` without `--section` must add `--content`.
- The JSON body of a new issue has only its title, until someone fills a section.
