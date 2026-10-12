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
	"runtime"
	"strings"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/airole"
	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
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
}

// Sisyphus runs the sisyphus command in a directory.
type Sisyphus func(dir string, args ...string) (string, error)

// TaskData is the part of the data of a task message that a tool agent reads.
type TaskData struct {
	// Workspace is the directory where the command runs, and where sisyphus finds the issue.
	// Default: the directory of the agent.
	Workspace string `json:"workspace,omitempty"`
}

// Agent is one worker of a tool role.
type Agent struct {
	Role roles.Role
	// Runner is the command line that runs a model for an AI role, with the prompt of the task on its
	// standard input; {role} in it becomes the name of the role. Empty for a tool role.
	Runner   string
	Worker   string
	Dir      string
	Queue    Queue
	Sisyphus Sisyphus
	Out, Err io.Writer
	Now      func() time.Time
	handled  map[string]bool
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

// Handle does the task of one message, and tells its sender the result.
func (a *Agent) Handle(ctx context.Context, m mq.Message) Result {
	result := a.handle(ctx, m)
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

func (a *Agent) handle(ctx context.Context, m mq.Message) Result {
	result := Result{Message: m.ID, Issue: m.Issue}
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
	if m.Issue != "" {
		if _, err := a.Sisyphus(dir, "start", m.Issue, "--owner", a.Worker); err != nil {
			result.Status, result.Detail = "refused", err.Error()
			if !strings.Contains(err.Error(), "in progress") {
				result.Status = "error"
			}
			return result
		}
	}
	exit, output, err := a.command(ctx, dir, m)
	result.Exit, result.Output = exit, tail(output, outputLines)
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
func (a *Agent) command(ctx context.Context, dir string, m mq.Message) (int, string, error) {
	input, err := json.Marshal(m)
	if err != nil {
		return -1, "", err
	}
	input = append(input, '\n')
	if a.Role.Kind == roles.AI {
		input = []byte(airole.TaskPrompt(a.Role, a.Worker, m, dir))
	}
	cmd := Shell(ctx, a.commandLine())
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(input)
	var output bytes.Buffer
	cmd.Stdout = io.MultiWriter(&output, a.Out)
	cmd.Stderr = io.MultiWriter(&output, a.Err)
	cmd.Env = append(os.Environ(),
		ownerVariable+"="+a.Worker,
		"HEPHAESTUS_ROLE="+a.Role.Name,
		"HEPHAESTUS_MESSAGE_ID="+m.ID,
		"HEPHAESTUS_MESSAGE_KIND="+m.Kind,
		"HEPHAESTUS_MESSAGE_FROM="+m.From,
		"HEPHAESTUS_MESSAGE_ISSUE="+m.Issue,
	)
	err = cmd.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
		return 0, output.String(), nil
	case errors.As(err, &exit):
		return exit.ExitCode(), output.String(), nil
	}
	return -1, output.String(), fmt.Errorf("`%s` could not start: %v", a.commandLine(), err)
}

// commandLine is the command of a tool role, or the runner of an AI role with its role.
func (a *Agent) commandLine() string {
	if a.Role.Kind == roles.AI {
		return strings.ReplaceAll(a.Runner, "{role}", a.Role.Name)
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

// Shell returns the command that runs a command line in the shell of the platform.
func Shell(ctx context.Context, line string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", line)
	}
	return exec.CommandContext(ctx, "sh", "-c", line)
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
// ([[referendum]]), and returns its ballot. A tool reviewer never abstains.
func Review(ctx context.Context, role roles.Role, dir string, out io.Writer) (Ballot, error) {
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
	cmd.Env = append(os.Environ(), "HEPHAESTUS_ROLE="+role.Name, "HEPHAESTUS_REVIEW=1")
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
