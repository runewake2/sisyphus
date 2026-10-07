package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

// run is the entry point. findRoot and the writers are parameters, so that tests can run commands in-process.
func run(args []string, findRoot func() (string, error), stdout, stderr io.Writer) int {
	root := newRootCommand(findRoot)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", err)
		return 1
	}
	return 0
}

func newRootCommand(findRoot func() (string, error)) *cobra.Command {
	root := &cobra.Command{
		Use:           "sisyphus",
		Short:         "Manage the issues in issues/ and resolve wikilinks in the repo.",
		Version:       version(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("{{.Version}}\n")
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(
		newCommand(findRoot),
		updateCommand(findRoot),
		parentCommand(findRoot),
		remoteCommand(findRoot),
		dependsOnCommand(findRoot),
		showCommand(findRoot),
		listCommand(findRoot),
		searchCommand(findRoot),
		resolveCommand(findRoot),
		linksCommand(findRoot),
		initCommand(),
	)
	return root
}

func newCommand(findRoot func() (string, error)) *cobra.Command {
	var title, state, resolution, priority, effort, tags, bookmark, deferredFrom, parent string
	var owner, approver, workspace, agentSession, remote, dependsOn string
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create an issue from issues/TEMPLATE.md in the directory of its state.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := firstError(
				oneOf("state", state, states),
				oneOf("priority", priority, priorities),
				oneOf("effort", effort, efforts),
				resolutionFlag(cmd, resolution),
			); err != nil {
				return err
			}
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return newIssue(root, newOptions{
					name:         args[0],
					state:        state,
					resolution:   resolution,
					priority:     priority,
					effort:       effort,
					tags:         tags,
					title:        optional(cmd, "title", title),
					bookmark:     optional(cmd, "bookmark", bookmark),
					deferredFrom: optional(cmd, "deferred-from", deferredFrom),
					parent:       optional(cmd, "parent", parent),
					owner:        optional(cmd, "owner", owner),
					approver:     optional(cmd, "approver", approver),
					workspace:    optional(cmd, "workspace", workspace),
					agentSession: optional(cmd, "agent-session", agentSession),
					remote:       optional(cmd, "remote", remote),
					dependsOn:    optional(cmd, "depends-on", dependsOn),
				}, warnings)
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&title, "title", "t", "", "Issue title. The default is made from the name.")
	flags.StringVarP(&state, "state", "s", "open", "The state of the new issue: "+strings.Join(states, ", ")+".")
	flags.StringVarP(&resolution, "resolution", "r", "", "When the state is closed: "+strings.Join(resolutions, ", ")+". Defaults to completed.")
	flags.StringVarP(&priority, "priority", "p", "medium", "The priority of the issue: "+strings.Join(priorities, ", ")+".")
	flags.StringVarP(&effort, "effort", "e", "medium", "The effort of the issue: "+strings.Join(efforts, ", ")+".")
	flags.StringVar(&tags, "tags", "", `Comma-separated component names, for example "scheduler,plan".`)
	flags.StringVarP(&bookmark, "bookmark", "b", "", "The jj bookmark of the work on the issue.")
	flags.StringVarP(&deferredFrom, "deferred-from", "d", "", "Issue name, wikilink, or bookmark that deferred this work.")
	flags.StringVar(&parent, "parent", "", "The parent issue, if the new issue is a sub-issue. The parent must exist.")
	flags.StringVar(&owner, "owner", "", "The person or agent working on the issue.")
	flags.StringVar(&approver, "approver", "", "The person or agent who accepts the issue when it closes.")
	flags.StringVar(&workspace, "workspace", "", "The jj workspace where local work on the issue is happening. Appended to the issue's workspace history.")
	flags.StringVar(&agentSession, "agent-session", "", "The AI agent session id working on the issue, if available.")
	flags.StringVar(&remote, "remote", "", "A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.")
	flags.StringVar(&dependsOn, "depends-on", "", "An issue that must close before this one can start. The issue must exist.")
	return cmd
}

func updateCommand(findRoot func() (string, error)) *cobra.Command {
	var resolution, bookmark, owner, approver, workspace, agentSession, priority, effort, tags string
	cmd := &cobra.Command{
		Use:   "update <name> <state>",
		Short: "Change the state of an issue and move it to the directory of the new state.",
		Long: "Change the state of an issue and move it to the directory of the new state.\n\n" +
			"<name> is an issue name, [[issue-name]], #issue-name, or a path to the issue. " +
			"<state> is " + strings.Join(states, ", ") + ".",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := firstError(oneOf("state", args[1], states), resolutionFlag(cmd, resolution)); err != nil {
				return err
			}
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return updateIssue(root, args[0], args[1], updateOptions{
					resolution:   resolution,
					bookmark:     optional(cmd, "bookmark", bookmark),
					owner:        optional(cmd, "owner", owner),
					approver:     optional(cmd, "approver", approver),
					workspace:    optional(cmd, "workspace", workspace),
					agentSession: optional(cmd, "agent-session", agentSession),
					priority:     optional(cmd, "priority", priority),
					effort:       optional(cmd, "effort", effort),
					tags:         optional(cmd, "tags", tags),
				}, warnings)
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&resolution, "resolution", "r", "", "When the new state is closed: "+strings.Join(resolutions, ", ")+". Defaults to completed.")
	flags.StringVarP(&bookmark, "bookmark", "b", "", "The jj bookmark of the work. Set it when work starts.")
	flags.StringVar(&owner, "owner", "", "The person or agent working on the issue.")
	flags.StringVar(&approver, "approver", "", "The person or agent who accepts the issue when it closes.")
	flags.StringVar(&workspace, "workspace", "", "The jj workspace where local work on the issue is happening. Appended to the issue's workspace history.")
	flags.StringVar(&agentSession, "agent-session", "", "The AI agent session id working on the issue, if available.")
	flags.StringVar(&priority, "priority", "", "Change the priority: "+strings.Join(priorities, ", ")+".")
	flags.StringVar(&effort, "effort", "", "Change the effort: "+strings.Join(efforts, ", ")+".")
	flags.StringVar(&tags, "tags", "", `Replace the tags, comma-separated, for example "scheduler,plan".`)
	return cmd
}

func parentCommand(findRoot func() (string, error)) *cobra.Command {
	var clear bool
	cmd := &cobra.Command{
		Use:   "parent <name> [<parent>]",
		Short: "Set or remove the parent of an issue. The parent must exist and cannot be a sub-issue of the issue.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 2) == clear {
				return errors.New("Give a parent issue or --clear, but not both.")
			}
			var parent *string
			if len(args) == 2 {
				parent = &args[1]
			}
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return setParent(root, args[0], parent, warnings)
			})
		},
	}
	cmd.Flags().BoolVar(&clear, "clear", false, "Remove the parent of the issue.")
	return cmd
}

