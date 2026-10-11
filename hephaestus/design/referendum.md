# Referendum

- Date: 2026-10-11
- Issue: [[referendum-voting]]
- Status: Draft. The open questions at the end are not answered yet.

## Summary

A **referendum** lets a series of delegated agents vote on an implementation before it goes on.
Each agent votes **yes**, **no**, or submits a **modification**. A modification changes the
contents of the referendum, so every agent must vote again on the modified referendum. Each agent
evaluates the contents with its own scope and intents, against the current implementation.

A referendum is hephaestus work. sisyphus only stores the issue and its data
([[agent-delegation#What hephaestus needs from sisyphus]]).

## Terms

| Term | Meaning |
|---|---|
| Referendum | One vote on one implementation, from the call to the result. It has one or more rounds. |
| Contents | What the agents vote on: the issue, the question, and the implementation (a bookmark and the change ids of its head), as one version. |
| Voter | An agent with one role ([[agent-delegation]]) that hephaestus gives a ballot. |
| Scope | The part of the work that a voter judges, from the charter of its role: for example the docs for `technical-docs`. |
| Intents | What a voter wants the work to achieve in its scope: for example "the docs agree with the behavior". |
| Ballot | The vote of one voter in one round: `yes`, `no`, or `modify`, with its reasons. |
| Round | One vote of every voter on one version of the contents. |
| Current implementation | The code and docs that the contents would change: the base of the bookmark, normally `main`. |

## How a referendum runs

1. **Call.** A human, or an agent that may call one, calls a referendum on an issue: the question
   (for example "Merge `samw/ai/fix-login` as the fix of `fix-login-bug`?"), the implementation,
   and the voters. hephaestus records the call (see [[#Record]]).
2. **Round.** hephaestus gives each voter the contents of the current version and starts the
   voters at the same time. A voter does not see the ballots of the other voters in the same
   round, so that no voter follows another.
3. **Evaluate.** Each voter compares the contents with the current implementation, only in its own
   scope, and against its own intents:
   - Does the change do what the issue asks, in my scope?
   - Does it break something in my scope that works now?
   - What is missing in my scope?
4. **Ballot.** Each voter casts one ballot:
   - `yes`: in my scope, the change is right as it is.
   - `no`: in my scope, the change must not go on, with the reasons. A `no` tells what is wrong; it
     does not change the contents.
   - `modify`: in my scope, the change can go on with this modification. The ballot has the
     modification: a change to the implementation (new changes on the bookmark), or to the text of
     the issue (for example its acceptance criteria), and the reason.
5. **Count.** When every voter has cast a ballot, or the time of the round is over:
   - At least one `modify`: hephaestus applies a modification ([[#Modifications]]), makes a new
     version of the contents, and starts a new round with **every** voter. All the ballots of the
     earlier round are void, also the `yes` ballots: they were for other contents.
   - No `modify`: the result is the count of `yes` and `no` ([[#Result]]).
6. **Result.** hephaestus records the result in the issue, and the next step goes on: for example,
   the bookmark goes to the merge queue, or the issue goes back to its owner.

## Modifications

- **In scope.** A voter may only modify what is in its scope: `technical-docs` changes the docs,
  `ux-researcher` the help text or the messages, and so on. A modification outside its scope is a
  `no` with the change as its reason.
- **One for each round.** When more than one voter submits a modification in the same round,
  hephaestus applies the first one in the order of the voters, and gives the others back to their
  voters. In the next round, each of them can submit its modification again, against the new
  version. This keeps each version the result of one change that every voter has seen.
- **A modification is a new version.** It gets the next version number, the change ids of the new
  implementation, and the name of its voter. The voter that submitted it also votes again.
- **Rounds are limited.** After a maximum number of rounds (proposed: 3), hephaestus stops the
  referendum with the result `no result`, and asks a human, with every version and every ballot.

## Result

The proposed rule (R1): a referendum **passes** when every voter votes `yes` in a round with no
`modify`. It **fails** when at least one voter votes `no` in such a round. A referendum that fails
goes back to the owner of the issue with the reasons of each `no`. A human can always overrule a
result, and the record shows that a human did.

## Record

Every referendum is recorded in the repo, as [AGENTS.md](../../AGENTS.md) asks for every result:

- The issue gets the state of its referendum in `metadata`, for example
  `referendum=round-2` and then `referendum=passed`, with
  `sisyphus update <name> --metadata referendum=passed`.
- The `Notes` of the issue get one dated line for each round, with
  `sisyphus edit <name> notes "<line>" --append`: the version, the ballots, and what changed.
- The full record (the contents of each version, and each ballot with its reasons) is a document
  that the Notes link to (R3).

sisyphus needs no new field or command for this: `metadata`, `edit --append`, and wikilinks are
enough. A filter on `metadata` (sisyphus issue `filter-by-metadata`) lets hephaestus find the open
referendums.

## Example

The issue `fix-login-bug` has the bookmark `samw/ai/fix-login`. A human calls a referendum with
the voters `engineer`, `technical-docs`, and `ux-researcher`.

| Round | Version | engineer | technical-docs | ux-researcher | Count |
|---|---|---|---|---|---|
| 1 | v1: the bookmark as it is | `yes` | `modify`: add the new `--reset` flag to `docs/login.md` | `yes` | A modification: v2, a new round |
| 2 | v2: v1 + the docs change | `yes` | `yes` | `modify`: the error message names the wrong flag | A modification: v3, a new round |
| 3 | v3: v2 + the message change | `yes` | `yes` | `yes` | Passed |

In round 2, `engineer` votes again although it voted `yes` in round 1, because the contents
changed.

## Open questions

A human answers a question with a reply such as "R1 yes", or another answer.

| Id | Question | Proposed answer |
|---|---|---|
| R1 | What passes a referendum: every voter `yes` (unanimous), a majority, or a weight for each role? | Unanimous in a round with no `modify`. A voter that has a concern must say it, as a `no` or a `modify`. |
| R2 | When several voters modify in one round, apply one and rerun, or combine all that do not conflict? | Apply one, in the order of the voters, and give the others back. Each version then has one change that every voter saw. |
| R3 | Where is the full record of a referendum: a section of the issue, a document beside the issue, or a file of hephaestus? | A document beside the issue, for example `referendums/<issue-name>.md` in the repo of the issue, linked from its Notes. The issue stays readable. |
| R4 | Who can call a referendum: only a human, or also a role agent, for example the merge queue before each merge? | A human, and the roles that a human allows in their role description. |
| R5 | What is the maximum number of rounds? | 3, then a human decides. |
| R6 | Can a voter abstain, for example when the change has nothing in its scope? | Yes: an `abstain` ballot, which counts as neither `yes` nor `no`. A referendum with only abstentions has no result. |
