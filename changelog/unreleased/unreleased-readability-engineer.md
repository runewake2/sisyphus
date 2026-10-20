# Unreleased: A readability engineer

- Date: 2026-10-20
- Bookmark: samw/ai/readability-engineer
- Components: hephaestus
- Changelog: The AI role `readability-engineer` measures the spread of each change — the distance in lines between its related edits, its files, and its directories — ranks the sprawling changes, and proposes options for condensing the code under them, so that one behavior can be edited from a single screen.

## Summary

A role whose whole objective is condensed changes. A change that spans directories ranks above one
that spans files, and one that spans files above one that leaves gaps of unchanged lines longer than
a screen between its edits. The role proposes moves; it never changes behavior, and never crosses a
boundary that a decision record keeps.

## Changes

- `hephaestus/roles/readability-engineer.md`:
  - For a change, often from a `bookmark-ready` notice, it reads the hunks of the diff and records
    the spread as numbers: the directories, the files, and the largest gap, against a screen of
    about 40 lines.
  - For a sprawling change, it proposes the options to condense: move the helper next to its caller,
    group what changes together, split a file that mixes behaviors, join files that always change
    together — and keeps the boundaries of the decision records.
  - For a part of the repo, it ranks the recent changes by spread, and files one issue for each
    recurring sprawl, with the evidence and the options.
  - On request, it makes one small move on its own bookmark with no change of behavior.
  - As a reviewer in a referendum, it judges the locality of the change.
- [[agent-delegation#The government]] has the role. Its subagent is in `.claude/agents/`.

## Issues

- Created [[readability-engineer-agent]], with the trial run that decides whether the role stays.
