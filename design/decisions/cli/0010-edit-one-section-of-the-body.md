# 0010: Edit one section of an issue's body with sisyphus edit

- Status: Accepted
- Date: 2026-10-09
- Version: [[0.0.31]]

## Context

From the human, 2026-10-09: a test run of Claude as a sisyphus task manager found that no command
edits the body of an issue. Six times, Claude reported that the `Summary` or `Resolution` section
still held the template placeholder after it created or closed an issue. `sisyphus new --context`
fills only the `Context` section, and only at creation. An agent that uses only sisyphus-mcp cannot
edit the file by hand.

## Decision

Add `sisyphus edit <name> <section> <text> [--append]`, and the MCP tool `sisyphus_edit`.

- `<section>` is a heading, matched case-insensitively in the same way as a wikilink's `#heading`.
  The title heading (`# ...`) is not a section.
- The text replaces the section. With `--append`, the text goes after the current text, but
  replaces a section that holds only the template placeholder. So the first dated line in `Notes`
  does not keep the placeholder above it.
- `-` as `<text>` reads standard input, for text with many lines.
- `sisyphus update <name> closed` warns when `Resolution` still holds the placeholder, and names
  the command to fix it.
- sisyphus-mcp now returns the warnings of a command after its output. Before, it dropped them on
  success, so an agent never saw a warning.

## Alternatives considered

- **A `--resolution-text` flag on `update`.** It fixes only `Resolution`. The test run also found
  `Summary` with the placeholder, and `Notes` needs appends.
- **One flag per section on `new` and `update`, like `--context`.** Many flags, and a new section in
  a repo's template would need a new flag.
- **Text in a `--text` flag instead of an argument.** The text is required, so an argument is
  shorter. `-` covers text with many lines.
- **Read standard input when `<text>` is missing.** An agent that forgets the text then waits for
  input that never comes.

## Consequences

- An agent can complete every section of an issue with sisyphus alone.
- `edit` cannot change the title. The title is in the frontmatter and in the first heading.
- Tests that close an issue and expect no warning must complete `Resolution` first.
- [[CONTRIBUTING]] lists `sisyphus edit` with the other commands.
