# Unreleased: A referendum that times out fails

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: `hephaestus referendum call --timeout` (1 hour by default): a referendum that times out or is stopped fails, its reviewers stop, and the issue stays open; a referendum that crashed is recorded as failed after its deadline.

## Summary

A referendum that was stopped left `referendum=round-<n>` on its issue, and every later call was
refused until a human decided it. Now a referendum that times out fails, and the issue stays open
for its owner.

## Changes

- `hephaestus referendum call --timeout <duration>`, 1 hour by default. A referendum that times out,
  or that Ctrl-C or SIGTERM stops, has the result `failed`, with the reason in the Notes and the
  record. The state of the issue does not change, and it gets no `needs-human` tag.
- A stopped command of a role, for example a model runner, stops with every program that its shell
  started, so that it spends no more tokens or CPU. This is so for referendum reviewers and for
  workers.
- The issue gets `referendum-deadline` in its metadata. A call that finds a referendum still in
  `round-<n>` after its deadline, because its process ended with no record, records it as failed
  and goes on. Before the deadline, the call is refused with the deadline.

## Issues

- Closed [[referendum-timeout]].
