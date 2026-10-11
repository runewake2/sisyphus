# 0013: Let update change fields without a state

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.40]]

## Context

`sisyphus update <name> <state>` needed the state even to change one field, so a user had to look up
the current state and give it again. A wrong state moved the issue by mistake, and small models had
to be told this rule ([[update-without-state-change]], [[ollama-local-models]]). Restating
`closed` also reset an abandoned issue to `completed` and its closed date to today.

## Decision

The state is optional: `sisyphus update <name> [<state>] [flags]`. Without a state, the state does
not change, and at least one flag must be given. When the state does not change and `--resolution`
is not given, the resolution and the closed date stay. The warnings about closing an issue come only
when the issue goes from another state to `closed`. In `sisyphus_update`, `state` is optional too.

## Alternatives considered

- **A separate `sisyphus set` command for fields.** It would split one concept over two commands,
  and `update` would still need the state for a field change together with a state change.
- **Keep the state required.** It keeps the old, error-prone form.

## Consequences

- Scripts and agents can change a field with one call and no lookup.
- Restating the current state still works, and no longer resets the resolution or the closed date.
