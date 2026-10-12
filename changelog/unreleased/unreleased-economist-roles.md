# Unreleased: Economist roles

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Three economist roles analyze cost against a budget: `economist/tokens` for AI agent tokens, `economist/cloud` for cloud services such as AWS, and `economist/compute` for CPU, RAM, GPU, and storage.

## Summary

The government of hephaestus gets a family of economists. Each analyzes one kind of cost against
a budget, reports it with evidence, and proposes savings as issues. A human decides.

## Changes

- `hephaestus/roles/economist/tokens/role.md`: the tokens and money of each AI role, referendum
  round, and gate, from the usage that the runners report. It reads only counts, never the text of
  a prompt.
- `hephaestus/roles/economist/cloud/role.md`: the cost of cloud services from the reports of the
  provider, through read-only access only. It asks before a paid cost API call.
- `hephaestus/roles/economist/compute/role.md`: the time, CPU, RAM, GPU and VRAM, and disk of each
  command, worker, and stored file. It runs one heavy measurement at a time, with at most 20
  threads.
- `hephaestus/roles/economist/README.md`: the rules of every economist.
- [[agent-delegation#The government]] has the family. Its subagents are in `.claude/agents/`.

## Issues

- Created [[economist-roles]].
