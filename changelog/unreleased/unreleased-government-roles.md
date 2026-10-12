# Unreleased: More roles for the government, and how the roles work together

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Seven more hephaestus roles to try (triage, compatibility-reviewer, dependency-check, release-manager, performance-engineer, operations-auditor, model-eval), a `Works with` section that each role file uses to tell its hand-offs, and a `merge-queue-done` broadcast.

## Summary

A human asked for more roles that may be useful, for a government of agents that coordinate and
work together. Each new role has an issue for a trial run, and may be removed after a review. Each
role file now tells how its role works with the others, and hephaestus checks that the roles that
it names exist.

## Changes

- New AI roles:
  - `triage` makes each new issue ready for work, and proposes its priority and role.
  - `compatibility-reviewer` finds the breaks of commands, flags, exit codes, JSON output, formats,
    MCP tools, and messages.
  - `release-manager` checks a line of the merge queue, and writes the summary that a human reads
    before moving `main`.
  - `performance-engineer` measures, profiles, and fixes slow commands on large repos, with
    evidence before and after.
  - `operations-auditor` reviews how the roles work, and proposes changes to them.
- New tool roles:
  - `dependency-check` runs `govulncheck` and `go mod verify` on each Go module, and lists newer
    versions of direct dependencies. A failure goes to `security-researcher`.
  - `model-eval` runs the task suite of sisyphus-mcp with one local model, one at a time.
- `## Works with`, an optional section of a role file: one item for each role or family that the
  role works with. `hephaestus roles` reports an item that names no role or family, and an AI role
  gets the section with its instructions. Every role file has it now.
- `hephaestus merge-queue` broadcasts `merge-queue-done` after a run that makes a line, with the
  line, its base, and the bookmarks that it merged and left out. `release-manager` listens for it.
- [[agent-delegation#The government]] has the new roles, and the path of a piece of work through
  them, from intake to release and oversight. [[role-file]] describes `Works with`.

## Issues

- Created [[triage-agent]], [[compatibility-reviewer-agent]], [[dependency-check-tool]],
  [[release-manager-agent]], [[performance-engineer-agent]], [[operations-auditor-agent]], and
  [[model-eval-tool]].
