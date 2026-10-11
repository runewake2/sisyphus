# hephaestus

hephaestus is a worker manager: it provisions and manages agents, each with one role, and gives
them sisyphus issues to resolve. It depends on sisyphus, and uses sisyphus issues as its record of
work. It is a separate project that lives beside sisyphus, in this directory of the sisyphus repo.

> **Status:** design. No code yet. [[agent-delegation]] is the design, and
> [[delegate-issues-to-agents]] is the first epic.

## hephaestus and sisyphus

The two tools must not overlap:

| sisyphus | hephaestus |
|---|---|
| A tool to interact with issues: create, change, find, link, and check them. | A worker manager: provisions agents, starts them, gives them work, and records their runs. |
| Knows nothing about roles, agents, or runners. | Knows the roles, and runs the agents. |
| Gives hephaestus data: fields, `metadata`, claims, filters, and JSON. | Uses only the sisyphus CLI or `sisyphus-mcp` to read and change issues. |

All agent-specific work is in hephaestus. The only sisyphus work that hephaestus can ask for is
generic issue data that any client can use, for example a filter on a `metadata` key. See
[0029: sisyphus does not manage agents](../design/decisions/cli/0029-sisyphus-does-not-manage-agents.md).

## Layout

```
hephaestus/
  README.md     This page
  design/       Design documents of hephaestus
  issues/       The issues of hephaestus, a sisyphus tracker of its own
```

The issues of hephaestus are in their own tracker. Run `sisyphus` in this directory to work on
them, for example `sisyphus list`. The rules of the repo apply ([AGENTS.md](../AGENTS.md) and
[CONTRIBUTING.md](../CONTRIBUTING.md)): `sisyphus lint` must report no problems, here and at the
root of the repo.

Links to sisyphus documents are relative Markdown links, and sisyphus issues are named in code
text, for example `cli-and-docs-ux`, because `sisyphus lint` in this tracker does not resolve
wikilinks to files outside it (sisyphus issue `nested-tracker-links`).
