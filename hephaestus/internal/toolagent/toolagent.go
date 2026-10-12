// Package toolagent runs the workers of a role ([[agent-delegation#AI agents and tool agents]]). A
// worker gets the context of each task as a message ([[message-queue]]) and claims the issue of the
// task. A worker of a tool role runs the command of its role, and closes or stops the issue from the
// exit code: it never decides anything, and work that its command cannot do goes to a human or to an
// AI role. A worker of an AI role runs the runner command (a model, for example claude -p) with the
// prompt of the task, and the model closes or stops the issue itself.
package toolagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/airole"
	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
	"github.com/runewake2/sisyphus/hephaestus/internal/scratch"
	"github.com/runewake2/sisyphus/hephaestus/internal/taskspace"
	"github.com/runewake2/sisyphus/hephaestus/internal/watch"
)

// ownerVariable is the owner name of the worker, as sisyphus reads it.
const ownerVariable = "SISYPHUS_OWNER"

// NeedsHuman is the tag of an issue that a worker stopped, for a human to look at.
const NeedsHuman = "needs-human"

// outputLines is the number of the last lines of the output that go into an issue or a message.
const outputLines = 40

// Queue is what a tool agent needs from the message queue.
type Queue interface {
	Receive(ctx context.Context, r mq.Receiver) ([]mq.Message, error)
	Ack(worker string, ids []string) error
	Send(m mq.Message) (mq.Message, error)
	Renew(worker string, ids []string) (mq.Renewal, error)
}

// Notice reports whether a message is a notice: a broadcast or a message to every worker of a role.
// It tells many workers about an issue, so a worker does not claim, close, or reopen the issue; it
// records its result in the Notes of the issue. A message to one worker, or to one worker of a role,
// is a task: the worker claims its issue.
func Notice(m mq.Message) bool {
	return m.To.Broadcast || m.To.All
}

// Sisyphus runs the sisyphus command in a directory.
type Sisyphus func(dir string, args ...string) (string, error)

// TaskData is the part of the data of a task message that a tool agent reads.
type TaskData struct {
	// Workspace is the directory where the command runs, and where sisyphus finds the issue.
	// Default: a workspace of its own for a role with workspace: task, or the directory of the agent.
	Workspace string `json:"workspace,omitempty"`
	// Bookmark is the bookmark that the task is about. A task workspace starts on it, and a notice
	// about it runs in a scratch copy of it.
	Bookmark string `json:"bookmark,omitempty"`
	// Head is the commit of the bookmark when the notice was sent.
	Head string `json:"head,omitempty"`
	// Base is the base of a task workspace that has no bookmark. Default: the base of the agent.
	Base string `json:"base,omitempty"`
}

// Agent is one worker of a tool role.
type Agent struct {
	Role roles.Role
	// Runner is the command line that runs a model for an AI role, with the prompt of the task on its
	// standard input; {role} in it becomes the name of its subagent (Role.Flat). Empty for a tool role.
	Runner   string
	Worker   string
	Dir      string
	Queue    Queue
	Sisyphus Sisyphus
	Out, Err io.Writer
	Now      func() time.Time
	// TaskTimeout is the longest time that one task may run; then it is stopped, with every program
	// that it started. Zero: the timeout of the role, or DefaultTaskTimeout.
	TaskTimeout time.Duration
	// MaxTokens is the most tokens that one task of an AI role may use. Zero: the max-tokens of the
	// role, or no limit.
	MaxTokens int
	// Repo is the repo of the worker. With it, a task of a role with workspace: task gets a workspace
	// of its own, and a notice about a bookmark runs in a scratch copy of the bookmark.
	Repo repo.Repo
	// Workspaces is the directory of the task workspaces. Default: workspaces/ beside the main root.
	Workspaces string
	// BookmarkPrefix starts the name of the bookmark of each task, for example ai/.
	BookmarkPrefix string
	// Base is the base of a task workspace, normally main.
	Base    string
	handled map[string]bool
}

