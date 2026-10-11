# sisyphus list

List and filter issues, as a table or as JSON.

## Usage

```bash
sisyphus list [flags]
```

## Flags

| Flag | Meaning |
|---|---|
| `-s`, `--state` | Comma-separated states to include, or `all`. Default: `open,in-progress`. |
| `-p`, `--priority` | Comma-separated priorities to include. |
| `--tags` | Comma-separated tags. An issue matches if it has any of them. |
| `--owner` | Only issues with exactly this owner. |
| `--mine` | Only issues whose owner is the value of `SISYPHUS_OWNER`. Not with `--owner`. An error when `SISYPHUS_OWNER` is not set. |
| `--parent` | Only direct sub-issues of this issue (any form of its name). A name that gives no issue is an error. |
| `--blocked` | Only issues that depend on an issue that is not closed yet. |
| `--ready` | Only issues that can start now: not closed, with no dependency and no sub-issue that is not closed. Not with `--blocked`. |
| `--wide` | Do not shorten titles to fit the terminal. |
| `--json` | Print JSON instead of a table. |

## Behavior

Within one flag, the values are alternatives (OR). Different flags must all match (AND).
Closed issues are listed only when `--state` includes `closed`.

`--blocked` counts a dependency as blocking while it is open or in-progress. A dependency that no
longer exists does not block.

`--ready` gives the issues that [[sisyphus-graph]] calls available, with the same rule: an epic
waits on its open sub-issues, so it is not ready until they close. With the default state filter,
ready in-progress issues are listed too; use `--state open --ready` for only the issues that nobody
has started. See [[0014-ready-means-available]].

## Output

On a terminal, the title column is shortened, with `…`, so that each row fits the width of the
terminal (but not below 10 characters). When stdout is not a terminal, for example in a pipe or a
file, and with `--wide`, titles are not shortened.

Sorted by state (open, in-progress, closed), then priority (critical to low), then name:

```
name           title          state        priority  owner  tags
-------------  -------------  -----------  --------  -----  ---------
fix-login-bug  Fix login bug  open         high             bug
add-dark-mode  Add dark mode  in-progress  medium    alice  feature, ui
```

With `--json`, an array of objects with `name`, `title`, `state`, `priority`, `owner`, and `tags`.
An empty result is an empty table, or `[]`.

Warnings, on stderr: one for each issue file that sisyphus cannot read, or whose frontmatter has a
line that is not a `key: value` field, a value that opens `[`, `{`, or a quote and does not close
it, a `state` that does not match its directory, or a `priority`, `effort`, or `resolution` that is
not valid. `list` checks every issue in the `--state` filter, also the ones that other filters hide. The warnings do not change stdout, so `--json` output stays valid. See
[[0012-warn-about-invalid-issues-on-read]].

## Examples

```bash
sisyphus list
sisyphus list --state open,in-progress,closed --tags bug
sisyphus list --priority critical,high --owner alice
sisyphus list --parent auth-overhaul
sisyphus list --blocked
sisyphus list --ready --state open          # what can start now
sisyphus list --state in-progress --json | jq -r '.[].name'   # what is claimed right now
```
