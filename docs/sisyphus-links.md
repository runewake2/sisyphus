# sisyphus links

List every wikilink in a document, and check that each one resolves.

## Usage

```bash
sisyphus links <document> [--json] [--absolute]
```

`<document>` is a path, or a wikilink to the document, resolved the way [[sisyphus-resolve]] does.

## What it checks

Every wikilink in the document, in the frontmatter and the body, is resolved the way
[[sisyphus-resolve]] resolves a link. Each gets a status:

| Status | Meaning |
|---|---|
| `ok` | The link resolves to exactly one file, and its `#heading`, if any, exists there. |
| `missing` | No file matches. |
| `ambiguous` | More than one file matches. |
| `missing-heading` | The file exists, but has no heading with that name (ignoring case). |

- Links inside fenced code blocks and inline code are skipped, as Obsidian skips them. Write an
  example link in backticks so it is not checked.
- `[[#Heading]]` checks a heading in the same document.
- A block reference (`[[name#^block-id]]`) is only checked for the file, not for the block.

## Output

A table with `line`, `link`, `status`, and `resolved`. A document with no wikilinks prints
`No wikilinks in <document>.` to stderr. With `--json`, an array of objects with the same four
fields.

`links` reports problems but does not fail because of them: it exits 0 whenever the document can
be read.

## Examples

```bash
sisyphus links README.md
sisyphus links "[[fix-login-bug]]" --json | jq '.[] | select(.status != "ok")'
```