// DefaultTaskTimeout is the time limit of a task when neither the worker nor the role gives one.
const DefaultTaskTimeout = time.Hour

// timeout returns the time limit of one task.
func (a *Agent) timeout() time.Duration {
	switch {
	case a.TaskTimeout > 0:
		return a.TaskTimeout
	case a.Role.Timeout > 0:
		return a.Role.Timeout
	}
	return DefaultTaskTimeout
}

// maxTokens returns the token limit of one task of an AI role, or 0 for no limit.
func (a *Agent) maxTokens() int {
	if a.MaxTokens > 0 {
		return a.MaxTokens
	}
	return a.Role.MaxTokens
}

// Result is what a tool agent did with one message.
type Result struct {
	Message string `json:"message"`
	Issue   string `json:"issue,omitempty"`
	// Status is done (exit code 0), failed (another exit code), refused (another owner has the
	// issue), or error (the agent could not run the task).
	Status string `json:"status"`
	Exit   int    `json:"exit"`
	// Output is the end of the output of the command.
	Output string `json:"output,omitempty"`
	// Next is where the work goes after a failure: needs-human, or an AI role.
	Next string `json:"next,omitempty"`
	// Detail is the reason of a refusal or an error.
	Detail string `json:"detail,omitempty"`
	// Seconds is how long the command ran.
	Seconds float64 `json:"seconds,omitempty"`
	// Usage is what the model of an AI role used, as its runner reports it.
	Usage *Usage `json:"usage,omitempty"`
	// Workspace is the task workspace, and whether it was removed after the task.
	Workspace string `json:"workspace,omitempty"`
	Removed   bool   `json:"removed,omitempty"`
	// Fixes is the bookmark with the changes that a notice made in its scratch copy.
	Fixes string `json:"fixes,omitempty"`
}

// RunnerVariable names the environment variable of the runner of AI roles.
const RunnerVariable = "HEPHAESTUS_AI_RUNNER"

// Validate checks that a worker of the role can run: a tool role has its command, and an AI role
// needs a runner, because a model costs tokens and hephaestus never starts one by itself.
func Validate(role roles.Role, runner string) error {
	if role.Kind == roles.AI && strings.TrimSpace(runner) == "" {
		return fmt.Errorf("The role %s is an AI role: a model does its work, and a model costs tokens. Give the command that runs it with --runner or %s, for example: claude -p --agent {role}", role.Name, RunnerVariable)
	}
	return nil
}

// Run receives the messages of the worker and handles each, oldest first, until ctx ends. With once,
// it handles the messages that wait now, and returns.
func (a *Agent) Run(ctx context.Context, once bool) ([]Result, error) {
	if err := Validate(a.Role, a.Runner); err != nil {
		return nil, err
	}
	if a.handled == nil {
		a.handled = map[string]bool{}
	}
	var results []Result
	for {
		wait := 30 * time.Second
		if once {
			wait = 0
		}
		messages, err := a.Queue.Receive(ctx, mq.Receiver{Worker: a.Worker, Role: a.Role.Name, Kinds: a.Role.Listens, Wait: wait})
		if ctx.Err() != nil {
			return results, nil
		}
		if err != nil {
			return results, err
		}
		for _, m := range messages {
			if a.handled[m.ID] {
				// A copy of a message that this worker handled: acknowledge it again, and do nothing.
				if err := a.Queue.Ack(a.Worker, []string{m.ID}); err != nil {
					return results, err
				}
				continue
			}
			result := a.Handle(ctx, m)
			results = append(results, result)
			if result.Status == "error" && ctx.Err() != nil {
				return results, nil
			}
			a.handled[m.ID] = true
			if err := a.Queue.Ack(a.Worker, []string{m.ID}); err != nil {
				return results, err
			}
		}
		if once {
			return results, nil
		}
	}
}

