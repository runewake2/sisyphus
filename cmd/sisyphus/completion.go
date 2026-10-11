package main

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// completionShells are the shells that sisyphus completion writes a script for.
var completionShells = []string{"bash", "zsh", "fish", "powershell", "nushell"}

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion <shell>",
		Short: "Print the shell completion script for a shell.",
		Long: "Print the script that completes sisyphus commands, flags, issue names, and flag values in " +
			"a shell: " + strings.Join(completionShells, ", ") + ". Load the script in the start-up file of " +
			"the shell. docs/sisyphus-completion.md tells how for each shell.",
		Args:      cobra.ExactArgs(1),
		ValidArgs: completionShells,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			root := cmd.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(out, true)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(out)
			case "nushell":
				_, err := io.WriteString(out, nushellCompletion)
				return err
			}
			return oneOf("shell", args[0], completionShells)
		},
	}
}

// nushellCompletion is an external completer for Nushell. cobra writes no script for Nushell, so the
// completer asks sisyphus itself with the hidden __complete command, which the other scripts use too.
// It reads one candidate per line, "value<TAB>description", and a last line ":<directive>". Directive
// bit 4 (cobra.ShellCompDirectiveNoFileComp) means: do not complete file names. The completer keeps the
// external completer that was set before, for every other command.
const nushellCompletion = `# Nushell completion for sisyphus. Load it from config.nu, for example:
#   sisyphus completion nushell | save --force ($nu.default-config-dir | path join sisyphus.nu)
#   source ($nu.default-config-dir | path join sisyphus.nu)

let sisyphus_completer = {|spans: list<string>|
    let lines = (^sisyphus __complete ...($spans | skip 1) | complete | get stdout | lines)
    let directives = ($lines | where {|line| $line starts-with ":" })
    let directive = if ($directives | is-empty) { 0 } else { $directives | last | str substring 1.. | into int }
    let candidates = ($lines
        | where {|line| $line != "" and not ($line starts-with ":") }
        | each {|line|
            let parts = ($line | split row "\t")
            if ($parts | length) > 1 { {value: $parts.0, description: $parts.1} } else { {value: $parts.0} }
        })
    if ($candidates | is-empty) and ($directive bit-and 4) == 0 { null } else { $candidates }
}

let sisyphus_previous_completer = ($env.config.completions.external.completer? | default null)
$env.config.completions.external.enable = true
$env.config.completions.external.completer = {|spans: list<string>|
    if ($spans | first) == "sisyphus" {
        do $sisyphus_completer $spans
    } else if $sisyphus_previous_completer != null {
        do $sisyphus_previous_completer $spans
    } else {
        null
    }
}
`

// argumentKind is what one positional argument of a command holds, for completion.
type argumentKind int

const (
	anyIssue argumentKind = iota
	openIssue
	unclosedIssue
	stateArgument
	sectionArgument
)

// commandArguments lists, for each command that takes an issue, what its positional arguments hold.
var commandArguments = map[string][]argumentKind{
	"update":     {anyIssue, stateArgument},
	"start":      {openIssue},
	"close":      {unclosedIssue},
	"abandon":    {unclosedIssue},
	"parent":     {anyIssue, anyIssue},
	"depends-on": {anyIssue, anyIssue},
	"remote":     {anyIssue},
	"edit":       {anyIssue, sectionArgument},
	"rename":     {anyIssue},
	"show":       {anyIssue},
	"graph":      {anyIssue},
}

// filterCommands are the commands whose --state and --priority take a comma-separated list.
var filterCommands = []string{"list", "search"}

// describeLimit is the largest number of issue names that completion shows with their titles. A title
// costs a read of the issue file, so a longer list has only the names.
const describeLimit = 200

