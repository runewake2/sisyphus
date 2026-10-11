# 0027: Rename changes every link

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.105]]

## Context

The full name of an issue was permanent: no command changed it, and the docs said never to rename
or move an issue, because links to it would break. A human asked for a way to organize the issues
of a repo in subdirectories ([[rename-issue-command]]).

## Decision

- `sisyphus rename <name> <new-name>` moves the file in its state directory, and changes every
  wikilink that resolves to the issue in every Markdown file of the repo, with the rules of
  [[sisyphus-links]]: links in code do not change.
- A link with only the file name stays so while the new file name is unique, so a move to a
  directory changes few lines. A link with a path, and every link in the frontmatter, gets the new
  full name, as sisyphus writes links in the frontmatter.
- The claim of the issue moves. An issue that another workspace has in progress is refused.
- `--dry-run` shows every change first.
- The docs recommend a move to a directory over a new file name: commit messages, other branches,
  and other repos refer to the file name, and a rename cannot change them.

## Alternatives considered

- **Only moves to a directory, never a new file name.** A misspelled or misleading name could then
  never be corrected.
- **Leave links with the old name, and let them resolve through an alias.** sisyphus and Obsidian
  have no aliases for files, and lint would report every old link.
- **Change Markdown links (`[text](path)`) too.** sisyphus does not resolve them anywhere else, and
  a wrong change of a path in a README is worse than a link that lint can find.

## Consequences

- Issues can move to directories with no broken links.
- A rename changes many files when many documents link to the issue with a path. `--dry-run` shows
  them.