// Handle does the task of one message, and tells its sender the result. While it works, it renews
// the lease of the message (the keep-alive), so that the queue does not give it to another worker.
func (a *Agent) Handle(ctx context.Context, m mq.Message) Result {
	stop := mq.KeepAlive(a.Queue.Renew, a.Worker, []string{m.ID}, a.Err)
	result := a.handle(ctx, m)
	stop()
	fmt.Fprintf(a.Err, "%s: message %s: %s%s\n", a.Worker, m.ID, result.Status, map[bool]string{true: " (" + result.Issue + ")", false: ""}[result.Issue != ""])
	if m.From != a.Worker {
		data, _ := json.Marshal(result)
		body := fmt.Sprintf("%s %s the task of message %s.", a.Worker, map[string]string{
			"done": "did", "failed": "could not do", "refused": "did not start", "error": "could not run",
		}[result.Status], m.ID)
		if result.Detail != "" {
			body += " " + result.Detail
		}
		if _, err := a.Queue.Send(mq.Message{ID: "result-" + m.ID, From: a.Worker, To: mq.Address{Worker: m.From}, Kind: "result", Issue: result.Issue, Body: body, Data: data}); err != nil {
			fmt.Fprintf(a.Err, "Warning: %s could not send the result of message %s to %s: %v\n", a.Worker, m.ID, m.From, err)
		}
	}
	return result
}

// It names its result, so that its deferred steps (the usage, the workspace, the fixes of a notice)
// can complete it.
func (a *Agent) handle(ctx context.Context, m mq.Message) (result Result) {
	result = Result{Message: m.ID, Issue: m.Issue}
	var data TaskData
	if len(m.Data) > 0 {
		if err := json.Unmarshal(m.Data, &data); err != nil {
			result.Status, result.Detail = "error", fmt.Sprintf("Its data is not an object: %v.", err)
			return result
		}
	}
	dir := a.Dir
	if data.Workspace != "" {
		dir = data.Workspace
	}
	// tracker is where the issue of the message is recorded: the task workspace, or else the
	// directory of the worker. A scratch copy is not: it goes away.
	tracker := dir
	var space *taskspace.Space
	var copy *scratch.Worktree
	switch {
	case Notice(m) && data.Bookmark != "" && a.Repo.VCS != repo.None:
		// A notice about a bookmark runs in a scratch copy of it, never in the workspace of another.
		w, base, err := a.scratchCopy(data)
		if err != nil {
			result.Status, result.Detail = "error", err.Error()
			return result
		}
		defer w.Remove()
		copy, dir = w, w.Dir
		defer func() { a.keepFixes(m, data, copy, base, &result) }()
	case !Notice(m) && data.Workspace == "" && a.Role.Workspace == "task" && a.Repo.VCS != repo.None:
		s, err := a.taskSpace(m, data)
		if err != nil {
			result.Status, result.Detail = "error", "Its workspace could not be made: "+err.Error()
			return result
		}
		space, dir, tracker = &s, s.Dir, s.Dir
		result.Workspace = s.Dir
		defer func() { a.closeSpace(m, data, s, &result) }()
	}
	defer func() { a.recordUsage(tracker, m, result) }() // before the workspace goes
	if m.Issue != "" && !Notice(m) {
		start := []string{"start", m.Issue, "--owner", a.Worker}
		if space != nil {
			start = append(start, "--bookmark", space.Bookmark, "--workspace", space.Name)
		}
		if _, err := a.Sisyphus(dir, start...); err != nil {
			result.Status, result.Detail = "refused", err.Error()
			if !strings.Contains(err.Error(), "in progress") {
				result.Status = "error"
			}
			return result
		}
	}
	exit, output, err := a.command(ctx, dir, m, &result, space)
	result.Exit, result.Output = exit, tail(output, outputLines)
	if Notice(m) {
		return a.noticeResult(tracker, m, result, err)
	}
	if err != nil {
		result.Status, result.Detail = "error", err.Error()
		a.stop(dir, m, &result)
		return result
	}
	if exit == 0 && a.Role.Kind == roles.AI {
		// The model closes or stops its issue itself. An issue that it left in progress needs a human.
		result.Status = "done"
		if m.Issue != "" {
			state := a.issueState(dir, m.Issue)
			switch state {
			case "closed":
			case "open":
				result.Status, result.Detail = "stopped", "The worker stopped the issue."
			default:
				result.Status = "failed"
				result.Detail = "The model ended with the issue still in progress."
				a.stop(dir, m, &result)
				result.Next = NeedsHuman
			}
		}
		return result
	}
	if exit == 0 {
		result.Status = "done"
		if m.Issue != "" {
			summary := strings.TrimSpace(result.Output)
			if summary == "" {
				summary = fmt.Sprintf("`%s` finished with exit code 0.", a.Role.Command)
			}
			resolution := fmt.Sprintf("Done by the tool agent %s (role `%s`): `%s`.\n\n```\n%s\n```", a.Worker, a.Role.Name, a.Role.Command, summary)
			if _, err := a.Sisyphus(dir, "close", m.Issue, "--message="+resolution); err != nil {
				result.Status, result.Detail = "error", "The command succeeded, but the issue did not close: "+err.Error()
			}
		}
		return result
	}
	result.Status = "failed"
	a.stop(dir, m, &result)
	return result
}

