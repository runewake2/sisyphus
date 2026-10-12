# Unreleased: An abstraction engineer

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: The AI role `abstraction-engineer` reviews changes for code that the repo already has, and proposes clear abstractions for code that repeats, where the boundary makes sense.

## Summary

A role whose whole objective is reuse and clear abstractions. It weighs each abstraction against
its cost, and never shares code across a boundary that a decision record keeps.

## Changes

- `hephaestus/roles/abstraction-engineer.md`:
  - For a change, often from a `bookmark-ready` notice, it records each place where the change
    does again what the repo or the standard library has.
  - For a part of the repo, it files one issue for each proposed abstraction: every place, the name,
    the boundary, the contract, what it removes, what it costs, and the order of the moves.
  - On request, it makes one small abstraction on its own bookmark with no change of behavior.
  - Its rules: reuse first, three or more copies (or copies that must stay the same), one reason to
    change, respect the boundaries, clearer and not only shorter, and behavior stays.
- [[agent-delegation#The government]] has the role. Its subagent is in `.claude/agents/`.

## Issues

- Created [[abstraction-engineer-agent]], with the copies in hephaestus that a trial run should find.
