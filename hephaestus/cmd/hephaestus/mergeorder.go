package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/mergeorder"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
)

func mergeOrderCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var onto string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "merge-order <task>...",
		Short: "Print the order in which the merge queue merges the work of a set of tasks.",
		Long: "Print the order in which the merge queue merges the work of a set of tasks: sisyphus issues, each " +
			"with the bookmark of its work in its bookmark field. Tasks with the same bookmark merge together. The " +
			"order keeps two rules first: when a task depends on a task of another bookmark, that bookmark goes " +
			"first, and a bookmark based on another one goes after it. Then the bookmark that shares changed files " +
			"with the fewest others goes first, then the one with the oldest change, then by name. Each place " +
			"tells why. A bookmark with no change from --onto is left out. It changes nothing.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			tasks, err := resolveTasks(w.directory(), args)
			if err != nil {
				return err
			}
			plan, err := mergeorder.Compute(r, tasks, onto)
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), plan)
			}
			out := cmd.OutOrStdout()
			for _, place := range plan.Order {
				fmt.Fprintf(out, "%d. %s (%s)\n", place.Place, place.Bookmark, strings.Join(place.Tasks, ", "))
				for _, reason := range place.After {
					fmt.Fprintf(out, "   after %s: %s\n", reason.Bookmark, reason.Why)
				}
				if len(place.Overlaps) == 0 {
					fmt.Fprintln(out, "   changes no file that another bookmark changes")
				}
				for _, overlap := range place.Overlaps {
					fmt.Fprintf(out, "   shares files with %s: %s\n", overlap.Bookmark, strings.Join(overlap.Files, ", "))
				}
			}
			for _, skipped := range plan.Skipped {
				fmt.Fprintf(out, "left out: %s (%s): %s\n", skipped.Bookmark, strings.Join(skipped.Tasks, ", "), skipped.Why)
			}
			for _, warning := range plan.Warnings {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s.\n", warning)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&onto, "onto", "main", "The base of the merge.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the order as JSON.")
	return cmd
}

// resolveTasks reads the tasks of the set with sisyphus: their full names, states, bookmarks, and
// dependencies, with the state of each dependency.
func resolveTasks(dir string, references []string) ([]mergeorder.Task, error) {
	states := map[string]string{}
	all, err := sisyphus.List(dir, "--state", "all")
	if err != nil {
		return nil, err
	}
	for _, issue := range all {
		states[issue.Name] = issue.State
	}
	var tasks []mergeorder.Task
	seen := map[string]bool{}
	for _, reference := range references {
		detail, err := sisyphus.Show(dir, reference)
		if err != nil {
			return nil, err
		}
		if seen[detail.Name] {
			continue
		}
		seen[detail.Name] = true
		task := mergeorder.Task{Name: detail.Name, State: detail.State, Bookmark: detail.Bookmark}
		for _, dependency := range detail.DependsOn {
			task.DependsOn = append(task.DependsOn, mergeorder.Dependency{Name: dependency, State: states[dependency]})
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