// taskSpace makes the workspace of a task: <role>-<issue>, with the bookmark <prefix><name>, on the
// bookmark of the task, or on the base.
func (a *Agent) taskSpace(m mq.Message, data TaskData) (taskspace.Space, error) {
	about := m.ID
	if m.Issue != "" {
		about = path.Base(m.Issue)
	}
	name := taskspace.Name(a.Role.Flat(), about)
	base := data.Bookmark
	if base == "" {
		base = data.Base
	}
	if base == "" {
		base = a.base()
	}
	if a.Repo.VCS == repo.JJ {
		if err := exportJJ(a.Repo); err != nil {
			return taskspace.Space{}, err
		}
	}
	return taskspace.Create(a.Repo, a.Workspaces, name, base, a.BookmarkPrefix+name)
}

func (a *Agent) base() string {
	if a.Base != "" {
		return a.Base
	}
	return "main"
}

// closeSpace ends the workspace of a task. When the task is done, it announces the close of its
// issue (the same notices that hephaestus watch sends, so the queue keeps one of each), and removes
// the workspace; its commits and bookmark stay. Otherwise it keeps the workspace for a human, and
// tells where it is.
func (a *Agent) closeSpace(m mq.Message, data TaskData, s taskspace.Space, result *Result) {
	if result.Status != "done" {
		if m.Issue != "" && result.Status != "refused" {
			note := fmt.Sprintf("%s: the workspace of the task is kept for a human: %s (bookmark %s).", a.Now().Format("2006-01-02"), s.Dir, s.Bookmark)
			a.Sisyphus(s.Dir, "edit", m.Issue, "notes", "--append", "--", note)
		}
		return
	}
	if m.Issue != "" {
		a.announce(m.Issue, s)
	}
	if taskspace.Dirty(a.Repo, s) {
		note := fmt.Sprintf("%s: the task is done, but its workspace %s has changes that no commit keeps, so it is kept for a human.", a.Now().Format("2006-01-02"), s.Dir)
		if m.Issue != "" {
			a.Sisyphus(s.Dir, "edit", m.Issue, "notes", "--append", "--", note)
		}
		fmt.Fprintln(a.Err, "Warning: "+note)
		return
	}
	if err := taskspace.Remove(a.Repo, s); err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not remove the workspace %s: %v\n", a.Worker, s.Dir, err)
		return
	}
	result.Removed = true
}

