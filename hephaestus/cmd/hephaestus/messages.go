package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/repo"
	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
)

// ownerVariable names the worker of a command when --from or --as is not given: the owner name that
// hephaestus gives each worker, as sisyphus uses it.
const ownerVariable = "SISYPHUS_OWNER"

func messagesCommand(w workspace, findRepo func() (repo.Repo, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Send and receive messages between workers, through the message queue of the repo.",
		Long: "The message queue relays messages between workers (humans, AI agents, and tool agents) in every " +
			"workspace of the repo: to one worker, to one worker of a role, to every worker of a role, or as a " +
			"broadcast to every listener. A message stays pending until its receiver acknowledges it, and comes again " +
			"after the lease if it does not; the messages of one sender keep their order. Every event is kept in an " +
			"append-only log. Start the service with: hephaestus messages serve",
	}
	connect := func() (*mq.Client, error) {
		r, err := findRepo()
		if err != nil {
			return nil, err
		}
		return mq.Connect(r.SharedDir())
	}
	cmd.AddCommand(serveCommand(findRepo), sendCommand(w, connect), receiveCommand(connect), ackCommand(connect),
		renewCommand(connect), workersCommand(connect), logCommand(findRepo), listenCommand(connect))
	return cmd
}

func serveCommand(findRepo func() (repo.Repo, error)) *cobra.Command {
	var lease time.Duration
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the message queue of the repo until it is stopped.",
		Long: "Run the message queue service of the repo: it listens on 127.0.0.1, and writes its address and a token " +
			"to mq.json in the hephaestus directory that every workspace of the repo shares (in .jj/repo or the git " +
			"common directory). It keeps every event in messages.log there, and reads the log again when it starts. " +
			"Stop it with Ctrl-C or SIGTERM.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			dir := r.SharedDir()
			server, err := mq.Open(dir, lease)
			if err != nil {
				return err
			}
			if err := server.Start(); err != nil {
				server.Close()
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "The message queue listens on %s. Log: %s\n", server.Address(), filepath.Join(dir, mq.LogFile))
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			<-ctx.Done()
			return server.Close()
		},
	}
	cmd.Flags().DurationVar(&lease, "lease", mq.DefaultLease, "How long a receiver has to acknowledge a message before it comes again.")
	return cmd
}

// address reads the receiver flags of send: exactly one of --to, --role, and --broadcast.
func address(to, role string, all, broadcast bool) (mq.Address, error) {
	given := 0
	for _, set := range []bool{to != "", role != "", broadcast} {
		if set {
			given++
		}
	}
	if given != 1 {
		return mq.Address{}, errors.New("Give exactly one receiver: --to <worker>, --role <role> (with --all for every worker of the role), or --broadcast.")
	}
	if all && role == "" {
		return mq.Address{}, errors.New("--all is only for --role: every worker of the role.")
	}
	return mq.Address{Worker: to, Role: role, All: all, Broadcast: broadcast}, nil
}

// familyWarning returns a warning when a message to one worker of a role names a family, for example
// formatter: no worker has the role of a family, so the message reaches no worker. The roles come
// from the role files; with none to read, it says nothing.
func familyWarning(w workspace, to mq.Address) string {
	if to.Role == "" || to.All {
		return ""
	}
	dir, err := w.roles()
	if err != nil {
		return ""
	}
	all, _ := roles.Load(dir)
	var members []string
	for _, role := range all {
		if role.Name == to.Role {
			return ""
		}
		if strings.HasPrefix(role.Name, to.Role+"/") {
			members = append(members, role.Name)
		}
	}
	if len(members) == 0 {
		return ""
	}
	return fmt.Sprintf("%s is a family of roles (%s), not a role. A message to one worker of a family reaches no worker: it waits for a worker of the role %s, and none has it. Send it to one role of the family, for example --role %s, or to every worker of every role of the family with --role %s --all.",
		to.Role, strings.Join(members, ", "), to.Role, members[0], to.Role)
}

