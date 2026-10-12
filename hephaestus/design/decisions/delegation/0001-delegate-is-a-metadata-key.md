# 0001: delegate is a metadata key

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-delegation-model]]

## Context

hephaestus must record which role owns a task, and find the tasks of each role. sisyphus does not
manage agents ([[0029-sisyphus-does-not-manage-agents]]), so it must not get a field for agents. Q1
of [[agent-delegation]] asked between a `metadata` key, a tag, and the `owner`.

## Decision

From the human, 2026-10-11: `delegate` is a key of the sisyphus `metadata` field, for example
`metadata: {delegate: merge-queue}`. It names the role assigned to own the task. The `owner` stays
the worker that works on it. hephaestus sets it with `sisyphus update --metadata delegate=<role>`,
and finds the tasks of a role with the generic sisyphus filter `--metadata delegate=<role>`
([[filter-by-metadata]]).

## Alternatives considered

- **A frontmatter field `delegate` in sisyphus.** A field that only agents use, in a tool that
  does not manage agents.
- **A tag.** Tags name the components of the work; a role in them mixes two meanings.
- **The `owner`.** The owner is the worker instance, and a role has many.

## Consequences

- sisyphus needs only a generic filter on `metadata`.
- `hephaestus delegate` and `hephaestus tasks` are thin wrappers of sisyphus commands.
