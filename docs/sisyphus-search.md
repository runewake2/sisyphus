# sisyphus search

Find issues by text, by frontmatter filters, or both, and print what matched.

## Usage

```bash
sisyphus search [<query>] [flags]
```

## Flags

`search` takes every filter of [[sisyphus-list]] (`--state`, `--priority`, `--tags`, `--owner`,
`--mine`, `--parent`, `--blocked`, `--ready`), with the same defaults and meaning, plus:

| Flag | Meaning |
|---|---|
| `--section` | Search, and print, only one section of each issue, for example `summary`. |
| `--content` | Print the whole body of each match, instead of the matched lines. |
| `--limit` | Print at most this many issues, and tell on stderr how many were left out. `0` (the default): all. |
| `--json` | Print JSON instead of text. |

## Behavior

- `<query>` matches an issue whose title and body have every word of the query, in any order,
  without regard to case: `docs site theme update` finds "Update the docs site theme". Each word
  matches as a part of a word, so `cache` also matches `caches`. Without `--section`, the words are
  matched against the title and the body.
- Words in double quotes match as one exact phrase: `sisyphus search '"session cache" login'`
  finds an issue that has the phrase "session cache" and the word "login". Sections that hold only template text (the `<...>`
  placeholders) are not searched. Without a query, every issue that passes the filters matches.
- `--section <heading>` matches a heading by name at any level, ignoring case, the same way a
  wikilink's `#heading` does. The section runs until the next heading of the same or a higher
  level. With `--section`, the query is matched against that section only (not the title), and an
  issue without that section does not match.
- A `--section` that `issues/TEMPLATE.md` does not have, and that no issue in the filtered set has,
  is an error that lists the sections of the template. It is almost always a typo.
- Every word, every phrase, and every filter must match (AND). There are no operators such as OR
  or NOT.

## Output

Sorted the same way as `list`. For each match, a line with the name and title, then:

- with a query, each line that has a word or phrase of the query, with the heading of its section
  (`title` for the title).
  At most 5 lines for each issue, then a count of the others. A line longer than 200 characters is
  cut around the match, with `…`. If a phrase spans two lines of the body, the lines that contain
  a word of the phrase are shown.
- with `--section`, the section.
- with `--content`, the body without the sections that hold only template text.

A blank line separates matches. No match prints nothing and is not an error.

```
fix-login-bug: Fix login bug
  Summary: The login form rejects a valid password after a session times out.
  Notes: 2026-10-09: the timeout is in the session cache.
```

Read a whole issue with [[sisyphus-show]]. With `--json`, an array of objects with `name`, `title`,
`state`, `priority`, `owner`, `bookmark`, `tags` (the empty ones are left out), and:

- `matches`: the matched lines, each with `section` and `text`, and `more-matches`, the number of
  lines after the first 5. Left out when there are none.
- `section` and `content`: with `--section`, or `content` with `--content`.

With `--limit`, stderr tells how many issues were left out:
`Showing 20 of 57 issues. Narrow the filters, or raise --limit.`

Warnings, on stderr: one for each issue file that sisyphus cannot read, or whose frontmatter has a
line that is not a `key: value` field, a value that opens `[`, `{`, or a quote and does not close
it, a `state` that does not match its directory, or a `priority`, `effort`, or `resolution` that is
not valid. `search` checks every issue in the `--state` filter, also the ones that other filters hide. The warnings do not change stdout, so `--json` output stays valid. See
[[0012-warn-about-invalid-issues-on-read]].

## Examples

```bash
sisyphus search timeout
sisyphus search timeout --section summary
sisyphus search '"session cache" login'   # an exact phrase and a word
sisyphus search timeout --content --limit 3
sisyphus search --section "acceptance criteria" --tags bug --json
sisyphus search cache --state open,in-progress,closed --priority critical,high
```
