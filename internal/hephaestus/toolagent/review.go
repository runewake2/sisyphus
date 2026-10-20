// The referendum review of a tool role: its ballot comes from its exit code and the changes that
// its command makes, so that voting stays one file away from the task lifecycle of the agent.

package toolagent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/runewake2/sisyphus/internal/hephaestus/repo"
	"github.com/runewake2/sisyphus/internal/hephaestus/roles"
	"github.com/runewake2/sisyphus/internal/hephaestus/shell"
)

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
	cmd := shell.Command(ctx, role.Command)
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
		if !slices.Contains(beforeFiles, file) || fileDiff(before, file) != fileDiff(after, file) {
			ballot.Files = append(ballot.Files, file)
		}
	}
	return ballot, nil
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
