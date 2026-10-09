package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// buildSisyphusOnce builds the sisyphus binary this test's tools shell out to, once per test run,
// and prepends its directory to PATH so exec.LookPath("sisyphus") finds it.
var buildSisyphusOnce = sync.OnceFunc(func() {
	dir, err := os.MkdirTemp("", "sisyphus-mcp-test-bin")
	if err != nil {
		panic(err)
	}
	bin := filepath.Join(dir, "sisyphus")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/runewake2/sisyphus/cmd/sisyphus")
	if output, err := cmd.CombinedOutput(); err != nil {
		panic("building sisyphus for tests: " + err.Error() + "\n" + string(output))
	}
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		panic(err)
	}
})

// connectedSession starts a server with every tool registered, connects a client to it over an
// in-memory transport, and returns the client session, cleaned up when the test ends.
func connectedSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	buildSisyphusOnce()
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "sisyphus-test"}, nil)
	registerTools(server)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatal(err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

// callTool calls name with arguments and fails the test if the call itself errors or the tool
// reports IsError. It returns the concatenated text content.
func callTool(t *testing.T, session *mcp.ClientSession, name string, arguments map[string]any) string {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var text strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			text.WriteString(tc.Text)
		}
	}
	if res.IsError {
		t.Fatalf("%s reported an error: %s", name, text.String())
	}
	return text.String()
}

// callToolExpectingError is like callTool, but asserts the call reports IsError and returns the
// error text instead of failing the test.
func callToolExpectingError(t *testing.T, session *mcp.ClientSession, name string, arguments map[string]any) string {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var text strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			text.WriteString(tc.Text)
		}
	}
	if !res.IsError {
		t.Fatalf("%s: want an error, got: %s", name, text.String())
	}
	return text.String()
}

func TestToolsCreateShowListSearchGraph(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})

	callTool(t, session, "sisyphus_new", map[string]any{
		"dir": dir, "name": "fix-login-bug", "title": "Fix the login bug",
		"priority": "high", "tags": "auth,bug",
	})
	callTool(t, session, "sisyphus_new", map[string]any{
		"dir": dir, "name": "sub-issue-test", "parent": "fix-login-bug",
	})

	show := callTool(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "fix-login-bug"})
	var shown map[string]any
	if err := json.Unmarshal([]byte(show), &shown); err != nil {
		t.Fatalf("show did not return JSON: %v\n%s", err, show)
	}
	if shown["title"] != "Fix the login bug" {
		t.Errorf("want title 'Fix the login bug', got %v", shown["title"])
	}

	list := callTool(t, session, "sisyphus_list", map[string]any{"dir": dir})
	var rows []map[string]any
	if err := json.Unmarshal([]byte(list), &rows); err != nil {
		t.Fatalf("list did not return JSON: %v\n%s", err, list)
	}
	if len(rows) != 2 {
		t.Errorf("want 2 issues, got %d: %s", len(rows), list)
	}

	search := callTool(t, session, "sisyphus_search", map[string]any{"dir": dir, "query": "login"})
	if !strings.Contains(search, "fix-login-bug") {
		t.Errorf("search for 'login' did not find fix-login-bug: %s", search)
	}

	graph := callTool(t, session, "sisyphus_graph", map[string]any{"dir": dir, "name": "sub-issue-test", "full": true})
	if !strings.Contains(graph, "fix-login-bug") || !strings.Contains(graph, "sub-issue-test") {
		t.Errorf("graph did not show both issues: %s", graph)
	}

	source := callTool(t, session, "sisyphus_graph", map[string]any{"dir": dir, "name": "sub-issue-test", "mermaid": true, "full": true})
	if !strings.HasPrefix(source, "graph LR") || !strings.Contains(source, "sub-issue-test [") || !strings.Contains(source, "classDef focus") {
		t.Errorf("graph did not return Mermaid source: %s", source)
	}
}

