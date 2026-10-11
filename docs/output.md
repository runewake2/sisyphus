# Output

The rules that every sisyphus command follows for its output, so that scripts and agents can rely
on them, and the shape of each JSON output.

## Streams

- **stdout** has the result only: a path, a table, an issue, or JSON. With `--json`, stdout is
  always one valid JSON value.
- **stderr** has everything else: warnings, errors, and notes such as
  `Showing 20 of 57 issues. Narrow the filters, or raise --limit.`
- A **warning** is one line that starts with `Warning:`. It tells about a problem that did not stop
  the command, for example a close while an issue still depends on the issue. Read it.
- An **error** is one line that starts with `Error:`. The command changed nothing, except where its
  page says otherwise. The message tells what to change, and a value that is not valid gets the
  valid values and, when one is near, a suggestion.
- The commands that change an issue print the path of the file, from the repo root, for example
  `issues/in-progress/fix-login-bug.md`. With `--json`, they print the changed issue instead, and
  the warnings go into its `warnings` list, not to stderr.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success, also with warnings, and also when a list or search finds nothing. |
| 1 | An error. Also [[sisyphus-lint]] when it finds problems: then stdout still has the report. |

[[sisyphus-links]] exits 0 also when a link does not resolve: read the `status` of each link.

## JSON

`--json` gives JSON on these commands:

| Commands | Shape |
|---|---|
| [[sisyphus-show]] | [[#An issue]] |
| [[sisyphus-new]], [[sisyphus-update]], [[sisyphus-start]], [[sisyphus-close]], [[sisyphus-abandon]], [[sisyphus-edit]], [[sisyphus-parent]], [[sisyphus-depends-on]], [[sisyphus-remote]] | [[#A changed issue]] |
| [[sisyphus-list]] | An array of [[#A list row]] |
| [[sisyphus-search]] | An array of [[#A search row]] |
| [[sisyphus-graph]] | [[#A graph]] |
| [[sisyphus-links]] | An array of [[#A link]] |
| [[sisyphus-lint]] | [[#A lint report]] |
| [[sisyphus-claims]] | An array of [[#A claim]] |

A field that holds an empty value is left out, unless the tables below say "always". A list is a
JSON array, and a map is a JSON object.

### An issue

| Field | Value |
|---|---|
| `name` | The full name, for example `web/fix-login-bug`. Always. |
| `title`, `state`, `priority`, `effort`, `created` | The frontmatter fields. Always. |
| `resolution`, `closed` | Only for a closed issue. |
| `tags`, `workspaces` | Arrays. |
| `owner`, `approver`, `bookmark`, `deferred-from`, `parent`, `remote` | Strings, as in the frontmatter. `parent` is a wikilink. |
| `depends-on` | An array of plain issue names. |
| `metadata` | An object. |
| `body` | The body, without the sections that hold only template text. `show --json --all` gives it as it is in the file. Always. |

### A changed issue

[[#An issue]], plus:

| Field | Value |
|---|---|
| `path` | The path of the file, from the repo root. Always. |
| `warnings` | An array of the warnings, without `Warning:`. Always; empty when there are none. |

### A list row

| Field | Value |
|---|---|
| `name`, `title`, `state`, `priority` | Always. |
| `owner`, `bookmark`, `tags` | When set. |

### A search row

[[#A list row]], plus:

| Field | Value |
|---|---|
| `matches` | With a query: an array of the matched lines, each with `section` (a heading, or `title`) and `text`. At most 5. |
| `more-matches` | The number of matched lines after the first 5. |
| `section`, `content` | With `--section`: the section name and its text. `content` also with `--content`: the body. |

### A graph

| Field | Value |
|---|---|
| `focus` | The full name of the issue that the graph is drawn for. |
| `nodes` | An array of issues, each with `name`, `title`, `state`, and when it applies `resolution`, `available` (true: work can start now), and `indirect` (true: only with `--full`, not below the focus or on the path above it). |
| `edges` | An array, each with `from`, `to`, and `kind`: `parent` (from a parent to a sub-issue) or `depends-on` (from a dependency to the issue that depends on it). |
| `hidden` | The number of linked issues that are not drawn. |

### A link

| Field | Value |
|---|---|
| `line` | The line of the link in the document. |
| `link` | The link as it is written, for example `[[fix-login-bug#Notes]]`. |
| `status` | `ok`, `missing`, `ambiguous`, or `missing-heading`. |
| `resolved` | The path of the target from the repo root, or `""`. Always. |

### A lint report

| Field | Value |
|---|---|
| `issues`, `documents` | The number of issue files and Markdown files checked. |
| `problems` | An array, each with `file`, `problem`, and for a link problem `line`. Empty when there are none. |

### A claim

| Field | Value |
|---|---|
| `name`, `workspace`, `since`, `status` | The issue, the absolute path of the workspace, the start time (RFC 3339), and `here`, `active`, or `stale`. Always. |
| `owner`, `bookmark` | When set. |

## Changes to the JSON

- Fields are added, never removed or renamed, without a note.
- A change that can break a script, for example a field that moves or is no longer returned by
  default, is marked **Changed:** in the changelog of its version, and has a decision record. For
  example, [[0.0.94]] returns `content` of `search` only with `--section` or `--content`.
- Do not depend on the order of the fields, or on the indentation.

## MCP

[[sisyphus-mcp]] returns the stdout of the command as text, followed by the warnings, and JSON also
as structured content. An error result has the error message. See [[sisyphus-mcp#Results]].
