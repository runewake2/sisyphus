# sisyphus lint

Check every issue and every wikilink in the repo at once, and write nothing.

## Usage

```bash
sisyphus lint [--json]
```

| Flag | Meaning |
|---|---|
| `--json` | Print the result as JSON: the counts and the list of problems. |

## What it checks

For each issue file in `issues/open`, `issues/in-progress`, and `issues/closed`, also in
subdirectories:

- The file can be read: it starts with YAML frontmatter.
- Each frontmatter line is one `key: value` field, and a value that opens a quote, `[`, or `{`
  closes it. These are the same checks as the warnings of `list`, `search`, and `show` (see
  [[0012-warn-about-invalid-issues-on-read]]).
- `state` matches the state directory.
- `priority`, `effort`, and `resolution` hold valid values.
- The full name is in one state directory only.
- No tag contains a space.
- `parent` names one existing issue, and the chain of parents above it does not come back to the
  issue.
- Each `depends-on` entry names one existing issue that is not the issue itself, and the
  dependencies do not come back to the issue.

For each Markdown file in the repo: every wikilink is `ok`, the same check as [[sisyphus-links]].
A frontmatter link that does not resolve, for example a `parent` that does not exist, is reported
by both checks.

`lint` never writes to a file. It does not repair problems: use [[sisyphus-update]] (which corrects
a `state` field that does not match its directory) or edit the file by hand.

## Output

One line for each problem, sorted by file:

```
design/notes.md:3: the link [[nowhere-at-all]] is missing
issues/open/fix-login-bug.md: its priority 'urgent' is not one of: critical, high, medium, low
issues/open/fix-login-bug.md: its depends-on entry 'old-cache-layer' does not exist
```

and then, on stderr, `Error: 3 problems in 62 issues and 148 Markdown files.` The exit code is 1.

With no problem, one line on stdout, for example `62 issues and 148 Markdown files: no problems.`,
and exit code 0.

With `--json`, one object: `issues` and `documents` (the counts), and `problems`, a list of objects
with `file`, `line` (only for a link), and `problem`. The exit code is the same as without
`--json`.

## Examples

```bash
sisyphus lint
sisyphus lint --json | jq -r '.problems[].file' | sort -u
```

The CI workflow of this repo runs `sisyphus lint` next to `go test`.
