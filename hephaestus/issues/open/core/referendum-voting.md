---
title: "Let delegated agents vote on an implementation in a referendum"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[core/delegation-model]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Let delegated agents vote on an implementation in a referendum

## Summary

Add referendums: a series of delegated agents, required and non-required reviewers, vote yes or no on an implementation, or submit a modification. The required reviewers must accept a modification unanimously, and then every reviewer votes again on the modified version. Each agent evaluates the contents with its own scope and intents, against the current implementation. [[referendum]] is the design.

## Context

Requested by a human on 2026-10-11. One agent of one role judges a change only in its own scope; a referendum gathers the judgments of several roles, for example the engineer, the technical docs agent, and the UX researcher, before a change goes on, for example to the merge queue ([[merge-queue-agent]]). The design has the terms, the steps of a round, the modification gate, the result, the record, an example, and the open questions R1-R8. A referendum uses only generic sisyphus data: `metadata`, `edit --append`, and wikilinks.

## Acceptance criteria

- [ ] A decision record answers R1-R8.
- [ ] hephaestus can call a referendum on an issue with a set of voters, run its rounds, and record each round in the issue.
- [ ] A referendum has required and non-required reviewers; it passes only when every required reviewer votes yes, and a non-required no is a concern, not a fail.
- [ ] A modification goes on only when every required reviewer accepts it; it then makes a new version and a new round with every reviewer, and earlier ballots are void.
- [ ] Reviewers do not see the ballots of other reviewers in the same round.
- [ ] After the maximum number of rounds, or with no result, a human decides.
- [ ] A test run with three role agents on a real bookmark, checked by a human.

## Out of scope

- Votes of humans. A human can overrule a result, but does not cast a ballot.
- New sisyphus fields or commands.

## Notes

2026-10-11, from the human: a referendum has required reviewers, which must all vote yes to pass it, and non-required reviewers, which are only related at its edges: they can suggest modifications, but their no cannot fail it. Every modification must first be accepted unanimously by the required reviewers; then the larger vote, of all reviewers, decides whether the edit is accepted at all. This answers R1 and R6 in [[referendum]], and adds the modification gate and the open questions R7 and R8.

2026-10-11, from the human: the entire referendum fails when the proposal does not pass on the edited version (R7), and required reviewers always vote again (R8). Both proposals of [[referendum]] are accepted.

2026-10-11, from the human: modifications occur one by one, and each is reviewed individually (R2 in [[referendum]]).

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
