package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// appendFlag appends flag and value to args, unless value is empty.
func appendFlag(args []string, flag, value string) []string {
	if value == "" {
		return args
	}
	return append(args, flag, value)
}

// appendBoolFlag appends flag to args if value is true.
func appendBoolFlag(args []string, flag string, value bool) []string {
	if !value {
		return args
	}
	return append(args, flag)
}

// appendMetadataFlags appends one --metadata key=value pair per entry.
func appendMetadataFlags(args []string, metadata map[string]string) []string {
	for key, value := range metadata {
		args = append(args, "--metadata", key+"="+value)
	}
	return args
}

func registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_new",
		Description: "Create a sisyphus issue from issues/TEMPLATE.md.",
	}, newHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_update",
		Description: "Change a sisyphus issue's state (and move it between issues/open, issues/in-progress, and issues/closed), or edit its priority, effort, tags, owner, approver, bookmark, workspace, or metadata.",
	}, updateHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_show",
		Description: "Show one sisyphus issue: its frontmatter fields and body.",
	}, showHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_list",
		Description: "List and filter sisyphus issues by state, priority, tags, owner, parent, or whether they are blocked by an open dependency.",
	}, listHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_search",
		Description: "Search sisyphus issues by frontmatter filters and a free-text query over title and body, optionally scoped to one section (for example only the Summary).",
	}, searchHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_graph",
		Description: "Draw a sisyphus issue, everything below it (sub-issues and dependents, all the way down), and the path above it (parents and dependencies, all the way up), with their states, and say how many other linked issues are not drawn. Leaves (not closed, with no open dependency or sub-issue, so ready to start) are marked with \"🍃\" (and the issue itself with \"📍\") and have leaf: true in JSON. With full, draw every linked issue. Arrows point from the issue that comes first: solid from a parent to a sub-issue, dotted from a dependency to the issue that depends on it.",
	}, graphHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_parent",
		Description: "Set or clear a sisyphus issue's parent (making it a sub-issue of another issue).",
	}, parentHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_remote",
		Description: "Set or clear a sisyphus issue's remote reference: the URL of a GitHub issue or Jira ticket it corresponds to.",
	}, remoteHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_depends_on",
		Description: "Add or remove an issue that a sisyphus issue depends on (must close before it can start).",
	}, dependsOnHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_slug",
		Description: "Turn arbitrary text (for example a title) into a unique, valid sisyphus issue name.",
	}, slugHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_resolve",
		Description: "Find the file a wikilink points to, anywhere in the repo.",
	}, resolveHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_links",
		Description: "List every wikilink in a document and whether each one resolves.",
	}, linksHandler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "sisyphus_init",
		Description: "Set up issue tracking (issues/TEMPLATE.md and the open/in-progress/closed directories) in a repo.",
	}, initHandler)
}

