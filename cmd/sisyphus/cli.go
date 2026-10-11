package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/runewake2/sisyphus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

func newRootCommand(defaultFindRoot func() (string, error)) *cobra.Command {
	// -C makes every command start its search for the repo root in another directory, as in git.
	var directory string
	findRoot := func() (string, error) {
		if directory == "" {
			return defaultFindRoot()
		}
		start, err := filepath.Abs(directory)
		if err != nil {
			return "", err
		}
		if info, err := os.Stat(start); err != nil || !info.IsDir() {
			return "", fmt.Errorf("-C %s is not a directory.", directory)
		}
		return findRootFrom(start)
	}
	root := &cobra.Command{
		Use:   "sisyphus",
		Short: "Track issues as Markdown files in your repo.",
		Long: "sisyphus tracks the issues of a project as Markdown files in the repo, next to the code. Each " +
			"issue is one file with YAML frontmatter in issues/open/, issues/in-progress/, or issues/closed/, " +
			"and the directory always matches its state. Wikilinks such as [[fix-login-bug]] link issues and " +
			"documents.\n\n" +
			"To start, run sisyphus init, then sisyphus new --title \"<title>\".",
		Version:       sisyphus.Version(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVarP(&directory, "directory", "C", "", "Run as if started in this directory, as git -C does.")
	root.SetVersionTemplate("{{.Version}}\n")
	// sisyphus completion replaces the default command of cobra, to add Nushell.
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(
		addJSONFlag(newCommand(findRoot)),
		addJSONFlag(updateCommand(findRoot)),
		addJSONFlag(parentCommand(findRoot)),
		addJSONFlag(remoteCommand(findRoot)),
		addJSONFlag(dependsOnCommand(findRoot)),
		addJSONFlag(editCommand(findRoot)),
		addJSONFlag(startCommand(findRoot)),
		addJSONFlag(closeCommand(findRoot, "close", "completed")),
		addJSONFlag(closeCommand(findRoot, "abandon", "abandoned")),
		slugCommand(findRoot),
		showCommand(findRoot),
		graphCommand(findRoot),
		listCommand(findRoot),
		claimsCommand(findRoot),
		searchCommand(findRoot),
		resolveCommand(findRoot),
		linksCommand(findRoot),
		lintCommand(findRoot),
		initCommand(&directory),
		completionCommand(),
	)
	addHelp(root)
	addCompletions(root, findRoot)
	return root
}

func newCommand(findRoot func() (string, error)) *cobra.Command {
	var title, state, resolution, priority, effort, tags, bookmark, deferredFrom, parent string
	var owner, approver, workspace, remote, context string
	var dependsOn []string
	var metadata map[string]string
	cmd := &cobra.Command{
		Use:   "new [<name>]",
		Short: "Create an issue from issues/TEMPLATE.md.",
		Long: "Create an issue from issues/TEMPLATE.md in the directory of its state.\n\n" +
			"<name> is 2-6 lowercase words in kebab-case, for example fix-login-bug, optionally below " +
			"kebab-case directories, for example web/auth/fix-login-bug. The directories stay with the " +
			"issue when its state changes. Without <name>, the name is made from --title, the way " +
			"sisyphus slug makes it.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			} else if !cmd.Flags().Changed("title") {
				return errors.New("Give a name, or a --title to make the name from, for example: sisyphus new --title \"Fix the login bug\"")
			}
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
					name:         name,
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
					remote:       optional(cmd, "remote", remote),
					dependsOn:    dependsOn,
					context:      optional(cmd, "context", context),
					metadata:     metadata,
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
	flags.StringVar(&tags, "tags", "", `Comma-separated tags, for example "bug,scheduler". Built-in: `+strings.Join(builtinTags, ", ")+`; any other tag works too.`)
	flags.StringVarP(&bookmark, "bookmark", "b", "", "The jj bookmark or git branch of the work on the issue.")
	flags.StringVarP(&deferredFrom, "deferred-from", "d", "", "Issue name, wikilink, or bookmark that deferred this work.")
	flags.StringVar(&parent, "parent", "", "The parent issue, if the new issue is a sub-issue. The parent must exist.")
	flags.StringVar(&owner, "owner", "", "The person or agent working on the issue.")
	flags.StringVar(&approver, "approver", "", "The person or agent who accepts the issue when it closes.")
	flags.StringVar(&workspace, "workspace", "", "The jj workspace or git worktree where local work on the issue happens. Appended to the issue's workspace history.")
	flags.StringVar(&remote, "remote", "", "A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.")
	flags.StringSliceVar(&dependsOn, "depends-on", nil, "Issues that must close before this one can start, comma-separated or repeated. Each issue must exist.")
	flags.StringVar(&context, "context", "", "Replaces the Context section's placeholder with this text.")
	flags.StringToStringVar(&metadata, "metadata", nil, `Arbitrary key=value notes, for example "session-id=abc123". Repeatable, or comma-separated.`)
	return cmd
}

func updateCommand(findRoot func() (string, error)) *cobra.Command {
	var resolution, bookmark, owner, approver, workspace, priority, effort, tags, title, addTags, removeTags, message string
	var reopen, take bool
	var metadata map[string]string
	cmd := &cobra.Command{
		Use:   "update <name> [<state>]",
		Short: "Change the state or the fields of an issue.",
		Long: "Change the state of an issue and move it to the directory of the new state, or change its fields.\n\n" +
			"<name> is an issue's full name, its file name, [[issue-name]], #issue-name, or a path to the issue. " +
			"A file name that more than one issue has is refused. <state> is " + strings.Join(states, ", ") + ". " +
			"Without <state>, the state does not change, and at least one flag must be given.",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			state := ""
			if len(args) == 2 {
				state = args[1]
				if err := oneOf("state", state, states); err != nil {
					return err
				}
			} else if cmd.Flags().NFlag() == 0 {
				return errors.New("Give a new state, or at least one flag to change, for example --priority high.")
			}
			if err := resolutionFlag(cmd, resolution); err != nil {
				return err
			}
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return updateIssue(root, args[0], state, updateOptions{
					resolution: resolution,
					bookmark:   optional(cmd, "bookmark", bookmark),
					owner:      optional(cmd, "owner", owner),
					approver:   optional(cmd, "approver", approver),
					workspace:  optional(cmd, "workspace", workspace),
					priority:   optional(cmd, "priority", priority),
					effort:     optional(cmd, "effort", effort),
					tags:       optional(cmd, "tags", tags),
					title:      optional(cmd, "title", title),
					addTags:    addTags,
					removeTags: removeTags,
					reopen:     reopen,
					take:       take,
					message:    optional(cmd, "message", message),
					metadata:   metadata,
				}, warnings)
			})
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&resolution, "resolution", "r", "", "When the new state is closed: "+strings.Join(resolutions, ", ")+". Defaults to completed.")
	flags.StringVarP(&bookmark, "bookmark", "b", "", "The jj bookmark or git branch of the work. Set it when work starts.")
	flags.StringVar(&owner, "owner", "", "The person or agent working on the issue.")
	flags.StringVar(&approver, "approver", "", "The person or agent who accepts the issue when it closes.")
	flags.StringVar(&workspace, "workspace", "", "The jj workspace or git worktree where local work on the issue happens. Appended to the issue's workspace history.")
	flags.StringVar(&priority, "priority", "", "Change the priority: "+strings.Join(priorities, ", ")+".")
	flags.StringVar(&effort, "effort", "", "Change the effort: "+strings.Join(efforts, ", ")+".")
	flags.StringVarP(&title, "title", "t", "", "Change the title, in the title field and in the first # heading of the body.")
	flags.StringVar(&tags, "tags", "", `Replace the tags, comma-separated, for example "bug,scheduler". Built-in: `+strings.Join(builtinTags, ", ")+`; any other tag works too.`)
	flags.StringVarP(&message, "message", "m", "", "The text of the Resolution section, for an issue that closes: what was done, or why it stopped.")
	flags.BoolVar(&reopen, "reopen", false, "Reopen an issue whose close was a mistake: move a closed issue back to open or in-progress, clear its resolution and closed date, and note it in Notes. For work that needs more after a real close, create a new issue instead.")
	flags.BoolVar(&take, "take", false, takeHelp)
	flags.StringVar(&addTags, "add-tags", "", `Add tags, comma-separated, and keep the others, for example "bug".`)
	flags.StringVar(&removeTags, "remove-tags", "", `Remove tags, comma-separated, and keep the others. A tag that is not there is not an error.`)
	flags.StringToStringVar(&metadata, "metadata", nil, `Arbitrary key=value notes to add or update, for example "session-id=abc123". Repeatable, or comma-separated. Never cleared automatically.`)
	cmd.MarkFlagsMutuallyExclusive("tags", "add-tags")
	cmd.MarkFlagsMutuallyExclusive("tags", "remove-tags")
	return cmd
}

