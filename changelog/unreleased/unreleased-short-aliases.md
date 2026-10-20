# Unreleased: sis and forge, the short names of the commands

- Date: 2026-10-20
- Bookmark: samw/ai/short-aliases
- Components: sisyphus, hephaestus
- Changelog: The documented short names of the two commands are `sis` (sisyphus) and `forge` (hephaestus): symlinks, with one completion line per alias, as the Aliases section of the README shows. The docs and the role files keep the full names.

## Summary

Symlinks instead of shell aliases, because neither binary cares about the name it is called by, and
a symlink works in scripts, CI, and agent runners, where an alias does not expand. On a machine with
Foundry, its `forge` collides; the README names `hep` as the fallback there.

## Issues

- Closed [[short-command-aliases]].
