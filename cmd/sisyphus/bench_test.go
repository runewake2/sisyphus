package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// benchmarkRepo writes issues issues below project and component directories. Each component has an
// epic, and each other issue names the epic as its parent and the previous issue as a dependency, by
// bare name, so that every reference needs a lookup.
func benchmarkRepo(b *testing.B, issues int) string {
	b.Helper()
	root := b.TempDir()
	template, err := kit.ReadFile("kit/issues/TEMPLATE.md.tmpl")
	if err != nil {
		b.Fatal(err)
	}
	write := func(relative, content string) {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	write("issues/TEMPLATE.md", string(template))
	const perComponent = 50
	for i := 0; i < issues; i++ {
		directory := fmt.Sprintf("project-%d/component-%d", i/(perComponent*10), i/perComponent)
		epic := fmt.Sprintf("epic-of-component-%d", i/perComponent)
		if i%perComponent == 0 {
			write(fmt.Sprintf("issues/open/%s/%s.md", directory, epic),
				"---\ntitle: \"Epic\"\nstate: open\npriority: medium\n---\n\n# Epic\n")
			continue
		}
		dependsOn := "[]"
		if i%perComponent > 1 {
			dependsOn = fmt.Sprintf(`["[[issue-number-%d]]"]`, i-1)
		}
		state := "open"
		if i%7 == 0 {
			state = "closed"
		}
		write(fmt.Sprintf("issues/%s/%s/issue-number-%d.md", state, directory, i), fmt.Sprintf(
			"---\ntitle: \"Issue %d\"\nstate: %s\npriority: medium\nparent: \"[[%s]]\"\ndepends-on: %s\n---\n\n# Issue %d\n",
			i, state, epic, dependsOn, i))
	}
	return root
}

func benchmarkCommand(b *testing.B, args ...string) {
	root := benchmarkRepo(b, 3000)
	findRoot := func() (string, error) { return root, nil }
	b.ResetTimer()
	for range b.N {
		if exit := run(args, findRoot, io.Discard, io.Discard); exit != 0 {
			b.Fatalf("sisyphus %v exited with %d", args, exit)
		}
	}
}

func BenchmarkListBlocked(b *testing.B) {
	benchmarkCommand(b, "list", "--blocked")
}

func BenchmarkGraphFull(b *testing.B) {
	benchmarkCommand(b, "graph", "--full", "epic-of-component-3")
}

func BenchmarkUpdate(b *testing.B) {
	benchmarkCommand(b, "update", "issue-number-2048", "open", "--priority", "high")
}

func BenchmarkListParent(b *testing.B) {
	benchmarkCommand(b, "list", "--parent", "epic-of-component-3")
}
