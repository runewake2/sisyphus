package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/runewake2/sisyphus"
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
	return connectedSessionWith(t, serverOptions{allowDir: true}) // the tests give each call its own repo
}

// connectedSessionWith is connectedSession for a server with the options o.
func connectedSessionWith(t *testing.T, o serverOptions) *mcp.ClientSession {
	t.Helper()
	buildSisyphusOnce()
	ctx := context.Background()

	server := newServer(o)
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

// Free text that starts with "-", for example a Markdown list item, must reach sisyphus as text,
// not as a flag.
func TestToolsPassTextThatStartsWithADash(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug", "title": "-leading dash title"})

	callTool(t, session, "sisyphus_edit", map[string]any{
		"dir": dir, "name": "fix-login-bug", "section": "Acceptance criteria", "text": "- [ ] The login works.",
	})
	show := callTool(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "fix-login-bug"})
	if !strings.Contains(show, "- [ ] The login works.") || !strings.Contains(show, "-leading dash title") {
		t.Errorf("show does not have the dash text: %s", show)
	}

	search := callTool(t, session, "sisyphus_search", map[string]any{"dir": dir, "query": "-leading"})
	if !strings.Contains(search, "fix-login-bug") {
		t.Errorf("search for '-leading' did not find fix-login-bug: %s", search)
	}

	slug := callTool(t, session, "sisyphus_slug", map[string]any{"dir": dir, "text": "-dash first words"})
	if slug != "dash-first-words" {
		t.Errorf("want slug dash-first-words, got %q", slug)
	}
}

func TestUpdateToolChangesFieldsWithoutAState(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})

	callTool(t, session, "sisyphus_update", map[string]any{"dir": dir, "name": "fix-login-bug", "priority": "critical"})

	show := callTool(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "fix-login-bug"})
	if !strings.Contains(show, `"priority": "critical"`) || !strings.Contains(show, `"state": "open"`) {
		t.Errorf("want priority critical and state open: %s", show)
	}
}

func TestNewToolTakesMoreThanOneDependency(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "first-issue-test"})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "second-issue-test"})

	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "third-issue-test", "dependsOn": "first-issue-test,second-issue-test"})

	show := callTool(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "third-issue-test"})
	if !strings.Contains(show, `"first-issue-test"`) || !strings.Contains(show, `"second-issue-test"`) {
		t.Errorf("want both dependencies: %s", show)
	}
}

func TestLintToolReturnsTheProblems(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})

	clean := callTool(t, session, "sisyphus_lint", map[string]any{"dir": dir})
	if !strings.Contains(clean, `"problems": []`) {
		t.Errorf("want no problems: %s", clean)
	}

	path := filepath.Join(dir, "issues", "open", "fix-login-bug.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(content), "priority: medium", "priority: urgent", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	found := callTool(t, session, "sisyphus_lint", map[string]any{"dir": dir})
	if !strings.Contains(found, "its priority 'urgent'") {
		t.Errorf("want the priority problem: %s", found)
	}
}

func TestGraphToolRefusesAnUnreadableDrawing(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "big-epic-test"})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "one-sub-test", "parent": "big-epic-test"})

	message := callToolExpectingError(t, session, "sisyphus_graph", map[string]any{"dir": dir, "name": "big-epic-test", "maxIssues": 1})
	if !strings.Contains(message, "more than --max-issues (1)") {
		t.Errorf("want the size error: %s", message)
	}
	callTool(t, session, "sisyphus_graph", map[string]any{"dir": dir, "name": "big-epic-test", "json": true, "maxIssues": 1})
}

func TestNewToolMakesTheNameFromTheTitle(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})

	out := callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "title": "Add a dark mode"})

	if !strings.Contains(out, "issues/open/add-dark-mode.md") {
		t.Errorf("want the name made from the title: %s", out)
	}
}

func TestWriteToolsReturnJSONWhenAsked(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})

	out := callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug", "json": true})

	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	if got["name"] != "fix-login-bug" || got["path"] != "issues/open/fix-login-bug.md" {
		t.Errorf("want the issue and its path: %s", out)
	}
}

// TestToolDescriptionsDoNotAssumeJJ checks that no description names jj without git: sisyphus works
// the same in a repo that uses only git.
func TestToolDescriptionsDoNotAssumeJJ(t *testing.T) {
	session := connectedSession(t)
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		texts := map[string]string{"description": tool.Description}
		schema, _ := json.Marshal(tool.InputSchema)
		var parsed struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(schema, &parsed); err != nil {
			t.Fatal(err)
		}
		for name, property := range parsed.Properties {
			texts[name] = property.Description
		}
		for name, text := range texts {
			if strings.Contains(text, "jj ") && !strings.Contains(text, "git") {
				t.Errorf("%s %s names jj without git: %q", tool.Name, name, text)
			}
		}
	}
}