// addCompletions sets the completion of positional arguments and flag values on every command below
// root. findRoot must see the -C flag, which cobra parses before it asks for completions.
func addCompletions(root *cobra.Command, findRoot func() (string, error)) {
	registerFlagCompletion(root, "directory", directoriesOnly)
	for _, cmd := range root.Commands() {
		if kinds, ok := commandArguments[cmd.Name()]; ok {
			cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
				if len(args) >= len(kinds) {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}
				switch kinds[len(args)] {
				case stateArgument:
					return prefixed(states, toComplete), cobra.ShellCompDirectiveNoFileComp
				case sectionArgument:
					return completeSections(findRoot, args[0], toComplete)
				default:
					return completeIssues(findRoot, kinds[len(args)], toComplete)
				}
			}
		} else if cmd.ValidArgsFunction == nil && cmd.ValidArgs == nil && cmd.Name() != "links" {
			// links takes a file. Every other command takes text, which completion cannot guess.
			cmd.ValidArgsFunction = cobra.NoFileCompletions
		}
		multiple := slices.Contains(filterCommands, cmd.Name())
		values := map[string][]string{
			"state":      states,
			"priority":   priorities,
			"effort":     efforts,
			"resolution": resolutions,
			"color":      {"auto", "always", "never"},
		}
		if multiple {
			values["state"] = append(slices.Clone(states), "all")
		}
		for flag, choices := range values {
			if cmd.Flags().Lookup(flag) == nil {
				continue
			}
			complete := func(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
				if multiple && (flag == "state" || flag == "priority") {
					if strings.Contains(toComplete, ",") {
						// all is only valid alone.
						choices = slices.DeleteFunc(slices.Clone(choices), func(c string) bool { return c == "all" })
					}
					return completeList(choices, toComplete)
				}
				return prefixed(choices, toComplete), cobra.ShellCompDirectiveNoFileComp
			}
			registerFlagCompletion(cmd, flag, complete)
		}
		for _, flag := range []string{"tags", "add-tags", "remove-tags"} {
			if cmd.Flags().Lookup(flag) != nil {
				registerFlagCompletion(cmd, flag, func(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
					return completeList(tagsInUse(findRoot), toComplete)
				})
			}
		}
		for _, flag := range []string{"parent", "depends-on", "deferred-from"} {
			if cmd.Flags().Lookup(flag) != nil {
				registerFlagCompletion(cmd, flag, func(_ *cobra.Command, _ []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
					return completeIssues(findRoot, anyIssue, toComplete)
				})
			}
		}
		if cmd.Flags().Lookup("dir") != nil {
			registerFlagCompletion(cmd, "dir", directoriesOnly)
		}
	}
}

// registerFlagCompletion sets the completion of one flag. A failure is a programming error, such as a
// second registration for the same flag, so it panics in tests and at start-up.
func registerFlagCompletion(cmd *cobra.Command, flag string, complete cobra.CompletionFunc) {
	if err := cmd.RegisterFlagCompletionFunc(flag, complete); err != nil {
		panic(fmt.Sprintf("completion of --%s of %s: %v", flag, cmd.Name(), err))
	}
}

func directoriesOnly(*cobra.Command, []string, string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveFilterDirs
}

// prefixed returns the choices that start with toComplete.
func prefixed(choices []string, toComplete string) []cobra.Completion {
	var result []cobra.Completion
	for _, choice := range choices {
		if strings.HasPrefix(choice, toComplete) {
			result = append(result, choice)
		}
	}
	return result
}

// completeList completes the last item of a comma-separated list, and keeps the items before it. It
// does not offer an item that the list already has.
func completeList(choices []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	done, last := "", toComplete
	if i := strings.LastIndex(toComplete, ","); i >= 0 {
		done, last = toComplete[:i+1], toComplete[i+1:]
	}
	have := strings.Split(done, ",")
	var result []cobra.Completion
	for _, choice := range prefixed(choices, last) {
		if !slices.Contains(have, choice) {
			result = append(result, done+choice)
		}
	}
	return result, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

// completeIssues returns the full names of the issues of kind that start with toComplete, with their
// titles when there are not too many.
func completeIssues(findRoot func() (string, error), kind argumentKind, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	root, err := findRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	index := loadIndex(root)
	var issues []issueFile
	for _, issue := range index.all() {
		if !strings.HasPrefix(issue.name, toComplete) ||
			kind == openIssue && issue.state != "open" ||
			kind == unclosedIssue && issue.state == "closed" {
			continue
		}
		issues = append(issues, issue)
	}
	describe := len(issues) <= describeLimit
	if describe {
		index.preload(issues, false)
	}
	result := make([]cobra.Completion, 0, len(issues))
	for _, issue := range issues {
		title := ""
		if describe {
			if doc, err := index.header(issue); err == nil {
				title = doc.get("title")
			}
		}
		result = append(result, cobra.CompletionWithDesc(issue.name, title))
	}
	return slices.Compact(result), cobra.ShellCompDirectiveNoFileComp
}

// completeSections returns the section headings of the issue reference that start with toComplete,
// matched case-insensitively, as sisyphus edit matches them.
func completeSections(findRoot func() (string, error), reference, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	root, err := findRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	_, _, doc, err := loadIndex(root).load(reference)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var result []cobra.Completion
	for _, name := range sectionNames(doc.body) {
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(toComplete)) {
			result = append(result, name)
		}
	}
	return result, cobra.ShellCompDirectiveNoFileComp
}

// tagsInUse returns the built-in tags and every tag that an issue has, sorted.
func tagsInUse(findRoot func() (string, error)) []string {
	tags := slices.Clone(builtinTags)
	if root, err := findRoot(); err == nil {
		index := loadIndex(root)
		issues := index.all()
		index.preload(issues, false)
		for _, issue := range issues {
			if doc, err := index.header(issue); err == nil {
				tags = append(tags, parseList(doc.get("tags"))...)
			}
		}
	}
	slices.Sort(tags)
	return slices.Compact(tags)
}
