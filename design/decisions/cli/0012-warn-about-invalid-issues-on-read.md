# 0012: Warn about invalid issues on read, with a structural check

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.39]]

## Context

sisyphus reads frontmatter line by line with its own key/value reader (`document.go`), so that it
can edit one field and keep the inline comments. That reader never fails: `title: [broken` gives
the title `[broken`. `list`, `show`, and `search` showed such files, files whose `state` field did
not match their directory, and files with an invalid priority as if they were good, and a priority
filter could hide an issue with an invalid priority ([[warn-invalid-issues-on-read]]).
[[lint-issues-command]] will check the whole tree on request, but people and agents run the read
commands every day.

## Decision

`list`, `search`, and `show` write one warning to stderr for each issue file with a problem:

- The file cannot be read (no frontmatter).
- A frontmatter line is not a `key: value` field.
- A value opens a double or single quote, a `[` list, or a `{` map, and does not close it, or a
  double-quoted value is not a valid string.
- `state` does not match the state directory.
- `priority`, `effort`, or `resolution` is not a valid value.

`list` and `search` check every issue in the `--state` filter before the other filters, so that a
filter cannot hide a broken issue without a warning. The commands still succeed, and stdout does
not change.

## Alternatives considered

- **Parse the frontmatter with a YAML library.** It would catch every YAML error, but it adds a
  dependency, and the line-based reader would still be the one that reads the values, so the two
  could disagree. The structural check finds the errors that a hand edit makes in practice, with
  the same reader that the commands use.
- **Fail the command.** One broken file would stop `list` for the whole repo, which is worse than a
  warning for a read command.
- **Leave it to `sisyphus lint`.** Lint runs on request; the read commands show bad data every day.

## Consequences

- A broken issue is visible at the next `list`, `search`, or `show`.
- [[lint-issues-command]] can reuse `issueProblems` for the same checks.
- Valid YAML that the line reader cannot read, for example a block-style list, gets a warning too.
  That is correct: sisyphus would read such a field wrong.