func TestUpdateToolRefusesAClaimedIssueWithoutTake(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})
	callTool(t, session, "sisyphus_update", map[string]any{"dir": dir, "name": "fix-login-bug", "state": "in-progress", "owner": "a1"})

	refused := callToolExpectingError(t, session, "sisyphus_update", map[string]any{"dir": dir, "name": "fix-login-bug", "owner": "a2"})
	taken := callTool(t, session, "sisyphus_update", map[string]any{"dir": dir, "name": "fix-login-bug", "owner": "a2", "take": true})

	if !strings.Contains(refused, "owned by a1") {
		t.Errorf("want the owner in the error, got: %s", refused)
	}
	if !strings.Contains(taken, "was owned by a1") {
		t.Errorf("want a warning with the old owner, got: %s", taken)
	}
}

func TestClaimsToolListsTheClaims(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})
	callTool(t, session, "sisyphus_update", map[string]any{"dir": dir, "name": "fix-login-bug", "state": "in-progress", "owner": "a1"})

	out := callTool(t, session, "sisyphus_claims", map[string]any{"dir": dir})

	var claims []map[string]any
	if err := json.Unmarshal([]byte(out), &claims); err != nil {
		t.Fatalf("claims did not return JSON: %v\n%s", err, out)
	}
	if len(claims) != 1 || claims[0]["name"] != "fix-login-bug" || claims[0]["status"] != "here" {
		t.Errorf("want one claim of this workspace, got %s", out)
	}
}

func TestLifecycleToolsStartCloseAndAbandon(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "old-login-form"})

	started := callTool(t, session, "sisyphus_start", map[string]any{"dir": dir, "name": "fix-login-bug", "owner": "a1", "bookmark": "ai/fix"})
	refused := callToolExpectingError(t, session, "sisyphus_start", map[string]any{"dir": dir, "name": "fix-login-bug", "owner": "a2"})
	closed := callTool(t, session, "sisyphus_close", map[string]any{"dir": dir, "name": "fix-login-bug", "message": "- Fixed the redirect.", "json": true})
	abandoned := callTool(t, session, "sisyphus_abandon", map[string]any{"dir": dir, "name": "old-login-form", "message": "Replaced by SSO."})

	if started != "issues/in-progress/fix-login-bug.md" {
		t.Errorf("start: got %q", started)
	}
	if !strings.Contains(refused, "owned by a1") {
		t.Errorf("a second start: got %q", refused)
	}
	var issue map[string]any
	if err := json.Unmarshal([]byte(closed), &issue); err != nil {
		t.Fatalf("close --json: %v\n%s", err, closed)
	}
	if issue["state"] != "closed" || issue["resolution"] != "completed" || !strings.Contains(issue["body"].(string), "- Fixed the redirect.") {
		t.Errorf("close: got %s", closed)
	}
	if abandoned != "issues/closed/old-login-form.md" {
		t.Errorf("abandon: got %q", abandoned)
	}
	shown := callTool(t, session, "sisyphus_show", map[string]any{"dir": dir, "name": "old-login-form"})
	if !strings.Contains(shown, `"resolution": "abandoned"`) {
		t.Errorf("abandon: want resolution abandoned, got %s", shown)
	}
}

func TestCloseToolsNeedAMessage(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})

	for _, tool := range []string{"sisyphus_close", "sisyphus_abandon"} {
		text := callToolExpectingError(t, session, tool, map[string]any{"dir": dir, "name": "fix-login-bug", "message": " "})
		if !strings.Contains(text, "Give message") {
			t.Errorf("%s: got %q", tool, text)
		}
	}
}

// TestAnOtherVersionOfSisyphusIsReportedInErrors runs the tools against a fake sisyphus that has
// another version and fails every command, as an old binary fails on a flag it does not know.
func TestAnOtherVersionOfSisyphusIsReportedInErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake sisyphus is a shell script")
	}
	session := connectedSession(t)
	bin := t.TempDir()
	script := "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 0.0.1; exit 0; fi\necho \"Error: unknown flag: --take\" >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "sisyphus"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	text := callToolExpectingError(t, session, "sisyphus_list", map[string]any{"dir": t.TempDir()})

	if !strings.Contains(text, "unknown flag: --take") || !strings.Contains(text, "runs sisyphus 0.0.1 from "+filepath.Join(bin, "sisyphus")) {
		t.Errorf("want the error and the version difference, got: %s", text)
	}
}

func TestTheSameVersionAddsNothingToErrors(t *testing.T) {
	session := connectedSession(t)

	text := callToolExpectingError(t, session, "sisyphus_show", map[string]any{"dir": t.TempDir(), "name": "missing-issue"})

	if strings.Contains(text, "Install the same version") {
		t.Errorf("the test binary has the version of sisyphus-mcp, got: %s", text)
	}
}

func TestInstructionsAreShortAndNameOnlyToolsThatExist(t *testing.T) {
	session := connectedSession(t)
	text := session.InitializeResult().Instructions
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}

	if words := len(strings.Fields(text)); words == 0 || words > 200 {
		t.Errorf("want 1-200 words of instructions, got %d", words)
	}
	for _, word := range strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == ',' || r == '.' || r == '\n' }) {
		if strings.HasPrefix(word, "sisyphus_") && !names[word] {
			t.Errorf("the instructions name %s, which is not a tool", word)
		}
	}
	for _, step := range []string{"sisyphus_start", "sisyphus_close", "sisyphus_list"} {
		if !strings.Contains(text, step) {
			t.Errorf("the instructions do not name %s", step)
		}
	}
}

