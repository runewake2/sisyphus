package main

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
)

// delegateKey is the key of the sisyphus metadata field that names the role assigned to own a task
// (Q1 of the agent delegation design). sisyphus does not know what it means.
const delegateKey = "delegate"

// workspace is where hephaestus finds the repo, its roles, and the sisyphus tracker of a command.
type workspace struct {
	// directory returns the -C directory: sisyphus runs there, so that it finds the tracker of the
	// user. It is a function, because the flag is read after the commands are made.
	directory func() string
	// roles is the roles directory, from --roles or the repo.
	roles func() (string, error)
}

func rolesCommand(w workspace) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "roles",
		Short: "List the roles, and check their role files.",
		Long: "List the roles of hephaestus/roles/<role>.md: each role file describes one role, its kind (ai or tool), " +
			"its triggers, and what it does when it cannot finish a task. A role file with a problem is an error that " +
			"names the problem.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := w.roles()
			if err != nil {
				return err
			}
			all, loadErr := roles.Load(dir)
			if asJSON {
				if all == nil {
					all = []roles.Role{}
				}
				if err := writeJSON(cmd.OutOrStdout(), all); err != nil {
					return err
				}
				return loadErr
			}
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(table, "role\tkind\ttrigger\ton-failure\tcharter")
			for _, r := range all {
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.Kind, strings.Join(r.Triggers, ","), r.OnFailure, r.Charter)
			}
			table.Flush()
			return loadErr
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print JSON.")
	return cmd
}

func delegateCommand(w workspace) *cobra.Command {
	var clear bool
	cmd := &cobra.Command{
		Use:   "delegate <issue> [<role>]",
		Short: "Delegate a sisyphus issue to a role: the role that owns the task.",
		Long: "Delegate a sisyphus issue to a role, with the sisyphus metadata key delegate: sisyphus update <issue> " +
			"--metadata delegate=<role>. The role must have a role file. The owner of the issue stays the worker that " +
			"works on it. --clear removes the delegation.",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			value := ""
			switch {
			case clear && len(args) == 2:
				return fmt.Errorf("Give a role or --clear, not both.")
			case !clear && len(args) == 1:
				return fmt.Errorf("Give the role to delegate %s to, or --clear.", args[0])
			case !clear:
				dir, err := w.roles()
				if err != nil {
					return err
				}
				role, err := roles.Find(dir, args[1])
				if err != nil {
					return err
				}
				value = role.Name
			}
			out, err := sisyphus.Run(w.directory(), "update", args[0], "--metadata", delegateKey+"="+value)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().BoolVar(&clear, "clear", false, "Remove the delegation of the issue.")
	return cmd
}

func tasksCommand(w workspace) *cobra.Command {
	var asJSON, all bool
	cmd := &cobra.Command{
		Use:   "tasks <role>",
		Short: "List the tasks delegated to a role that can start now.",
		Long: "List the sisyphus issues delegated to a role (metadata delegate=<role>) that can start now: sisyphus " +
			"list --metadata delegate=<role> --state open --ready. --all lists every issue delegated to the role, in " +
			"any state.",
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
			filter := []string{"--metadata", delegateKey + "=" + role.Name}
			if all {
				filter = append(filter, "--state", "all")
			} else {
				filter = append(filter, "--state", "open", "--ready")
			}
			issues, err := sisyphus.List(w.directory(), filter...)
			if err != nil {
				return err
			}
			if asJSON {
				if issues == nil {
					issues = []sisyphus.Issue{}
				}
				return writeJSON(cmd.OutOrStdout(), issues)
			}
			if len(issues) == 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "No task delegated to %s %s.\n", role.Name, map[bool]string{true: "in any state", false: "can start now"}[all])
				return nil
			}
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(table, "name\tstate\tpriority\towner\ttitle")
			for _, issue := range issues {
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", issue.Name, issue.State, issue.Priority, issue.Owner, issue.Title)
			}
			return table.Flush()
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "List the tasks of the role in every state.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print JSON: the rows of sisyphus list --json.")
	return cmd
}
