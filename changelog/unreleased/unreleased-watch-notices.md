# Unreleased: hephaestus watch sends the notices of a close

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: `hephaestus watch` broadcasts `bookmark-ready` and `issue-closed` from the state of the issues and their bookmarks, once each, so the reviewers and the docs start without a model that remembers to send them.

## Summary

The notices that start the reviewers and the docs come from code now
([[0007-the-watcher-sends-the-notices-of-a-close]]).

## Changes

- `hephaestus watch` (`--once`, or every `--every`, 30 seconds by default):
  - when an issue closes, it broadcasts `bookmark-ready` (if the bookmark has work that `--onto` does
    not have), and then `issue-closed`;
  - when the bookmark of a closed issue moves, it broadcasts `bookmark-ready` again;
  - each notice is sent once; the first run only records the state, in `watch.json`.
- The implementer no longer broadcasts `bookmark-ready`; it closes its issue. The `Works with`
  sections of the roles say that the notice comes from `hephaestus watch`.

## Decisions

- [[0007-the-watcher-sends-the-notices-of-a-close]]

## Issues

- Closed [[watch-notices]].