// announce broadcasts bookmark-ready and issue-closed for an issue that the task closed in its
// workspace, when the queue runs.
func (a *Agent) announce(issue string, s taskspace.Space) {
	if a.issueState(s.Dir, issue) != "closed" {
		return
	}
	head, err := a.Repo.CommitOf(s.Bookmark)
	review := err == nil && head != ""
	if review {
		if merged, err := a.Repo.IsAncestor(head, a.base()); err == nil && merged {
			review = false
		}
	} else {
		fmt.Fprintf(a.Err, "Warning: the task of %s closed, but its bookmark %s does not exist: no review can start.\n", issue, s.Bookmark)
	}
	for _, e := range watch.Notices(issue, s.Bookmark, head, a.base(), review, true) {
		if _, err := a.Queue.Send(e.Message(a.Worker)); err != nil {
			fmt.Fprintf(a.Err, "Warning: %s could not broadcast %s for %s: %v\n", a.Worker, e.Kind, issue, err)
		}
	}
}

// scratchCopy makes a scratch copy of the bookmark of a notice, at the head that the notice names.
func (a *Agent) scratchCopy(data TaskData) (*scratch.Worktree, string, error) {
	if a.Repo.VCS == repo.JJ {
		if err := exportJJ(a.Repo); err != nil {
			return nil, "", err
		}
	}
	gitDir, err := a.Repo.GitDir()
	if err != nil {
		return nil, "", err
	}
	commit := data.Head
	if commit == "" {
		if commit, err = a.Repo.CommitOf(data.Bookmark); err != nil {
			return nil, "", fmt.Errorf("the bookmark %s of the notice does not exist", data.Bookmark)
		}
	}
	w, err := scratch.New(gitDir, commit, "hephaestus-notice-")
	return w, commit, err
}

// keepFixes keeps the changes that a notice made in its scratch copy: it commits them on the bookmark
// <prefix>fix-<role>-<issue>, and records it in the Notes of the issue, for the author to squash.
func (a *Agent) keepFixes(m mq.Message, data TaskData, w *scratch.Worktree, base string, result *Result) {
	head, _, err := w.Commit(base, fmt.Sprintf("%s: changes of the notice %s on %s", a.Role.Name, m.ID, data.Bookmark))
	if err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not keep the changes of its notice: %v\n", a.Worker, err)
		return
	}
	if head == "" {
		return
	}
	about := m.ID
	if m.Issue != "" {
		about = path.Base(m.Issue)
	}
	// A flat name: a git branch cannot be a directory of another one, for example ai/fix/... beside ai/fix.
	fixes := a.BookmarkPrefix + taskspace.Name("fix", a.Role.Flat(), about)
	if _, err := w.Git("branch", "--force", fixes, head); err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not keep the changes of its notice: %v\n", a.Worker, err)
		return
	}
	if a.Repo.VCS == repo.JJ {
		importJJ(a.Repo)
	}
	result.Fixes = fixes
	if m.Issue != "" {
		note := fmt.Sprintf("%s: %s (role %s) made changes to %s in the bookmark %s, on top of it. The author squashes them into the change that they belong to.", a.Now().Format("2006-01-02"), a.Worker, a.Role.Name, data.Bookmark, fixes)
		a.Sisyphus(a.Dir, "edit", m.Issue, "notes", "--append", "--", note)
	}
}

func exportJJ(r repo.Repo) error {
	cmd := exec.Command("jj", "git", "export")
	cmd.Dir = r.Root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("jj git export: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func importJJ(r repo.Repo) {
	cmd := exec.Command("jj", "git", "import")
	cmd.Dir = r.Root
	cmd.Run()
}

// recordUsage adds a line about what the model of an AI role used to the Notes of the issue of the
// message: the record of the run, for the economists and the operations auditor.
func (a *Agent) recordUsage(dir string, m mq.Message, result Result) {
	if result.Usage == nil || !result.Usage.Reported || result.Issue == "" {
		return
	}
	u := result.Usage
	note := fmt.Sprintf("%s: the run of %s (role %s) for message %s used %d tokens", a.Now().Format("2006-01-02"), a.Worker, a.Role.Name, m.ID, u.Tokens)
	if u.CostUSD > 0 {
		note += fmt.Sprintf(", $%.2f", u.CostUSD)
	}
	if u.Turns > 0 {
		note += fmt.Sprintf(", %d turns", u.Turns)
	}
	note += fmt.Sprintf(", in %s: %s.", time.Duration(result.Seconds*float64(time.Second)).Round(time.Second), result.Status)
	if _, err := a.Sisyphus(dir, "edit", result.Issue, "notes", "--append", "--", note); err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not record the usage of its run in %s: %v\n", a.Worker, result.Issue, err)
	}
}

