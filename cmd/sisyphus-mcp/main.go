// Command sisyphus-mcp is an MCP server that gives an agent sisyphus's issue-tracking commands as
// tools, over stdio. It is a thin wrapper: every tool shells out to the sisyphus binary (which must
// be installed and on PATH) and returns its output, so the tools stay exactly in step with the CLI
// they wrap.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/runewake2/sisyphus"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "sisyphus", Version: sisyphus.Version()}, nil)
	registerTools(server)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

// errNotInstalled is the error of every tool when the sisyphus binary is not on PATH.
var errNotInstalled = errors.New("sisyphus is not installed or not on PATH. Install it with " +
	"`go install github.com/runewake2/sisyphus/cmd/sisyphus@latest`.")

// execSisyphus runs the sisyphus binary with args, in dir if given (otherwise sisyphus-mcp's own
// working directory), and returns its trimmed stdout and stderr. A failed run gives sisyphus's own
// stderr message as the error, the same message a human running the command would see.
func execSisyphus(dir string, args ...string) (stdout, stderr string, err error) {
	path, err := exec.LookPath("sisyphus")
	if err != nil {
		return "", "", errNotInstalled
	}
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	stdout, stderr = strings.TrimSpace(out.String()), strings.TrimSpace(errOut.String())
	if err != nil {
		if stderr != "" {
			err = errors.New(stderr)
		} else {
			err = fmt.Errorf("sisyphus %s: %w", strings.Join(args, " "), err)
		}
	}
	return stdout, stderr, err
}

// runSisyphus runs sisyphus and returns its stdout, followed by any warnings it wrote to stderr.
func runSisyphus(dir string, args ...string) (string, error) {
	stdout, stderr, err := execSisyphus(dir, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout + "\n" + stderr), nil
}

// runSisyphusReport runs sisyphus for a command that reports findings on stdout and exits 1 when
// it finds any, such as lint. A run that wrote to stdout is a result, even with exit code 1; only a
// run that wrote nothing to stdout is an error.
func runSisyphusReport(dir string, args ...string) (string, error) {
	stdout, _, err := execSisyphus(dir, args...)
	if err != nil && stdout == "" {
		return "", err
	}
	return stdout, nil
}

// textResult wraps text as a successful CallToolResult.
func textResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, nil, nil
}
