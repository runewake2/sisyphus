// Package watch finds the events that the roles of hephaestus listen for in the state of the issues
// and their bookmarks: an issue that closes (issue-closed), and the work of a bookmark that is ready
// for review (bookmark-ready). It compares the state with the state that it saw before, so that each
// event is sent once, and the flow does not depend on a worker that remembers to send it.
package watch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
)

// File is the name of the state of the watcher in the shared directory of the repo.
const File = "watch.json"

// Kinds of the events.
const (
	BookmarkReady = "bookmark-ready"
	IssueClosed   = "issue-closed"
)

// Seen is what the watcher saw of one issue.
type Seen struct {
	State    string `json:"state"`
	Bookmark string `json:"bookmark,omitempty"`
	// Head is the commit of the bookmark when the issue was closed, or later.
	Head string `json:"head,omitempty"`
}

// State is what the watcher saw at its last run.
type State struct {
	Since  time.Time       `json:"since"`
	Issues map[string]Seen `json:"issues"`
}

// Event is one notice to broadcast.
type Event struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Issue    string `json:"issue"`
	Bookmark string `json:"bookmark,omitempty"`
	Head     string `json:"head,omitempty"`
	Base     string `json:"base"`
	Body     string `json:"body"`
}

// Repo is what the watcher needs from the repo: the commit of a bookmark, and whether the base has a
// commit already.
type Repo interface {
	Head(bookmark string) (string, bool)
	InBase(commit string) bool
}

// Plan returns the events of the issues since prev, and the new state. With no earlier state (the
// first run), it returns no event: it only records what it sees, so that it does not announce the
// whole history of the repo.
func Plan(prev *State, issues []sisyphus.Issue, repo Repo, base string, now time.Time) (State, []Event) {
	next := State{Since: now, Issues: map[string]Seen{}}
	if prev != nil {
		next.Since = prev.Since
	}
	var events []Event
	issues = append([]sisyphus.Issue(nil), issues...) // sort a copy: the caller keeps its order
	sort.Slice(issues, func(i, j int) bool { return issues[i].Name < issues[j].Name })
	for _, issue := range issues {
		seen := Seen{State: issue.State, Bookmark: issue.Bookmark}
		head, hasHead := "", false
		if issue.Bookmark != "" {
			head, hasHead = repo.Head(issue.Bookmark)
		}
		if hasHead {
			seen.Head = head
		}
		next.Issues[issue.Name] = seen
		if prev == nil || issue.State != "closed" {
			continue
		}
		before, known := prev.Issues[issue.Name]
		closedNow := !known || before.State != "closed"
		moved := known && before.State == "closed" && hasHead && before.Head != head
		if closedNow || moved {
			events = append(events, Notices(issue.Name, issue.Bookmark, head, base, hasHead && !repo.InBase(head), closedNow)...)
		}
	}
	return next, events
}

// Notices returns the notices of a close: bookmark-ready when the bookmark has work for review
// (review), and issue-closed when the issue closed now (closed). The ids come from the issue and the
// head, so that the watcher and a worker that announce the same close make one message.
func Notices(issue, bookmark, head, base string, review, closed bool) []Event {
	var events []Event
	if review {
		body := fmt.Sprintf("The work of %s is ready for review: %s at %.12s, on %s.", issue, bookmark, head, base)
		if !closed {
			body = fmt.Sprintf("The work of %s changed after its close: %s is now at %.12s, on %s.", issue, bookmark, head, base)
		}
		events = append(events, Event{
			ID: fmt.Sprintf("%s-%s-%.12s", BookmarkReady, flat(issue), head), Kind: BookmarkReady,
			Issue: issue, Bookmark: bookmark, Head: head, Base: base, Body: body,
		})
	}
	if closed {
		events = append(events, Event{
			ID: fmt.Sprintf("%s-%s", IssueClosed, flat(issue)), Kind: IssueClosed,
			Issue: issue, Bookmark: bookmark, Head: head, Base: base,
			Body: fmt.Sprintf("The issue %s closed.", issue),
		})
	}
	return events
}

// Message returns the broadcast of a notice.
func (e Event) Message(from string) mq.Message {
	data, _ := json.Marshal(map[string]string{"issue": e.Issue, "bookmark": e.Bookmark, "head": e.Head, "base": e.Base})
	return mq.Message{ID: e.ID, From: from, To: mq.Address{Broadcast: true}, Kind: e.Kind, Issue: e.Issue, Body: e.Body, Data: data}
}

// flat returns a name with "-" for "/", for an id.
func flat(name string) string {
	out := []rune(name)
	for i, r := range out {
		if r == '/' {
			out[i] = '-'
		}
	}
	return string(out)
}

// Load reads the state in the shared directory dir. No state yet gives nil.
func Load(dir string) (*State, error) {
	data, err := os.ReadFile(filepath.Join(dir, File))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("%s is not a valid state of the watcher: %v. Remove it to start again from now.", filepath.Join(dir, File), err)
	}
	if state.Issues == nil {
		state.Issues = map[string]Seen{}
	}
	return &state, nil
}

// Save writes the state in the shared directory dir, through a new file, so that a crash leaves the
// old state or the new one.
func Save(dir string, state State) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, File+".tmp")
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, File))
}
