# sisyphus depends-on

Record that an issue cannot start until another issue closes, or remove that.

## Usage

```bash
sisyphus depends-on <name> <blocking-issue>           # add
sisyphus depends-on <name> <blocking-issue> --clear   # remove one
sisyphus depends-on <name> --clear                    # remove every dependency
```

Both arguments take any form of an issue name (see [[commands#Naming an issue]]).

## Flags

| Flag | Meaning |
|---|---|
| `--clear` | With `<blocking-issue>`, remove that dependency; without it, remove every dependency. |
| `--json` | Print the changed issue as JSON, as [[sisyphus-show]] `--json` does, with its path and the warnings. |

## Behavior

- An issue can depend on any number of issues. Adding one that is already there changes nothing.
- The blocking issue must exist. A closed blocking issue is allowed, with a warning.
- An issue cannot depend on itself, or on an issue that already depends on it directly or through
  other issues: that would make a cycle.
- `--clear` with a `<blocking-issue>` that the issue does not depend on changes nothing, and warns
  with the dependencies that the issue has.
- Unlike `parent`, `depends-on` is not a hierarchy: it can link any two issues, anywhere in the
  tree, and an issue can depend on several.

The list is written as quoted wikilinks: `depends-on: ["[[fix-login-bug]]", "[[session-timeouts]]"]`.

### Where it shows up

- `sisyphus list --blocked` lists issues that depend on an issue that is not closed yet.
- Closing an issue that open or in-progress issues still depend on gives a warning.
- [[sisyphus-graph]] draws each dependency as a dotted arrow to the issue that depends on it.

## Output

With `--json`, the changed issue as [[sisyphus-show]] prints it with `--json`, plus `path` (the file)
and `warnings` (a list; the warnings do not go to stderr then). See [[0.0.72]].

The path of the issue's file.

## Examples

```bash
sisyphus depends-on split-login-form fix-login-bug
sisyphus depends-on split-login-form fix-login-bug --clear
sisyphus depends-on split-login-form --clear
```