func parentCommand(findRoot func() (string, error)) *cobra.Command {
	var clear bool
	cmd := &cobra.Command{
		Use:   "parent <name> [<parent>]",
		Short: "Set or remove the parent of an issue.",
		Long:  "Set or remove the parent of an issue. The parent must exist and cannot be a sub-issue of the issue.",
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
		Short: "Set or remove the URL of an external ticket for an issue.",
		Long:  "Set or remove the remote reference of an issue: the URL of the GitHub issue or Jira ticket that tracks it outside the repo.",
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
		Short: "Add or remove an issue that must close before an issue can start.",
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

func editCommand(findRoot func() (string, error)) *cobra.Command {
	var appendText bool
	cmd := &cobra.Command{
		Use:   "edit <name> [<section> <text>]",
		Short: "Edit an issue in your editor, or replace one section of it.",
		Long: "With only <name>, open the file of the issue in your editor: $VISUAL, or else $EDITOR. The value " +
			"can have arguments, for example \"code --wait\". When the editor exits, sisyphus checks the file as " +
			"lint does, and warns about problems, for example a frontmatter line that it cannot read.\n\n" +
			"With <section> and <text>, replace the text of one section of the issue's body, for example " +
			"Summary or Resolution, or append to it with --append.\n\n" +
			"<name> is an issue name, [[issue-name]], #issue-name, or a path to the issue. <section> is " +
			"the heading of the section, matched case-insensitively, for example \"resolution\". " +
			"<text> is the new text; \"-\" reads it from standard input. The heading line and every other " +
			"section stay as they are. An append to a section that holds only the template placeholder " +
			"replaces the placeholder.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 && len(args) != 3 {
				return fmt.Errorf("Give <name> to open the issue in your editor, or <name> <section> <text> to replace a section. Got %d arguments.", len(args))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				for _, flag := range []string{"append", "json"} {
					if cmd.Flags().Changed(flag) {
						return fmt.Errorf("--%s needs <section> and <text>. With only <name>, edit opens your editor.", flag)
					}
				}
				return editInEditor(cmd, findRoot, args[0])
			}
			text := args[2]
			if text == "-" {
				input, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				text = string(input)
			}
			return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
				return editSection(root, args[0], args[1], text, appendText)
			})
		},
	}
	cmd.Flags().BoolVarP(&appendText, "append", "a", false, "Add the text at the end of the section, after a blank line, instead of replacing the section.")
	return cmd
}

