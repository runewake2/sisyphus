# Referendum

- Date: 2026-10-11
- Issue: [[referendum-voting]]
- Status: Draft. All open questions (R1-R8) are answered.

## Summary

The referendum is how hephaestus coordinates many concurrent agents (from the human, 2026-10-11).
When the work of several agents must agree, a human or an allowed agent calls a referendum, with
the agents whose scope the work touches as its reviewers. The agents then agree, or the referendum
fails, through one recorded procedure. Messages between agents carry information
([[agent-delegation#Messages between workers]]); the referendum is how they decide.

A **referendum** lets a series of delegated agents vote on an implementation before it goes on.
The agents are **reviewers** of two kinds:

- **Required reviewers** must all vote `yes` for the referendum to pass. One `no` of a required
  reviewer fails it.
- **Non-required reviewers** are only related to the referendum at its edges. They can vote, and
  can submit modifications, but their `no` cannot fail the referendum.

Any reviewer can submit a **modification** of the contents. A modification goes on only when every
required reviewer accepts it. Then the larger vote, of all reviewers, runs again on the modified
contents, and decides whether the edit is accepted at all.

Each reviewer evaluates the contents with its own scope and intents, against the current
implementation. A referendum is hephaestus work. sisyphus only stores the issue and its data
([[agent-delegation#What hephaestus needs from sisyphus]]).

## Terms

| Term | Meaning |
|---|---|
| Referendum | One vote on one implementation, from the call to the result. It has one or more rounds. |
| Contents | What the reviewers vote on: the issue, the question, and the implementation (a bookmark and the change ids of its head), as one numbered version: v1, v2, and so on. |
| Reviewer | An agent with one role ([[agent-delegation]]) that hephaestus gives a ballot: an AI agent, or a tool agent, for example a lint, which votes with its exit code (Q9 of [[agent-delegation]]). The call makes each reviewer required or non-required. |
| Required reviewer | A reviewer whose `yes` the referendum needs. It also decides, with the other required reviewers, which modifications go on. There is at least one. |
| Non-required reviewer | A reviewer that is only related to the referendum at its edges. Its ballot is advice: it can submit modifications, but its `no` does not fail the referendum. |
| Scope | The part of the work that a reviewer judges, from the charter of its role: for example the docs for `technical-docs`. |
| Intents | What a reviewer wants the work to achieve in its scope: for example "the docs agree with the behavior". |
| Ballot | The vote of one reviewer in one round: `yes`, `no`, or `modify`, with its reasons. A non-required reviewer can also `abstain`. |
| Round | One vote of every reviewer on one version of the contents: the larger vote. |
| Modification gate | The vote of the required reviewers on one modification: `accept` or `reject`. It must be unanimous to accept. |
| Current implementation | The code and docs that the contents would change: the base of the bookmark, normally `main`. |

## How a referendum runs

1. **Call.** A human, or an agent that a human allows to call one (R4), calls a referendum on an
   issue: the
   question (for example "Merge `samw/ai/fix-login` as the fix of `fix-login-bug`?"), the
   implementation, the required reviewers, and the non-required reviewers. hephaestus records the
   call (see [[#Record]]).
2. **Round.** hephaestus gives every reviewer the contents of the current version and starts them
   at the same time. A reviewer does not see the ballots of the other reviewers in the same round,
   so that no reviewer follows another.
3. **Evaluate.** Each reviewer compares the contents with the current implementation, only in its
   own scope, and against its own intents:
   - Does the change do what the issue asks, in my scope?
   - Does it break something in my scope that works now?
   - What is missing in my scope?
4. **Ballot.** Each reviewer casts one ballot:
   - `yes`: in my scope, the change is right as it is.
   - `no`: in my scope, the change must not go on, with the reasons. It does not change the
     contents. From a non-required reviewer, it is recorded as a concern.
   - `modify`: in my scope, the change can go on with this modification. The ballot has the
     modification: a change to the implementation (new changes on the bookmark), or to the text of
     the issue (for example its acceptance criteria), and the reason.
   - `abstain`: only a non-required reviewer, when nothing of the change is in its scope.
5. **Modifications.** When the round has at least one `modify`, the modifications go through the
   gate first ([[#Modifications]]). When the required reviewers accept one, it makes a new version,
   and a new round starts with every reviewer on it: the larger vote. All ballots of the earlier
   round are void, also the `yes` ballots: they were for other contents.
6. **Count.** A round with no modification that the gate accepted has a result ([[#Result]]).
7. **Result.** hephaestus records the result in the issue, and the next step goes on: for example,
   the bookmark goes to the merge queue, or the issue goes back to its owner.

## Modifications

Any reviewer, required or not, can submit a modification. A modification goes on in two steps:

1. **The gate: the required reviewers accept it unanimously.** hephaestus gives the modification
   to every required reviewer, also to its author when the author is a required reviewer. Each
   votes `accept` or `reject`, in its own scope, with the reasons. One `reject` stops the
   modification: it is recorded with the reasons, and the contents do not change.
2. **The larger vote decides whether the edit is accepted at all.** When the gate accepts the
   modification, it makes a new version, and every reviewer votes on it in a new round. The edit is
   accepted only when that round passes ([[#Result]]): every required reviewer votes `yes` on the
   modified contents. When it fails, the edit is not accepted, and the whole referendum fails (R7).

The other rules:

- **In scope.** A reviewer may only modify what is in its scope: `technical-docs` changes the docs,
  `ux-researcher` the help text or the messages, and so on. A modification outside its scope is a
  `no` (or a concern, from a non-required reviewer) with the change as its reason.
- **One by one, each reviewed alone** (R2, answered by the human on 2026-10-11). Modifications are
  never combined. When a round has more than one, they go to the gate one at a time, in the order of
  the reviewers, required reviewers first, and the gate reviews each one alone. The first that the
  gate accepts makes the next version; the others go back to their authors, who can submit them
  again against the new version, where each again gets its own gate.
- **A rejected modification.** When the gate rejects every modification of a round, the round is
  counted as it is, and a `modify` ballot of a required reviewer counts as not `yes`. So a required
  reviewer whose modification was rejected must vote again on the same version, `yes` or `no`, in
  a new round (R8). A rejected modification of a non-required reviewer is recorded as a concern,
  and goes with the result.
- **A modification is a new version.** It gets the next version number, the change ids of the new
  implementation, and the name of its author.
- **Rounds are limited.** After 3 rounds (R5, answered by the human on 2026-10-11: try 3 for now),
  hephaestus stops the referendum
  with the result `no result`, and asks a human, with every version, every gate, and every ballot.

## Result

Answered by the human on 2026-10-11 (R1):

- A referendum **passes** when every required reviewer votes `yes` in a round with no modification
  that the gate accepted.
- It **fails** when at least one required reviewer votes `no` in such a round. This includes the
  larger vote on a modified version: then the whole referendum fails, and does not go back to the
  version before the edit (R7, answered by the human on 2026-10-11).
- A non-required reviewer cannot fail a referendum. Its `no` is recorded as a concern, and goes with
  the result to the owner of the issue.

A referendum that fails goes back to the owner of the issue with the reasons of each `no`. A human
can always overrule a result, and the record shows that a human did.

## Record

Every referendum is recorded in the repo, as [[AGENTS]] asks for every result:

- The issue gets the state of its referendum in `metadata`, for example
  `referendum=round-2` and then `referendum=passed`, with
  `sisyphus update <name> --metadata referendum=passed`.
- The `Notes` of the issue get one dated line for each round and each gate, with
  `sisyphus edit <name> notes "<line>" --append`: the version, the ballots, and what changed.
- The full record (the reviewers and their kind, the contents of each version, each gate, and each
  ballot with its reasons) is a historical record that the Notes link to (R3, answered by the human
  on 2026-10-11, for now): one file for each referendum in `referendums/` of the repo, for example
  `referendums/2026-10-11-fix-login-bug.md`. hephaestus adds to it while the referendum runs, and
  nobody changes it after the result. A later referendum on the same issue gets a new file.

sisyphus needs no new field or command for this: `metadata`, `edit --append`, and wikilinks are
enough. A filter on `metadata` (sisyphus issue [[filter-by-metadata]]) lets hephaestus find the open
referendums.

## Example

The issue `fix-login-bug` has the bookmark `samw/ai/fix-login`. A human calls a referendum with
the required reviewers `engineer` and `technical-docs`, and the non-required reviewer
`ux-researcher`.

| Step | Version | engineer (required) | technical-docs (required) | ux-researcher (not required) | Outcome |
|---|---|---|---|---|---|
| Round 1 | v1: the bookmark as it is | `yes` | `modify`: add the new `--reset` flag to `docs/login.md` | `no`: the error message is unclear | A modification: to the gate |
| Gate | the docs change | `accept` | `accept` | (does not vote) | Accepted: v2 |
| Round 2 | v2: v1 + the docs change | `yes` | `yes` | `modify`: the error message names the wrong flag | A modification: to the gate |
| Gate | the message change | `reject`: the message is in the scope of another issue | `accept` | (does not vote) | Rejected: round 2 is counted as it is |
| Result | v2 | | | | Passed, with two concerns of `ux-researcher` |

- In round 1, the `no` of `ux-researcher` does not fail the referendum: it is not a required
  reviewer. Its concern goes with the result.
- In round 2, `engineer` votes again although it voted `yes` in round 1, because the contents
  changed.
- The gate on the message change needs both required reviewers. `engineer` rejects it, so it does
  not go on. Round 2 is then counted as it is: both required reviewers voted `yes` on v2, so the
  referendum passes, and the docs edit is accepted. The rejected modification of `ux-researcher`
  goes with the result as a concern.

## Open questions

A human answers a question with a reply such as "R2 yes", or another answer.

| Id | Question | Proposed answer |
|---|---|---|
| R1 | What passes a referendum? | Answered by the human on 2026-10-11: every required reviewer votes `yes`; a non-required reviewer cannot fail it. |
| R2 | When several reviewers modify in one round, gate and apply one at a time, or combine all that do not conflict? | Answered by the human on 2026-10-11: one by one, each reviewed alone, never combined. The order is the order of the reviewers, required reviewers first. Each version then has one change that every reviewer saw. |
| R3 | Where is the full record of a referendum: a section of the issue, a document beside the issue, or a file of hephaestus? | Answered by the human on 2026-10-11, for now: a historical record, one file for each referendum in `referendums/`, for example `referendums/2026-10-11-fix-login-bug.md`, linked from the Notes of the issue and never changed after the result. |
| R4 | Who can call a referendum, and choose its required and non-required reviewers: only a human, or also a role agent, for example the merge queue before each merge? | Answered by the human on 2026-10-11: humans, and the agents that a human allows. The allowance is in the role description of the agent, which can also propose default reviewers. |
| R5 | What is the maximum number of rounds? | Answered by the human on 2026-10-11: try 3 rounds, then a human decides. |
| R6 | Can a reviewer abstain? | Answered with R1: a non-required reviewer can, when nothing is in its scope. A required reviewer cannot: anything but `yes` stops the pass. |
| R7 | When the larger vote on a modified version fails, does the referendum fail, or does it go back to the version before the edit and vote on it again? | Answered by the human on 2026-10-11: the whole referendum fails, with the reasons. It does not go back to the version before the edit. The owner of the issue can call a new referendum. |
| R8 | After the gate rejects a modification of a required reviewer, must that reviewer vote again on the same version, or does its `modify` count as `no`? | Answered by the human on 2026-10-11: required reviewers always vote again. The reviewer votes on the same version, `yes` or `no`, in a new round. The rounds stay limited by R5. |
