# sisyphus completion

Print the script that completes sisyphus commands, flags, issue names, and flag values in a shell.

## Usage

```bash
sisyphus completion <shell>
```

`<shell>` is `bash`, `zsh`, `fish`, `powershell`, or `nushell`.

## Flags

None.

## Behavior

The script completes:

- Command names and flags.
- Issue names, with their titles when 200 or fewer match. Each command offers only the issues that
  it can change: [[sisyphus-start]] offers open issues, [[sisyphus-close]] and [[sisyphus-abandon]]
  offer open and in-progress issues, and every other command offers all issues. `--parent`,
  `--depends-on`, and `--deferred-from` offer all issues.
- The new state of [[sisyphus-update]], and the section of [[sisyphus-edit]] from the headings of
  the issue.
- The values of `--state`, `--priority`, `--effort`, `--resolution`, and `--color`. On
  [[sisyphus-list]] and [[sisyphus-search]], `--state` and `--priority` complete each item of a
  comma-separated list.
- `--tags`, `--add-tags`, and `--remove-tags` complete the built-in tags and every tag that an issue
  has, as a comma-separated list.
- `-C`/`--directory` and the `--dir` of [[sisyphus-init]] complete directories. With `-C`, the issue
  names come from the repo in that directory.

Each script calls sisyphus to get the candidates, so the completions always show the issues as they
are now. The `sisyphus` that the shell finds on its `PATH` must be the one you complete.

### Install

Load the script from the start-up file of the shell:

| Shell | Add this |
|---|---|
| bash | `source <(sisyphus completion bash)` in `~/.bashrc`. It needs the bash-completion package. |
| zsh | `source <(sisyphus completion zsh)` in `~/.zshrc`, after `compinit`. |
| fish | `sisyphus completion fish \| source` in `~/.config/fish/config.fish`. |
| PowerShell | `sisyphus completion powershell \| Out-String \| Invoke-Expression` in the file that `$PROFILE` names. |
| Nushell | See below. |

Nushell cannot load a script from the output of a command at start-up, so save the script once, and
load the file from `config.nu`:

```nu
sisyphus completion nushell | save --force ($nu.default-config-dir | path join sisyphus.nu)
```

Then add this line to `config.nu`:

```nu
source ($nu.default-config-dir | path join sisyphus.nu)
```

The Nushell script sets an external completer. It keeps the external completer that `config.nu` set
before it, for every other command, so source it after that completer. Save the script again after
you update sisyphus.

## Output

The script, on stdout. An unknown shell is an error, with a suggestion when the name is close to a
shell, for example `bsh`.

## Examples

```bash
source <(sisyphus completion bash)
sisyphus completion fish > ~/.config/fish/completions/sisyphus.fish
```
