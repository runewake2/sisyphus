package main

import (
	"io"

	"github.com/spf13/cobra"
)

// startCommand is "sisyphus update <name> in-progress" in fewer words.
func startCommand(findRoot func() (string, error)) *cobra.Command {
	var owner, bookmark, workspace string
	var take bool
	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Start work on an issue: update <name> in-progress.",
		Long: "Start work on an issue. It is sisyphus update <name> in-progress, with the same rules and " +
			"warnings: for example, a warning when an issue that the issue depends on is not closed.\n\n" +
			"An issue that is in progress with a different owner is refused, unless you add --take. A start " +
			"by the same owner changes nothing, so it is safe to repeat. Without --owner, the owner is the " +
			"value of " + ownerVariable + ".",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return updateIssue(root, args[0], "in-progress", updateOptions{
					owner:     optional(cmd, "owner", owner),
					bookmark:  optional(cmd, "bookmark", bookmark),
					workspace: optional(cmd, "workspace", workspace),
					start:     true,
					take:      take,
				}, warnings)
			})
		},
	}
	cmd.Flags().StringVar(&owner, "owner", "", "The person or agent working on the issue. Default: the value of "+ownerVariable+".")
	cmd.Flags().BoolVar(&take, "take", false, takeHelp)
	cmd.Flags().StringVarP(&bookmark, "bookmark", "b", "", "The jj bookmark or git branch of the work.")
	cmd.Flags().StringVar(&workspace, "workspace", "", "The jj workspace or git worktree where the work happens. Appended to the issue's workspace history.")
	return cmd
}

// closeCommand is "sisyphus update <name> closed --resolution <resolution>" in fewer words, as close
// (completed) or abandon (abandoned), with -m for the text of the Resolution section.
func closeCommand(findRoot func() (string, error), use, resolution string) *cobra.Command {
	var message string
	short := "Close an issue as completed: update <name> closed."
	long := "Close an issue as completed. It is sisyphus update <name> closed, with the same rules and warnings. " +
		"-m fills the Resolution section: what was done."
	if resolution == "abandoned" {
		short = "Close an issue as abandoned: the work will not happen."
		long = "Close an issue as abandoned: the work will not happen. It is sisyphus update <name> closed " +
			"--resolution abandoned, with the same rules and warnings. -m fills the Resolution section: why the " +
			"work stopped, and what replaces it."
	}
	cmd := &cobra.Command{
		Use:   use + " <name>",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return updateIssue(root, args[0], "closed", updateOptions{
					resolution: resolution,
					message:    optional(cmd, "message", message),
				}, warnings)
			})
		},
	}
	cmd.Flags().StringVarP(&message, "message", "m", "", "The text of the Resolution section.")
	return cmd
}