func remoteCommand(findRoot func() (string, error)) *cobra.Command {
	var clear bool
	cmd := &cobra.Command{
		Use:   "remote <name> [<url>]",
		Short: "Set or remove the remote reference (a GitHub issue or Jira ticket URL) of an issue.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (len(args) == 2) == clear {
				return errors.New("Give a remote URL or --clear, but not both.")
			}
			var remoteURL *string
			if len(args) == 2 {
				remoteURL = &args[1]
			}
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return setRemote(root, args[0], remoteURL, warnings)
			})
		},
	}
	cmd.Flags().BoolVar(&clear, "clear", false, "Remove the remote reference of the issue.")
	return cmd
}

func dependsOnCommand(findRoot func() (string, error)) *cobra.Command {
	var clear bool
	cmd := &cobra.Command{
		Use:   "depends-on <name> [<blocking-issue>]",
		Short: "Add or remove an issue that must close before <name> can start.",
		Long: "Add or remove an issue that must close before <name> can start.\n\n" +
			"With a <blocking-issue> and no --clear, adds it (the blocking issue must exist, and the " +
			"dependency cannot be or create a cycle). With a <blocking-issue> and --clear, removes just " +
			"that one. With --clear alone, removes every dependency of <name>.",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 && !clear {
				return errors.New("Give a blocking issue to add, or --clear to remove one or all.")
			}
			var blocking *string
			if len(args) == 2 {
				blocking = &args[1]
			}
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return setDependsOn(root, args[0], blocking, clear, warnings)
			})
		},
	}
	cmd.Flags().BoolVar(&clear, "clear", false, "Remove one dependency (with a <blocking-issue>) or every dependency (without one).")
	return cmd
}

