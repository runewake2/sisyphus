package main

import (
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
)

func filesCommand(findRepo func() (repo.Repo, error)) *cobra.Command {
	var changed, root, null bool
	var base string
	var exclude []string
	cmd := &cobra.Command{
		Use:   "files [<pattern>...]",
		Short: "Print the files of the working copy, or the files that the work changes, for a tool role.",
		Long: "Print the paths, from the root of the working copy, of the files that the VCS tracks, one on each line. " +
			"With --changed, only the files that the work since --base changes and that still exist: the commits from " +
			"the fork point of the base, and the changes that are not committed. A pattern, for example '*.go', " +
			"matches the file name, or the full path when it has a /. --exclude leaves out paths below a directory or " +
			"that match a pattern. --root prints the root of the working copy. A tool role, for example a formatter, " +
			"uses it to work only on the files of a change, with jj or git.",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			if root {
				fmt.Fprintln(cmd.OutOrStdout(), r.Root)
				return nil
			}
			if base == "" {
				base = os.Getenv("HEPHAESTUS_BASE")
			}
			if base == "" {
				base = "main"
			}
			var files []string
			if changed {
				files, err = r.ChangedSince(base)
			} else {
				files, err = r.Files()
			}
			if err != nil {
				return err
			}
			slices.Sort(files)
			end := "\n"
			if null {
				end = "\x00"
			}
			for _, file := range files {
				if matches(file, args) && !excluded(file, exclude) {
					fmt.Fprint(cmd.OutOrStdout(), file+end)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&changed, "changed", false, "Only the files that the work since --base changes.")
	cmd.Flags().StringVar(&base, "base", "", "The base of --changed. Default: HEPHAESTUS_BASE, or main.")
	cmd.Flags().StringSliceVar(&exclude, "exclude", nil, "Leave out the files below these directories or that match these patterns.")
	cmd.Flags().BoolVarP(&null, "null", "0", false, "End each path with a NUL character, for xargs -0.")
	cmd.Flags().BoolVar(&root, "root", false, "Print the root of the working copy.")
	return cmd
}

// matches reports whether file matches one of the patterns. No pattern matches every file.
func matches(file string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		target := path.Base(file)
		if strings.Contains(pattern, "/") {
			target = file
		}
		if ok, _ := path.Match(pattern, target); ok {
			return true
		}
	}
	return false
}

// excluded reports whether file is below one of the directories, or matches one of the patterns.
func excluded(file string, excludes []string) bool {
	for _, e := range excludes {
		e = strings.TrimSuffix(e, "/")
		if file == e || strings.HasPrefix(file, e+"/") || matches(file, []string{e}) {
			return true
		}
	}
	return false
}
