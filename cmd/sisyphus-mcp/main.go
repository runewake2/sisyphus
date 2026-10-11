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
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/runewake2/sisyphus"
)

func main() {
	if err := newServer().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

// instructions tell the agent how to use the tools together. The client shows them to the model in
// every session, so they are short.
const instructions = `sisyphus tracks the issues of this repo as Markdown files in issues/. Work on one issue at a time:

1. Find work: sisyphus_list with ready true and state open. sisyphus_claims shows the issues that other workspaces have started.
2. Read the issue with sisyphus_show.
3. Claim it with sisyphus_start and your owner name. If the start fails because another owner or workspace has the issue, select another issue. Do not use take unless a human tells you to.
4. Record progress with sisyphus_edit, section notes, and append true.
5. For work that you find but that is out of scope, create an issue with sisyphus_new and deferredFrom.
6. When the work is complete, use sisyphus_close with a message that tells what was done, or sisyphus_abandon with a message that tells why the work stopped. Then run sisyphus_lint.

Read the warnings at the end of each result. Lists, for example tags and dependsOn, are comma-separated strings. Leave dir empty unless the user names another repo.`

// newServer returns the server with every tool. Its instructions also tell about a sisyphus binary
// with another version, so that the agent can tell the user before a tool fails.
func newServer() *mcp.Server {
	text := instructions
	if path, err := exec.LookPath("sisyphus"); err == nil {
		if mismatch := versionMismatch(path); mismatch != "" {
			text += "\n\nWarning: " + mismatch
		}
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "sisyphus", Version: sisyphus.Version()}, &mcp.ServerOptions{Instructions: text})
	registerTools(server)
	return server
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
	mismatch := versionMismatch(path)
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
		if mismatch != "" {
			// A flag that the other version does not know is the most likely cause.
			err = fmt.Errorf("%w\n%s", err, mismatch)
		}
	}
	return stdout, stderr, err
}

var (
	versionsMu sync.Mutex
	// versions holds the result of versionMismatch for each sisyphus binary, so that each binary
	// runs --version once.
	versions = map[string]string{}
)

// versionMismatch returns a message when the sisyphus binary at path has another version than
// sisyphus-mcp, and "" when the versions are the same or the version cannot be read. The two
// binaries are installed separately, so one of them can be old. The message is also written to the
// log of the server once.
func versionMismatch(path string) string {
	versionsMu.Lock()
	defer versionsMu.Unlock()
	if message, ok := versions[path]; ok {
		return message
	}
	message := ""
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, path, "--version").Output(); err == nil {
		if theirs, ours := strings.TrimSpace(string(out)), sisyphus.Version(); theirs != ours {
			message = fmt.Sprintf("sisyphus-mcp %s runs sisyphus %s from %s. Install the same version of both, for "+
				"example the latest: go install github.com/runewake2/sisyphus/cmd/sisyphus@latest and "+
				"go install github.com/runewake2/sisyphus/cmd/sisyphus-mcp@latest", ours, theirs, path)
			log.Print(message)
		}
	}
	versions[path] = message
	return message
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
