package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const module = "github.com/runewake2/sisyphus"

// deps returns the packages of this module that the package pattern depends on.
func deps(t *testing.T, pattern string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", pattern)
	cmd.Dir = filepath.Join("..", "..")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list %s: %v", pattern, err)
	}
	var own []string
	for _, p := range strings.Fields(string(out)) {
		if p == module || strings.HasPrefix(p, module+"/") {
			own = append(own, p)
		}
	}
	return own
}

// hephaestus and sisyphus share one Go module, so that their commands live side by side, but
// hephaestus uses sisyphus only through its command, and sisyphus knows nothing of agents
// ([[0003-one-go-module-with-the-commands-side-by-side]]).
func TestHephaestusAndSisyphusShareNoCode(t *testing.T) {
	for _, p := range deps(t, "./cmd/hephaestus") {
		if p != module+"/cmd/hephaestus" && !strings.HasPrefix(p, module+"/internal/hephaestus/") {
			t.Errorf("hephaestus imports %s: it may use sisyphus only through the sisyphus command", p)
		}
	}
	for _, pattern := range []string{"./cmd/sisyphus", "./cmd/sisyphus-mcp"} {
		for _, p := range deps(t, pattern) {
			if strings.HasPrefix(p, module+"/internal/hephaestus") || p == module+"/cmd/hephaestus" {
				t.Errorf("%s imports %s: sisyphus does not manage agents", pattern, p)
			}
		}
	}
}
