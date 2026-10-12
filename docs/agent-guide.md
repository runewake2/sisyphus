# AI agents

How to let AI agents use sisyphus in your repo. sisyphus keeps the issues in the repo, so an agent
reads and changes them with the same commands as a person. This page gives text to paste into the
instructions of your agents, and tells how several agents work together.

## Set up

1. Install `sisyphus`, and run `sisyphus init` in the repo.
2. Give the agent the commands, in one of two ways:
   - **Shell**: an agent that can run shell commands uses `sisyphus` directly. Paste
     [[#Instructions for a shell]] into the `AGENTS.md` or `CLAUDE.md` of your repo.
   - **MCP**: connect [[sisyphus-mcp]]. The server gives the agent short instructions itself. Paste
     [[#Instructions for MCP]] for the rules of your repo.
3. Give each agent its own owner name in `SISYPHUS_OWNER`, for example in the environment of the
   agent harness, or in the `env` of the MCP server. Then each start has an owner, also when the
   agent forgets `--owner`. Use a different name for each agent that runs at the same time.

## Instructions for a shell

````markdown
## Issues

This repo tracks its issues with sisyphus, in `issues/`. Do not edit the issue files by hand; use
the commands. Each command that changes an issue prints the path, and warnings on stderr. Read the
warnings.

1. Find work that can start now, and see what other agents work on:

   ```sh
   sisyphus list --ready --state open
   sisyphus claims
   ```

2. Read the issue before you start: `sisyphus show <name>`.
3. Start it with your name and branch. This claims the issue:

   ```sh
   sisyphus start <name> --owner <you> --bookmark <branch>
   ```

   If `start` fails because another owner or workspace has the issue, select another issue. Do not
   use `--take` unless a human tells you to.
4. Record progress and findings in the Notes section:

   ```sh
   sisyphus edit <name> notes "<date>: <text>" --append
   ```

   Never run `sisyphus edit <name>` without a section: it opens an editor and waits.
5. For work that you find but that is out of scope, create an issue:

   ```sh
   sisyphus new --title "<title>" --deferred-from <name>
   ```

6. When the work is complete, close the issue and tell what was done. If the work stops, abandon
   it and tell why:

   ```sh
   sisyphus close <name> -m "<what was done>"
   sisyphus abandon <name> -m "<why the work stopped>"
   ```

7. Check the repo: `sisyphus lint`.
````

The short version, for a file that has little space:

```markdown
Track work with sisyphus: `sisyphus list --ready --state open` to find work,
`sisyphus start <name> --owner <you>` to claim it (select another issue if it fails),
`sisyphus edit <name> notes "<text>" --append` for progress, and
`sisyphus close <name> -m "<what was done>"` when it is complete.
```

## Instructions for MCP

The server already tells the agent the workflow ([[sisyphus-mcp#Instructions]]). Add only the
rules of your repo, for example:

```markdown
## Issues

Use the sisyphus tools for all issue work. Start an issue with sisyphus_start before you change
code for it, and give it your branch as bookmark. If sisyphus_start fails, select another issue.
Close each issue with sisyphus_close and a message that links to the commit.
```

## Several agents at the same time

sisyphus stops two agents from taking one issue:

- **One working copy**: a start of an issue that another owner has in progress fails
  ([[sisyphus-start]]). Commands lock the repo while they change issues, so two starts at the same
  moment cannot both succeed.
- **Several workspaces**: give each agent its own jj workspace or git worktree. A start records a
  claim that every workspace of the repo sees before the work merges, and a start of an issue that
  another workspace has claimed fails ([[sisyphus-claims]]). `list --ready` leaves out claimed
  issues.
- **Several machines**: claims are local to one machine. Agents on different machines can start
  the same issue until the work merges.

When an agent stops without closing its issue, the issue stays in progress. A human can stop it
with `sisyphus update <name> open`, or give it to another agent with
`sisyphus start <name> --owner <other> --take`. A claim of a removed workspace is stale and does not
block anyone; `sisyphus claims --prune` removes it.

## What sisyphus does not do for agents

- It does not run agents, and it does not select the issue for an agent.
- It does not stop an agent that changes the issue files by hand. `sisyphus lint` finds the
  problems that such a change makes.
