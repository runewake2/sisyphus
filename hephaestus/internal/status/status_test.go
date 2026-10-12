package status

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/runewake2/sisyphus/hephaestus/internal/mq"
	"github.com/runewake2/sisyphus/hephaestus/internal/sisyphus"
)

func contains(t *testing.T, s, substring string) {
	t.Helper()
	if !strings.Contains(s, substring) {
		t.Errorf("want %q to contain %q", s, substring)
	}
}

func TestTheReportShowsWhatNeedsAHuman(t *testing.T) {
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	data := func(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
	in := Input{
		Now:   now,
		Stale: 2 * time.Minute,
		NeedsHuman: []sisyphus.Issue{
			{Name: "fix-login-bug", Title: "Fix the login bug", State: "open", Priority: "high"},
		},
		WithReferendum: []sisyphus.Issue{
			{Name: "running-one", State: "in-progress", Metadata: map[string]string{"referendum": "round-2", "referendum-deadline": now.Add(time.Hour).Format(time.RFC3339)}},
			{Name: "crashed-one", State: "in-progress", Metadata: map[string]string{"referendum": "round-1", "referendum-deadline": now.Add(-time.Hour).Format(time.RFC3339)}},
			{Name: "stuck-one", State: "open", Metadata: map[string]string{"referendum": "no-result"}},
			{Name: "failed-open", State: "open", Metadata: map[string]string{"referendum": "failed"}},
			{Name: "failed-closed", State: "closed", Metadata: map[string]string{"referendum": "failed"}},
			{Name: "passed-one", State: "open", Metadata: map[string]string{"referendum": "passed"}},
		},
		Queue: "127.0.0.1:4000",
		Workers: []mq.WorkerStatus{
			{Worker: "check-1", Role: "repo-check", LastSeen: now.Add(-10 * time.Second), Holds: []mq.Held{{ID: "m1"}}},
			{Worker: "docs-1", Role: "technical-docs", LastSeen: now.Add(-time.Hour)},
		},
		Pending: []mq.Pending{
			{ID: "m1", Kind: "bookmark-ready", For: "role:repo-check", Sent: now.Add(-time.Minute), LeasedBy: "check-1"},
			{ID: "m2", Kind: "issue-closed", For: "role:technical-docs", Sent: now.Add(-30 * time.Minute)},
			{ID: "m3", Kind: "task", For: "role:technical-docs", Sent: now.Add(-5 * time.Minute)},
		},
		Log: []mq.Message{
			{Kind: LineKind, From: "merge-queue", Time: now.Add(-2 * time.Hour), Data: data(map[string]any{"bookmark": "merge-queue/old", "base": "main", "head": "aaa"})},
			{Kind: LineKind, From: "merge-queue", Time: now.Add(-time.Hour), Data: data(map[string]any{"bookmark": "merge-queue/2026-10-12", "base": "main", "head": "bbb", "merged": []string{"ai/a"}, "left-out": []string{"ai/b"}})},
			{Kind: RecommendationKind, From: "release-manager-1", Time: now.Add(-50 * time.Minute), Data: data(map[string]any{"bookmark": "merge-queue/2026-10-12", "recommend": "move", "summary": "0.0.120 and 0.0.121, no problems."})},
		},
		Merged: func(commit, base string) bool { return commit == "aaa" },
	}

	r := Build(in)
	text := r.Text(now)

	if len(r.Referendums) != 4 {
		t.Fatalf("want 4 referendums, got %+v", r.Referendums)
	}
	for _, part := range []string{
		"Issues that wait for a human (1)\n  fix-login-bug  open, high: Fix the login bug",
		"crashed-one  overdue (round-1, deadline ",
		"running-one  running (round-2, deadline ",
		"stuck-one  no-result. A human decides it: hephaestus referendum decide stuck-one",
		"failed-open  failed.",
		"Workers at 127.0.0.1:4000 (1 live, 1 stopped)",
		"live     check-1 (repo-check), last seen 10s ago, holds 1",
		"stopped  docs-1 (technical-docs), last seen 1h ago",
		"Messages that wait (3)",
		"role:repo-check  1 (bookmark-ready), oldest 1m ago, 1 leased: a live worker has them",
		"role:technical-docs  2 (issue-closed, task), oldest 30m ago, 0 leased: no live worker: start one",
		"merge-queue/2026-10-12 on main, made 1h ago: 1 merged, 1 left out. main does not have it yet",
		"left out: ai/b",
		"release manager (release-manager-1, 50m ago): move. 0.0.120 and 0.0.121, no problems.",
		"6 items need a human.",
	} {
		contains(t, text, part)
	}
	if strings.Contains(text, "failed-closed") || strings.Contains(text, "passed-one") {
		t.Errorf("a passed referendum, or a failed one on a closed issue, needs nobody:\n%s", text)
	}
}

func TestAnEmptyReportSaysSo(t *testing.T) {
	r := Build(Input{Now: time.Now(), Stale: time.Minute})

	text := r.Text(time.Now())

	contains(t, text, "Message queue: not running. Start it with: hephaestus messages serve")
	contains(t, text, "none in the message log")
	contains(t, text, "Nothing needs a human now.")
}
