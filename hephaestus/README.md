# hephaestus

hephaestus is a worker manager: it provisions and manages agents, each with one role, and gives
them sisyphus issues to resolve. It depends on sisyphus, and uses sisyphus issues as its record of
work. It is a separate project that lives beside sisyphus, in this directory of the sisyphus repo.

> **Status:** early. The versioning tools work; the workers do not exist yet. [[agent-delegation]] is the design, and
> [[delegate-issues-to-agents]] is the first epic.

hephaestus is a Go module of its own, in this directory. It runs as a CLI tool, `hephaestus`, and as
a background scheduler daemon that starts the workers on their triggers.

## hephaestus and sisyphus

The two tools must not overlap:

| sisyphus | hephaestus |
|---|---|
| A tool to interact with issues: create, change, find, link, and check them. | A worker manager: provisions agents, starts them, gives them work, and records their runs. |
| Knows nothing about roles, agents, or runners. | Knows the roles, and runs the agents. |
| Gives hephaestus data: fields, `metadata`, claims, filters, and JSON. | Uses only the sisyphus CLI or `sisyphus-mcp` to read and change issues. |

All agent-specific work is in hephaestus. The only sisyphus work that hephaestus can ask for is
generic issue data that any client can use, for example a filter on a `metadata` key. See
[[0029-sisyphus-does-not-manage-agents]].

## Commands

Build it with `go build ./cmd/hephaestus` in this directory. Every command takes `-C <directory>`,
as `git -C` does.

| Command | What it does |
|---|---|
| `hephaestus version list` | List the unreleased changelog entries, in the order that `assign` uses by default. |
| `hephaestus version assign [<entry>...]` | Give the next SHAME versions to unreleased changelog entries, in the given order: move each to `changelog/<version>.md`, add the `CHANGELOG.md` entry, set `VERSION`, and change every link to it. `--dry-run` and `--json` show the plan. |
| `hephaestus decisions renumber [--onto main]` | Give a new number to each decision record that has the number of another one in its directory and that the base does not have, and change every link to it. `--dry-run` and `--json` show the plan. |

The merge queue runs `version assign` and `decisions renumber` as it lands work: versions come
from the merge queue ([[0001-versions-come-from-the-merge-queue]]).
A human who lands work by hand runs them too.

## Layout

```
hephaestus/
  README.md     This page
  go.mod        The Go module of hephaestus
  cmd/          The hephaestus command
  internal/     Its packages
  design/       Design documents of hephaestus
  issues/       The issues of hephaestus, a sisyphus tracker of its own
```

The issues of hephaestus are in their own tracker. Run `sisyphus` in this directory to work on
them, for example `sisyphus list`. The rules of the repo apply ([[AGENTS]] and
[[CONTRIBUTING]]): `sisyphus lint` must report no problems, here and at the
root of the repo.
