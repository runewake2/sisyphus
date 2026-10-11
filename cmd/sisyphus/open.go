package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// openCommand opens an issue file in the editor of the user, and checks the file when the editor exits.
func openCommand(findRoot func() (string, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "open <name>",
		Short: "Open an issue in your editor ($VISUAL or $EDITOR), and check it after.",
		Long: "Open the file of an issue in your editor: $VISUAL, or else $EDITOR. The value can have arguments, " +
			"for example \"code --wait\". When the editor exits, sisyphus checks the file as lint does, and warns " +
			"about problems, for example a frontmatter line that it cannot read.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := strings.Fields(os.Getenv("VISUAL"))
			if len(editor) == 0 {
				editor = strings.Fields(os.Getenv("EDITOR"))
			}
			if len(editor) == 0 {
				return errors.New("Set VISUAL or EDITOR to the command of your editor, for example: export EDITOR=vim")
			}
			root, err := findRoot()
			if err != nil {
				return err
			}
			x := loadIndex(root)
			issue, err := x.find(issueName(args[0]))
			if err != nil {
				return err
			}
			run := exec.Command(editor[0], append(editor[1:], issue.path)...)
			run.Stdin, run.Stdout, run.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
			if err := run.Run(); err != nil {
				return fmt.Errorf("The editor %q failed: %w", strings.Join(editor, " "), err)
			}
			doc, err := loadDocument(issue.path)
			warnIssueProblems(x, issue, doc, err, cmd.ErrOrStderr())
			return nil
		},
	}
}
