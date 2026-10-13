// Package sisyphus runs the sisyphus command for hephaestus. hephaestus uses sisyphus only through
// its command, as any other client does: it never reads or writes issue files itself.
package sisyphus

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

// ErrNotInstalled is the error when the sisyphus command is not on PATH.
var ErrNotInstalled = errors.New("sisyphus is not installed or not on PATH. Install it with `go install github.com/runewake2/sisyphus/cmd/sisyphus@latest`.")

// Run runs sisyphus with args in dir, and returns its stdout. A failure gives the error message of
// sisyphus, without its "Error: " prefix.
func Run(dir string, args ...string) (string, error) {
	path, err := exec.LookPath("sisyphus")
	if err != nil {
		return "", ErrNotInstalled
	}
	cmd := exec.Command(path, append([]string{"-C", dir}, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		message := strings.TrimPrefix(strings.TrimSpace(errOut.String()), "Error: ")
		if message == "" {
			message = err.Error()
		}
		return "", errors.New(message)
	}
	return out.String(), nil
}

// Issue is a row of sisyphus list --json.
type Issue struct {
	Name     string            `json:"name"`
	Title    string            `json:"title"`
	State    string            `json:"state"`
	Priority string            `json:"priority"`
	Owner    string            `json:"owner,omitempty"`
	Bookmark string            `json:"bookmark,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// List runs sisyphus list --json with args, and returns its rows.
func List(dir string, args ...string) ([]Issue, error) {
	out, err := Run(dir, append([]string{"list", "--json"}, args...)...)
	if err != nil {
		return nil, err
	}
	var issues []Issue
	if err := json.Unmarshal([]byte(out), &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// Detail is the part of sisyphus show --json that hephaestus reads.
type Detail struct {
	Name      string            `json:"name"`
	Title     string            `json:"title"`
	State     string            `json:"state"`
	Owner     string            `json:"owner,omitempty"`
	Bookmark  string            `json:"bookmark,omitempty"`
	DependsOn []string          `json:"depends-on,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Show runs sisyphus show --json for the issue reference.
func Show(dir, reference string) (Detail, error) {
	out, err := Run(dir, "show", "--json", "--", reference)
	if err != nil {
		return Detail{}, err
	}
	var detail Detail
	if err := json.Unmarshal([]byte(out), &detail); err != nil {
		return Detail{}, err
	}
	return detail, nil
}