type newArgs struct {
	Dir          string            `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Name         string            `json:"name" jsonschema:"The issue name: 2-6 lowercase kebab-case words, for example fix-login-bug. Must be unique in the repo."`
	Title        string            `json:"title,omitempty" jsonschema:"The issue title. Defaults to a title made from the name."`
	State        string            `json:"state,omitempty" jsonschema:"open, in-progress, or closed. Defaults to open."`
	Priority     string            `json:"priority,omitempty" jsonschema:"critical, high, medium, or low. Defaults to medium."`
	Effort       string            `json:"effort,omitempty" jsonschema:"small, medium, large, or x-large. Defaults to medium."`
	Resolution   string            `json:"resolution,omitempty" jsonschema:"completed or abandoned. Only valid when state is closed; defaults to completed."`
	Tags         string            `json:"tags,omitempty" jsonschema:"Comma-separated tags, for example bug,scheduler. Built-in: bug, feature; any other tag works too."`
	Bookmark     string            `json:"bookmark,omitempty" jsonschema:"The jj bookmark of the work on the issue."`
	DeferredFrom string            `json:"deferredFrom,omitempty" jsonschema:"An issue name, wikilink, or bookmark that deferred this work."`
	Parent       string            `json:"parent,omitempty" jsonschema:"The parent issue, if this is a sub-issue. Must already exist."`
	Owner        string            `json:"owner,omitempty" jsonschema:"The person or agent working on the issue."`
	Approver     string            `json:"approver,omitempty" jsonschema:"The person or agent who accepts the issue when it closes."`
	Workspace    string            `json:"workspace,omitempty" jsonschema:"The jj workspace where local work on the issue is happening."`
	Remote       string            `json:"remote,omitempty" jsonschema:"A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo."`
	DependsOn    string            `json:"dependsOn,omitempty" jsonschema:"An issue that must close before this one can start. Must already exist."`
	Context      string            `json:"context,omitempty" jsonschema:"Replaces the Context section's placeholder text with this text."`
	Metadata     map[string]string `json:"metadata,omitempty" jsonschema:"Arbitrary key-value notes, for example a session id so work can be resumed with context later."`
}

func newHandler(_ context.Context, _ *mcp.CallToolRequest, in newArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"new", in.Name}
	args = appendFlag(args, "--title", in.Title)
	args = appendFlag(args, "--state", in.State)
	args = appendFlag(args, "--priority", in.Priority)
	args = appendFlag(args, "--effort", in.Effort)
	args = appendFlag(args, "--resolution", in.Resolution)
	args = appendFlag(args, "--tags", in.Tags)
	args = appendFlag(args, "--bookmark", in.Bookmark)
	args = appendFlag(args, "--deferred-from", in.DeferredFrom)
	args = appendFlag(args, "--parent", in.Parent)
	args = appendFlag(args, "--owner", in.Owner)
	args = appendFlag(args, "--approver", in.Approver)
	args = appendFlag(args, "--workspace", in.Workspace)
	args = appendFlag(args, "--remote", in.Remote)
	args = appendFlag(args, "--depends-on", in.DependsOn)
	args = appendFlag(args, "--context", in.Context)
	args = appendMetadataFlags(args, in.Metadata)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type updateArgs struct {
	Dir        string            `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Name       string            `json:"name" jsonschema:"The issue name, [[name]], #name, or a path to the issue."`
	State      string            `json:"state" jsonschema:"The new state: open, in-progress, or closed."`
	Resolution string            `json:"resolution,omitempty" jsonschema:"completed or abandoned. Only valid when state is closed; defaults to completed."`
	Bookmark   string            `json:"bookmark,omitempty" jsonschema:"The jj bookmark of the work. Set it when work starts."`
	Owner      string            `json:"owner,omitempty" jsonschema:"The person or agent working on the issue."`
	Approver   string            `json:"approver,omitempty" jsonschema:"The person or agent who accepts the issue when it closes."`
	Workspace  string            `json:"workspace,omitempty" jsonschema:"The jj workspace where local work is happening. Appended to the issue's workspace history."`
	Priority   string            `json:"priority,omitempty" jsonschema:"Change the priority: critical, high, medium, or low."`
	Effort     string            `json:"effort,omitempty" jsonschema:"Change the effort: small, medium, large, or x-large."`
	Tags       string            `json:"tags,omitempty" jsonschema:"Replace the tags, comma-separated, for example bug,scheduler. Built-in: bug, feature; any other tag works too."`
	Metadata   map[string]string `json:"metadata,omitempty" jsonschema:"Arbitrary key-value notes to add or update, for example a session id. Merged into the existing notes; never cleared automatically."`
}

func updateHandler(_ context.Context, _ *mcp.CallToolRequest, in updateArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"update", in.Name, in.State}
	args = appendFlag(args, "--resolution", in.Resolution)
	args = appendFlag(args, "--bookmark", in.Bookmark)
	args = appendFlag(args, "--owner", in.Owner)
	args = appendFlag(args, "--approver", in.Approver)
	args = appendFlag(args, "--workspace", in.Workspace)
	args = appendFlag(args, "--priority", in.Priority)
	args = appendFlag(args, "--effort", in.Effort)
	args = appendFlag(args, "--tags", in.Tags)
	args = appendMetadataFlags(args, in.Metadata)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type showArgs struct {
	Dir  string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Name string `json:"name" jsonschema:"The issue name, [[name]], #name, or a path to the issue."`
	Text bool   `json:"text,omitempty" jsonschema:"Return the human-readable text form instead of JSON."`
}

