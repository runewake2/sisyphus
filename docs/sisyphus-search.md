# sisyphus search

Find issues by text, by frontmatter filters, or both, and print what matched.

## Usage

```bash
sisyphus search [<query>] [flags]
```

## Flags

`search` takes every filter of [[sisyphus-list]] (`--state`, `--priority`, `--tags`, `--owner`,
`--parent`, `--blocked`), with the same defaults and meaning, plus:

| Flag | Meaning |
|---|---|
| `--section` | Search, and print, only one section of each issue, for example `summary`. |
| `--json` | Print JSON instead of text. |

## What it matches

- `<query>` is a plain substring, matched without regard to case. Without `--section`, it is
  matched against the title and the whole body. Without a query, every issue that passes the
  filters matches.
- `--section <heading>` matches a heading by name at any level, ignoring case, the same way a
  wikilink's `#heading` does. The section runs until the next heading of the same or a higher
  level. With `--section`, the query is matched against that section only (not the title), and an
  issue without that section does not match.
- The query and every filter must all match (AND). There are no operators such as OR or NOT.

## Output

Sorted the same way as `list`. For each match, a line with the name and title, then the content:
the whole body, or only the section with `--section`. A blank line separates matches. No match
prints nothing and is not an error.

```
fix-login-bug: Fix login bug
The login form rejects a valid password after a session times out.
```

With `--json`, an array of objects with `name`, `title`, `state`, `priority`, `owner`, `tags`,
`section`, and `content`.

## Examples

```bash
sisyphus search timeout
sisyphus search timeout --section summary
sisyphus search --section "acceptance criteria" --tags bug --json
sisyphus search cache --state open,in-progress,closed --priority critical,high
```