func slugCommand(findRoot func() (string, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slug <text>",
		Short: "Print a valid issue name, made from text, that is not in use.",
		Long: "Print a valid issue name, made from text such as a GitHub issue title, that no Markdown file " +
			"in the repo has. It does not create anything.\n\n" +
			"1. Lowercase the text and spell its letters in ASCII: \"é\" becomes \"e\", \"ß\" becomes \"ss\", " +
			"and \"æ\" becomes \"ae\". The name is always ASCII.\n" +
			"2. Each run of ASCII letters and digits is a word. Every other character, for example a " +
			"Cyrillic letter or an emoji, separates words and is dropped.\n" +
			"3. Drop the filler words " + fillerWordList() + ", unless fewer than 2 words would remain.\n" +
			"4. Keep the first 6 words and drop the rest. Text with fewer than 2 words is an error.\n" +
			"5. Join the words with hyphens.\n" +
			"6. If a Markdown file in the repo has that name, ignoring case, add -2, -3, and so on until no " +
			"file has it. A 6-word name drops its last word to make room for the number.\n\n" +
			"Example: \"Build the issue index once per command\" gives build-issue-index-once-per-command.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			name, err := uniqueSlug(root, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), name)
			return nil
		},
	}
	return cmd
}

func showCommand(findRoot func() (string, error)) *cobra.Command {
	var asJSON, all bool
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Print one issue: its key fields and its body.",
		Long: "Print one issue: its key fields and its body, without first finding which state directory holds it.\n\n" +
			"<name> is an issue's full name, its file name, [[issue-name]], #issue-name, or a path to the issue. " +
			"A file name that more than one issue has is refused.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			x := loadIndex(root)
			name, issue, doc, err := x.load(args[0])
			if err != nil {
				return err
			}
			warnIssueProblems(x, issue, doc, nil, cmd.ErrOrStderr())
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(newIssueView(name, doc))
			}
			writeIssueText(cmd.OutOrStdout(), doc, all)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the issue as JSON instead of readable text.")
	cmd.Flags().BoolVar(&all, "all", false, "Print every field and section, also the empty fields and the sections that hold only the template text.")
	return cmd
}

