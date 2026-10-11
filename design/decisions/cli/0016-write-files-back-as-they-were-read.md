# 0016: Write an issue file back with the BOM and line endings it had

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.47]]

## Context

sisyphus edits frontmatter line by line, so that a change touches only the lines it must. But it
could not read a file that starts with a UTF-8 byte order mark ([[read-files-with-bom]]), and it
wrote every file back with LF line endings, so one field change in a CRLF file changed every line
([[keep-crlf-line-endings]]). Some Windows editors write a BOM or CRLF, and git and Obsidian read both.

## Decision

sisyphus reads a file with or without a BOM, and with LF or CRLF line endings. When it writes the
file back, it keeps what the file had: the BOM if there was one, and its line endings. A new issue
takes both from `issues/TEMPLATE.md`.

## Alternatives considered

- **Remove the BOM and use LF on every write.** The file would be normalized, but the first change
  by sisyphus would also be a change of every line, which hides the real change in a diff.
- **Refuse such files.** Users on Windows would have to change their editor settings for sisyphus.

## Consequences

- A field change is a one-line diff in every file.
- A file with mixed line endings gets the line ending of its first line.
