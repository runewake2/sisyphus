package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// hasTerminal reports whether in is a terminal, where an editor can run. Tests replace it.
var hasTerminal = func(in io.Reader) bool {
	f, ok := in.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// noTerminalError is the error of edit with only a name when no terminal is attached, for example
// in the shell tool of an agent or in sisyphus-mcp, where an editor would wait for ever.
const noTerminalError = "No terminal to run your editor in, so sisyphus edit <name> would wait for an editor that nobody can use. To change one section, use: sisyphus edit <name> <section> <text>"

// editInEditor opens the file of an issue in the editor of the user, and checks the file when the
// editor exits. Without a terminal on standard input, it fails at once.
func editInEditor(cmd *cobra.Command, findRoot func() (string, error), reference string) error {
	if !hasTerminal(cmd.InOrStdin()) {
		return errors.New(noTerminalError)
	}
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