// noticeResult records the result of a notice in the Notes of its issue, and changes no state: the
// issue belongs to another worker, or is closed. A failure tags the issue needs-human.
func (a *Agent) noticeResult(dir string, m mq.Message, result Result, runErr error) Result {
	switch {
	case runErr != nil:
		result.Status, result.Detail, result.Next = "error", runErr.Error(), NeedsHuman
	case result.Exit != 0:
		result.Status, result.Next = "failed", NeedsHuman
	default:
		result.Status = "done"
	}
	if m.Issue == "" {
		return result
	}
	verb := map[string]string{"done": "did", "failed": "failed at", "error": "could not run"}[result.Status]
	note := fmt.Sprintf("%s: %s (role %s) %s the %s notice of message %s: `%s`, exit code %d.", a.Now().Format("2006-01-02"), a.Worker, a.Role.Name, verb, m.Kind, m.ID, a.commandLine(), result.Exit)
	if output := strings.TrimSpace(result.Output); output != "" && a.Role.Kind == roles.Tool {
		note += "\n\n```\n" + output + "\n```"
	}
	if result.Detail != "" {
		note += "\n\n" + result.Detail
	}
	if _, err := a.Sisyphus(dir, "edit", m.Issue, "notes", "--append", "--", note); err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not record its result in %s: %v\n", a.Worker, m.Issue, err)
	}
	if result.Next == NeedsHuman {
		if _, err := a.Sisyphus(dir, "update", m.Issue, "--add-tags", NeedsHuman); err != nil {
			fmt.Fprintf(a.Err, "Warning: %s could not tag %s for a human: %v\n", a.Worker, m.Issue, err)
		}
	}
	return result
}

// stop records a failure in the issue of the task, and gives the work to the next one: a human, or
// the AI role of on-failure, with a message.
func (a *Agent) stop(dir string, m mq.Message, result *Result) {
	result.Next = a.Role.OnFailure
	if a.Role.Kind == roles.AI {
		result.Next = NeedsHuman // a model that failed hands its work to a human, not to another model
	}
	if m.Issue == "" {
		return
	}
	note := fmt.Sprintf("%s: the tool agent %s ran `%s`: exit code %d.\n\n```\n%s\n```", a.Now().Format("2006-01-02"), a.Worker, a.commandLine(), result.Exit, strings.TrimSpace(result.Output))
	if a.Role.Kind == roles.AI {
		note = strings.Replace(note, "the tool agent", "the AI worker", 1)
	}
	if result.Detail != "" {
		note += "\n\n" + result.Detail
	}
	if _, err := a.Sisyphus(dir, "edit", m.Issue, "notes", "--append", "--", note); err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not record the failure in %s: %v\n", a.Worker, m.Issue, err)
	}
	args := []string{"update", m.Issue, "open"}
	if result.Next == NeedsHuman {
		args = append(args, "--add-tags", NeedsHuman)
	} else {
		args = append(args, "--metadata", "delegate="+a.Role.OnFailure)
	}
	if _, err := a.Sisyphus(dir, args...); err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not stop %s: %v\n", a.Worker, m.Issue, err)
		return
	}
	if result.Next == NeedsHuman {
		return
	}
	data, _ := json.Marshal(map[string]any{"workspace": dir, "failed": result})
	body := fmt.Sprintf("The tool agent %s could not do this task: `%s` ended with exit code %d. Its output is in the Notes of the issue.", a.Worker, a.Role.Command, result.Exit)
	if _, err := a.Queue.Send(mq.Message{ID: "next-" + m.ID, From: a.Worker, To: mq.Address{Role: a.Role.OnFailure}, Kind: "task", Issue: m.Issue, Body: body, Data: data}); err != nil {
		fmt.Fprintf(a.Err, "Warning: %s could not send the task to %s: %v\n", a.Worker, a.Role.OnFailure, err)
	}
}

