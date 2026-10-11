# sisyphus open

Open an issue in your editor, and check the file when the editor exits.

## Usage

```bash
sisyphus open <name>
```

`<name>` takes any form of an issue name (see [[commands#Naming an issue]]), so you do not need to
know which state directory or subdirectory holds the file.

## Flags

None.

## Behavior

- The editor is `$VISUAL`, or else `$EDITOR`. The value can have arguments, for example
  `code --wait` or `subl -w`. A graphical editor must wait until you close the file, or sisyphus
  checks the file before you edit it. Without `$VISUAL` and `$EDITOR`, the command fails.
- When the editor exits, sisyphus checks the file as [[sisyphus-lint]] does for one issue, and
  warns about problems, for example a frontmatter line that it cannot read. It does not change the
  file.
- To change one section from a script or an agent, use [[sisyphus-edit]] instead.

## Output

The output of the editor. Warnings go to stderr. A failed editor is an error.

## Examples

```bash
EDITOR=vim sisyphus open fix-login-bug
VISUAL="code --wait" sisyphus open web/auth/fix-login-bug
```