func showCommand(findRoot func() (string, error)) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Print one issue: its key fields and its body.",
		Long: "Print one issue: its key fields and its body, without first finding which state directory holds it.\n\n" +
			"<name> is an issue name, [[issue-name]], #issue-name, or a path to the issue.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			name, _, doc, err := loadIssue(root, args[0])
			if err != nil {
				return err
			}
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(newIssueView(name, doc))
			}
			writeIssueText(cmd.OutOrStdout(), doc)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the issue as JSON instead of readable text.")
	return cmd
}

func listCommand(findRoot func() (string, error)) *cobra.Command {
	var state, priority, tags, owner, parent string
	var asJSON, blocked bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List and filter the issues in issues/.",
		Long: "List and filter the issues in issues/, sorted by state, then priority, then name.\n\n" +
			"Without --state, only open and in-progress issues are listed. --state and --priority accept a " +
			"comma-separated list and match any of the given values. --tags matches an issue that has any of " +
			"the given tags. Every given flag narrows the list together (AND).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stateValues := parseList(state)
			priorityValues := parseList(priority)
			if err := firstError(
				validateEach("state", stateValues, states),
				validateEach("priority", priorityValues, priorities),
			); err != nil {
				return err
			}
			root, err := findRoot()
			if err != nil {
				return err
			}
			rows := listIssues(root, listOptions{
				states:      stateValues,
				priorities:  priorityValues,
				tags:        parseList(tags),
				owner:       owner,
				parent:      parent,
				blockedOnly: blocked,
			})
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(rows)
			}
			cells := make([][]string, len(rows))
			for i, row := range rows {
				cells[i] = []string{row.Name, row.Title, row.State, row.Priority, row.Owner, strings.Join(row.Tags, ", ")}
			}
			writeTable(cmd.OutOrStdout(), []string{"name", "title", "state", "priority", "owner", "tags"}, cells)
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&state, "state", "s", "", "Comma-separated states to include: "+strings.Join(states, ", ")+". Default: open, in-progress.")
	flags.StringVarP(&priority, "priority", "p", "", "Comma-separated priorities to include: "+strings.Join(priorities, ", ")+".")
	flags.StringVar(&tags, "tags", "", "Comma-separated tags; matches an issue with any of them.")
	flags.StringVar(&owner, "owner", "", "Only issues with exactly this owner.")
	flags.StringVar(&parent, "parent", "", "Only direct sub-issues of this issue.")
	flags.BoolVar(&blocked, "blocked", false, "Only issues with a depends-on issue that is not yet closed.")
	flags.BoolVar(&asJSON, "json", false, "Print JSON instead of a table.")
	return cmd
}