func TestTheDocsShowTheInstructions(t *testing.T) {
	page, err := os.ReadFile("../../docs/sisyphus-mcp.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), instructions) {
		t.Error("docs/sisyphus-mcp.md does not show the instructions of the server as they are")
	}
}

func TestEditToolWithAnEmptySectionDoesNotStartAnEditor(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})
	t.Setenv("VISUAL", "sleep 30")

	text := callToolExpectingError(t, session, "sisyphus_edit", map[string]any{"dir": dir, "name": "fix-login-bug", "section": "", "text": ""})

	if !strings.Contains(text, "Give section") {
		t.Errorf("got %q", text)
	}
}

func TestEditToolWithEmptyTextSaysThatTheTextIsEmpty(t *testing.T) {
	session := connectedSession(t)
	dir := t.TempDir()
	callTool(t, session, "sisyphus_init", map[string]any{"dir": dir})
	callTool(t, session, "sisyphus_new", map[string]any{"dir": dir, "name": "fix-login-bug"})

	text := callToolExpectingError(t, session, "sisyphus_edit", map[string]any{"dir": dir, "name": "fix-login-bug", "section": "notes", "text": ""})

	if !strings.Contains(text, "The text is empty") {
		t.Errorf("got %q", text)
	}
}

// slowSisyphus puts a fake sisyphus first on PATH that has the version of sisyphus-mcp, and that
// waits for seconds and then creates the file marker for every other command.
func slowSisyphus(t *testing.T, seconds, marker string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake sisyphus is a shell script")
	}
	bin := t.TempDir()
	script := "#!/bin/sh\nif [ \"$1\" = --version ]; then echo " + sisyphus.Version() + "; exit 0; fi\nsleep " + seconds + "\ntouch " + marker + "\n"
	if err := os.WriteFile(filepath.Join(bin, "sisyphus"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestASlowCommandStopsAtTheTimeLimit(t *testing.T) {
	session := connectedSession(t)
	marker := filepath.Join(t.TempDir(), "finished")
	slowSisyphus(t, "3", marker)
	t.Setenv(timeoutVariable, "200ms")

	start := time.Now()
	text := callToolExpectingError(t, session, "sisyphus_list", map[string]any{"dir": t.TempDir()})

	if !strings.Contains(text, "did not finish in 200ms") || !strings.Contains(text, timeoutVariable) {
		t.Errorf("got %q", text)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("the call took %s", elapsed)
	}
}

func TestACancelledCallStopsTheCommand(t *testing.T) {
	session := connectedSession(t)
	marker := filepath.Join(t.TempDir(), "finished")
	slowSisyphus(t, "1", marker)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "sisyphus_list", Arguments: map[string]any{"dir": t.TempDir()}})
	if err == nil {
		t.Fatal("want the cancelled call to fail")
	}
	time.Sleep(1500 * time.Millisecond)

	if _, err := os.Stat(marker); err == nil {
		t.Error("the sisyphus process ran to its end after the call was cancelled")
	}
}

func TestAnInvalidTimeoutUsesTheDefault(t *testing.T) {
	t.Setenv(timeoutVariable, "soon")

	if got := timeout(); got != defaultTimeout {
		t.Errorf("want %s, got %s", defaultTimeout, got)
	}
}

func TestByDefaultNoToolHasADirArgument(t *testing.T) {
	session := connectedSessionWith(t, serverOptions{})
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		schema, _ := json.Marshal(tool.InputSchema)
		if strings.Contains(string(schema), `"dir"`) {
			t.Errorf("%s has a dir argument", tool.Name)
		}
	}
	if strings.Contains(session.InitializeResult().Instructions, "dir") {
		t.Error("the instructions name dir, which the tools do not have")
	}
}

func TestByDefaultADirArgumentIsRefused(t *testing.T) {
	session := connectedSessionWith(t, serverOptions{})

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "sisyphus_list", Arguments: map[string]any{"dir": "docs"}})

	if err == nil && !res.IsError {
		t.Error("want a call with dir to fail")
	}
}

func TestTheToolsActOnTheDirectoryOfTheServer(t *testing.T) {
	dir := t.TempDir()
	original := defaultDir
	defaultDir = dir
	t.Cleanup(func() { defaultDir = original })
	session := connectedSessionWith(t, serverOptions{})

	callTool(t, session, "sisyphus_init", map[string]any{})
	created := callTool(t, session, "sisyphus_new", map[string]any{"title": "Fix the login bug"})

	if created != "issues/open/fix-login-bug.md" {
		t.Errorf("got %q", created)
	}
	if _, err := os.Stat(filepath.Join(dir, "issues", "open", "fix-login-bug.md")); err != nil {
		t.Errorf("the issue is not in the directory of the server: %v", err)
	}
}