func showHandler(_ context.Context, _ *mcp.CallToolRequest, in showArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"show", in.Name}
	args = appendBoolFlag(args, "--json", !in.Text)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type listArgs struct {
	Dir      string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	State    string `json:"state,omitempty" jsonschema:"Comma-separated states to include: open, in-progress, closed. Defaults to open and in-progress."`
	Priority string `json:"priority,omitempty" jsonschema:"Comma-separated priorities to include: critical, high, medium, low."`
	Tags     string `json:"tags,omitempty" jsonschema:"Comma-separated tags, for example bug; matches an issue with any of them. Built-in: bug, feature."`
	Owner    string `json:"owner,omitempty" jsonschema:"Only issues with exactly this owner."`
	Parent   string `json:"parent,omitempty" jsonschema:"Only direct sub-issues of this issue."`
	Blocked  bool   `json:"blocked,omitempty" jsonschema:"Only issues with a depends-on issue that is not yet closed."`
	Text     bool   `json:"text,omitempty" jsonschema:"Return the human-readable table instead of JSON."`
}

func listHandler(_ context.Context, _ *mcp.CallToolRequest, in listArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"list"}
	args = appendFlag(args, "--state", in.State)
	args = appendFlag(args, "--priority", in.Priority)
	args = appendFlag(args, "--tags", in.Tags)
	args = appendFlag(args, "--owner", in.Owner)
	args = appendFlag(args, "--parent", in.Parent)
	args = appendBoolFlag(args, "--blocked", in.Blocked)
	args = appendBoolFlag(args, "--json", !in.Text)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type searchArgs struct {
	Dir      string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Query    string `json:"query,omitempty" jsonschema:"A case-insensitive query matched against title and body. Optional: filters alone also work."`
	State    string `json:"state,omitempty" jsonschema:"Comma-separated states to include: open, in-progress, closed. Defaults to open and in-progress."`
	Priority string `json:"priority,omitempty" jsonschema:"Comma-separated priorities to include: critical, high, medium, low."`
	Tags     string `json:"tags,omitempty" jsonschema:"Comma-separated tags, for example bug; matches an issue with any of them. Built-in: bug, feature."`
	Owner    string `json:"owner,omitempty" jsonschema:"Only issues with exactly this owner."`
	Parent   string `json:"parent,omitempty" jsonschema:"Only direct sub-issues of this issue."`
	Blocked  bool   `json:"blocked,omitempty" jsonschema:"Only issues with a depends-on issue that is not yet closed."`
	Section  string `json:"section,omitempty" jsonschema:"Restrict the query and the returned content to one section, for example summary."`
	Text     bool   `json:"text,omitempty" jsonschema:"Return human-readable text instead of JSON."`
}

func searchHandler(_ context.Context, _ *mcp.CallToolRequest, in searchArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"search"}
	args = appendFlag(args, "--state", in.State)
	args = appendFlag(args, "--priority", in.Priority)
	args = appendFlag(args, "--tags", in.Tags)
	args = appendFlag(args, "--owner", in.Owner)
	args = appendFlag(args, "--parent", in.Parent)
	args = appendBoolFlag(args, "--blocked", in.Blocked)
	args = appendFlag(args, "--section", in.Section)
	args = appendBoolFlag(args, "--json", !in.Text)
	if in.Query != "" {
		args = append(args, in.Query)
	}
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type graphArgs struct {
	Dir     string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Name    string `json:"name" jsonschema:"The issue name, [[name]], #name, or a path to the issue."`
	JSON    bool   `json:"json,omitempty" jsonschema:"Return the nodes and edges as JSON instead of a drawing."`
	Mermaid bool   `json:"mermaid,omitempty" jsonschema:"Return Mermaid flowchart source instead of a drawing."`
	Full    bool   `json:"full,omitempty" jsonschema:"Draw every linked issue, not only what is below this issue and the path above it."`
}

func graphHandler(_ context.Context, _ *mcp.CallToolRequest, in graphArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"graph", in.Name}
	args = appendBoolFlag(args, "--json", in.JSON)
	args = appendBoolFlag(args, "--mermaid", in.Mermaid)
	args = appendBoolFlag(args, "--full", in.Full)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type parentArgs struct {
	Dir    string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Name   string `json:"name" jsonschema:"The issue name, [[name]], #name, or a path to the issue."`
	Parent string `json:"parent,omitempty" jsonschema:"The parent issue to set. Omit and set clear to remove the current parent instead."`
	Clear  bool   `json:"clear,omitempty" jsonschema:"Remove the issue's current parent instead of setting a new one."`
}

