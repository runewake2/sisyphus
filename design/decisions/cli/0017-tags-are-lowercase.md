# 0017: Store tags in lowercase and compare them without regard to case

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.55]]

## Context

Tags were stored and compared with their case, so `--tags Bug` and `list --tags bug` did not
match, and a tag `Bug` dropped out of a filter for the built-in tag `bug` ([[tags-ignore-case]]).
Obsidian treats tags without regard to case.

## Decision

`new --tags`, `update --tags`, and `update --add-tags` store tags in lowercase. `list --tags`,
`search --tags`, `--add-tags`, and `--remove-tags` compare tags without regard to case, so that
tags that a person wrote by hand in another case still match.

## Alternatives considered

- **Keep the case, and only compare without it.** The same tag could then be stored as `Bug` in one
  issue and `bug` in another, and `list` would show both forms.
- **Reject tags with uppercase letters.** It is stricter than needed: the lowercase form is clear.

## Consequences

- Tags that are already in a file keep their case until a command changes them.
- [[migrate-command]] can lowercase existing tags.