// worker returns the name of the worker of a command: the flag, or SISYPHUS_OWNER.
func worker(flag, name string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if owner := strings.TrimSpace(os.Getenv(ownerVariable)); owner != "" {
		return owner, nil
	}
	return "", fmt.Errorf("Give --%s <worker>, or set %s.", name, ownerVariable)
}

func sendCommand(w workspace, connect func() (*mq.Client, error)) *cobra.Command {
	var from, to, role, kind, issue, data, id string
	var all, broadcast, asJSON bool
	cmd := &cobra.Command{
		Use:   "send [<body> | -]",
		Short: "Send a message to a worker, a role, or every listener.",
		Long: "Send a message. The body is the argument, or standard input with -. --data adds JSON for a tool agent. " +
			"--id gives the message an id, so that a send that you repeat after a failure does not make a second message.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sender, err := worker(from, "from")
			if err != nil {
				return err
			}
			receiver, err := address(to, role, all, broadcast)
			if err != nil {
				return err
			}
			body := ""
			if len(args) == 1 {
				body = args[0]
				if body == "-" {
					input, err := io.ReadAll(cmd.InOrStdin())
					if err != nil {
						return err
					}
					body = strings.TrimRight(string(input), "\n")
				}
			}
			if data != "" && !json.Valid([]byte(data)) {
				return errors.New("--data is not valid JSON.")
			}
			client, err := connect()
			if err != nil {
				return err
			}
			sent, err := client.Send(mq.Message{ID: id, From: sender, To: receiver, Kind: kind, Issue: issue, Body: body, Data: json.RawMessage(data)})
			if err != nil {
				return err
			}
			if warning := familyWarning(w, receiver); warning != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", warning)
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), sent)
			}
			fmt.Fprintln(cmd.OutOrStdout(), sent.ID)
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&from, "from", "", "The sender: a worker name. Default: "+ownerVariable+".")
	flags.StringVar(&to, "to", "", "Send to one worker, by its owner name, for example merge-queue-1.")
	flags.StringVar(&role, "role", "", "Send to one worker of the role, for example engineer or formatter/golang. A family, for example formatter, needs --all; without it, the message reaches no worker, and send warns.")
	flags.BoolVar(&all, "all", false, "With --role: send to every worker of the role.")
	flags.BoolVar(&broadcast, "broadcast", false, "Send to every listener.")
	flags.StringVar(&kind, "kind", "info", "The kind of the message, for example task, info, or referendum. Listeners filter broadcasts by it.")
	flags.StringVar(&issue, "issue", "", "The sisyphus issue that the message is about.")
	flags.StringVar(&data, "data", "", "JSON for a tool agent.")
	flags.StringVar(&id, "id", "", "The id of the message. Default: a new id.")
	flags.BoolVar(&asJSON, "json", false, "Print the message as JSON.")
	return cmd
}

func receiveCommand(connect func() (*mq.Client, error)) *cobra.Command {
	var as, role string
	var kinds []string
	var wait time.Duration
	var max int
	var acknowledge, asJSON bool
	cmd := &cobra.Command{
		Use:   "receive",
		Short: "Print the messages that wait for a worker.",
		Long: "Print the messages that wait for a worker, oldest first. The first receive of a worker registers it with " +
			"its role and the kinds of broadcast that it listens for (all kinds if no --kind). The messages stay " +
			"pending until the worker acknowledges them (ack, or --ack here), and come again after the lease if it " +
			"does not.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := worker(as, "as")
			if err != nil {
				return err
			}
			client, err := connect()
			if err != nil {
				return err
			}
			messages, err := client.Receive(cmd.Context(), mq.Receiver{Worker: name, Role: role, Kinds: kinds, Wait: wait, Max: max})
			if err != nil {
				return err
			}
			if asJSON {
				if messages == nil {
					messages = []mq.Message{}
				}
				err = writeJSON(cmd.OutOrStdout(), messages)
			} else {
				for _, m := range messages {
					printMessage(cmd.OutOrStdout(), m)
				}
			}
			if err != nil || !acknowledge || len(messages) == 0 {
				return err
			}
			var ids []string
			for _, m := range messages {
				ids = append(ids, m.ID)
			}
			return client.Ack(name, ids)
		},
	}
	addReceiverFlags(cmd, &as, &role, &kinds)
	cmd.Flags().DurationVar(&wait, "wait", 0, "Wait up to this long for a message when none waits now, for example 30s.")
	cmd.Flags().IntVar(&max, "max", 0, "Print at most this many messages. Default: 100.")
	cmd.Flags().BoolVar(&acknowledge, "ack", false, "Acknowledge the messages after they are printed.")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print JSON.")
	return cmd
}

