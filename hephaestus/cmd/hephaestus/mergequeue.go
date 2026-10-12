package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/mergequeue"
	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
)

// runRequest is the data of a message that asks the merge queue for a run.
type runRequest struct {
	Tasks  []string `json:"tasks"`
	Onto   string   `json:"onto,omitempty"`
	Name   string   `json:"name,omitempty"`
	Checks string   `json:"checks,omitempty"`
}

func mergeQueueCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var onto, name, checks, conflictsTo string
	var fromMessage, asJSON bool
	cmd := &cobra.Command{
		Use:   "merge-queue [<task>...]",
		Short: "Make one linear history from the work of a set of tasks.",
		Long: "Make one linear history from the work of a set of tasks: sisyphus issues, each with the bookmark of " +
			"its work. In a scratch git worktree on --onto, it merges the commits of each bookmark in the order of " +
			"merge-order, renumbers decision records that collide, gives the unreleased changelog entries of each " +
			"bookmark their versions in a commit \"merge-queue: give versions to <bookmark>\", and runs --checks. A " +
			"bookmark that conflicts, fails the checks, or comes after one that is left out is left out: its tasks " +
			"get a note, and a conflict a message to --conflicts-to. The result is the bookmark --name (default " +
			"merge-queue/<date>). It never moves --onto, and never pushes. With --from-message, the set of tasks " +
			"comes from the data of the message on standard input, as a tool agent gives it: " +
			`{"tasks": [...], "onto": ..., "name": ..., "checks": ...}.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			request := runRequest{Tasks: args, Onto: onto, Name: name, Checks: checks}
			if fromMessage {
				if len(args) > 0 {
					return errors.New("Give the tasks as arguments or in a message (--from-message), not both.")
				}
				var m mq.Message
				if err := json.NewDecoder(cmd.InOrStdin()).Decode(&m); err != nil {
					return fmt.Errorf("Standard input is not a message: %v.", err)
				}
				var data runRequest
				if len(m.Data) == 0 || json.Unmarshal(m.Data, &data) != nil || len(data.Tasks) == 0 {
					return errors.New(`The message has no tasks: give them in its data, for example {"tasks": ["fix-login-bug"]}.`)
				}
				request.Tasks = data.Tasks
				if data.Onto != "" {
					request.Onto = data.Onto
				}
				if data.Name != "" {
					request.Name = data.Name
				}
				if data.Checks != "" {
					request.Checks = data.Checks
				}
			}
			if len(request.Tasks) == 0 {
				return errors.New("Give the tasks to merge.")
			}
			r, err := findRepo()
			if err != nil {
				return err
			}
			if r.VCS == repo.JJ {
				// The merge happens with git. jj exports its bookmarks to git by itself only in a
				// colocated workspace, so export them first: the git branches must be the bookmarks.
				exporter := exec.Command("jj", "git", "export")
				exporter.Dir = r.Root
				if out, err := exporter.CombinedOutput(); err != nil {
					return fmt.Errorf("jj git export: %v: %s", err, strings.TrimSpace(string(out)))
				}
			}
			if request.Name == "" {
				if request.Name, err = freeName(r, "merge-queue/"+time.Now().Format("2006-01-02")); err != nil {
					return err
				}
			}
			tasks, err := resolveTasks(w.directory(), request.Tasks)
			if err != nil {
				return err
			}
			report, err := mergequeue.Run(cmd.Context(), mergequeue.Options{
				Repo: r, Tasks: tasks, Base: request.Onto, Name: request.Name, Checks: request.Checks,
				Out: cmd.ErrOrStderr(), Now: time.Now,
			})
			if err != nil {
				return err
			}
			if r.VCS == repo.JJ {
				// jj sees the new git branch as a bookmark.
				importer := exec.Command("jj", "git", "import")
				importer.Dir = r.Root
				importer.Run()
			}
			recordLeftOut(cmd, w, r, report, conflictsTo)
			announce(cmd, r, report)
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), report)
			}
			printReport(cmd.OutOrStdout(), cmd.ErrOrStderr(), report)
			return nil
		},
	}
	cmd.Flags().StringVar(&onto, "onto", "main", "The base of the new line. It does not change.")
	cmd.Flags().StringVar(&name, "name", "", "The bookmark of the result. Default: merge-queue/<date>, with -2, -3, ... when it exists.")
	cmd.Flags().StringVar(&checks, "checks", "", "A command line that must succeed after each bookmark, in the scratch worktree, for example the tests.")
	cmd.Flags().StringVar(&conflictsTo, "conflicts-to", "engineer", "The role that gets a message about each conflict, if the message queue runs.")
	cmd.Flags().BoolVar(&fromMessage, "from-message", false, "Read the set of tasks from the message on standard input, as a tool agent gives it.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the report as JSON.")
	return cmd
}

// freeName returns prefix, or prefix-2, prefix-3, and so on: the first that is not a branch yet.
func freeName(r repo.Repo, prefix string) (string, error) {
	gitDir, err := r.GitDir()
	if err != nil {
		return "", err
	}
	for i := 1; ; i++ {
		name := prefix
		if i > 1 {
			name = fmt.Sprintf("%s-%d", prefix, i)
		}
		check := exec.Command("git", "--git-dir="+gitDir, "rev-parse", "--verify", "--quiet", "refs/heads/"+name)
		if check.Run() != nil {
			return name, nil
		}
	}
}

// recordLeftOut adds a note to each task of a bookmark that the run left out, and sends a message
// about each conflict to the role conflictsTo. The tool agent that runs the merge queue sends the
// report to the sender of the request.
func recordLeftOut(cmd *cobra.Command, w workspace, r repo.Repo, report mergequeue.Report, conflictsTo string) {
	date := time.Now().Format("2006-01-02")
	client, queueErr := mq.Connect(r.SharedDir())
	for _, left := range report.LeftOut {
		if left.Reason == "nothing" {
			continue
		}
		detail := left.Detail
		if len(left.Files) > 0 {
			detail += ". Files: " + strings.Join(left.Files, ", ")
		}
		for _, task := range left.Tasks {
			note := fmt.Sprintf("%s: the merge queue left the work of this task (%s) out of %s: %s.", date, left.Bookmark, report.Name, strings.TrimSuffix(detail, "."))
			if _, err := sisyphus.Run(w.directory(), "edit", task, "notes", "--append", "--", note); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not record in %s why its work was left out: %v\n", task, err)
			}
		}
		if left.Reason != "conflict" || conflictsTo == "" {
			continue
		}
		if queueErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: the conflict of %s needs %s or a human, but no message was sent: %v\n", left.Bookmark, conflictsTo, queueErr)
			continue
		}
		data, _ := json.Marshal(left)
		body := fmt.Sprintf("The merge queue left %s out of %s: %s. It does not resolve a conflict that needs judgment.", left.Bookmark, report.Name, detail)
		if _, err := client.Send(mq.Message{From: "merge-queue", To: mq.Address{Role: conflictsTo}, Kind: "conflict", Issue: first(left.Tasks), Body: body, Data: data}); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not send the conflict of %s to %s: %v\n", left.Bookmark, conflictsTo, err)
		}
	}
}

// announce broadcasts merge-queue-done for a run that made a line, so that the roles that check a
// line before a human moves the base, for example the release manager, can start. With no message
// queue, it does nothing.
func announce(cmd *cobra.Command, r repo.Repo, report mergequeue.Report) {
	if !report.Created {
		return
	}
	client, err := mq.Connect(r.SharedDir())
	if err != nil {
		return
	}
	merged, leftOut := []string{}, []string{}
	for _, m := range report.Merged {
		merged = append(merged, m.Bookmark)
	}
	for _, l := range report.LeftOut {
		leftOut = append(leftOut, l.Bookmark)
	}
	data, _ := json.Marshal(map[string]any{"bookmark": report.Name, "base": report.Base, "head": report.Head, "merged": merged, "left-out": leftOut})
	body := fmt.Sprintf("The merge queue made %s on %s: %d bookmarks merged, %d left out. A human checks it before %s moves.", report.Name, report.Base, len(merged), len(leftOut), report.Base)
	if _, err := client.Send(mq.Message{ID: "merge-queue-done-" + report.Name, From: "merge-queue", To: mq.Address{Broadcast: true}, Kind: "merge-queue-done", Body: body, Data: data}); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not broadcast merge-queue-done for %s: %v\n", report.Name, err)
	}
}

func first(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}

func printReport(out, errOut io.Writer, report mergequeue.Report) {
	if report.Created {
		fmt.Fprintf(out, "%s at %.12s, on %s (%.12s):\n", report.Name, report.Head, report.Base, report.BaseID)
	} else {
		fmt.Fprintf(out, "Nothing merged: %s was not created. %s (%.12s) does not change.\n", report.Name, report.Base, report.BaseID)
	}
	for _, m := range report.Merged {
		fmt.Fprintf(out, "%d. %s (%s): %d %s\n", m.Place, m.Bookmark, strings.Join(m.Tasks, ", "), m.Commits, plural(m.Commits, "change", "changes"))
		for _, v := range m.Versions {
			fmt.Fprintf(out, "   %s -> %s\n", v.Entry, v.Version)
		}
		for _, rn := range m.Renumbered {
			fmt.Fprintf(out, "   %s -> %s\n", rn.From, rn.To)
		}
	}
	for _, left := range report.LeftOut {
		fmt.Fprintf(out, "left out: %s (%s): %s: %s\n", left.Bookmark, strings.Join(left.Tasks, ", "), left.Reason, strings.ReplaceAll(left.Detail, "\n", "\n    "))
		if len(left.Files) > 0 {
			fmt.Fprintf(out, "    files: %s\n", strings.Join(left.Files, ", "))
		}
	}
	for _, warning := range report.Warnings {
		fmt.Fprintf(errOut, "Warning: %s.\n", warning)
	}
}