func parentHandler(_ context.Context, _ *mcp.CallToolRequest, in parentArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"parent", in.Name}
	if in.Parent != "" {
		args = append(args, in.Parent)
	}
	args = appendBoolFlag(args, "--clear", in.Clear)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type remoteArgs struct {
	Dir   string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Name  string `json:"name" jsonschema:"The issue name, [[name]], #name, or a path to the issue."`
	URL   string `json:"url,omitempty" jsonschema:"The GitHub issue or Jira ticket URL to set. Omit and set clear to remove the current one instead."`
	Clear bool   `json:"clear,omitempty" jsonschema:"Remove the issue's current remote reference instead of setting a new one."`
}

func remoteHandler(_ context.Context, _ *mcp.CallToolRequest, in remoteArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"remote", in.Name}
	if in.URL != "" {
		args = append(args, in.URL)
	}
	args = appendBoolFlag(args, "--clear", in.Clear)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type dependsOnArgs struct {
	Dir           string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Name          string `json:"name" jsonschema:"The issue name, [[name]], #name, or a path to the issue."`
	BlockingIssue string `json:"blockingIssue,omitempty" jsonschema:"The issue that must close first, to add or (with clear) remove. Omit with clear to remove every dependency."`
	Clear         bool   `json:"clear,omitempty" jsonschema:"Remove blockingIssue (if given) or every dependency (if not) instead of adding one."`
}

func dependsOnHandler(_ context.Context, _ *mcp.CallToolRequest, in dependsOnArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"depends-on", in.Name}
	if in.BlockingIssue != "" {
		args = append(args, in.BlockingIssue)
	}
	args = appendBoolFlag(args, "--clear", in.Clear)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type slugArgs struct {
	Dir  string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Text string `json:"text" jsonschema:"Arbitrary text, for example a title, to turn into a unique issue name."`
}

func slugHandler(_ context.Context, _ *mcp.CallToolRequest, in slugArgs) (*mcp.CallToolResult, any, error) {
	out, err := runSisyphus(in.Dir, "slug", in.Text)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type resolveArgs struct {
	Dir      string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Link     string `json:"link" jsonschema:"A wikilink, issue name, or path to resolve, for example [[some-doc]]."`
	All      bool   `json:"all,omitempty" jsonschema:"Print every match. Without this, an ambiguous link is an error."`
	Absolute bool   `json:"absolute,omitempty" jsonschema:"Print absolute paths instead of paths from the repo root."`
}

func resolveHandler(_ context.Context, _ *mcp.CallToolRequest, in resolveArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"resolve", in.Link}
	args = appendBoolFlag(args, "--all", in.All)
	args = appendBoolFlag(args, "--absolute", in.Absolute)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type linksArgs struct {
	Dir      string `json:"dir,omitempty" jsonschema:"The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory."`
	Document string `json:"document" jsonschema:"The path of the Markdown file to check."`
	Absolute bool   `json:"absolute,omitempty" jsonschema:"Report absolute paths instead of paths from the repo root."`
}

func linksHandler(_ context.Context, _ *mcp.CallToolRequest, in linksArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"links", in.Document, "--json"}
	args = appendBoolFlag(args, "--absolute", in.Absolute)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}

type initArgs struct {
	Dir   string `json:"dir,omitempty" jsonschema:"The repo's root directory to set up. Defaults to sisyphus-mcp's own working directory."`
	Force bool   `json:"force,omitempty" jsonschema:"Replace issues/TEMPLATE.md if it already exists."`
}

func initHandler(_ context.Context, _ *mcp.CallToolRequest, in initArgs) (*mcp.CallToolResult, any, error) {
	args := []string{"init"}
	args = appendBoolFlag(args, "--force", in.Force)
	out, err := runSisyphus(in.Dir, args...)
	if err != nil {
		return nil, nil, err
	}
	return textResult(out)
}
