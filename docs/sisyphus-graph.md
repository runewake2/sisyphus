# sisyphus graph

Draw an issue, everything below it, and the path above it, so a large task and what it takes to
finish it can be reviewed at a glance.

## Usage

```bash
sisyphus graph <name> [--full] [--json | --mermaid]
```

`<name>` is any form of an existing issue (see [[commands#Naming an issue]]).

## What it does

Every arrow points from the issue that comes first to the issue that comes after it:

- A solid arrow goes from a parent to its sub-issue, so an epic points to its tasks and a task to
  its sub-tasks.
- A dotted arrow goes from an issue to the issue that depends on it. The issue it points to cannot
  start until the issue it comes from closes.

`sisyphus graph <name>` then draws:

1. `<name>`, in a double box.
2. Everything below `<name>`: its sub-issues and the issues that depend on it, and theirs, all the
   way down.
3. The path above `<name>`: its parent and the issues it depends on, and theirs, all the way up.
   This is what it takes to get to `<name>`.

Other issues linked to these, such as a sibling under the same parent, are not drawn. A note under
the legend says how many. With `--full`, every issue linked to `<name>` by any path of `parent` or
`depends-on` links is drawn.

So `graph` on an epic draws the whole epic. `graph` on an issue deep in the epic draws only that
issue's part of it.

Issues with no path of links to `<name>` are never drawn or counted. A `parent` or `depends-on` link
to an issue that does not exist is drawn as a box with the state `?`.

Each leaf is in a heavy box. A leaf is an issue that is not closed, whose dependencies are all
closed, and whose sub-issues are all closed: nothing is left that it waits on, so work on it can
start now. A dependency that does not exist does not block, as in `sisyphus list --blocked`.

| Box | Issue |
| --- | --- |
| `╔═╗` double | `<name>` |
| `┏━┓` heavy | A leaf |
| `┌─┐` light | Any other issue |
| `╭─╮` rounded | A directory, with its name in the top border |

No box is both double and heavy. So if `<name>` is a leaf, its box is double, and the legend says
`(a leaf)` after its name.

Each directory that holds a drawn issue is a rounded box around the issues and directories below
it, nested as the directories nest. So an issue box shows only the file name of its issue, and the
legend gives the full name of `<name>`.

## Output

On an issue inside an epic:

```
┌────────────────────┐    ┌──────────────────────┐
│auth-overhaul [open]├─┬─►│fix-login-bug [closed]├┄┐
└────────────────────┘ │  └──────────────────────┘ ┆
                       │                           ┆
                       │                           ┆  ╔══════════════════════════════╗    ┏━━━━━━━━━━━━━━━━━━━━━━┓
                       └───────────────────────────┴─►║split-login-form [in-progress]╟───►┃form-validation [open]┃
                                                      ╚══════════════════════════════╝    ┗━━━━━━━━━━━━━━━━━━━━━━┛

╔═╗ split-login-form   ┏━┓ leaf (ready to start)   ──► sub-issue   ┄┄► needed by
1 more linked issue is not drawn. Use --full to draw it.
```

`split-login-form` is a sub-issue of `auth-overhaul`, depends on `fix-login-bug`, and has the
sub-issue `form-validation`. Its sibling `session-timeouts` is not drawn. `form-validation` is
the only leaf drawn: `split-login-form` waits on it, and `auth-overhaul` waits on both.

On the epic:

```
╔════════════════════╗    ┌───────────────────────┐
║auth-overhaul [open]╟─┬─►│fix-login-bug [closed] ├┄┐
╚════════════════════╝ │  └───────────────────────┘ ┆
                       │                            ┆
                       │  ┏━━━━━━━━━━━━━━━━━━━━━━━┓ ┆
                       ├─►┃session-timeouts [open]┃ ┆
                       │  ┗━━━━━━━━━━━━━━━━━━━━━━━┛ ┆
                       │                            ┆
                       │                            ┆  ┌──────────────────────────────┐    ┏━━━━━━━━━━━━━━━━━━━━━━┓
                       └────────────────────────────┴─►│split-login-form [in-progress]├───►┃form-validation [open]┃
                                                       └──────────────────────────────┘    ┗━━━━━━━━━━━━━━━━━━━━━━┛

╔═╗ auth-overhaul   ┏━┓ leaf (ready to start)   ──► sub-issue   ┄┄► needed by
```

With issues in directories, here `web/auth/login-epic`, `web/auth/fix-login-bug`,
`web/split-login-form`, and `mobile/add-dark-mode`:

```
╭─ web ──────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                                │
│ ╭─ auth ────────────────────────────────────────────────╮                                      │
│ │                                                       │                                      │
│ │ ╔═════════════════╗            ┏━━━━━━━━━━━━━━━━━━━━┓ │                                      │
│ │ ║login-epic [open]╟─────┬─────►┃fix-login-bug [open]┝┄┼┄┄┄┐                                  │
│ │ ╚═════════════════╝     │      ┗━━━━━━━━━━━━━━━━━━━━┛ │   ┆                                  │
│ │                         │                             │   ┆                                  │
│ ╰─────────────────────────┼─────────────────────────────╯   ┆                                  │
│                           │                                 ┆                                  │
│                           │                                 ┆      ┌───────────────────────┐   │
│                           └─────────────────────────────────┴─────►│split-login-form [open]├┄┄┄┼┄┐
│                                                                    └───────────────────────┘   │ ┆
│                                                                                                │ ┆
╰────────────────────────────────────────────────────────────────────────────────────────────────╯ ┆
                                                                                                   ┆
                                                                                                   ┆  ╭─ mobile ───────────────────╮
                                                                                                   ┆  │                            │
                                                                                                   ┆  │   ┌────────────────────┐   │
                                                                                                   └┄┄┼┄┄►│add-dark-mode [open]│   │
                                                                                                      │   └────────────────────┘   │
                                                                                                      │                            │
                                                                                                      ╰────────────────────────────╯

╔═╗ web/auth/login-epic   ┏━┓ leaf (ready to start)   ──► sub-issue   ┄┄► needed by
```

The drawing flows left to right, so a large family grows down the terminal rather than across it.
Each issue is placed one column after the farthest issue that points to it, so a chain of
dependencies reads left to right, and most arrows reach only the next column. An arrow that
skips columns runs along the row of the issue it enters, and no other issue sits on that row in
the columns between, so the arrow is one straight line. Only an arrow in a cycle goes out to the
right and comes back along a blank line.

Each directory has lines of its own, so no box of a directory holds an issue from outside it, and
the boxes of two sibling directories do not overlap. An arrow that crosses a border joins it with
`┼`, and never runs along it. Arrows that leave one issue share a line, and so do arrows that enter
one issue. Use `--json` or `--mermaid` when the exact links matter.

## Options

| Option | Effect |
| --- | --- |
| `--full` | Draw every issue linked to `<name>`, not only what is below it and the path above it. |
| `--json` | Print `focus` (the issue name), `nodes` (each with `name`, `title`, `state`, and `leaf`, which is `true` for a leaf and left out otherwise), `edges` (each with `from`, `to`, and `kind`), and `hidden` (how many linked issues are not drawn). An edge goes from the issue that comes first to the one that comes after: for kind `parent`, from the parent to the sub-issue; for kind `depends-on`, from the dependency to the issue that depends on it. |
| `--mermaid` | Print the Mermaid flowchart source instead of drawing it. Paste it into a ` ```mermaid ` block in a Markdown file, and Obsidian or GitHub renders it. The note about issues not drawn becomes a `%%` comment. |

`--json` and `--mermaid` cannot be used together. `--full` works with either.

The epic with `--mermaid`. Each directory becomes a `subgraph`. The `focus` and `leaf` classes
give `<name>` and the leaves thicker borders:

```
graph LR
    n0["auth-overhaul [open]"]
    n1["fix-login-bug [closed]"]
    n2["session-timeouts [open]"]
    n3["split-login-form [in-progress]"]
    n4["form-validation [open]"]
    n0 --> n1
    n0 --> n2
    n0 --> n3
    n1 -.-> n3
    n3 --> n4
    classDef focus stroke-width:5px
    class n0 focus
    classDef leaf stroke-width:3px
    class n2,n4 leaf
```

## Examples

```bash
sisyphus graph auth-overhaul               # the whole epic
sisyphus graph split-login-form            # its part of the epic
sisyphus graph split-login-form --full     # everything linked to it
sisyphus graph auth-overhaul --json
sisyphus graph auth-overhaul --mermaid > auth-overhaul-graph.txt
```
