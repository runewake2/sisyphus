# sisyphus

> **DO NOT USE.** This is a personal project, built for my own workflow with no
> guarantees of stability, support, or correctness. Expect breaking changes
> without notice. Use at your own risk.

A small CLI for managing the issues in `issues/` and managing work in a more useful way.

## Usage

Install it, then initialize a repo. Init writes the workflow kit: [[CONTRIBUTING]], [[AGENTS]],
`issues/`, [[CHANGELOG]], and `VERSION`.

```bash
go install github.com/runewake2/sisyphus/cmd/sisyphus@latest
sisyphus init
```

Create and move issues through `issues/open/`, `issues/in-progress/`, and `issues/closed/`:

```bash
sisyphus new <issue-name> --title "<title>" --priority high --tags "<component>"
sisyphus update <issue-name> in-progress --bookmark <bookmark> --owner <you>
sisyphus update <issue-name> closed --resolution completed   # or: abandoned
sisyphus show <issue-name>
```

Find issues:

```bash
sisyphus list --state open,in-progress --priority critical,high
sisyphus search "timeout" --section summary
```

Relate an issue to another issue, or to an external tracker:

```bash
sisyphus parent <issue-name> <parent-issue-name>
sisyphus depends-on <issue-name> <blocking-issue-name>
sisyphus remote <issue-name> <url>
```

Check wikilinks:

```bash
sisyphus resolve "[[<issue-name>]]"
sisyphus links <file>
```

Run `sisyphus <command> --help` for all options on any command. See [[CONTRIBUTING]] for the full
workflow: the design log, decisions, versioning, the changelog, wikilinks, issues, and jj.