func addReceiverFlags(cmd *cobra.Command, as, role *string, kinds *[]string) {
	cmd.Flags().StringVar(as, "as", "", "The worker that receives: its owner name. Default: "+ownerVariable+".")
	cmd.Flags().StringVar(role, "role", "", "The role of the worker, for the messages to its role.")
	cmd.Flags().StringSliceVar(kinds, "kind", nil, "Listen only for broadcasts of these kinds, comma-separated or repeated. Default: every kind.")
}

func printMessage(out io.Writer, m mq.Message) {
	about := ""
	if m.Issue != "" {
		about = " (" + m.Issue + ")"
	}
	fmt.Fprintf(out, "%s  %s  %s -> %s  [%s]%s\n", m.ID, m.Time.Local().Format("2006-01-02 15:04:05"), m.From, m.To, m.Kind, about)
	if m.Body != "" {
		for _, line := range strings.Split(m.Body, "\n") {
			fmt.Fprintf(out, "    %s\n", line)
		}
	}
	if len(m.Data) > 0 {
		fmt.Fprintf(out, "    data: %s\n", m.Data)
	}
}

func ackCommand(connect func() (*mq.Client, error)) *cobra.Command {
	var as string
	cmd := &cobra.Command{
		Use:   "ack <id>...",
		Short: "Acknowledge messages that a worker handled, so that they do not come again.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := worker(as, "as")
			if err != nil {
				return err
			}
			client, err := connect()
			if err != nil {
				return err
			}
			return client.Ack(name, args)
		},
	}
	cmd.Flags().StringVar(&as, "as", "", "The worker that handled the messages. Default: "+ownerVariable+".")
	return cmd
}

func renewCommand(connect func() (*mq.Client, error)) *cobra.Command {
	var as string
	cmd := &cobra.Command{
		Use:   "renew <id>...",
		Short: "Renew the leases of messages that a worker still handles: its keep-alive.",
		Long: "Renew the leases of messages that the worker received and still handles, so that they do not come " +
			"again to another worker while it works, and record that the worker is alive. A worker that handles a " +
			"message for longer than the lease renews well within each lease. hephaestus agent run and messages " +
			"listen do it by themselves. It fails for a message whose lease the worker lost.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := worker(as, "as")
			if err != nil {
				return err
			}
			client, err := connect()
			if err != nil {
				return err
			}
			renewal, err := client.Renew(name, args)
			if err != nil {
				return err
			}
			if len(renewal.Lost) > 0 {
				return fmt.Errorf("%s does not hold %s now: the lease ended, and another worker may have it.", name, strings.Join(renewal.Lost, ", "))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Renewed for %s.\n", renewal.Lease)
			return nil
		},
	}
	cmd.Flags().StringVar(&as, "as", "", "The worker that handles the messages. Default: "+ownerVariable+".")
	return cmd
}

func workersCommand(connect func() (*mq.Client, error)) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "workers",
		Short: "List the workers, when each last spoke to the queue, and the messages that each holds.",
		Long: "List every worker that the queue knows: its role, the kinds that it listens for, when it last " +
			"received, acknowledged, or renewed (its keep-alive), and the messages that it holds now, with the end of " +
			"each lease. The keep-alive is not in the log: after a restart of the queue, it starts empty.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := connect()
			if err != nil {
				return err
			}
			workers, err := client.Workers()
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), workers)
			}
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(table, "worker\trole\tlast seen\tholds")
			for _, w := range workers {
				seen := "-"
				if !w.LastSeen.IsZero() {
					seen = w.LastSeen.Local().Format("2006-01-02 15:04:05")
				}
				var holds []string
				for _, h := range w.Holds {
					holds = append(holds, fmt.Sprintf("%s until %s", h.ID, h.Until.Local().Format("15:04:05")))
				}
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", w.Worker, w.Role, seen, strings.Join(holds, ", "))
			}
			return table.Flush()
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print JSON.")
	return cmd
}

