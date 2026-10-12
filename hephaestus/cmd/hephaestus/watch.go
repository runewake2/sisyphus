package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
	"github.com/runewake2/sisyphus/hephaestus/internal/watch"
)

// gitRepo answers the watcher from the git directory of the repo.
type gitRepo struct {
	gitDir, base string
	baseID       string
}

func (g gitRepo) Head(bookmark string) (string, bool) {
	head, err := revParse(g.gitDir, bookmark)
	return head, err == nil && head != ""
}

func (g gitRepo) InBase(commit string) bool {
	return g.baseID != "" && isAncestor(g.gitDir, commit, g.baseID)
}

func watchCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var onto string
	var every time.Duration
	var once bool
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Broadcast bookmark-ready and issue-closed from the state of the issues, once each.",
		Long: "Watch the issues of the repo, and broadcast the notices that the roles listen for, so that they do " +
			"not depend on a worker that remembers to send them:\n\n" +
			"- When an issue closes: bookmark-ready, if its bookmark has work that --onto does not have, and then " +
			"issue-closed.\n" +
			"- When the bookmark of a closed issue moves later, for example for the fixes of a review: bookmark-ready " +
			"again, for the new head.\n\n" +
			"Each notice has an id from the issue (and the head), so a repeated send makes no second message. The " +
			"first run records what it sees, and sends nothing. The state is watch.json in the shared hephaestus " +
			"directory of the repo; it changes only after the notices are sent, so a failure loses nothing. It needs " +
			"the message queue. --once looks once; otherwise it looks every --every until it is stopped.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			client, err := mq.Connect(r.SharedDir())
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			for {
				if err := watchOnce(cmd, w, r, client, onto); err != nil {
					return err
				}
				if once {
					return nil
				}
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(every):
				}
			}
		},
	}
	cmd.Flags().StringVar(&onto, "onto", "main", "The base: a bookmark whose work the base has already is not announced.")
	cmd.Flags().DurationVar(&every, "every", 30*time.Second, "How often to look.")
	cmd.Flags().BoolVar(&once, "once", false, "Look once, and stop.")
	return cmd
}

// watchOnce compares the issues with the last state, broadcasts the events, and saves the new state.
func watchOnce(cmd *cobra.Command, w workspace, r repo.Repo, client *mq.Client, onto string) error {
	if err := exportJJ(r); err != nil {
		return err
	}
	gitDir, err := r.GitDir()
	if err != nil {
		return err
	}
	baseID, _ := revParse(gitDir, onto)
	issues, err := sisyphus.List(w.directory(), "--state", "all")
	if err != nil {
		return err
	}
	prev, err := watch.Load(r.SharedDir())
	if err != nil {
		return err
	}
	next, events := watch.Plan(prev, issues, gitRepo{gitDir: gitDir, base: onto, baseID: baseID}, onto, time.Now())
	for _, e := range events {
		data, _ := json.Marshal(map[string]string{"issue": e.Issue, "bookmark": e.Bookmark, "head": e.Head, "base": e.Base})
		if _, err := client.Send(mq.Message{ID: e.ID, From: "hephaestus-watch", To: mq.Address{Broadcast: true}, Kind: e.Kind, Issue: e.Issue, Body: e.Body, Data: data}); err != nil {
			return fmt.Errorf("could not broadcast %s for %s: %v", e.Kind, e.Issue, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", e.Kind, e.Issue)
	}
	if prev == nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "The watcher recorded %d issues as they are now. From now on, it announces each close.\n", len(next.Issues))
	}
	return watch.Save(r.SharedDir(), next)
}
