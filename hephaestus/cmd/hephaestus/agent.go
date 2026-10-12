package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
	"github.com/runewake2/sisyphus/hephaestus/internal/toolagent"
)

func agentCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Run the workers of a tool role, which do their work with a program and no model.",
	}
	var as string
	var once, asJSON bool
	run := &cobra.Command{
		Use:   "run <role>",
		Short: "Run a worker of a tool role: do the task of each message that it gets.",
		Long: "Run a worker of a tool role. It gets the messages to it, to its role, and the broadcasts of the kinds " +
			"in listens of its role file. For each message with an issue, it claims the issue (sisyphus start, as " +
			"the worker), runs the command of the role in the workspace of the task (the data field workspace, or " +
			"-C), with the message as JSON on its standard input. Exit code 0 closes the issue with the output as its " +
			"Resolution. Another exit code records the output in Notes and stops the issue: tagged needs-human, or " +
			"delegated to the AI role of on-failure, with a message to that role. The worker sends the result to the " +
			"sender of the message, and acknowledges the message. It makes no model call. --once handles the messages " +
			"that wait now, and stops.",
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
			if err := toolagent.Validate(role); err != nil {
				return err
			}
			name := as
			if name == "" {
				name = role.Name + "-1"
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
				Role: role, Worker: name, Dir: w.directory(), Queue: client, Sisyphus: sisyphus.Run,
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
	run.Flags().BoolVar(&asJSON, "json", false, "Print the result of each message as JSON when it stops.")
	var reviewJSON bool
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
			ballot, err := toolagent.Review(cmd.Context(), role, w.directory(), cmd.ErrOrStderr())
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
	cmd.AddCommand(run, review)
	return cmd
}
