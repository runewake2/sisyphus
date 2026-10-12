# Unreleased: Report wikilink aliases in table rows

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: cli, links
- Changelog: `sisyphus links` and `sisyphus lint` report a wikilink with an alias in a Markdown table row (`alias-in-table`), because Markdown reads its `|` as the end of the cell.

## Summary

A wikilink with an alias, `[[target|text]]`, breaks a Markdown table row: GitHub-flavored Markdown
reads its `|` as the end of the cell, and formatters such as prettier split the cell there. The
repo no longer uses aliases in tables, and sisyphus reports each one.

## Changes

- `sisyphus links` gives a link the status `alias-in-table` when it resolves, has an alias, and is
  in a table row. The same link outside a table is `ok`.
- `sisyphus lint` reports it: "the link ... is in a table and has an alias: Markdown reads its | as
  the end of the cell. Remove the alias".
- The government table of [[agent-delegation]] links to the role files without aliases.
- [[CONTRIBUTING#Links between documents]] has the rule. [[sisyphus-links]] and [[output]] have the
  new status.

## Issues

- Closed [[links-alias-in-table]].
