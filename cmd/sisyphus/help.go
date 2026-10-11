package main

import (
	"strings"

	"github.com/spf13/cobra"
)

// commandGroups are the groups of the root help, in the order the help shows them. Each command is in
// the group of its first word.
var commandGroups = []struct {
	id, title string
	commands  []string
}{
	{"change", "Create and change issues:", []string{"new", "start", "update", "close", "abandon", "edit", "rename"}},
	{"find", "Read and find issues:", []string{"show", "list", "search", "graph", "claims"}},
	{"relate", "Relate issues:", []string{"parent", "depends-on", "remote"}},
	{"check", "Check names and links:", []string{"slug", "resolve", "links", "lint"}},
	{"setup", "Set up:", []string{"init", "completion", "help"}},
}

// commandExamples are the examples of each command's help. Every line is also on the docs page of the
// command, docs/sisyphus-<command>.md, so the help and the docs show the same use.
var commandExamples = map[string][]string{
	"new": {
		`sisyphus new --title "Fix the login bug" --tags bug --priority high`,
		`sisyphus new web/auth/fix-session-expiry --tags bug   # issues/open/web/auth/fix-session-expiry.md`,
		`sisyphus new split-login-form --parent add-dark-mode --depends-on fix-login-bug`,
	},
	"start": {
		`sisyphus start fix-login-bug`,
		`sisyphus start fix-login-bug --owner alice --bookmark ai/fix-login-bug`,
		`sisyphus start fix-login-bug --owner bob --take   # alice has stopped the work`,
	},
	"update": {
		`sisyphus update fix-login-bug in-progress --priority critical --tags bug,auth`,
		`sisyphus update fix-login-bug --add-tags bug --remove-tags triage`,
		`sisyphus update fix-login-bug closed`,
		`sisyphus update retry-backoff open`,
	},
	"close": {
		`sisyphus close fix-login-bug`,
		`sisyphus close fix-login-bug -m "Fixed in [[0.0.9]]: the session cache now expires."`,
	},
	"abandon": {
		`sisyphus abandon old-login-form`,
		`sisyphus abandon old-login-form -m "Login moved to SSO in [[sso-login]]."`,
	},
	"edit": {
		`sisyphus edit fix-login-bug                # in $VISUAL or $EDITOR`,
		`sisyphus edit fix-login-bug summary "Fix the login loop after a password reset."`,
		`sisyphus edit fix-login-bug notes "2026-10-09: the bug is in the session cache." --append`,
		`git log -1 --format=%B | sisyphus edit fix-login-bug resolution -`,
	},
	"rename": {
		`sisyphus rename fix-login-bug web/auth/fix-login-bug   # move it to a directory`,
		`sisyphus rename fix-login-bug fix-login-loop`,
		`sisyphus rename fix-login-bug web/fix-login-bug --dry-run`,
	},
	"show": {
		`sisyphus show fix-login-bug`,
		`sisyphus show "[[fix-login-bug]]" --json | jq -r .state`,
	},
	"list": {
		`sisyphus list`,
		`sisyphus list --priority critical,high --owner alice`,
		`sisyphus list --ready --state open          # what can start now`,
		`sisyphus list --state in-progress --json | jq -r '.[].name'   # what is claimed right now`,
	},
	"search": {
		`sisyphus search timeout`,
		`sisyphus search timeout --section summary`,
		`sisyphus search '"session cache" login'   # an exact phrase and a word`,
		`sisyphus search --section "acceptance criteria" --tags bug --json`,
	},
	"graph": {
		`sisyphus graph auth-overhaul               # the whole epic`,
		`sisyphus graph split-login-form --full     # everything linked to it`,
		`sisyphus graph auth-overhaul --mermaid > auth-overhaul-graph.txt`,
	},
	"claims": {
		`sisyphus claims`,
		`sisyphus claims --json | jq -r '.[] | select(.status == "active") | .name'`,
		`sisyphus claims --prune                    # remove the claims of removed workspaces`,
	},
	"parent": {
		`sisyphus parent split-login-form auth-overhaul`,
		`sisyphus parent split-login-form --clear`,
	},
	"depends-on": {
		`sisyphus depends-on split-login-form fix-login-bug`,
		`sisyphus depends-on split-login-form fix-login-bug --clear`,
		`sisyphus depends-on split-login-form --clear`,
	},
	"remote": {
		`sisyphus remote fix-login-bug https://github.com/acme/widgets/issues/42`,
		`sisyphus remote fix-login-bug --clear`,
	},
	"slug": {
		`sisyphus slug "Crash when the cache is empty"`,
		`name="$(sisyphus slug "$TITLE")" && sisyphus new "$name" --title "$TITLE"`,
	},
	"resolve": {
		`sisyphus resolve fix-login-bug`,
		`sisyphus resolve "[[CONTRIBUTING#Issues]]"`,
		`sisyphus resolve README --all`,
	},
	"links": {
		`sisyphus links README.md`,
		`sisyphus links "[[fix-login-bug]]" --json | jq '.[] | select(.status != "ok")'`,
	},
	"lint": {
		`sisyphus lint`,
		`sisyphus lint --json | jq -r '.problems[].file' | sort -u`,
	},
	"init": {
		`sisyphus init`,
		`sisyphus init --dir ../other-repo`,
		`sisyphus init --force    # restore the default issues/TEMPLATE.md`,
	},
	"completion": {
		`source <(sisyphus completion bash)`,
		`sisyphus completion fish > ~/.config/fish/completions/sisyphus.fish`,
	},
}

// addHelp puts every command of root in its group, gives it its examples, and words the help command
// the way the other commands are worded.
func addHelp(root *cobra.Command) {
	root.InitDefaultHelpCmd()
	group := map[string]string{}
	for _, g := range commandGroups {
		root.AddGroup(&cobra.Group{ID: g.id, Title: g.title})
		for _, name := range g.commands {
			group[name] = g.id
		}
	}
	for _, cmd := range root.Commands() {
		cmd.GroupID = group[cmd.Name()]
		if examples, ok := commandExamples[cmd.Name()]; ok {
			cmd.Example = "  " + strings.Join(examples, "\n  ")
		}
		if cmd.Name() == "help" {
			cmd.Short = "Print the help of a command."
			cmd.Long = "Print the help of a command, for example: sisyphus help update. It is the same as sisyphus update --help."
		}
	}
}
