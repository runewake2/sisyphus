package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/internal/hephaestus/mq"
	"github.com/runewake2/sisyphus/internal/hephaestus/repo"
	"github.com/runewake2/sisyphus/internal/hephaestus/roles"
	"github.com/runewake2/sisyphus/internal/hephaestus/scratch"
	"github.com/runewake2/sisyphus/internal/hephaestus/sisyphus"
	"github.com/runewake2/sisyphus/internal/hephaestus/toolagent"
)

// trialsFile is the record of the trials, one JSON line each, in the shared hephaestus directory.
const trialsFile = "trials.jsonl"

// trialSender is the sender of the message of a trial.
const trialSender = "trial"

// Trial is the record of one trial: what the role did with one issue in a scratch copy.
type Trial struct {
	Role    string    `json:"role"`
	Issue   string    `json:"issue"`
	Kind    string    `json:"kind"`
	Onto    string    `json:"onto"`
	Commit  string    `json:"commit"`
	Started time.Time `json:"started"`
	// Before and After are the states of the issue in the copy.
	Before string           `json:"before"`
	After  string           `json:"after"`
	Result toolagent.Result `json:"result"`
	// Changed are the files that the task changed in the copy, with the status of git status --short.
	Changed []string `json:"changed"`
	// Messages are the messages that the worker sent: in a trial, nobody gets them.
	Messages []mq.Message `json:"messages"`
	// Kept is the copy, when --keep keeps it for a human.
	Kept string `json:"kept,omitempty"`
}

// trialQueue is the message queue of a trial: it keeps what the worker sends, and delivers nothing.
type trialQueue struct {
	mu   sync.Mutex
	sent []mq.Message
}

func (q *trialQueue) Receive(context.Context, mq.Receiver) ([]mq.Message, error) { return nil, nil }
func (q *trialQueue) Ack(string, []string) error                                 { return nil }
func (q *trialQueue) Renew(_ string, ids []string) (mq.Renewal, error) {
	return mq.Renewal{Renewed: ids}, nil
}
func (q *trialQueue) Send(m mq.Message) (mq.Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.sent = append(q.sent, m)
	return m, nil
}

func trialCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var onto, runner, notice, bookmark string
	var taskTimeout time.Duration
	var maxTokens int
	var keep, asJSON bool
	cmd := &cobra.Command{
		Use:   "trial <role> <issue>",
		Short: "Try a role on one issue in a scratch copy, and record what it did and what it cost.",
		Long: "Try a role on one issue, as a dry run: one worker of the role does the task of the issue in a scratch " +
			"git worktree of --onto (default main), with the same prompt, limits, and steps as hephaestus agent run. " +
			"It claims, closes, or stops the issue only in the copy. Nothing reaches the repo: no bookmark, no workspace, " +
			"and no message; the messages that the worker would send are in the record. The copy goes away after the " +
			"trial, unless --keep keeps it for a human.\n\n" +
			"With --notice <kind>, the worker gets the issue as a broadcast of that kind instead of a task, as a reviewer " +
			"gets bookmark-ready; --bookmark gives the bookmark of the notice, and the copy starts on it.\n\n" +
			"Each trial is added to " + trialsFile + " in the shared hephaestus directory of the repo, with its result, " +
			"time, tokens, and cost, to compare the roles. hephaestus agent trials prints them. An AI role needs " +
			"--runner, as hephaestus agent run does: a trial of an AI role costs tokens.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := w.roles()
			if err != nil {
				return err
			}
			role, err := roles.Find(dir, args[0])
			if err != nil {
				return err
			}
			if runner == "" {
				runner = os.Getenv(toolagent.RunnerVariable)
			}
			if err := toolagent.Validate(role, runner); err != nil {
				return err
			}
			r, err := findRepo()
			if err != nil {
				return err
			}
			if r.VCS == repo.None {
				return errors.New("A trial needs a jj or git repo: it runs in a scratch worktree.")
			}
			start := onto
			if bookmark != "" && !cmd.Flags().Changed("onto") {
				start = bookmark
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			trial, err := runTrial(ctx, r, role, args[1], trialOptions{
				Onto: start, Runner: runner, Notice: notice, Bookmark: bookmark, TaskTimeout: taskTimeout,
				MaxTokens: maxTokens, Keep: keep, Err: cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}
			if err := recordTrial(r, trial); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: the trial is not recorded: %v\n", err)
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), trial)
			}
			printTrial(cmd.OutOrStdout(), trial)
			return nil
		},
	}
	cmd.Flags().StringVar(&onto, "onto", "main", "The commit or bookmark that the scratch copy starts on.")
	cmd.Flags().StringVar(&runner, "runner", "", "For an AI role: the command that runs the model, as for hephaestus agent run. Default: "+toolagent.RunnerVariable+".")
	cmd.Flags().StringVar(&notice, "notice", "", "Give the issue as a broadcast of this kind, for example bookmark-ready, instead of as a task.")
	cmd.Flags().StringVar(&bookmark, "bookmark", "", "The bookmark of a notice. The copy starts on it, unless --onto is given.")
	cmd.Flags().DurationVar(&taskTimeout, "task-timeout", 0, "The longest time that the task may run. Default: timeout of the role file, or "+toolagent.DefaultTaskTimeout.String()+".")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "For an AI role: the most tokens that the task may use. Default: max-tokens of the role file, or no limit.")
	cmd.Flags().BoolVar(&keep, "keep", false, "Keep the scratch copy for a human, and print where it is.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the record of the trial as JSON.")
	return cmd
}