func TestToolsUpdateParentRemoteDependsOn(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "blocking-issue-test"})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "parent-issue-test"})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "main-issue-test"})

	callTool(t, session, "sisyphus_update", map[string]any{
		"dir": dir, "name": "main-issue-test", "state": "in-progress",
		"bookmark": "ai/work", "owner": "alice",
		"metadata": map[string]any{"session-id": "abc123"},
	})
	callTool(t, session, "sisyphus_parent", map[string]any{"dir": dir, "name": "main-issue-test", "parent": "parent-issue-test"})
	callTool(t, session, "sisyphus_remote", map[string]any{"dir": dir, "name": "main-issue-test", "url": "https://github.com/acme/widgets/issues/1"})
	callTool(t, session, "sisyphus_depends_on", map[string]any{"dir": dir, "name": "main-issue-test", "blockingIssue": "blocking-issue-test"})

	show := callTool(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "main-issue-test"})
	var shown map[string]any
	if err := json.Unmarshal([]byte(show), &shown); err != nil {
		t.Fatalf("show did not return JSON: %v\n%s", err, show)
	}
	equalField(t, shown, "owner", "alice")
	equalField(t, shown, "bookmark", "ai/work")
	equalField(t, shown, "parent", "[[parent-issue-test]]")
	equalField(t, shown, "remote", "https://github.com/acme/widgets/issues/1")
	if deps, _ := shown["depends-on"].([]any); len(deps) != 1 || deps[0] != "blocking-issue-test" {
		t.Errorf("want depends-on [blocking-issue-test], got %v", shown["depends-on"])
	}
	if meta, _ := shown["metadata"].(map[string]any); meta["session-id"] != "abc123" {
		t.Errorf("want metadata session-id abc123, got %v", shown["metadata"])
	}
}

func equalField(t *testing.T, m map[string]any, key, want string) {
	t.Helper()
	if got, _ := m[key].(string); got != want {
		t.Errorf("want %s=%q, got %q", key, want, got)
	}
}

func TestToolEditFillsTheResolution(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})

	closed := callTool(t, session, "sisyphus_update", map[string]any{
		"dir": dir, "name": "fix-login-bug", "state": "closed", "resolution": "abandoned",
	})
	if !strings.Contains(closed, "issues/closed/fix-login-bug.md") || !strings.Contains(closed, "Warning: The Resolution section") {
		t.Errorf("update did not return the path and the warning: %s", closed)
	}

	callTool(t, session, "sisyphus_edit", map[string]any{
		"dir": dir, "name": "fix-login-bug", "section": "resolution", "text": "Abandoned. Login moved to SSO.",
	})
	callTool(t, session, "sisyphus_edit", map[string]any{
		"dir": dir, "name": "fix-login-bug", "section": "notes", "text": "2026-10-09: abandoned.", "append": true,
	})

	show := callTool(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "fix-login-bug"})
	var shown map[string]any
	if err := json.Unmarshal([]byte(show), &shown); err != nil {
		t.Fatalf("show did not return JSON: %v\n%s", err, show)
	}
	body, _ := shown["body"].(string)
	if shown["resolution"] != "abandoned" || !strings.Contains(body, "## Resolution\n\nAbandoned. Login moved to SSO.") ||
		!strings.Contains(body, "## Notes\n\n2026-10-09: abandoned.\n") {
		t.Errorf("the edits are not in the issue: %s", show)
	}
}

func TestToolsSlugResolveLinks(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})

	slug := callTool(t, session, "sisyphus_slug", map[string]any{"dir": dir, "text": "Fix the login bug"})
	if strings.TrimSpace(slug) != "fix-login-bug" {
		t.Errorf("want fix-login-bug, got %q", slug)
	}

	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-the-login-bug"})

	resolve := callTool(t, session, "sisyphus_resolve", map[string]any{"dir": dir, "link": "fix-the-login-bug"})
	if !strings.Contains(resolve, "issues/open/fix-the-login-bug.md") {
		t.Errorf("want resolve to find the issue, got %q", resolve)
	}

	links := callTool(t, session, "sisyphus_links", map[string]any{"dir": dir, "document": "issues/open/fix-the-login-bug.md"})
	if strings.TrimSpace(links) == "" {
		t.Errorf("want JSON (even if empty array), got empty string")
	}
}

func TestToolReportsAMissingIssueAsAnError(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})

	message := callToolExpectingError(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "no-such-issue"})
	if !strings.Contains(message, "No issue named 'no-such-issue'") {
		t.Errorf("want the missing-issue error, got %q", message)
	}
}
