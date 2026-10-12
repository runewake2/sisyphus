# Unreleased: Filter list and search on a metadata key

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: cli, sisyphus-mcp, docs
- Changelog: Filter list and search on metadata with --metadata key=value, and include metadata in their JSON rows.

## Summary

`list` and `search` take `--metadata key=value`, so that any client can find the issues that it
marked with a key, for example hephaestus with `delegate=merge-queue`. sisyphus does not know what a
key means ([[0029-sisyphus-does-not-manage-agents]]).

## Changes

- `--metadata key=value` on `list` and `search`, repeatable: every pair must match. `--metadata key`
  matches any value of the key. A pair with no key is an error.
- The JSON rows of `list` and `search` include `metadata` when it is set.
- `sisyphus_list` and `sisyphus_search` take `metadata` as an object; an empty value matches any
  value.
- [[sisyphus-list]], [[sisyphus-search]], and [[output]] describe the filter and the field.

## Issues

- Closed [[filter-by-metadata]].
