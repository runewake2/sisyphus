# Unreleased: the CI actions are pinned to their release commits

- Date: 2026-10-20
- Bookmark: samw/ai/pin-github-actions
- Components: repo
- Changelog: The GitHub Actions of ci.yml are pinned to the full commit SHAs of their releases (checkout v4.4.0, setup-go v5.6.0), with the version in a comment, so the workflow cannot run code that a moved tag points at.

## Summary

A tag can move to other code; a commit SHA cannot. Each SHA was verified against the published tags
of the action with git ls-remote, and dependency-check/github-actions votes yes on the repo.

## Issues

- Closed [[pin-github-actions]], found by the first run of dependency-check/github-actions and
  triaged into repo/ on 2026-10-20.
