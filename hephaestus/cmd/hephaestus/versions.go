package main

import (
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/decisions"
	"github.com/runewake2/sisyphus/hephaestus/internal/release"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
)

func versionCommand(findRepo func() (repo.Repo, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Give versions to the unreleased changelog entries.",
	}
	var dryRun, asJSON bool
	assign := &cobra.Command{
		Use:   "assign [<entry>...]",
		Short: "Give the next versions to unreleased changelog entries, and update every link to them.",
		Long: "A line of work does not pick a version: it writes an unreleased entry, " + release.UnreleasedDir + "/" +
			release.Prefix + "<topic>.md, with a first line \"# Unreleased: <title>\" and a line " +
			"\"- Changelog: <one line for CHANGELOG.md>\". assign gives each entry the next SHAME version after VERSION, " +
			"in the order of <entry>... (by default, every entry, by its Date, then its name). It moves the entry to " +
			"changelog/<version>.md, adds the CHANGELOG.md entry, sets VERSION, and changes every wikilink to the entry, " +
			"for example [[unreleased-fix-login]], to the version. The merge queue runs it as it lands the work.",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			plan, err := release.PlanAssign(r, args, time.Now())
			if err != nil {
				return err
			}
			if !dryRun {
				if err := release.Apply(r, plan); err != nil {
					return err
				}
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), plan)
			}
			out := cmd.OutOrStdout()
			for _, a := range plan.Assignments {
				fmt.Fprintf(out, "%s -> %s\n", a.Entry.Path, a.To)
			}
			fmt.Fprintf(out, "VERSION: %s -> %s\n", plan.From, plan.To)
			for _, file := range slices.Sorted(maps.Keys(plan.Links)) {
				fmt.Fprintf(out, "%s: %d %s\n", file, plan.Links[file], plural(plan.Links[file], "link", "links"))
			}
			return nil
		},
	}
	assign.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would change, and change nothing.")
	assign.Flags().BoolVar(&asJSON, "json", false, "Print the plan as JSON.")
	list := &cobra.Command{
		Use:   "list",
		Short: "List the unreleased changelog entries, in the default order of assign.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			entries, err := release.Entries(r.Root)
			if err != nil {
				return err
			}
			if asJSON {
				if entries == nil {
					entries = []release.Entry{}
				}
				return writeJSON(cmd.OutOrStdout(), entries)
			}
			for _, e := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", e.Date, e.Name, e.Title)
			}
			return nil
		},
	}
	list.Flags().BoolVar(&asJSON, "json", false, "Print JSON.")
	cmd.AddCommand(assign, list)
	return cmd
}

func decisionsCommand(findRepo func() (repo.Repo, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decisions",
		Short: "Keep the numbers of decision records unique.",
	}
	var dryRun, asJSON bool
	var onto string
	renumber := &cobra.Command{
		Use:   "renumber",
		Short: "Give a new number to each decision record that has the number of another one.",
		Long: "Two lines of work can each give a new decision record the same next number, for example " +
			"design/decisions/cli/0030-a.md and 0030-b.md. renumber finds such records in each decisions directory, " +
			"keeps the number of the record that --onto (the base of the merge) already has, gives each other record " +
			"the next free number of its directory, and changes its \"# NNNN:\" heading and every wikilink to it.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			plan, err := decisions.PlanRenumber(r, onto)
			if err != nil {
				return err
			}
			if !dryRun {
				if err := decisions.Apply(r, plan); err != nil {
					return err
				}
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), plan)
			}
			out := cmd.OutOrStdout()
			for _, rn := range plan.Renumberings {
				fmt.Fprintf(out, "%s -> %s\n", rn.From, rn.To)
			}
			for _, file := range slices.Sorted(maps.Keys(plan.Links)) {
				fmt.Fprintf(out, "%s: %d %s\n", file, plan.Links[file], plural(plan.Links[file], "link", "links"))
			}
			for _, problem := range plan.Problems {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", problem)
			}
			if len(plan.Renumberings) == 0 && len(plan.Problems) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "No two decision records have the same number.")
			}
			return nil
		},
	}
	renumber.Flags().StringVar(&onto, "onto", "main", "The base of the merge: a record that it has keeps its number.")
	renumber.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would change, and change nothing.")
	renumber.Flags().BoolVar(&asJSON, "json", false, "Print the plan as JSON.")
	cmd.AddCommand(renumber)
	return cmd
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
