package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// editInEditor opens the file of an issue in the editor of the user, and checks the file when the
// editor exits.
func editInEditor(cmd *cobra.Command, findRoot func() (string, error), reference string) error {
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
	issue, err := x.find(issueName(reference))
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
}
