# Unreleased: Warn when a message to one worker names a family

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: `hephaestus messages send --role <family>` without `--all` warns that the message reaches no worker, and names the roles of the family.

## Summary

A message to one worker of a family, for example `--role formatter`, reaches no worker, because one
worker has one role. `hephaestus messages send` now warns when it sends one.

## Changes

- When `--role` names a family of roles in the role files, and `--all` is not given, `send` prints a
  warning: the roles of the family, and the two ways to reach them (`--role formatter/golang`, or
  `--role formatter --all`). It sends the message all the same.

## Decisions

- [[0002-broadcasts-reach-each-role-once-and-workers-keep-their-leases-alive]] has the rule.

## Issues

- Closed [[warn-family-role-message]].
