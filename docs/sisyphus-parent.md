# sisyphus parent

Make an issue a sub-issue of another issue, or remove its parent.

## Usage

```bash
sisyphus parent <name> <parent>
sisyphus parent <name> --clear
```

Both `<name>` and `<parent>` take any form of an issue name (see [[commands#Naming an issue]]).
Give exactly one of `<parent>` and `--clear`.

## Rules

- An issue has at most one parent. Setting a parent replaces the old one.
- The parent must exist, in any state directory.
- An issue cannot be its own parent, and its parent cannot be one of its own sub-issues, at any
  depth: that would make a cycle.
- A closed parent is allowed, with a warning.

The parent is written as a quoted wikilink, `parent: "[[auth-overhaul]]"`, so it stays valid YAML
and resolves in Obsidian.

Use [[sisyphus-graph]] to see the whole tree, and `sisyphus list --parent <name>` to list the direct
sub-issues. Closing a parent with sub-issues that are not closed gives a warning (see
[[sisyphus-update]]).

## Output

The path of the issue's file.

## Examples

```bash
sisyphus parent split-login-form auth-overhaul
sisyphus parent split-login-form --clear
```