// command runs the command of the role in dir, with the message as JSON on its standard input. It
// returns the exit code and the output (stdout and stderr together). An error is a command that
// could not start.
// The command stops at the time limit of the task, or when an AI role passes its token limit; that
// is an error, and result gets the time and the usage.
func (a *Agent) command(ctx context.Context, dir string, m mq.Message, result *Result, space *taskspace.Space) (int, string, error) {
	limit := a.timeout()
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	overBudget := false
	var meter *Meter
	if a.Role.Kind == roles.AI {
		meter = NewMeter(a.maxTokens(), func() { overBudget = true; cancel() })
	}
	started := a.Now()
	defer func() { result.Seconds = a.Now().Sub(started).Seconds() }()
	input, err := json.Marshal(m)
	if err != nil {
		return -1, "", err
	}
	input = append(input, '\n')
	if a.Role.Kind == roles.AI {
		bookmark := ""
		if space != nil {
			bookmark = space.Bookmark
		}
		input = []byte(airole.TaskPrompt(a.Role, a.Worker, m, dir, bookmark))
	}
	cmd := Shell(ctx, a.commandLine())
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(input)
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(&output, a.Out)
	if meter != nil {
		cmd.Stdout = io.MultiWriter(&output, a.Out, meter)
	}
	cmd.Stderr = io.MultiWriter(&output, a.Err)
	cmd.Env = append(os.Environ(),
		ownerVariable+"="+a.Worker,
		"HEPHAESTUS_ROLE="+a.Role.Name,
		"HEPHAESTUS_ROLE_DIR="+a.Role.Dir(),
		"HEPHAESTUS_MESSAGE_ID="+m.ID,
		"HEPHAESTUS_MESSAGE_KIND="+m.Kind,
		"HEPHAESTUS_MESSAGE_FROM="+m.From,
		"HEPHAESTUS_MESSAGE_ISSUE="+m.Issue,
	)
	if space != nil {
		cmd.Env = append(cmd.Env, "HEPHAESTUS_WORKSPACE="+space.Dir, "HEPHAESTUS_BOOKMARK="+space.Bookmark)
	}
	err = cmd.Run()
	text := output.String()
	if meter != nil {
		usage := meter.Usage()
		result.Usage = &usage
		text = meter.Text(text)
		if a.maxTokens() > 0 && !usage.Reported {
			fmt.Fprintf(a.Err, "Warning: the runner of %s reported no usage, so its limit of %d tokens could not be kept. Give a runner that prints JSON lines with usage, for example claude -p --output-format stream-json --verbose.\n", a.Role.Name, a.maxTokens())
		}
	}
	var exit *exec.ExitError
	switch {
	case overBudget:
		return -1, text, fmt.Errorf("the task passed its limit of %d tokens, so it was stopped", a.maxTokens())
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return -1, text, fmt.Errorf("the task ran longer than its limit of %s, so it was stopped", limit)
	case err == nil:
		return 0, text, nil
	case errors.As(err, &exit):
		return exit.ExitCode(), text, nil
	}
	return -1, text, fmt.Errorf("`%s` could not start: %v", a.commandLine(), err)
}

// commandLine is the command of a tool role, or the runner of an AI role with its role.
func (a *Agent) commandLine() string {
	if a.Role.Kind == roles.AI {
		return strings.ReplaceAll(a.Runner, "{role}", a.Role.Flat())
	}
	return a.Role.Command
}

// issueState returns the state of an issue, or "" when sisyphus cannot read it.
func (a *Agent) issueState(dir, issue string) string {
	out, err := a.Sisyphus(dir, "show", "--json", "--", issue)
	if err != nil {
		return ""
	}
	var shown struct{ State string }
	json.Unmarshal([]byte(out), &shown)
	return shown.State
}