type trialOptions struct {
	Onto, Runner, Notice, Bookmark string
	TaskTimeout                    time.Duration
	MaxTokens                      int
	Keep                           bool
	Err                            io.Writer
}

// runTrial runs one task of role on issue in a scratch copy of o.Onto, and returns its record.
func runTrial(ctx context.Context, r repo.Repo, role roles.Role, issue string, o trialOptions) (Trial, error) {
	trial := Trial{Role: role.Name, Issue: issue, Kind: "task", Onto: o.Onto, Started: time.Now(), Changed: []string{}, Messages: []mq.Message{}}
	if o.Notice != "" {
		trial.Kind = o.Notice
	}
	if r.VCS == repo.JJ {
		if err := exportJJ(r); err != nil {
			return trial, err
		}
	}
	commit, err := r.CommitOf(o.Onto)
	if err != nil || commit == "" {
		return trial, fmt.Errorf("The commit %q does not exist.", o.Onto)
	}
	trial.Commit = commit
	gitDir, err := r.GitDir()
	if err != nil {
		return trial, err
	}
	copy, err := scratch.New(gitDir, commit, "hephaestus-trial-")
	if err != nil {
		return trial, err
	}
	if o.Keep {
		trial.Kept = copy.Dir
	} else {
		defer copy.Remove()
	}
	shown, err := sisyphus.Show(copy.Dir, issue)
	if err != nil {
		return trial, fmt.Errorf("The issue %s is not in %s (%.12s): %v", issue, o.Onto, commit, err)
	}
	trial.Issue, trial.Before = shown.Name, shown.State

	worker := role.Flat() + "-trial"
	queue := &trialQueue{}
	agent := &toolagent.Agent{
		Role: role, Runner: o.Runner, Worker: worker, Dir: copy.Dir, Queue: queue, Sisyphus: sisyphus.Run,
		Out: o.Err, Err: o.Err, Now: time.Now, TaskTimeout: o.TaskTimeout, MaxTokens: o.MaxTokens,
	}
	m := mq.Message{
		ID: fmt.Sprintf("trial-%s-%s", role.Flat(), trial.Started.Format("20060102T150405")), From: trialSender,
		To: mq.Address{Worker: worker}, Kind: "task", Issue: trial.Issue, Time: trial.Started,
		Body: fmt.Sprintf("Do the task of the issue %s, as the role %s. This is a trial.", trial.Issue, role.Name),
	}
	if o.Notice != "" {
		m.To, m.Kind = mq.Address{Broadcast: true}, o.Notice
		m.Body = fmt.Sprintf("%s: %s. This is a trial.", o.Notice, trial.Issue)
		if o.Bookmark != "" {
			m.Data, _ = json.Marshal(toolagent.TaskData{Bookmark: o.Bookmark, Head: commit})
		}
	}
	trial.Result = agent.Handle(ctx, m)
	for _, sent := range queue.sent {
		if sent.To.Worker != trialSender {
			trial.Messages = append(trial.Messages, sent)
		}
	}
	if after, err := sisyphus.Show(copy.Dir, trial.Issue); err == nil {
		trial.After = after.State
	}
	status, err := copy.Git("status", "--short", "--untracked-files=all")
	if err != nil {
		return trial, err
	}
	for _, line := range strings.Split(status, "\n") {
		if line = strings.TrimRight(line, " "); line != "" {
			trial.Changed = append(trial.Changed, line)
		}
	}
	return trial, nil
}

