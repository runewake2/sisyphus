---
title: "Add dependency check tool roles, one for each ecosystem"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, tool-agents, security] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add dependency check tool roles, one for each ecosystem

## Summary

Add a family of tool roles in `hephaestus/roles/dependency-check/`, one for each ecosystem: `golang` (govulncheck, go mod verify), `javascript` (the audit of npm, yarn, or pnpm), `dotnet` (dotnet list package --vulnerable), `java` (OSV-Scanner), `docker` (versions of images; trivy), and `github-actions` (actions pinned to a commit; zizmor). Each fails closed, and a failure goes to security-researcher.

## Context

Requested by a human on 2026-10-12: implement other agents that may be useful, for a government of agents who coordinate and work together. The roles may be removed after a review; a trial run shows how useful each is. The family is described in [[roles/dependency-check/README]]; on the same day the human asked to make dependency-check a family, to move the Go check to golang, and to add javascript, dotnet/csharp, java, docker, and GitHub Actions YAML files. Each role file has its section Works with tells how it works with the other roles.

## Acceptance criteria

- [x] Each ecosystem has a directory with a role file and a check script, and `lib.sh` is shared; the README has the rules of every dependency check.
- [x] On a `bookmark-ready` notice, a check runs only when the change touches a dependency file of its ecosystem.
- [x] Tests with no network: the GitHub Actions check finds unpinned actions and images, the docker check finds images with no version, and every check passes on a repo with nothing to check.
- [x] Tried on fixtures: the JavaScript check found the critical vulnerability of minimist 0.0.8 through npm audit.
- [ ] The Go check reads vuln.go.dev, and the Java check reads api.osv.dev and api.deps.dev: the network of the sandbox blocks them now.
- [ ] The .NET check runs where the .NET SDK is installed.
- [ ] A human checks the result of a trial run, and decides which roles stay.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
