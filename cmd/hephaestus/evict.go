package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/internal/hephaestus/mq"
	"github.com/runewake2/sisyphus/internal/hephaestus/repo"
)

func evictCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var reason string
	var force, asJSON bool
	cmd := &cobra.Command{
		Use:   "evict <worker>",
		Short: "Break glass: drop the leases of a worker, and stop its running task through its keep-alive.",
		Long: "Evict a worker that loops and never releases its lease: the one case that kick refuses, because the " +
			"worker is live. The queue drops every lease of the worker, so its messages are deliverable to another " +
			"worker at once; the next keep-alive of the worker tells it that it was evicted, and the worker stops its " +
			"task with every program that it started, releases its issue (open, no needs-human, the delegation " +
			"stays), does not acknowledge the message, and ends. Its next receive is refused once, so it takes no " +
			"more work.\n\n" +
			"This is a break-glass command: it always needs --force, and without it, it only shows what it would do. " +
			"Diagnose first (hephaestus doctor <issue>) and prefer hephaestus kick for a worker that stopped. Evict " +
			"does not kill a process: a worker whose keep-alive is dead cannot hear the eviction, and its leases end " +
			"by themselves; stop such a process by hand.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			worker := args[0]
			r, err := findRepo()
			if err != nil {
				return err
			}
			client, err := mq.Connect(r.SharedDir())
			if err != nil {
				return err
			}
			workers, err := client.Workers()
			if err != nil {
				return err
			}
			var found *mq.WorkerStatus
			for i := range workers {
				if workers[i].Worker == worker {
					found = &workers[i]
				}
			}
			if found == nil {
				return fmt.Errorf("The queue knows no worker %q: there is nothing to evict. hephaestus messages workers lists them.", worker)
			}
			if !force {
				holds := make([]string, 0, len(found.Holds))
				for _, held := range found.Holds {
					holds = append(holds, fmt.Sprintf("%s (lease until %s)", held.ID, held.Until.Local().Format("15:04:05")))
				}
				detail := "no lease"
				if len(holds) > 0 {
					detail = strings.Join(holds, ", ")
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s: last seen %s, holds %s.\n", worker, lastSeen(found.LastSeen), detail)
				fmt.Fprintf(cmd.OutOrStdout(), "An eviction would drop its leases, stop its running task, and end the worker.\n")
				return exitError{code: 1, message: fmt.Sprintf("Evict is a break-glass command: run it with --force. For a worker that stopped, use hephaestus kick <issue> instead.")}
			}
			evicted, err := client.Evict(worker, strings.TrimSpace(reason))
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), evicted)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s is evicted: %d %s dropped. Its task stops at its next keep-alive, within about %s.\n",
				worker, len(evicted.Dropped), plural(len(evicted.Dropped), "lease", "leases"), mq.KeepAliveInterval)
			for _, issue := range evicted.Issues {
				fmt.Fprintf(cmd.OutOrStdout(), "Check the task afterwards: hephaestus doctor %s\n", issue)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Why the worker is evicted; the worker records it in the Notes of its issue.")
	cmd.Flags().BoolVar(&force, "force", false, "Do it. Without --force, evict only shows what it would do.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the dropped leases as JSON.")
	return cmd
}

// lastSeen writes when a worker last spoke to the queue.
func lastSeen(t time.Time) string {
	if t.IsZero() {
		return "never since the queue started"
	}
	return t.Local().Format("15:04:05")
}
