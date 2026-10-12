package watch

import (
	"strings"
	"testing"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
)

type fakeRepo struct {
	heads  map[string]string
	inBase map[string]bool
}

func (f fakeRepo) Head(b string) (string, bool) { h, ok := f.heads[b]; return h, ok }
func (f fakeRepo) InBase(c string) bool         { return f.inBase[c] }

func kinds(events []Event) string {
	var parts []string
	for _, e := range events {
		parts = append(parts, e.Kind+" "+e.Issue)
	}
	return strings.Join(parts, ", ")
}

func TestTheWatcherAnnouncesEachCloseOnce(t *testing.T) {
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	repo := fakeRepo{heads: map[string]string{"ai/fix": "aaa", "ai/old": "ooo"}, inBase: map[string]bool{"ooo": true}}
	issues := []sisyphus.Issue{
		{Name: "fix-login-bug", State: "in-progress", Bookmark: "ai/fix"},
		{Name: "old-work", State: "closed", Bookmark: "ai/old"},
		{Name: "no-bookmark", State: "open"},
	}

	// The first run records the state, and announces nothing: old closes are history.
	state, events := Plan(nil, issues, repo, "main", now)
	if len(events) != 0 {
		t.Fatalf("want no event on the first run, got %s", kinds(events))
	}

	issues[0].State, issues[2].State = "closed", "closed"
	state, events = Plan(&state, issues, repo, "main", now)
	if got := kinds(events); got != "bookmark-ready fix-login-bug, issue-closed fix-login-bug, issue-closed no-bookmark" {
		t.Fatalf("got %s", got)
	}
	if events[0].ID != "bookmark-ready-fix-login-bug-aaa" || events[0].Head != "aaa" || events[1].ID != "issue-closed-fix-login-bug" {
		t.Errorf("got %+v", events[:2])
	}

	state, events = Plan(&state, issues, repo, "main", now)
	if len(events) != 0 {
		t.Fatalf("want each close announced once, got %s", kinds(events))
	}

	repo.heads["ai/fix"] = "bbb" // the fixes of a review
	state, events = Plan(&state, issues, repo, "main", now)
	if got := kinds(events); got != "bookmark-ready fix-login-bug" || events[0].ID != "bookmark-ready-fix-login-bug-bbb" {
		t.Fatalf("got %s", got)
	}

	repo.inBase["ccc"], repo.heads["ai/fix"] = true, "ccc" // merged: nothing to review
	_, events = Plan(&state, issues, repo, "main", now)
	if len(events) != 0 {
		t.Fatalf("want no review of merged work, got %s", kinds(events))
	}
}
