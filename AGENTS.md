# AGENTS.md

Rules for AI agents that work in sisyphus.

**First, read [[CONTRIBUTING]] (`CONTRIBUTING.md`) completely.** It is the shared guide for humans and agents: the design log, decisions, versioning, the changelog, wikilinks, issues and `sisyphus`, the writing style, jj, and change descriptions. All of it applies to you.

This file adds the rules that apply only to agents. If this file and `CONTRIBUTING.md` do not agree, follow this file.

## Code host and main

- Do not push to the code host, and do not open or change pull requests or issues on it, unless a human asks you to.
- Do not move `main`. A human moves `main` and merges finished work, or asks you to do it.
- Track all work in `issues/` (see [[CONTRIBUTING#Issues]]).

## Versioning

**Agents only ever increment the SHAME version.** Only a human bumps PROUD or DEFAULT (see [[CONTRIBUTING#Versioning]]).

## Writing style

Use Simplified Technical English (see [[CONTRIBUTING#Writing style]]) for all text that you write, including your reports to a human.

## Workspaces and bookmarks

Several agents can work in this repo at the same time. These rules keep their work apart.

### 1. Work in your own workspace

Each agent works in exactly one jj workspace, so that agents never edit the same working copy.

- Location: `../workspaces/sisyphus-<workspace-name>`
- `<workspace-name>`: 2–4 words in kebab-case that describe the objective.

```bash
mkdir -p ../workspaces
jj workspace add ../workspaces/sisyphus-<workspace-name> --name <workspace-name> -r main
cd ../workspaces/sisyphus-<workspace-name>
```

Do all of your edits, builds, and jj operations from inside that workspace. Do not edit files in the main checkout or in another agent's workspace. Do not run jj in the main checkout while a human works there.

### 2. Name your bookmark `samw/ai/<workspace-name>`

Put a bookmark on the head change of your work, and move it when you add changes:

```bash
jj bookmark create samw/ai/<workspace-name> -r @
jj bookmark set samw/ai/<workspace-name> -r @
```

Use this bookmark wherever `CONTRIBUTING.md` asks for a bookmark, for example in the `bookmark` field of an issue.

### 3. One linear line of changes per workspace, based on `main`

Within a workspace, keep exactly one chain of changes with exactly one bookmark on the head. Do not create side branches or merge changes. If you need an unrelated line of work, create a new workspace for it. If `main` moves ahead, rebase the whole chain onto it:

```bash
jj rebase -b samw/ai/<workspace-name> -d main
```

Base a workspace on another unmerged bookmark only when the new work needs that work. Say so in the change description.

### 4. Take an issue only if no other agent has it

```bash
sisyphus list --state in-progress
```

If `<issue-name>` appears in the output, another agent already has the issue. Select a different
issue. Otherwise, start it:

```bash
sisyphus update <issue-name> in-progress --bookmark samw/ai/<workspace-name> --owner <you> --workspace <workspace-name>
```

### 5. Remove your workspace when the work is complete

The work is complete only when `main..samw/ai/<workspace-name>` is empty. Then remove the workspace from outside it:

```bash
jj workspace forget <workspace-name>
rm -rf ../workspaces/sisyphus-<workspace-name>
```

The changes and the bookmark stay in the repo. Never forget or delete a workspace of another agent.

## Agent checklist

Complete the checklist in [[CONTRIBUTING#Checklist for every change]]. Also:

- [ ] Working inside `../workspaces/sisyphus-<workspace-name>`
- [ ] Exactly one bookmark, `samw/ai/<workspace-name>`, on the head change
- [ ] One linear chain of changes in this workspace
- [ ] The issue was free before you started it
- [ ] Only the SHAME version changed
- [ ] No push and no change on the code host, unless a human asked
- [ ] Reports to humans use Simplified Technical English
- [ ] Once `main..samw/ai/<workspace-name>` is empty: workspace forgotten and its directory removed
