// Command hephaestus is a worker manager: it gives sisyphus issues to agents that each have one
// role, and runs the tool agents, for example the merge queue. It uses sisyphus only through the
// sisyphus command, as any other client does.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run runs hephaestus with args, and returns the exit code. Errors go to stderr, as "Error: ...".
func run(args []string, stdout, stderr io.Writer) int {
	root := newRootCommand()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", err)
		return 1
	}
	return 0
}

func newRootCommand() *cobra.Command {
	var directory, rolesDir string
	root := &cobra.Command{
		Use:   "hephaestus",
		Short: "Give sisyphus issues to agents that each have one role, and run the tool agents.",
		Long: "hephaestus is a worker manager. It delegates sisyphus issues to roles, and runs the tool agents, " +
			"for example the merge queue. It uses sisyphus only through the sisyphus command.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().StringVarP(&directory, "directory", "C", ".", "Run as if started in this directory, as git -C does.")
	root.PersistentFlags().StringVar(&rolesDir, "roles", "", "The directory of the role files. Default: hephaestus/roles of the repo.")
	findRepo := func() (repo.Repo, error) { return repo.Find(directory) }
	w := workspace{
		directory: func() string { return directory },
		roles: func() (string, error) {
			if rolesDir != "" {
				return rolesDir, nil
			}
			r, err := findRepo()
			if err != nil {
				return "", err
			}
			return roles.Dir(r.Root), nil
		},
	}
	root.AddCommand(versionCommand(findRepo), decisionsCommand(findRepo), rolesCommand(w), delegateCommand(w), tasksCommand(w), mergeOrderCommand(w, findRepo), messagesCommand(w, findRepo), agentCommand(w, findRepo), mergeQueueCommand(w, findRepo), referendumCommand(w, findRepo), filesCommand(findRepo), statusCommand(w, findRepo))
	return root
}

// writeJSON writes value as indented JSON.
func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
