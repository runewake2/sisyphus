# 0021: Shell completion for five shells from one source

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.78]]

## Context

Issue names are long, and every command that changes an issue needs one, but shell completion was
off ([[shell-completion]]). The completion must work in bash, zsh, fish, PowerShell, and Nushell.
cobra writes scripts for the first four, but not for Nushell.

## Decision

- sisyphus has its own `completion <shell>` command in place of the default command of cobra. It
  calls the cobra generators for bash, zsh, fish, and PowerShell, and prints a Nushell script.
- The Nushell script is an external completer that calls the hidden `sisyphus __complete` command,
  which the cobra scripts call too. So all five shells get the same candidates from one place: the
  `ValidArgsFunction` and flag completion functions in `completion.go`.
- The Nushell completer keeps the external completer that was set before it, for other commands.
- Issue names come from the index each time. Titles come with the names only when 200 or fewer
  issues match, because a title costs a read of the issue file.

## Alternatives considered

- **The default completion command of cobra.** It has no Nushell, and a subcommand cannot be added
  to it before it exists. Its help text does not follow the rules of this repo.
- **A static Nushell script with `extern` definitions.** It would have the commands and flags, but
  not the issue names, and it would go out of date with each new flag.
- **A cache of issue names for completion.** Faster on large repos, but it can show issues that moved
  or no longer exist. Persistent state is a separate decision.

## Consequences

- New commands and flags complete in every shell with no change to the scripts. A new command that
  takes an issue must be in `commandArguments`.
- The Nushell script depends on the external completer API of Nushell (tested with 0.116). A user
  must save the script again after an update of sisyphus that changes it.