func searchCommand(findRoot func() (string, error)) *cobra.Command {
	var state, priority, tags, owner, parent, section string
	var asJSON, blocked bool
	cmd := &cobra.Command{
		Use:   "search [<query>]",
		Short: "Search issues by frontmatter filters and content.",
		Long: "Search issues by frontmatter filters and content, sorted by state, then priority, then name.\n\n" +
			"<query> matches case-insensitively against the title and body. The frontmatter filters are " +
			"the same as sisyphus list's, and combine with the query and with each other using AND. " +
			"--section restricts the query, when one is given, and the content returned for each match, " +
			"to one section of the body, matched the same way a wikilink's #heading is.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var query string
			if len(args) == 1 {
				query = args[0]
			}
			stateValues := parseList(state)
			priorityValues := parseList(priority)
			if err := firstError(
				validateEach("state", stateValues, states),
				validateEach("priority", priorityValues, priorities),
			); err != nil {
				return err
			}
			root, err := findRoot()
			if err != nil {
				return err
			}
			rows := searchIssues(root, searchOptions{
				filters: listOptions{
					states:      stateValues,
					priorities:  priorityValues,
					tags:        parseList(tags),
					owner:       owner,
					parent:      parent,
					blockedOnly: blocked,
				},
				query:   query,
				section: section,
			})
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(rows)
			}
			for i, row := range rows {
				if i > 0 {
					fmt.Fprintln(cmd.OutOrStdout())
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", row.Name, row.Title)
				fmt.Fprintln(cmd.OutOrStdout(), row.Content)
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&state, "state", "s", "", "Comma-separated states to include: "+strings.Join(states, ", ")+". Default: open, in-progress.")
	flags.StringVarP(&priority, "priority", "p", "", "Comma-separated priorities to include: "+strings.Join(priorities, ", ")+".")
	flags.StringVar(&tags, "tags", "", "Comma-separated tags; matches an issue with any of them.")
	flags.StringVar(&owner, "owner", "", "Only issues with exactly this owner.")
	flags.StringVar(&parent, "parent", "", "Only direct sub-issues of this issue.")
	flags.StringVar(&section, "section", "", `Restrict the query and the returned content to one section, for example "summary".`)
	flags.BoolVar(&blocked, "blocked", false, "Only issues with a depends-on issue that is not yet closed.")
	flags.BoolVar(&asJSON, "json", false, "Print JSON instead of readable text.")
	return cmd
}

func resolveCommand(findRoot func() (string, error)) *cobra.Command {
	var all, absolute bool
	cmd := &cobra.Command{
		Use:   "resolve <link>",
		Short: "Find the file that a wikilink points to, in any directory of the repo.",
		Long: "Find the file that a wikilink points to, in any directory of the repo. A link matches a file by name, " +
			`with or without ".md", in any directory. A link that contains "/" matches the end of the path. ` +
			"An exact path from the repo root wins over other matches.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			paths, err := resolveLink(root, args[0], all, absolute)
			if err != nil {
				return err
			}
			for _, path := range paths {
				fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Print every match. Without this, an ambiguous link is an error.")
	cmd.Flags().BoolVar(&absolute, "absolute", false, "Print absolute paths. The default is paths from the repo root.")
	return cmd
}

func linksCommand(findRoot func() (string, error)) *cobra.Command {
	var absolute, asJSON bool
	cmd := &cobra.Command{
		Use:   "links <document>",
		Short: "List every wikilink in a document and the file that each link resolves to.",
		Long: "List every wikilink in a document and the file that each link resolves to. The status of a link is ok, " +
			"missing, ambiguous, or missing-heading. Links in fenced code blocks and in inline code are ignored, " +
			"because Obsidian ignores them.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			file, rows, err := listLinks(root, args[0], absolute)
			if err != nil {
				return err
			}
			switch {
			case asJSON:
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(rows)
			case len(rows) == 0:
				fmt.Fprintf(cmd.ErrOrStderr(), "No wikilinks in %s.\n", file)
			default:
				cells := make([][]string, len(rows))
				for i, row := range rows {
					cells[i] = []string{strconv.Itoa(row.Line), row.Link, row.Status, row.Resolved}
				}
				writeTable(cmd.OutOrStdout(), []string{"line", "link", "status", "resolved"}, cells)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&absolute, "absolute", false, "Show absolute paths. The default is paths from the repo root.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print JSON instead of a table.")
	return cmd
}

func initCommand() *cobra.Command {
	var dir, project, bookmarkPrefix string
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write the workflow into a repo: CONTRIBUTING.md, AGENTS.md, issues/, the changelog, and VERSION.",
		Long: "Write the workflow into a repo: CONTRIBUTING.md, AGENTS.md, issues/, design/decisions/, the changelog, " +
			"and VERSION. Files that exist stay as they are, unless --force is given. After init, the other commands " +
			"of sisyphus work in the repo.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			if project == "" {
				project = filepath.Base(root)
			}
			values := kitValues{Project: project, BookmarkPrefix: strings.TrimSuffix(bookmarkPrefix, "/"), Date: today()}
			return initRepo(root, values, force, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".", "The root directory of the repo.")
	cmd.Flags().StringVar(&project, "project", "", "The name of the project. The default is the name of the directory.")
	cmd.Flags().StringVar(&bookmarkPrefix, "bookmark-prefix", "ai", `The prefix of the jj bookmarks of agents, for example "samw/ai".`)
	cmd.Flags().BoolVar(&force, "force", false, "Replace files that exist.")
	return cmd
}

// printPath finds the repo root, runs the operation, and prints the path of the file that the operation wrote.
func printPath(cmd *cobra.Command, findRoot func() (string, error), operation func(root string, warnings io.Writer) (string, error)) error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	path, err := operation(root, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), path)
	return nil
}

// writeIssueText prints an issue's key fields, then its body.
func writeIssueText(out io.Writer, doc *document) {
	fields := [][2]string{
		{"title", doc.get("title")},
		{"state", doc.get("state")},
		{"priority", doc.get("priority")},
		{"effort", doc.get("effort")},
		{"tags", doc.get("tags")},
		{"owner", doc.get("owner")},
		{"approver", doc.get("approver")},
		{"bookmark", doc.get("bookmark")},
		{"parent", doc.get("parent")},
	}
	width := 0
	for _, f := range fields {
		width = max(width, utf8.RuneCountInString(f[0]))
	}
	for _, f := range fields {
		label, value := f[0], f[1]
		fmt.Fprintf(out, "%s:%s %s\n", label, strings.Repeat(" ", width-utf8.RuneCountInString(label)), value)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, strings.TrimLeft(strings.Join(doc.body, "\n"), "\n"))
}

// writeTable prints cells as a table with header, padding every column to its widest cell.
func writeTable(out io.Writer, header []string, cells [][]string) {
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range cells {
		for i, cell := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}
	format := func(values []string) string {
		padded := make([]string, len(values))
		for i, v := range values {
			padded[i] = v + strings.Repeat(" ", widths[i]-utf8.RuneCountInString(v))
		}
		return strings.TrimRight(strings.Join(padded, "  "), " ")
	}
	dashes := make([]string, len(widths))
	for i, w := range widths {
		dashes[i] = strings.Repeat("-", w)
	}
	fmt.Fprintln(out, format(header))
	fmt.Fprintln(out, format(dashes))
	for _, row := range cells {
		fmt.Fprintln(out, format(row))
	}
}

func optional(cmd *cobra.Command, name, value string) *string {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	return &value
}

// resolutionFlag checks --resolution only when it is given, because its default is empty.
func resolutionFlag(cmd *cobra.Command, resolution string) error {
	if !cmd.Flags().Changed("resolution") {
		return nil
	}
	return oneOf("resolution", resolution, resolutions)
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func oneOf(what, value string, choices []string) error {
	if slices.Contains(choices, value) {
		return nil
	}
	return fmt.Errorf("Invalid %s '%s'. Use one of: %s.", what, value, strings.Join(choices, ", "))
}

func validateEach(what string, values, choices []string) error {
	for _, value := range values {
		if err := oneOf(what, value, choices); err != nil {
			return err
		}
	}
	return nil
}
