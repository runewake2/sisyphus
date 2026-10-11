# 0023: Lock and write issues atomically

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.88]]

## Context

Agents often run sisyphus in one checkout at the same time. A command wrote an issue with
`os.WriteFile`, which truncates the file and then writes it, so a command that read at the same
time could see an empty or short file. Two commands that changed one issue each read it, changed
it, and wrote it, so the second write removed the change of the first. Two starts of one open
issue both succeeded, and [[0022-refuse-to-take-a-claimed-issue]] could not stop it
([[atomic-issue-writes]]).

## Decision

- Every issue write goes to a temporary file in the same directory, which is then renamed over the
  issue file. A rename is atomic on the file systems that sisyphus supports.
- Every command that changes issues holds one lock for the repo from its first read to its last
  write. The lock is a file that the command creates with `O_EXCL`.
- The lock file is outside the working copy, in a `sisyphus` directory of the shared VCS directory:
  the jj repo directory, or the git common directory. So no VCS records it, and every jj workspace
  and git worktree of the repo uses the same lock. Without jj or git, the lock is `issues/.lock`.
- A command waits up to 10 seconds for the lock. A lock older than 30 seconds is removed: no
  command holds the lock that long, so the command that took it has stopped.
- Commands that only read take no lock: the atomic rename is sufficient for them.

## Alternatives considered

- **`flock` or `LockFileEx`.** The operating system releases the lock when a process stops, but the
  calls are different on each platform, and a file lock on a network file system is not reliable.
- **One lock for each issue.** A command such as `depends-on` reads and changes more than one
  issue, and a change is short, so one lock costs nothing.
- **A lock file in `issues/`.** jj records every file of the working copy when it snapshots it, so a
  lock there could get into a commit. It is used only when there is no jj or git.

## Consequences

- Two changes of one issue at the same time never lose a change, and a reader never sees a part of
  a file.
- [[claims-across-workspaces]] can keep its records in the same shared directory, under the same
  lock.
- A command that stops while it holds the lock makes the others wait up to 30 seconds once.
