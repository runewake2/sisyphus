package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/referendum"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
	"github.com/runewake2/sisyphus/hephaestus/internal/toolagent"
)

// Metadata keys of an issue with a referendum.
const (
	referendumKey = "referendum"
	recordKey     = "referendum-record"
)

// roleVariable names the role of the agent that runs hephaestus, as a worker sets it.
const roleVariable = "HEPHAESTUS_ROLE"

func referendumCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "referendum",
		Short: "Let role agents vote on an implementation, and record the result in the issue.",
	}
	cmd.AddCommand(callCommand(w, findRepo), decideCommand(w, findRepo))
	return cmd
}

func callCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var question, bookmark, onto, runner string
	var required, optional []string
	var maxRounds, parallel int
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "call <issue>",
		Short: "Call a referendum on the implementation of an issue, and run its rounds to the result.",
		Long: "Call a referendum on the implementation of an issue ([[referendum]]): the bookmark of the issue (or " +
			"--bookmark), compared with --onto. Each round gives every reviewer the contents at the same time, each in " +
			"a scratch git worktree of its own, so that no reviewer sees another ballot. A reviewer votes yes, no, " +
			"modify, or (non-required only) abstain. A tool reviewer votes with the exit code of its command, and " +
			"modifies with the changes that its command makes. An AI reviewer runs --runner (or " + toolagent.RunnerVariable +
			") with the prompt of its ballot, and ends its output with its ballot as JSON.\n\n" +
			"Each modification goes to the gate, one at a time: every required reviewer must accept it. The first that " +
			"they accept makes a new version, on the branch referendum/<id>/v<n>, and every reviewer votes again on it. " +
			"The referendum passes when every required reviewer votes yes, and fails when one votes no. A no of a " +
			"non-required reviewer is a concern. After --max-rounds, or when a reviewer gives no valid ballot, the " +
			"result is no-result, and the issue gets the tag needs-human.\n\n" +
			"The record: metadata referendum (round-<n>, then the result) and referendum-record of the issue, a Notes " +
			"line for each round and for the result, and the full record in referendums/<id>.md. When it passes on a " +
			"modified version, the issue edits of that version are applied, and the bookmark of the issue becomes the " +
			"branch of that version. A human, or an agent whose role file has can-call-referendum: true, can call one. " +
			"The exit code is 0 only when the referendum passes.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue := args[0]
			if strings.TrimSpace(question) == "" {
				return errors.New(`Give the question of the referendum with --question, for example "Merge samw/ai/fix-login as the fix of fix-login-bug?".`)
			}
			rolesDir, err := w.roles()
			if err != nil {
				return err
			}
			calledBy, err := caller(rolesDir)
			if err != nil {
				return err
			}
			var reviewers []referendum.Reviewer
			for _, role := range splitList(required) {
				reviewers = append(reviewers, referendum.Reviewer{Role: role, Required: true})
			}
			for _, role := range splitList(optional) {
				reviewers = append(reviewers, referendum.Reviewer{Role: role})
			}
			if err := referendum.Check(reviewers); err != nil {
				return err
			}
			roleMap, err := referendum.RoleMap(rolesDir, reviewers)
			if err != nil {
				return err
			}
			if runner == "" {
				runner = os.Getenv(toolagent.RunnerVariable)
			}
			for _, reviewer := range reviewers {
				if err := toolagent.Validate(roleMap[reviewer.Role], runner); err != nil {
					return err
				}
			}

			detail, err := sisyphus.Show(w.directory(), issue)
			if err != nil {
				return err
			}
			if state := detail.Metadata[referendumKey]; strings.HasPrefix(state, "round-") {
				return fmt.Errorf("The issue %s has a referendum that runs (%s=%s, record %s). Wait for its result. If it stopped, a human decides it: hephaestus referendum decide %s passed|failed --reason <why>.", detail.Name, referendumKey, state, detail.Metadata[recordKey], detail.Name)
			}
			if bookmark == "" {
				bookmark = detail.Bookmark
			}
			if bookmark == "" {
				return fmt.Errorf("The issue %s has no bookmark. Give the implementation with --bookmark.", detail.Name)
			}

			r, err := findRepo()
			if err != nil {
				return err
			}
			if err := exportJJ(r); err != nil {
				return err
			}
			gitDir, err := r.GitDir()
			if err != nil {
				return err
			}
			commit, err := revParse(gitDir, bookmark)
			if err != nil {
				return fmt.Errorf("The bookmark %q does not exist.", bookmark)
			}
			baseID, err := revParse(gitDir, onto)
			if err != nil {
				return fmt.Errorf("The base %q does not exist.", onto)
			}

			now := time.Now()
			ref := &referendum.Referendum{
				Date: now.Format("2006-01-02"), Issue: detail.Name, Question: strings.TrimSpace(question), CalledBy: calledBy,
				Bookmark: bookmark, Base: onto, BaseID: baseID, Reviewers: reviewers, MaxRounds: maxRounds,
			}
			ref.ID = freeRecordID(r.Root, gitDir, ref.Date+"-"+path.Base(detail.Name))
			voter := &referendum.RoleVoter{GitDir: gitDir, Roles: roleMap, Runner: runner, Out: cmd.ErrOrStderr()}
			v1 := referendum.Version{Number: 1, Commit: commit}
			if v1.Ref, err = voter.NewVersion(ref, v1); err != nil {
				return err
			}
			ref.Versions = []referendum.Version{v1}

			rec := recorder{dir: w.directory(), root: r.Root, ref: ref, out: cmd.ErrOrStderr()}
			if err := rec.write(); err != nil {
				return err
			}
			rec.sisyphus("update", ref.Issue, "--metadata", referendumKey+"=round-1", "--metadata", recordKey+"="+ref.ID)
			fmt.Fprintf(cmd.ErrOrStderr(), "referendum %s: %s on %s (%.12s), %d reviewers\n", ref.ID, ref.Issue, bookmark, commit, len(reviewers))

			runErr := referendum.Run(cmd.Context(), ref, voter, referendum.Options{Parallel: parallel, On: rec.on})
			importJJ(r)
			if runErr != nil {
				rec.sisyphus("update", ref.Issue, "--metadata", referendumKey+"="+referendum.NoResult, "--add-tags", toolagent.NeedsHuman)
				rec.note(fmt.Sprintf("%s, the referendum stopped: %v. A human decides. Record: [[%s]].", ref.Date, runErr, ref.ID))
				return runErr
			}
			rec.notify(r, detail.Owner)
			if asJSON {
				if err := writeJSON(cmd.OutOrStdout(), ref); err != nil {
					return err
				}
			} else {
				printReferendum(cmd.OutOrStdout(), ref)
			}
			if ref.Result != referendum.Passed {
				return fmt.Errorf("The referendum on %s has the result %s. %s", ref.Issue, ref.Result, ref.Reasons)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&question, "question", "q", "", `The question, for example "Merge samw/ai/fix-login as the fix of fix-login-bug?".`)
	cmd.Flags().StringSliceVar(&required, "required", nil, "The required reviewers: roles, comma-separated. Every one must vote yes.")
	cmd.Flags().StringSliceVar(&optional, "optional", nil, "The non-required reviewers: roles, comma-separated. Their no is a concern.")
	cmd.Flags().StringVar(&bookmark, "bookmark", "", "The implementation. Default: the bookmark of the issue.")
	cmd.Flags().StringVar(&onto, "onto", "main", "The current implementation, to compare with.")
	cmd.Flags().StringVar(&runner, "runner", "", "The command that runs the model of an AI reviewer, for example \"claude -p --agent {role}\". Default: "+toolagent.RunnerVariable+".")
	cmd.Flags().IntVar(&maxRounds, "max-rounds", referendum.DefaultMaxRounds, "The maximum number of rounds. Then a human decides.")
	cmd.Flags().IntVar(&parallel, "parallel", 0, "The maximum number of reviewers that run at the same time. 0: all of them.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the full referendum as JSON.")
	return cmd
}

func decideCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	var reason string
	cmd := &cobra.Command{
		Use:   "decide <issue> passed|failed",
		Short: "Decide the referendum of an issue as a human: after no result, or to overrule its result.",
		Long: "Decide the last referendum of an issue as a human: when it has no result, or to overrule its result. " +
			"It sets metadata referendum of the issue, adds a Notes line, and adds a section \"Human decision\" to the " +
			"record, so that the record shows that a human decided. An agent cannot decide.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			issue, result := args[0], args[1]
			if result != referendum.Passed && result != referendum.Failed {
				return fmt.Errorf("The decision is passed or failed, not %q.", result)
			}
			if role := os.Getenv(roleVariable); role != "" {
				return fmt.Errorf("Only a human decides a referendum, and this runs as the role %s (%s).", role, roleVariable)
			}
			if strings.TrimSpace(reason) == "" {
				return errors.New("Give the reason of the decision with --reason.")
			}
			detail, err := sisyphus.Show(w.directory(), issue)
			if err != nil {
				return err
			}
			id := detail.Metadata[recordKey]
			if id == "" {
				return fmt.Errorf("The issue %s has no referendum: its metadata has no %s.", detail.Name, recordKey)
			}
			r, err := findRepo()
			if err != nil {
				return err
			}
			file := filepath.Join(r.Root, referendum.Dir, id+".md")
			record, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("The record of the referendum %s is missing: %v", id, err)
			}
			date := time.Now().Format("2006-01-02")
			previous := detail.Metadata[referendumKey]
			section := fmt.Sprintf("\n## Human decision\n\n%s: a human decided **%s** (the result was %s). %s\n", date, result, previous, strings.TrimSpace(reason))
			if err := os.WriteFile(file, append(record, []byte(section)...), 0o644); err != nil {
				return err
			}
			if _, err := sisyphus.Run(w.directory(), "update", detail.Name, "--metadata", referendumKey+"="+result); err != nil {
				return err
			}
			note := fmt.Sprintf("%s, a human decided the referendum: %s (the result was %s). %s Record: [[%s]].", date, result, previous, strings.TrimSpace(reason), id)
			if _, err := sisyphus.Run(w.directory(), "edit", detail.Name, "notes", "--append", "--", note); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: the referendum %s is %s, by a human decision.\n", detail.Name, id, result)
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Why the human decided so.")
	return cmd
}