// recordTrial adds the trial to the record of the trials of the repo.
func recordTrial(r repo.Repo, trial Trial) error {
	dir := r.SharedDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(dir, trialsFile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	line, err := json.Marshal(trial)
	if err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(append(line, '\n')); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// readTrials returns the trials of the repo, oldest first.
func readTrials(r repo.Repo) ([]Trial, error) {
	file, err := os.Open(filepath.Join(r.SharedDir(), trialsFile))
	if errors.Is(err, os.ErrNotExist) {
		return []Trial{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	trials := []Trial{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var trial Trial
		if err := json.Unmarshal(scanner.Bytes(), &trial); err != nil {
			continue // a line that a crash cut short
		}
		trials = append(trials, trial)
	}
	return trials, scanner.Err()
}

func printTrial(out io.Writer, t Trial) {
	fmt.Fprintf(out, "%s on %s (%s, in a copy of %s at %.12s): %s", t.Role, t.Issue, t.Kind, t.Onto, t.Commit, t.Result.Status)
	if t.Result.Exit != 0 {
		fmt.Fprintf(out, ", exit code %d", t.Result.Exit)
	}
	fmt.Fprintf(out, ", in %s\n", seconds(t.Result.Seconds))
	if t.Result.Usage != nil && t.Result.Usage.Reported {
		fmt.Fprintf(out, "usage: %d tokens", t.Result.Usage.Tokens)
		if t.Result.Usage.CostUSD > 0 {
			fmt.Fprintf(out, ", $%.2f", t.Result.Usage.CostUSD)
		}
		if t.Result.Usage.Turns > 0 {
			fmt.Fprintf(out, ", %d turns", t.Result.Usage.Turns)
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintf(out, "issue: %s -> %s\n", t.Before, t.After)
	if t.Result.Detail != "" {
		fmt.Fprintf(out, "detail: %s\n", t.Result.Detail)
	}
	if t.Result.Next != "" {
		fmt.Fprintf(out, "next: %s\n", t.Result.Next)
	}
	for _, line := range t.Changed {
		fmt.Fprintf(out, "changed: %s\n", line)
	}
	for _, m := range t.Messages {
		to := m.To.Worker
		switch {
		case m.To.Broadcast:
			to = "every listener"
		case m.To.Role != "":
			to = "role " + m.To.Role
		}
		fmt.Fprintf(out, "would send: %s to %s: %s\n", m.Kind, to, m.Body)
	}
	if t.Kept != "" {
		fmt.Fprintf(out, "kept: %s\n", t.Kept)
	}
	if output := strings.TrimSpace(t.Result.Output); output != "" {
		fmt.Fprintf(out, "\n%s\n", output)
	}
}

func seconds(s float64) string {
	return time.Duration(s * float64(time.Second)).Round(time.Second).String()
}

func trialsCommand(findRepo func() (repo.Repo, error)) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "trials [<role>...]",
		Short: "List the trials of the roles, oldest first, to compare what each did and what it cost.",
		Long: "List the trials that hephaestus agent trial recorded, oldest first: the role, the issue, the result, the " +
			"state of the issue before and after, the time, the tokens, and the cost. With roles, only the trials of " +
			"those roles, or of the roles of those families.",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			all, err := readTrials(r)
			if err != nil {
				return err
			}
			trials := []Trial{}
			for _, t := range all {
				if len(args) == 0 || matchesRole(t.Role, args) {
					trials = append(trials, t)
				}
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), trials)
			}
			if len(trials) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No trials.")
				return nil
			}
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(table, "started\trole\tissue\tkind\tresult\tissue state\ttime\ttokens\tcost\tchanged")
			for _, t := range trials {
				tokens, cost := "-", "-"
				if u := t.Result.Usage; u != nil && u.Reported {
					tokens = fmt.Sprint(u.Tokens)
					if u.CostUSD > 0 {
						cost = fmt.Sprintf("$%.2f", u.CostUSD)
					}
				}
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s -> %s\t%s\t%s\t%s\t%d\n", t.Started.Format("2006-01-02 15:04"), t.Role, t.Issue, t.Kind,
					t.Result.Status, t.Before, t.After, seconds(t.Result.Seconds), tokens, cost, len(t.Changed))
			}
			return table.Flush()
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the trials as JSON.")
	return cmd
}

// matchesRole reports whether role is one of names, or a role of one of their families.
func matchesRole(role string, names []string) bool {
	for _, name := range names {
		if role == name || strings.HasPrefix(role, name+"/") {
			return true
		}
	}
	return false
}
