package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
	"github.com/runewake2/sisyphus/hephaestus/internal/status"
)

func statusCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var stale time.Duration
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show what needs a human: issues, referendums, workers, waiting messages, and the last merge queue line.",
		Long: "Show what needs a human in the government of hephaestus, in one view:\n\n" +
			"- the open and in-progress issues tagged needs-human;\n" +
			"- the referendums that run, that are past their deadline, that failed on an issue that is still open, " +
			"and that have no result;\n" +
			"- the workers of the message queue, live or stopped (no keep-alive for --stale);\n" +
			"- the messages that wait, by receiver, and whether a live worker can take them;\n" +
			"- the last line of the merge queue (its merge-queue-done message), whether the base has it, and the " +
			"last release-recommendation of the release manager for it.\n\n" +
			"It changes nothing. Without a message queue, it reads the message log for the last line, and says that " +
			"the queue does not run.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			in := status.Input{Now: time.Now(), Stale: stale}
			if in.NeedsHuman, err = sisyphus.List(w.directory(), "--state", "open,in-progress", "--tags", "needs-human"); err != nil {
				return err
			}
			if in.WithReferendum, err = sisyphus.List(w.directory(), "--state", "all", "--metadata", referendumKey); err != nil {
				return err
			}
			if client, err := mq.Connect(r.SharedDir()); err == nil {
				in.Queue = client.Address()
				if in.Workers, err = client.Workers(); err != nil {
					in.QueueErr = err.Error()
				}
				if in.Pending, err = client.Pending(); err != nil {
					in.QueueErr = err.Error()
				}
			} else if err != mq.ErrNotRunning {
				in.QueueErr = err.Error()
			}
			if in.Log, err = mq.ReadLog(r.SharedDir()); err != nil {
				return err
			}
			if r.VCS != repo.None {
				exportJJ(r)
				if gitDir, err := r.GitDir(); err == nil {
					in.Merged = func(commit, base string) bool {
						baseID, err := revParse(gitDir, base)
						return err == nil && isAncestor(gitDir, commit, baseID)
					}
				}
			}
			report := status.Build(in)
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), report)
			}
			fmt.Fprint(cmd.OutOrStdout(), report.Text(in.Now))
			return nil
		},
	}
	cmd.Flags().DurationVar(&stale, "stale", 2*time.Minute, "A worker that has not spoken to the queue for this long counts as stopped.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the report as JSON.")
	return cmd
}
