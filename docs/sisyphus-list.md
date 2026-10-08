# sisyphus list

List and filter issues, as a table or as JSON.

## Usage

```bash
sisyphus list [flags]
```

## Flags

| Flag | Meaning |
|---|---|
| `-s`, `--state` | Comma-separated states to include. Default: `open,in-progress`. |
| `-p`, `--priority` | Comma-separated priorities to include. |
| `--tags` | Comma-separated tags. An issue matches if it has any of them. |
| `--owner` | Only issues with exactly this owner. |
| `--parent` | Only direct sub-issues of this issue (any form of its name). |
| `--blocked` | Only issues that depend on an issue that is not closed yet. |
| `--json` | Print JSON instead of a table. |

Within one flag, the values are alternatives (OR). Different flags must all match (AND).
Closed issues are listed only when `--state` includes `closed`.

`--blocked` counts a dependency as blocking while it is open or in-progress. A dependency that no
longer exists does not block.

## Output

Sorted by state (open, in-progress, closed), then priority (critical to low), then name:

```
name           title          state        priority  owner  tags
-------------  -------------  -----------  --------  -----  ---------
fix-login-bug  Fix login bug  open         high             bug
add-dark-mode  Add dark mode  in-progress  medium    alice  feature, ui
```

With `--json`, an array of objects with `name`, `title`, `state`, `priority`, `owner`, and `tags`.
An empty result is an empty table, or `[]`.

## Examples

```bash
sisyphus list
sisyphus list --state open,in-progress,closed --tags bug
sisyphus list --priority critical,high --owner alice
sisyphus list --parent auth-overhaul
sisyphus list --blocked
sisyphus list --state in-progress --json | jq -r '.[].name'   # what is claimed right now
```
