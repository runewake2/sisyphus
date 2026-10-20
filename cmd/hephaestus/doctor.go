package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/internal/hephaestus/doctor"
	"github.com/runewake2/sisyphus/internal/hephaestus/mq"
	"github.com/runewake2/sisyphus/internal/hephaestus/repo"
	"github.com/runewake2/sisyphus/internal/hephaestus/sisyphus"
	"github.com/runewake2/sisyphus/internal/hephaestus/taskspace"
	"github.com/runewake2/sisyphus/internal/hephaestus/toolagent"
)

// taskInput gathers what the doctor reads about the task of one issue: the issue, the queue, the
// message log, the task workspace, and the bookmark.
func taskInput(w workspace, r repo.Repo, issue, onto string, stale time.Duration) (doctor.Input, error) {
	detail, err := sisyphus.Show(w.directory(), issue)
	if err != nil {
		return doctor.Input{}, err
	}
	in := doctor.Input{Now: time.Now(), Stale: stale, Issue: detail}
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
		return in, err
	}
	if role := doctor.Role(in); role != "" && r.VCS != repo.None {
		name := taskspace.Name(strings.ReplaceAll(role, "/", "-"), path.Base(detail.Name))
		space := taskspace.Space{Name: name, Dir: taskspace.Dir(r, "", name)}
		if info, err := os.Stat(space.Dir); err == nil && info.IsDir() {
			in.Workspace = &doctor.Workspace{Name: name, Dir: space.Dir, Dirty: taskspace.Dirty(r, space)}
		}
	}
	if detail.Bookmark != "" && r.VCS != repo.None {
		b := &doctor.Bookmark{Name: detail.Bookmark}
		if head, err := r.CommitOf(detail.Bookmark); err == nil && head != "" {
			b.Exists, b.Head = true, head
			if merged, err := r.IsAncestor(head, onto); err == nil {
				b.Merged = merged
			}
		}
		in.Bookmark = b
	}
	return in, nil
}

func doctorCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var stale time.Duration
	var onto string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor <issue>",
		Short: "Diagnose the task of an issue: whether a worker works on it, lost it, or stopped on it.",
		Long: "Diagnose the task of an issue that hephaestus manages, and say what makes it run again. It reads the " +
			"issue, the workers and pending deliveries of the message queue, the message log, the task workspace, and " +
			"the bookmark, and reports the facts and a finding with its fix for each problem: an owner that stopped, a " +
			"task message that no one will take, a delegated issue with no task message, a referendum past its " +
			"deadline, or a bookmark that does not exist. It changes nothing.\n\n" +
			"Exit codes: 0 when the task is not stuck; 2 when there are findings; 1 when the doctor could not look.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			in, err := taskInput(w, r, args[0], onto, stale)
			if err != nil {
				return err
			}
			report := doctor.Diagnose(in)
			if asJSON {
				if err := writeJSON(cmd.OutOrStdout(), report); err != nil {
					return err
				}
			} else {
				fmt.Fprint(cmd.OutOrStdout(), report.Text())
			}
			if n := len(report.Findings); n > 0 {
				return exitError{code: 2, message: fmt.Sprintf("The task of %s has %d %s.", report.Issue, n, plural(n, "finding", "findings"))}
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&stale, "stale", 2*time.Minute, "A worker that has not spoken to the queue for this long counts as stopped.")
	cmd.Flags().StringVar(&onto, "onto", "main", "The base: the doctor tells whether it has the bookmark of the issue.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the report as JSON.")
	return cmd
}

func kickCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var role, reason, as string
	var force, asJSON bool
	var stale time.Duration
	cmd := &cobra.Command{
		Use:   "kick <issue>",
		Short: "Make a stuck task run again: reopen its issue, and give it back to its role.",
		Long: "Make a stuck task run again, for example after a worker crashed, ran out of API keys, or hung until its " +
			"time limit. It reopens an in-progress issue (which drops the owner that stopped), removes the tag " +
			"needs-human, keeps the delegation to the role, and sends a new task message to the role, so that a live " +
			"worker takes the task again. When a task message already waits, it sends no second one: the reopened " +
			"issue lets a worker take it. The task workspace of the issue stays, and the kicked task continues there.\n\n" +
			"A task that a live worker holds is slow, not stuck: kick refuses it without --force. Kick does not stop a " +
			"running worker, and it does not start one: hephaestus never starts a model by itself. The kick is " +
			"recorded in the Notes of the issue. Diagnose first with: hephaestus doctor <issue>.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue := args[0]
			r, err := findRepo()
			if err != nil {
				return err
			}
			client, err := mq.Connect(r.SharedDir())
			if err != nil {
				return err // the kick sends and inspects messages: it needs the queue
			}
			in, err := taskInput(w, r, issue, "main", stale)
			if err != nil {
				return err
			}
			plan, err := doctor.PlanKick(in, role, force)
			if err != nil {
				return err
			}
			update := []string{"update", plan.Issue}
			if plan.Reopen {
				update = append(update, "open")
			}
			if plan.RemoveNeedsHuman {
				update = append(update, "--remove-tags", toolagent.NeedsHuman)
			}
			update = append(update, "--metadata", delegateKey+"="+plan.Role)
			if _, err := sisyphus.Run(w.directory(), update...); err != nil {
				return err
			}
			kicker := as
			if kicker == "" {
				kicker = os.Getenv(ownerVariable)
			}
			if kicker == "" {
				kicker = "human"
			}
			if plan.Send {
				body := fmt.Sprintf("The task was kicked by %s: do it again, or continue it.", kicker)
				if strings.TrimSpace(reason) != "" {
					body += " " + strings.TrimSpace(reason)
				}
				var data json.RawMessage
				if plan.Bookmark != "" {
					data, _ = json.Marshal(map[string]string{"bookmark": plan.Bookmark})
				}
				if _, err := client.Send(mq.Message{ID: plan.MessageID, From: kicker, To: mq.Address{Role: plan.Role}, Kind: "task", Issue: plan.Issue, Body: body, Data: data}); err != nil {
					return err
				}
			}
			note := fmt.Sprintf("%s: %s kicked the task to %s", time.Now().Format("2006-01-02"), kicker, plan.Role)
			if plan.Send {
				note += fmt.Sprintf(" (message %s)", plan.MessageID)
			}
			note += "."
			if strings.TrimSpace(reason) != "" {
				note += " " + strings.TrimSpace(reason)
			}
			if _, err := sisyphus.Run(w.directory(), "edit", plan.Issue, "notes", "--append", "--", note); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: the kick could not be recorded in %s: %v\n", plan.Issue, err)
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), plan)
			}
			did := []string{}
			if plan.Reopen {
				did = append(did, "reopened the issue")
			}
			if plan.RemoveNeedsHuman {
				did = append(did, "removed needs-human")
			}
			if plan.Send {
				did = append(did, fmt.Sprintf("sent the task message %s to role:%s", plan.MessageID, plan.Role))
			}
			if len(did) == 0 {
				did = append(did, "kept the issue as it is")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s.\n", plan.Issue, strings.Join(did, ", "))
			for _, warning := range plan.Warnings {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", warning)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "The role that gets the task. Default: the delegate metadata of the issue, or the receiver of its last task message.")
	cmd.Flags().StringVar(&reason, "reason", "", "Why the task is kicked; it goes into the message and the Notes of the issue.")
	cmd.Flags().StringVar(&as, "as", "", "Who kicks, for the message and the Notes. Default: "+ownerVariable+", or human.")
	cmd.Flags().BoolVar(&force, "force", false, "Kick also a task that a live worker holds.")
	cmd.Flags().DurationVar(&stale, "stale", 2*time.Minute, "A worker that has not spoken to the queue for this long counts as stopped.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the plan of the kick as JSON.")
	return cmd
}
