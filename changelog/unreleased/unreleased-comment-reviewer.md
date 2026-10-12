# Unreleased: A comment reviewer for code comments and commit messages

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: The AI role `technical-docs/comments` reviews the code comments and commit messages of a change with no knowledge of the solution, and makes them brief, concise, and correct; the implementer and the engineer run it in a review loop.

## Summary

A variant of the technical docs role reviews only what a change produced: the diff, the code around
it, and the commit descriptions. It does not read the issue or the design, so it reviews as a later
reader does. The implementer and the engineer ask it for a review in a loop.

## Changes

- `hephaestus/roles/technical-docs/comments/role.md`:
  - It checks each new or changed comment against the code, and each commit description against its
    diff, with the rules of [[CONTRIBUTING#Writing style]] and [[CONTRIBUTING#Change descriptions]].
  - It fixes comments only, on its own bookmark on top of the one under review, and names the change
    that each fix belongs to.
  - It writes the full corrected description of each commit that must change, for the author to
    apply.
  - It ends with a verdict, `clean` or `changes`, as one line of JSON.
- The implementer describes its changes, then asks for a review with only the bookmark, the base,
  and the issue in the data. It squashes the fixes, applies the descriptions, runs the checks, and
  asks again, up to three reviews. Then it broadcasts `bookmark-ready` and closes the issue.
- The engineer has its conflict resolutions reviewed the same way.
- [[agent-delegation#The government]] has the role. Its subagent is in `.claude/agents/`.

## Issues

- Created [[comment-reviewer-agent]].
