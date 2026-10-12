# 0001: A referendum runs each reviewer in a scratch worktree

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-referendum-voting]]

## Context

The human answered R1-R8 of [[referendum]]:

- every required reviewer must vote `yes`, and a non-required reviewer cannot fail a referendum (R1, R6);
- modifications go one by one, and each is reviewed alone (R2);
- the full record is a historical file in `referendums/` (R3);
- humans and the agents that a human allows call referendums (R4);
- a referendum has at most 3 rounds (R5);
- a failed vote on an edited version fails the whole referendum (R7);
- a required reviewer whose modification the gate rejected votes again (R8).

[[referendum-voting]] asks hephaestus to run referendums. This record holds the choices that the
answers left open.

## Decision

- **The command.** `hephaestus referendum call <issue>` runs a referendum from the call to the
  result, in one process. It takes the question, `--required` and `--optional` reviewers, and the
  implementation: the bookmark of the issue, or `--bookmark`. The base is `--onto`, `main` by
  default. `hephaestus referendum decide <issue> passed|failed --reason` is for a human: it decides
  after `no-result`, or overrules a result.
- **Isolation.** Each reviewer works in its own scratch git worktree at the commit of the version.
  Worktrees are made in the git directory of the repo, as the merge queue makes them
  ([[0001-merge-with-git-in-a-scratch-worktree]]). All the reviewers of a round start at the same
  time (`--parallel` limits them). The adds and removes of worktrees are serialized, because one
  reviewer's `git worktree prune` can remove the worktree that another reviewer is making. Thus no reviewer sees another ballot, and none changes the working
  copy of the user.
- **Ballots.** A tool reviewer votes with `hephaestus agent review`: `yes` for exit code 0, `no` for
  another code, and `modify` with the changes that its command made. Its command gets the base in
  `HEPHAESTUS_BASE`. At the gate, a tool runs on the version and on the modified version. It
  rejects only when the modification makes its result worse. The order from best to worst is
  `yes`, `modify`, `no`. A `modify` is also worse when it needs to change a new file. Without this
  rule, one formatter would reject the modification of another, because of the files of its own
  language that are still not formatted. An AI reviewer runs the runner
  ([[0004-ai-roles-run-as-claude-code-subagents]]) with a ballot prompt. Its output must end with one
  line of JSON: `{"vote": ..., "reasons": ..., "modification": ...}`. Optionally the line has
  `"section"` and `"text"` for an edit of the issue text. At the gate, the line is
  `{"gate": "accept"|"reject", "reasons": ...}`.
- **A modification is commits.** The commits and changes that a reviewer leaves in its worktree
  become its modification: hephaestus commits what is not committed yet. A modification must be a
  linear line of commits on the version. A `modify` ballot with no change and no issue edit counts as
  `no`. A required reviewer that abstains also counts as `no` (R6).
- **One round of modifications** (from the human, 2026-10-12). Every modification of a round goes to
  the gate alone, against the version of the round. Then the accepted ones are applied in the order
  of the reviewers: the commits of each are cherry-picked on the ones before it, in a scratch
  worktree. One that conflicts, or that edits a section of the issue that an earlier one edits, is
  left out and goes back to its author. The others make one new version for the next round.
- **Versions are branches.** Each version is kept on the branch `referendum/<id>/v<n>`. A new
  version gets the accepted modifications of its round: new commits, issue edits, or both. Issue edits are
  applied to the issue only when the referendum passes, because R7 discards an edited version that
  fails. When the referendum passes on a modified version, the bookmark of the issue becomes the
  branch of that version, so the merge queue takes the accepted implementation.
- **No valid ballot means no result.** When a reviewer gives no valid ballot or gate vote, the result
  is `no-result`, and a human decides. Examples are a runner that fails, or output with no ballot
  line. A missing ballot never counts as a `no`, because a broken runner must not fail a
  referendum.
- **Who calls.** A worker sets `HEPHAESTUS_ROLE`. When it is set, the role file must have
  `can-call-referendum: true`, or the call is refused. When it is not set, the caller is a human.
  `decide` refuses to run under a role.
- **The record.** The issue gets this record:
  - `metadata` `referendum=round-<n>`, then the result;
  - `referendum-record=<id>`;
  - a Notes line for each round and for the result;
  - on `no-result`, the tag `needs-human`.

  The full record is `referendums/<date>-<issue>.md`, with `-2`, `-3`, and so on for a later
  referendum. hephaestus rewrites it after each round. Only `decide` adds to it after the result.
  When the message queue runs, the owner of the issue gets a `referendum` message with the result.
- **Timeout** (from the human, 2026-10-12: a referendum that times out fails, and the issue remains
  open). `--timeout` (1 hour by default), Ctrl-C, and SIGTERM stop the reviewers, with every program
  that they started, and the result is `failed`, with no `needs-human` tag. The state of the issue
  does not change. The issue gets `referendum-deadline` in its metadata. When a call finds a
  referendum that still has `round-<n>` after its deadline, because its process ended with no
  record, it records that referendum as failed and goes on.
- **Exit code.** It is 0 only when the referendum passes, so that a script can do
  `hephaestus referendum call ... && hephaestus merge-queue ...`.

## Alternatives considered

- **Reviewers in the workspace of the issue.** The reviewers would see each other's changes and
  could change the work of the owner.
- **Reviewers as workers on the message queue.** This needs running workers of every reviewer role,
  which [[delegation-runner]] does not start yet. It can come later behind the same `Voter`
  interface.
- **A missing ballot counts as `no`.** A broken runner would then fail good work.
- **Apply issue edits when the gate accepts them.** The edit would stay in the issue when the vote
  on the edited version fails (R7).
- **Move the bookmark of the owner.** That bookmark belongs to another workspace. Changing the
  `bookmark` field of the issue does not touch it.

## Consequences

- A human or an allowed role can call a referendum now. A reviewer costs its runner (the tokens of a
  model) or its command, for each round and each gate.
- The branches `referendum/<id>/v<n>` stay until a human removes them.
- Ballots that come through the message queue, and default reviewers in a role file, need new work.