// Shell returns the command that runs a command line in the shell of the platform. When ctx ends,
// it stops the whole command, also the programs that the shell started, for example a model runner,
// so that a stopped task spends no more tokens or CPU.
func Shell(ctx context.Context, line string) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", line)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", line)
	}
	stopTree(cmd)
	cmd.WaitDelay = 5 * time.Second
	return cmd
}

func tail(text string, n int) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) > n {
		lines = append([]string{fmt.Sprintf("(%d earlier lines are not shown)", len(lines)-n)}, lines[len(lines)-n:]...)
	}
	return strings.Join(lines, "\n")
}

// Ballot is the vote of a tool agent in a referendum: its exit code, and the changes that it made.
type Ballot struct {
	Role string `json:"role"`
	// Vote is yes (exit code 0, no change), modify (exit code 0, and the command changed files, as a
	// formatter does), or no (another exit code).
	Vote string `json:"vote"`
	Exit int    `json:"exit"`
	// Reasons is the end of the output of the command.
	Reasons string `json:"reasons,omitempty"`
	// Files and Diff are the modification of a modify ballot.
	Files []string `json:"files,omitempty"`
	Diff  string   `json:"diff,omitempty"`
}

// Review runs the command of a tool role in dir, as a reviewer of a referendum
// ([[referendum]]), and returns its ballot. A tool reviewer never abstains. base is the current
// implementation, which the command gets in HEPHAESTUS_BASE, so that it can work only on the files
// that the change touches; empty: main.
func Review(ctx context.Context, role roles.Role, dir, base string, out io.Writer) (Ballot, error) {
	if role.Kind != roles.Tool {
		return Ballot{}, fmt.Errorf("The role %s is an AI role: its ballot comes from a referendum, not from an exit code.", role.Name)
	}
	r, err := repo.Find(dir)
	if err != nil {
		return Ballot{}, err
	}
	before, err := r.WorkingDiff()
	if err != nil {
		return Ballot{}, err
	}
	cmd := Shell(ctx, role.Command)
	cmd.Dir = dir
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(&output, out)
	cmd.Stderr = io.MultiWriter(&output, out)
	cmd.Env = append(os.Environ(), "HEPHAESTUS_ROLE="+role.Name, "HEPHAESTUS_ROLE_DIR="+role.Dir(), "HEPHAESTUS_REVIEW=1")
	if base != "" {
		cmd.Env = append(cmd.Env, "HEPHAESTUS_BASE="+base)
	}
	ballot := Ballot{Role: role.Name, Reasons: tail(output.String(), outputLines)}
	runErr := cmd.Run()
	ballot.Reasons = tail(output.String(), outputLines)
	var exit *exec.ExitError
	switch {
	case errors.As(runErr, &exit):
		ballot.Vote, ballot.Exit = "no", exit.ExitCode()
		return ballot, nil
	case runErr != nil:
		return Ballot{}, fmt.Errorf("`%s` could not start: %v", role.Command, runErr)
	}
	after, err := r.WorkingDiff()
	if err != nil {
		return Ballot{}, err
	}
	if after == before {
		ballot.Vote = "yes"
		return ballot, nil
	}
	ballot.Vote = "modify"
	ballot.Diff = after
	beforeFiles := repo.DiffFiles(before)
	for _, file := range repo.DiffFiles(after) {
		if !contains(beforeFiles, file) || fileDiff(before, file) != fileDiff(after, file) {
			ballot.Files = append(ballot.Files, file)
		}
	}
	return ballot, nil
}

func contains(items []string, item string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}

// fileDiff returns the part of a git diff for one file.
func fileDiff(diff, file string) string {
	start := strings.Index(diff, "diff --git a/"+file+" b/")
	if start < 0 {
		return ""
	}
	rest := diff[start+1:]
	if end := strings.Index(rest, "\ndiff --git "); end >= 0 {
		return diff[start : start+1+end]
	}
	return diff[start:]
}