func logCommand(findRepo func() (repo.Repo, error)) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Print every message of the log of the repo, oldest first.",
		Long:  "Print every message of the append-only log of the repo, oldest first. It reads the log file, so it works also when the service does not run.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := findRepo()
			if err != nil {
				return err
			}
			messages, err := mq.ReadLog(r.SharedDir())
			if err != nil {
				return err
			}
			if messages == nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "The log is empty: no message was sent in this repo.")
				messages = []mq.Message{}
				if !asJSON {
					return nil
				}
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), messages)
			}
			for _, m := range messages {
				printMessage(cmd.OutOrStdout(), m)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print JSON.")
	return cmd
}

func listenCommand(connect func() (*mq.Client, error)) *cobra.Command {
	var as, role string
	var kinds []string
	var once bool
	cmd := &cobra.Command{
		Use:   "listen -- <command> [<argument>...]",
		Short: "Run a command for each message that a worker receives, as a tool agent does.",
		Long: "Receive the messages of a worker, and run <command> for each, oldest first: the message is its standard " +
			"input as JSON, and HEPHAESTUS_MESSAGE_ID, HEPHAESTUS_MESSAGE_KIND, HEPHAESTUS_MESSAGE_FROM, " +
			"HEPHAESTUS_MESSAGE_ISSUE, and " + ownerVariable + " (the worker) are in its environment. Exit code 0 " +
			"acknowledges the message. Another exit code leaves it pending, so that it comes again after the lease. " +
			"While the command runs, the worker renews the lease of the message (its keep-alive). " +
			"A copy of a message that the command already handled is acknowledged and not run again. --once handles " +
			"the messages that wait now, and stops.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := worker(as, "as")
			if err != nil {
				return err
			}
			client, err := connect()
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			handled := map[string]bool{}
			for {
				wait := 30 * time.Second
				if once {
					wait = 0
				}
				messages, err := client.Receive(ctx, mq.Receiver{Worker: name, Role: role, Kinds: kinds, Wait: wait})
				if ctx.Err() != nil {
					return nil
				}
				if err != nil {
					return err
				}
				for _, m := range messages {
					if handled[m.ID] {
						if err := client.Ack(name, []string{m.ID}); err != nil {
							return err
						}
						continue
					}
					stopKeepAlive := mq.KeepAlive(client.Renew, name, []string{m.ID}, cmd.ErrOrStderr())
					err := runForMessage(ctx, cmd, args, name, m)
					stopKeepAlive()
					if err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s failed for message %s: %v. It comes again after the lease.\n", args[0], m.ID, err)
						continue
					}
					handled[m.ID] = true
					if err := client.Ack(name, []string{m.ID}); err != nil {
						return err
					}
				}
				if once {
					return nil
				}
			}
		},
	}
	addReceiverFlags(cmd, &as, &role, &kinds)
	cmd.Flags().BoolVar(&once, "once", false, "Handle the messages that wait now, and stop.")
	return cmd
}

func runForMessage(ctx context.Context, cmd *cobra.Command, args []string, name string, m mq.Message) error {
	input, err := json.Marshal(m)
	if err != nil {
		return err
	}
	run := exec.CommandContext(ctx, args[0], args[1:]...)
	run.Stdin = strings.NewReader(string(input) + "\n")
	run.Stdout, run.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
	run.Env = append(os.Environ(),
		"HEPHAESTUS_MESSAGE_ID="+m.ID,
		"HEPHAESTUS_MESSAGE_KIND="+m.Kind,
		"HEPHAESTUS_MESSAGE_FROM="+m.From,
		"HEPHAESTUS_MESSAGE_ISSUE="+m.Issue,
		ownerVariable+"="+name,
	)
	return run.Run()
}