func graphCommand(findRoot func() (string, error)) *cobra.Command {
	var asJSON, asMermaid, full bool
	maxIssues := defaultMaxDrawnIssues
	color := "auto"
	cmd := &cobra.Command{
		Use:   "graph <name>",
		Short: "Draw an issue, everything below it, and the path above it.",
		Long: "Draw <name>, everything below it, and the path above it as boxes and arrows in the " +
			"terminal. Arrows point from the issue that comes first to the one that comes after: solid " +
			"arrows from a parent to its sub-issues, dotted arrows from an issue to the issues that " +
			"depend on it. Below <name> are its sub-issues and dependents, all the way down; above it " +
			"are its parents and dependencies, all the way up. A note tells how many other linked " +
			"issues are not drawn; --full draws them too. Each box shows the issue's file name, with its state on the line below, " +
			"inside a rounded box for each directory above it. <name> has a double box (╔═╗), and each " +
			"available issue has a heavy box (┏━┓): an issue that is not closed and has no open dependency or " +
			"sub-issue, so work on it can start now. With --full, an issue that is not below <name> " +
			"or on its path above has a dashed box (┌╌┐, or ┏╍┓ if it is available). Heavy arrows (━━►) " +
			"trace the work path: what <name> waits on, down to the available issues to start now, " +
			"which the legend names.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			x := loadIndex(root)
			issue, err := x.find(issueName(args[0]))
			if err != nil {
				return err
			}
			graph := buildGraph(x, issue.name, full)
			switch {
			case asJSON:
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(graph)
			case asMermaid:
				source := mermaidSource(graph)
				if note := hiddenNote(graph); note != "" {
					source += "    %% " + note + "\n"
				}
				_, err := io.WriteString(cmd.OutOrStdout(), source)
				return err
			}
			if n := len(graph.Nodes); maxIssues > 0 && n > maxIssues {
				return fmt.Errorf("The drawing would have %d issues, more than --max-issues (%d), and could not be read. "+
					"Use --json or --mermaid, which have no limit, draw a narrower issue, or raise --max-issues (0 means no limit).", n, maxIssues)
			}
			colored, err := useColor(color, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			_, err = io.WriteString(cmd.OutOrStdout(), drawGraph(graph, colored))
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the nodes and edges as JSON instead of a drawing.")
	cmd.Flags().BoolVar(&asMermaid, "mermaid", false, "Print Mermaid flowchart source instead of a drawing, for example to paste into a Markdown file.")
	cmd.Flags().BoolVar(&full, "full", false, "Draw every issue linked to <name> by parent or depends-on, not only what is below it and the path above it.")
	cmd.Flags().StringVar(&color, "color", color, "Color the drawing: auto (on a terminal, unless NO_COLOR is set), always, or never.")
	cmd.Flags().IntVar(&maxIssues, "max-issues", maxIssues, "Do not draw more issues than this; use --json or --mermaid for a larger graph. 0 means no limit.")
	cmd.MarkFlagsMutuallyExclusive("json", "mermaid")
	return cmd
}

// defaultMaxDrawnIssues is the largest number of issues that graph draws by default. A drawing of
// thousands of boxes is megabytes of text that nobody can read, and takes a long time to make.
const defaultMaxDrawnIssues = 200

// useColor decides from the --color value whether a drawing written to out has color. auto means
// color only on a terminal, and never when NO_COLOR is set (see https://no-color.org) or TERM is dumb.
func useColor(mode string, out io.Writer) (bool, error) {
	switch mode {
	case "always":
		return true, nil
	case "never":
		return false, nil
	case "auto":
		if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
			return false, nil
		}
		f, isFile := out.(*os.File)
		if !isFile {
			return false, nil
		}
		info, err := f.Stat()
		return err == nil && info.Mode()&os.ModeCharDevice != 0, nil
	}
	return false, fmt.Errorf("--color must be auto, always, or never, not %q", mode)
}

// readyHelp is the help text of --ready in list and search.
const readyHelp = "Only issues that can start now: not closed, with no depends-on issue and no sub-issue that is not closed. These are the available issues of sisyphus graph."

func listCommand(findRoot func() (string, error)) *cobra.Command {
	var state, priority, tags, owner, parent string
	var asJSON, blocked, ready, wide, mine bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List and filter issues.",
		Long: "List and filter the issues in issues/, sorted by state, then priority, then name.\n\n" +
			"Without --state, only open and in-progress issues are listed. --state and --priority accept a " +
			"comma-separated list and match any of the given values. --tags matches an issue that has any of " +
			"the given tags. Every given flag narrows the list together (AND).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stateValues := stateFilter(state)
			priorityValues := parseList(priority)
			if err := firstError(
				validateEach("state", stateValues, states),
				validateEach("priority", priorityValues, priorities),
			); err != nil {
				return err
			}
			owner, err := ownerFilter(owner, mine)
			if err != nil {
				return err
			}
			root, err := findRoot()
			if err != nil {
				return err
			}
			rows, err := listIssues(root, listOptions{
				states:      stateValues,
				priorities:  priorityValues,
				tags:        parseList(tags),
				owner:       owner,
				parent:      parent,
				blockedOnly: blocked,
				readyOnly:   ready,
				warnings:    cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}
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
			maxWidth := 0
			if width, isTerminal := terminalWidth(cmd.OutOrStdout()); isTerminal && !wide {
				maxWidth = width
			}
			writeTableWithin(cmd.OutOrStdout(), []string{"name", "title", "state", "priority", "owner", "tags"}, cells, 1, maxWidth)
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&state, "state", "s", "", "Comma-separated states to include: "+strings.Join(states, ", ")+", or all. Default: open, in-progress.")
	flags.StringVarP(&priority, "priority", "p", "", "Comma-separated priorities to include: "+strings.Join(priorities, ", ")+".")
	flags.StringVar(&tags, "tags", "", "Comma-separated tags, for example \"bug\"; matches an issue with any of them. Built-in: "+strings.Join(builtinTags, ", ")+".")
	flags.StringVar(&owner, "owner", "", "Only issues with exactly this owner.")
	flags.StringVar(&parent, "parent", "", "Only direct sub-issues of this issue.")
	flags.BoolVar(&mine, "mine", false, "Only issues whose owner is the value of "+ownerVariable+".")
	flags.BoolVar(&blocked, "blocked", false, "Only issues with a depends-on issue that is not yet closed.")
	flags.BoolVar(&ready, "ready", false, readyHelp)
	flags.BoolVar(&wide, "wide", false, "Do not shorten titles to fit the terminal.")
	flags.BoolVar(&asJSON, "json", false, "Print JSON instead of a table.")
	cmd.MarkFlagsMutuallyExclusive("blocked", "ready")
	cmd.MarkFlagsMutuallyExclusive("owner", "mine")
	return cmd
}

func searchCommand(findRoot func() (string, error)) *cobra.Command {
	var state, priority, tags, owner, parent, section string
	var asJSON, blocked, ready, mine bool
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
			stateValues := stateFilter(state)
			priorityValues := parseList(priority)
			if err := firstError(
				validateEach("state", stateValues, states),
				validateEach("priority", priorityValues, priorities),
			); err != nil {
				return err
			}
			owner, err := ownerFilter(owner, mine)
			if err != nil {
				return err
			}
			root, err := findRoot()
			if err != nil {
				return err
			}
			rows, err := searchIssues(root, searchOptions{
				filters: listOptions{
					states:      stateValues,
					priorities:  priorityValues,
					tags:        parseList(tags),
					owner:       owner,
					parent:      parent,
					blockedOnly: blocked,
					readyOnly:   ready,
					warnings:    cmd.ErrOrStderr(),
				},
				query:   query,
				section: section,
			})
			if err != nil {
				return err
			}
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
	flags.StringVarP(&state, "state", "s", "", "Comma-separated states to include: "+strings.Join(states, ", ")+", or all. Default: open, in-progress.")
	flags.StringVarP(&priority, "priority", "p", "", "Comma-separated priorities to include: "+strings.Join(priorities, ", ")+".")
	flags.StringVar(&tags, "tags", "", "Comma-separated tags, for example \"bug\"; matches an issue with any of them. Built-in: "+strings.Join(builtinTags, ", ")+".")
	flags.StringVar(&owner, "owner", "", "Only issues with exactly this owner.")
	flags.StringVar(&parent, "parent", "", "Only direct sub-issues of this issue.")
	flags.BoolVar(&mine, "mine", false, "Only issues whose owner is the value of "+ownerVariable+".")
	flags.StringVar(&section, "section", "", `Restrict the query and the returned content to one section, for example "summary".`)
	flags.BoolVar(&blocked, "blocked", false, "Only issues with a depends-on issue that is not yet closed.")
	flags.BoolVar(&ready, "ready", false, readyHelp)
	flags.BoolVar(&asJSON, "json", false, "Print JSON instead of readable text.")
	cmd.MarkFlagsMutuallyExclusive("blocked", "ready")
	cmd.MarkFlagsMutuallyExclusive("owner", "mine")
	return cmd
}

// takeHelp is the help of --take of start and update.
const takeHelp = "Take an issue that is in progress from its owner, and warn with the old owner. Only when the owner has stopped the work."

// ownerFilter returns the owner that list and search filter on: owner, or with mine the value of
// SISYPHUS_OWNER.
func ownerFilter(owner string, mine bool) (string, error) {
	if !mine {
		return owner, nil
	}
	if owner = environmentOwner(); owner == "" {
		return "", fmt.Errorf("--mine lists the issues of the owner in %s, but %s is not set. Set it, or use --owner.", ownerVariable, ownerVariable)
	}
	return owner, nil
}

func resolveCommand(findRoot func() (string, error)) *cobra.Command {
	var all, absolute bool
	cmd := &cobra.Command{
		Use:   "resolve <link>",
		Short: "Find the file that a wikilink points to.",
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
		Short: "Check every wikilink in a document.",
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

func lintCommand(findRoot func() (string, error)) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Check every issue and every wikilink in the repo.",
		Long: "Check every issue and every wikilink in the repo, and write nothing.\n\n" +
			"For each issue in issues/: the file can be read; each frontmatter line is a key: value field, " +
			"and each value that opens a quote, [, or { closes it; state matches the state directory; " +
			"priority, effort, and resolution are valid; the full name is in one state directory only; no " +
			"tag has a space; the parent exists and has no cycle above it; and each depends-on issue exists, " +
			"is not the issue itself, and makes no cycle. For each Markdown file in the repo: every wikilink " +
			"is ok, not missing, ambiguous, or missing-heading.\n\n" +
			"Each problem is one line, \"<file>[:<line>]: <problem>\", and the exit code is 1. With no " +
			"problem, one summary line and exit code 0.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := findRoot()
			if err != nil {
				return err
			}
			result := lintRepo(root)
			out := cmd.OutOrStdout()
			if asJSON {
				encoder := json.NewEncoder(out)
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(result); err != nil {
					return err
				}
			} else {
				for _, problem := range result.Problems {
					fmt.Fprintln(out, problem)
				}
			}
			if n := len(result.Problems); n > 0 {
				return fmt.Errorf("%d %s in %d issues and %d Markdown files.", n, plural(n, "problem", "problems"), result.Issues, result.Documents)
			}
			if !asJSON {
				fmt.Fprintf(out, "%d issues and %d Markdown files: no problems.\n", result.Issues, result.Documents)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the result as JSON: the counts and a list of problems.")
	return cmd
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func initCommand(directory *string) *cobra.Command {
	var dir string
	var force, nested bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up issues/ in a repo: the template and the state directories.",
		Long: "Write issues/ into a repo: TEMPLATE.md and the open/in-progress/closed directories. Files that " +
			"exist stay as they are, unless --force is given. After init, the other commands of sisyphus work " +
			"in the repo.\n\n" +
			"init sets up issue tracking only. A contributing guide, agent rules, versioning, changelog, " +
			"decision records, and CI or GitHub Actions workflows are a separate concern: add them from a " +
			"project-scaffolding template if you want them.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --dir is taken from -C, as every other path of a command is.
			if *directory != "" && !filepath.IsAbs(dir) {
				dir = filepath.Join(*directory, dir)
			}
			root, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			return initRepo(root, force, nested, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".", "The root directory of the repo.")
	cmd.Flags().BoolVar(&force, "force", false, "Replace files that exist.")
	cmd.Flags().BoolVar(&nested, "nested", false, "Make a tracker below a directory that already has one. Without it, that is an error, because the new tracker would hide the other one.")
	return cmd
}

// printPath finds the repo root, runs the operation while it holds the lock of the repo (see
// lockIssues), and prints the path of the file that the operation
// wrote. With --json (see addJSONFlag), it prints the changed issue as show --json does, with its path
// and the warnings, instead of the path and warnings on stderr.
func printPath(cmd *cobra.Command, findRoot func() (string, error), operation func(root string, warnings io.Writer) (string, error)) error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	warnings := cmd.ErrOrStderr()
	var collected bytes.Buffer
	if asJSON {
		warnings = &collected
	}
	unlock, err := lockIssues(root)
	if err != nil {
		return err
	}
	path, err := operation(root, warnings)
	unlock()
	if err != nil {
		return err
	}
	if !asJSON {
		fmt.Fprintln(cmd.OutOrStdout(), path)
		return nil
	}
	doc, err := loadDocument(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return err
	}
	result := writeResult{issueView: newIssueView(issueName(path), doc), Path: path, Warnings: []string{}}
	for _, line := range splitLines(collected.String()) {
		result.Warnings = append(result.Warnings, strings.TrimPrefix(line, "Warning: "))
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// writeResult is the --json output of a command that changes an issue: the issue as show --json
// prints it, the path of its file, and the warnings of the command.
type writeResult struct {
	issueView
	Path     string   `json:"path"`
	Warnings []string `json:"warnings"`
}

// addJSONFlag gives a command that changes an issue the --json flag that printPath reads.
func addJSONFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool("json", false, "Print the changed issue as JSON, as show --json does, with its path and the warnings.")
	return cmd
}

// writeIssueText prints an issue's key fields, then its body. Without all, it leaves out the empty
// fields and the sections that hold only the template text, which tell an author what to write, not a
// reader what the issue is.
func writeIssueText(out io.Writer, doc *document, all bool) {
	fields := [][2]string{
		{"title", doc.get("title")},
		{"state", doc.get("state")},
		{"resolution", doc.get("resolution")},
		{"priority", doc.get("priority")},
		{"effort", doc.get("effort")},
		{"tags", doc.get("tags")},
		{"created", doc.get("created")},
		{"closed", doc.get("closed")},
		{"owner", doc.get("owner")},
		{"approver", doc.get("approver")},
		{"bookmark", doc.get("bookmark")},
		{"parent", doc.get("parent")},
		{"depends-on", doc.get("depends-on")},
		{"remote", doc.get("remote")},
		{"metadata", doc.get("metadata")},
	}
	// The width of every label, also the hidden ones, so that the values start in the same column in
	// every issue.
	width := 0
	for _, f := range fields {
		width = max(width, utf8.RuneCountInString(f[0]))
	}
	if !all {
		alwaysShown := []string{"title", "state", "priority", "effort", "created"}
		fields = slices.DeleteFunc(fields, func(f [2]string) bool {
			return !slices.Contains(alwaysShown, f[0]) && (f[1] == "" || f[1] == "[]" || f[1] == "{}")
		})
	}
	for _, f := range fields {
		label, value := f[0], f[1]
		fmt.Fprintf(out, "%s:%s %s\n", label, strings.Repeat(" ", width-utf8.RuneCountInString(label)), value)
	}
	body := doc.body
	if !all {
		body = withoutTemplateSections(body)
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, strings.TrimLeft(strings.Join(body, "\n"), "\n"))
}

// withoutTemplateSections leaves out each "##" section that holds only template text (see
// isTemplateText), with its heading.
func withoutTemplateSections(body []string) []string {
	var starts []int
	for _, index := range outsideCodeFences(body) {
		if headingPattern.MatchString(body[index]) && headingLevel(body[index]) <= 2 {
			starts = append(starts, index)
		}
	}
	drop := map[int]bool{}
	for i, start := range starts {
		end := len(body)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		if headingLevel(body[start]) == 2 && isTemplateText(body[start+1:end]) {
			for line := start; line < end; line++ {
				drop[line] = true
			}
		}
	}
	var keep []string
	for index, line := range body {
		if !drop[index] {
			keep = append(keep, line)
		}
	}
	return keep
}

// isTemplateText reports whether lines hold only the placeholders of the template: "<...>" paragraphs
// and list items such as "- [ ] <Condition>", or nothing.
func isTemplateText(lines []string) bool {
	inPlaceholder := false
	for _, line := range lines {
		text := strings.TrimSpace(line)
		switch {
		case inPlaceholder:
			inPlaceholder = !strings.HasSuffix(text, ">")
		case text == "":
		case strings.HasPrefix(text, "<"):
			inPlaceholder = !strings.HasSuffix(text, ">")
		default:
			item := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(text, "-"), " [ ]"), " "))
			if !(strings.HasPrefix(item, "<") && strings.HasSuffix(item, ">")) {
				return false
			}
		}
	}
	return true
}

// terminalWidth returns the width of the terminal that out writes to, and false when out is not a
// terminal. Tests replace it.
var terminalWidth = func(out io.Writer) (int, bool) {
	f, ok := out.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return 0, false
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil || width <= 0 {
		return 0, false
	}
	return width, true
}

// minimumShrunkWidth is the narrowest that writeTableWithin makes the column it shortens.
const minimumShrunkWidth = 10

// writeTable prints cells as a table with header, padding every column to its widest cell. Widths
// are terminal display widths, so a wide character such as a CJK letter counts as two columns.
func writeTable(out io.Writer, header []string, cells [][]string) {
	writeTableWithin(out, header, cells, -1, 0)
}

// writeTableWithin is writeTable with a maximum width: when maxWidth is more than 0 and the table is
// wider, the column shrink becomes narrower, down to minimumShrunkWidth, and its cells end in "…".
func writeTableWithin(out io.Writer, header []string, cells [][]string, shrink, maxWidth int) {
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = runewidth.StringWidth(h)
	}
	for _, row := range cells {
		for i, cell := range row {
			widths[i] = max(widths[i], runewidth.StringWidth(cell))
		}
	}
	if total := sum(widths) + 2*(len(widths)-1); maxWidth > 0 && shrink >= 0 && total > maxWidth {
		widths[shrink] = max(minimumShrunkWidth, widths[shrink]-(total-maxWidth))
	}
	format := func(values []string) string {
		padded := make([]string, len(values))
		for i, v := range values {
			if runewidth.StringWidth(v) > widths[i] {
				v = runewidth.Truncate(v, widths[i], "…")
			}
			padded[i] = v + strings.Repeat(" ", widths[i]-runewidth.StringWidth(v))
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

func sum(values []int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
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
	message := fmt.Sprintf("Invalid %s '%s'. Use one of: %s.", what, value, strings.Join(choices, ", "))
	if suggestion := suggestValue(value, choices); suggestion != "" {
		message += fmt.Sprintf(" Did you mean '%s'?", suggestion)
	}
	return errors.New(message)
}

// valueSynonyms are words that people use for a valid value. They give a suggestion, never the value.
var valueSynonyms = map[string]string{
	"done": "closed", "close": "closed", "finished": "closed", "complete": "closed", "resolved": "closed",
	"start": "in-progress", "started": "in-progress", "doing": "in-progress", "wip": "in-progress",
	"active": "in-progress", "progress": "in-progress", "inprogress": "in-progress", "in_progress": "in-progress",
	"new": "open", "todo": "open", "backlog": "open", "reopen": "open",
	"urgent": "critical", "blocker": "critical", "med": "medium", "normal": "medium",
	"xl": "x-large", "xlarge": "x-large", "x_large": "x-large", "huge": "x-large",
	"cancelled": "abandoned", "canceled": "abandoned", "dropped": "abandoned",
}

// suggestValue returns the valid value that value most likely meant, or "": a synonym, else the only
// choice that starts with value, else the only choice within two edits.
func suggestValue(value string, choices []string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	if synonym, ok := valueSynonyms[lower]; ok && slices.Contains(choices, synonym) {
		return synonym
	}
	if slices.Contains(choices, lower) {
		return lower
	}
	var prefixed []string
	for _, choice := range choices {
		if lower != "" && strings.HasPrefix(choice, lower) {
			prefixed = append(prefixed, choice)
		}
	}
	if len(prefixed) == 1 {
		return prefixed[0]
	}
	var near []string
	for _, choice := range choices {
		if editDistance(lower, choice) <= 2 {
			near = append(near, choice)
		}
	}
	if len(near) == 1 {
		return near[0]
	}
	return ""
}

// editDistance is the Levenshtein distance between a and b.
func editDistance(a, b string) int {
	previous := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current := make([]int, len(b)+1)
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = min(previous[j]+1, current[j-1]+1, previous[j-1]+cost)
		}
		previous = current
	}
	return previous[len(b)]
}

// stateFilter reads a --state value of list or search: a comma-separated list of states, or "all".
func stateFilter(value string) []string {
	values := parseList(value)
	if slices.Contains(values, "all") {
		return slices.Clone(states)
	}
	return values
}

func validateEach(what string, values, choices []string) error {
	for _, value := range values {
		if err := oneOf(what, value, choices); err != nil {
			return err
		}
	}
	return nil
}
