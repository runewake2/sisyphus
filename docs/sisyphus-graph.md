# sisyphus graph

Print the whole family tree of an issue, so a large task and its sub-issues can be reviewed at a
glance.

## Usage

```bash
sisyphus graph <name> [--json]
```

`<name>` is any form of an existing issue (see [[commands#Naming an issue]]).

## What it does

1. Walks up from `<name>` through `parent` to the top-most ancestor.
2. Prints that ancestor and every issue below it, at any depth, with its state.
3. Marks `<name>` itself, and shows each issue's `depends-on` issues next to it.

Sub-issues are sorted by name. A `depends-on` issue is named even when it is outside the tree.
If the parent links contain a cycle (only possible after editing files by hand), the walk stops at
the repeated issue, which is shown with the state `?` (in JSON, with the title `(cycle; see parent)`).

## Output

```
auth-overhaul [open]
├── fix-login-bug [closed]
├── split-login-form [in-progress]  <-- you asked about this one  (depends on: fix-login-bug)
│   └── form-validation [open]
└── session-timeouts [open]
```

With `--json`, nested objects with `name`, `title`, `state`, `focus` (true for `<name>`),
`depends-on`, and `children`.

## Examples

```bash
sisyphus graph auth-overhaul
sisyphus graph split-login-form     # same tree, marked at split-login-form
sisyphus graph auth-overhaul --json
```
