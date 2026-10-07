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

// runSisyphus runs the sisyphus binary with args, in dir if given (otherwise sisyphus-mcp's own
// working directory), and returns its trimmed stdout. On failure, the error is sisyphus's own
// stderr message, the same message a human running the command would see.
func runSisyphus(dir string, args ...string) (string, error) {
	path, err := exec.LookPath("sisyphus")
	if err != nil {
		return "", errors.New("sisyphus is not installed or not on PATH. Install it with " +
			"`go install github.com/runewake2/sisyphus/cmd/sisyphus@latest`.")
	}
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", errors.New(message)
		}
		return "", fmt.Errorf("sisyphus %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// textResult wraps text as a successful CallToolResult.
func textResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, nil, nil
}