// caller returns who calls the referendum: the role of the agent, when it runs as a worker, which
// its role file must allow (R4 of [[referendum]]), or a human.
func caller(rolesDir string) (string, error) {
	name := os.Getenv(roleVariable)
	if name == "" {
		if owner := os.Getenv(ownerVariable); owner != "" {
			return "a human (" + owner + ")", nil
		}
		return "a human", nil
	}
	role, err := roles.Find(rolesDir, name)
	if err != nil {
		return "", err
	}
	if !role.CanCallReferendum {
		return "", fmt.Errorf("The role %s cannot call a referendum: its role file does not have can-call-referendum: true. A human allows a role to call one.", role.Name)
	}
	who := "the role " + role.Name
	if owner := os.Getenv(ownerVariable); owner != "" {
		who += " (" + owner + ")"
	}
	return who, nil
}

func splitList(values []string) []string {
	var result []string
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			if item = strings.TrimSpace(item); item != "" {
				result = append(result, item)
			}
		}
	}
	return result
}

// exportJJ exports the jj bookmarks to git: the referendum reads them with git.
func exportJJ(r repo.Repo) error {
	if r.VCS != repo.JJ {
		return nil
	}
	exporter := exec.Command("jj", "git", "export")
	exporter.Dir = r.Root
	if out, err := exporter.CombinedOutput(); err != nil {
		return fmt.Errorf("jj git export: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// importJJ lets jj see the branches of the versions as bookmarks.
func importJJ(r repo.Repo) {
	if r.VCS == repo.JJ {
		importer := exec.Command("jj", "git", "import")
		importer.Dir = r.Root
		importer.Run()
	}
}

func revParse(gitDir, rev string) (string, error) {
	out, err := exec.Command("git", "--git-dir="+gitDir, "rev-parse", "--verify", "--quiet", rev+"^{commit}").Output()
	return strings.TrimSpace(string(out)), err
}

// freeRecordID returns id, or id-2, id-3, and so on: the first with no record and no branches.
func freeRecordID(root, gitDir, id string) string {
	for i := 1; ; i++ {
		candidate := id
		if i > 1 {
			candidate = fmt.Sprintf("%s-%d", id, i)
		}
		if _, err := os.Stat(filepath.Join(root, referendum.Dir, candidate+".md")); err == nil {
			continue
		}
		if _, err := revParse(gitDir, "refs/heads/referendum/"+candidate+"/v1"); err == nil {
			continue
		}
		return candidate
	}
}

// recorder records a referendum in the repo while it runs.
type recorder struct {
	dir, root string
	ref       *referendum.Referendum
	out       io.Writer
}

func (rec recorder) file() string {
	return filepath.Join(rec.root, referendum.Dir, rec.ref.ID+".md")
}

func (rec recorder) write() error {
	if err := os.MkdirAll(filepath.Dir(rec.file()), 0o755); err != nil {
		return err
	}
	return os.WriteFile(rec.file(), []byte(referendum.Markdown(rec.ref)), 0o644)
}

func (rec recorder) sisyphus(args ...string) {
	if _, err := sisyphus.Run(rec.dir, args...); err != nil {
		fmt.Fprintf(rec.out, "Warning: could not record the referendum in %s (sisyphus %s): %v\n", rec.ref.Issue, args[0], err)
	}
}

func (rec recorder) note(text string) {
	rec.sisyphus("edit", rec.ref.Issue, "notes", "--append", "--", text)
}

// on records each event of the referendum.
func (rec recorder) on(e referendum.Event) {
	if err := rec.write(); err != nil {
		fmt.Fprintf(rec.out, "Warning: could not write the record %s: %v\n", rec.file(), err)
	}
	ref := rec.ref
	switch e.Kind {
	case "round":
		fmt.Fprintf(rec.out, "referendum %s: round %d on v%d: %s\n", ref.ID, e.Round.Number, e.Round.Version, e.Round.Outcome)
		rec.note(referendum.RoundNote(ref, *e.Round, ref.ID))
		rec.sisyphus("update", ref.Issue, "--metadata", fmt.Sprintf("%s=round-%d", referendumKey, e.Round.Number+1))
	case "result":
		fmt.Fprintf(rec.out, "referendum %s: %s\n", ref.ID, ref.Result)
		final := ref.Current()
		if ref.Result == referendum.Passed {
			for _, edit := range final.IssueEdits {
				rec.sisyphus("edit", ref.Issue, edit.Section, "--", edit.Text)
			}
			if final.Number > 1 {
				rec.sisyphus("update", ref.Issue, "--bookmark", final.Ref)
			}
		}
		args := []string{"update", ref.Issue, "--metadata", referendumKey + "=" + ref.Result}
		if ref.Result == referendum.NoResult {
			args = append(args, "--add-tags", toolagent.NeedsHuman)
		}
		rec.sisyphus(args...)
		rec.note(referendum.ResultNote(ref, ref.ID))
	}
}

// notify tells the owner of the issue the result, when the message queue runs.
func (rec recorder) notify(r repo.Repo, owner string) {
	if owner == "" {
		return
	}
	client, err := mq.Connect(r.SharedDir())
	if err != nil {
		return // no message queue: the result is in the issue
	}
	data, _ := json.Marshal(map[string]string{"referendum": rec.ref.ID, "result": rec.ref.Result})
	body := fmt.Sprintf("The referendum %s on %s has the result %s. %s", rec.ref.ID, rec.ref.Issue, rec.ref.Result, rec.ref.Reasons)
	if _, err := client.Send(mq.Message{ID: "referendum-" + rec.ref.ID, From: "hephaestus", To: mq.Address{Worker: owner}, Kind: "referendum", Issue: rec.ref.Issue, Body: body, Data: data}); err != nil {
		fmt.Fprintf(rec.out, "Warning: could not send the result to %s: %v\n", owner, err)
	}
}

func printReferendum(out io.Writer, ref *referendum.Referendum) {
	fmt.Fprintf(out, "Referendum %s on %s: %s\n", ref.ID, ref.Issue, ref.Result)
	for _, round := range ref.Rounds {
		var votes []string
		for _, b := range round.Ballots {
			vote := b.Vote
			if b.Error != "" {
				vote = "no ballot"
			}
			votes = append(votes, b.Role+" "+vote)
		}
		fmt.Fprintf(out, "  round %d (v%d): %s. %s\n", round.Number, round.Version, strings.Join(votes, ", "), round.Outcome)
	}
	fmt.Fprintf(out, "  %s\n", ref.Reasons)
	for _, concern := range ref.Concerns {
		fmt.Fprintf(out, "  concern: %s\n", concern)
	}
	fmt.Fprintf(out, "  record: %s\n", filepath.ToSlash(filepath.Join(referendum.Dir, ref.ID+".md")))
}
