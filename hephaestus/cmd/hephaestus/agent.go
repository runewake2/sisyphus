package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/airole"
	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
	"github.com/runewake2/sisyphus/hephaestus/internal/toolagent"
)

func agentCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Run the workers of a role, and make Claude Code subagents of the AI roles.",
	}
	var as, runner string
	var once, asJSON bool
	run := &cobra.Command{
		Use:   "run <role>",
		Short: "Run a worker of a role: do the task of each message that it gets.",
		Long: "Run a worker of a tool role. It gets the messages to it, to its role, and the broadcasts of the kinds " +
			"in listens of its role file. For each message with an issue, it claims the issue (sisyphus start, as " +
			"the worker), runs the command of the role in the workspace of the task (the data field workspace, or " +
			"-C), with the message as JSON on its standard input. Exit code 0 closes the issue with the output as its " +
			"Resolution. Another exit code records the output in Notes and stops the issue: tagged needs-human, or " +
			"delegated to the AI role of on-failure, with a message to that role. The worker sends the result to the " +
			"sender of the message, and acknowledges the message. A tool role makes no model call.\n\n" +
			"For an AI role, the worker runs --runner (or " + toolagent.RunnerVariable + ") instead of a command, with the " +
			"prompt of the task on its standard input: the role file, the rules of every role, and the message. {role} " +
			"in the runner becomes the name of the role, for example: claude -p --agent {role}. The model closes or stops " +
			"the issue itself; an issue that it leaves in progress is stopped for a human. hephaestus never starts a " +
			"model without a runner. --once handles the messages that wait now, and stops.",
		Args: cobra.ExactArgs(1),
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
			name := as
			if name == "" {
				name = role.Flat() + "-1"
			}
			r, err := findRepo()
			if err != nil {
				return err
			}
			client, err := mq.Connect(r.SharedDir())
			if err != nil {
				return err
			}
			agent := &toolagent.Agent{
				Role: role, Runner: runner, Worker: name, Dir: w.directory(), Queue: client, Sisyphus: sisyphus.Run,
				Out: cmd.ErrOrStderr(), Err: cmd.ErrOrStderr(), Now: time.Now,
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			results, err := agent.Run(ctx, once)
			if asJSON {
				if results == nil {
					results = []toolagent.Result{}
				}
				if jsonErr := writeJSON(cmd.OutOrStdout(), results); jsonErr != nil {
					return jsonErr
				}
			}
			return err
		},
	}
	run.Flags().StringVar(&as, "as", "", "The owner name of the worker. Default: <role>-1.")
	run.Flags().BoolVar(&once, "once", false, "Handle the messages that wait now, and stop.")
	run.Flags().StringVar(&runner, "runner", "", "For an AI role: the command that runs the model, with the prompt on its standard input, for example \"claude -p --agent {role}\". Default: "+toolagent.RunnerVariable+".")
	run.Flags().BoolVar(&asJSON, "json", false, "Print the result of each message as JSON when it stops.")
	var reviewJSON bool
	var reviewBase string
	review := &cobra.Command{
		Use:   "review <role>",
		Short: "Run a tool role as a reviewer of a referendum, and print its ballot.",
		Long: "Run the command of a tool role in the working copy (-C), as a reviewer of a referendum, and print its " +
			"ballot: no for an exit code other than 0, modify when the command changed files (as a formatter does), " +
			"with the files and the diff, and yes otherwise. A tool reviewer never abstains.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := w.roles()
			if err != nil {
				return err
			}
			role, err := roles.Find(dir, args[0])
			if err != nil {
				return err
			}
			ballot, err := toolagent.Review(cmd.Context(), role, w.directory(), reviewBase, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if reviewJSON {
				return writeJSON(cmd.OutOrStdout(), ballot)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s (exit code %d)\n", role.Name, ballot.Vote, ballot.Exit)
			for _, file := range ballot.Files {
				fmt.Fprintf(cmd.OutOrStdout(), "  changed %s\n", file)
			}
			return nil
		},
	}
	review.Flags().BoolVar(&reviewJSON, "json", false, "Print the ballot as JSON, with the diff of a modification.")
	review.Flags().StringVar(&reviewBase, "base", "", "The current implementation, which the command gets in HEPHAESTUS_BASE. Default: none, and the command uses main.")
	cmd.AddCommand(run, review, exportCommand(w, findRepo))
	return cmd
}

func exportCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var dir string
	var check bool
	cmd := &cobra.Command{
		Use:   "export claude",
		Short: "Write a Claude Code subagent for each AI role, from its role file.",
		Long: "Write .claude/agents/<role>.md for each AI role: a Claude Code subagent with the description, tools, " +
			"model, and instructions of its role file, and the rules of every role. A person in Claude Code can then " +
			"give a task to the role, and claude -p --agent <role> can run it. --check writes nothing, and fails when a " +
			"subagent is missing or differs from its role file.",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"claude"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "claude" {
				return fmt.Errorf("hephaestus can export to claude only, not %q.", args[0])
			}
			rolesDir, err := w.roles()
			if err != nil {
				return err
			}
			all, err := roles.Load(rolesDir)
			if err != nil {
				return err
			}
			if dir == "" {
				r, err := findRepo()
				if err != nil {
					return err
				}
				dir = filepath.Join(r.Root, ".claude", "agents")
			}
			var stale []string
			for _, role := range all {
				if role.Kind != roles.AI {
					continue
				}
				path := filepath.Join(dir, role.Flat()+".md")
				content := airole.ClaudeAgent(role)
				if check {
					if current, err := os.ReadFile(path); err != nil || string(current) != content {
						stale = append(stale, path)
					}
					continue
				}
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			if len(stale) > 0 {
				return fmt.Errorf("These subagents are missing or differ from their role files: %s. Run: hephaestus agent export claude", strings.Join(stale, ", "))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "The directory of the subagents. Default: .claude/agents of the repo.")
	cmd.Flags().BoolVar(&check, "check", false, "Write nothing; fail when a subagent is missing or differs from its role file.")
	return cmd
}
