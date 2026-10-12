package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/runewake2/sisyphus/hephaestus/internal/toolagent"
)

// newDepCheckRepo makes a git repo with the dependency check roles of this repo, and hephaestus on
// PATH for their scripts.
func newDepCheckRepo(t *testing.T) *testRepo {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the checks are sh")
	}
	hephaestusOnPath(t)
	r := newGitRepo(t)
	roles, err := filepath.Abs(filepath.Join("..", "..", "roles", "dependency-check"))
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(r.root, "hephaestus", "roles")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cp", "-r", roles, target).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	return r
}

func (r *testRepo) depCheck(variant string) toolagent.Ballot {
	r.t.Helper()
	var b toolagent.Ballot
	if err := json.Unmarshal([]byte(r.mustRun("agent", "review", "dependency-check/"+variant, "--json").output), &b); err != nil {
		r.t.Fatal(err)
	}
	return b
}

func TestTheGitHubActionsCheckNeedsPinnedActions(t *testing.T) {
	r := newDepCheckRepo(t)
	r.write(".github/workflows/ci.yml", `jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@0aaccfd150d50ccaeb58ebd88d36e91967a5f35b # v5.4.0
      - uses: ./local-action
      - uses: "docker://alpine:3.20"
`)
	r.commit("workflow")

	b := r.depCheck("github-actions")

	equal(t, "no", b.Vote)
	contains(t, b.Reasons, ".github/workflows/ci.yml:4: actions/checkout@v4 is not pinned to a full commit SHA.")
	contains(t, b.Reasons, ".github/workflows/ci.yml:7: docker://alpine:3.20 is not pinned to a digest")
	notContains(t, b.Reasons, "setup-go")
	notContains(t, b.Reasons, "local-action")
}

func TestTheDockerCheckNeedsVersionedImages(t *testing.T) {
	r := newDepCheckRepo(t)
	r.write("Dockerfile", "FROM golang:1.23 AS build\nFROM --platform=linux/amd64 alpine\nFROM build\nFROM scratch\n")
	r.write("compose.yaml", "services:\n  db:\n    image: \"postgres:latest\"\n  cache:\n    image: redis:7.2@sha256:3f2b64ef97bd285e36132c684e6b2ae8fb2fd6e9a8a29ea3ef35a1c6c9f7e3d1\n")
	r.commit("images")

	b := r.depCheck("docker")

	equal(t, "no", b.Vote)
	contains(t, b.Reasons, "Dockerfile:1: the image golang:1.23 has no digest, so its content can change (information only).")
	contains(t, b.Reasons, "Dockerfile:2: the image alpine names no version (latest).")
	contains(t, b.Reasons, "compose.yaml:3: the image postgres:latest names no version (latest).")
	notContains(t, b.Reasons, "image build")
	notContains(t, b.Reasons, "redis")
	equal(t, 0, len(b.Files)) // the check changed no file of the working copy
}

func TestADependencyCheckWithNothingToCheckPasses(t *testing.T) {
	r := newDepCheckRepo(t)
	r.write("README.md", "# Readme\n")
	r.commit("readme")

	for _, variant := range []string{"golang", "javascript", "dotnet", "java", "docker", "github-actions"} {
		b := r.depCheck(variant)
		equal(t, "yes", b.Vote)
		contains(t, b.Reasons, "nothing to check")
	}
}

func TestADependencyCheckSkipsANoticeThatTouchesNoDependencyFile(t *testing.T) {
	r := newDepCheckRepo(t)
	r.write(".github/workflows/ci.yml", "jobs:\n  t:\n    steps:\n      - uses: actions/checkout@v4\n")
	r.commit("workflow")
	r.git("checkout", "-q", "-b", "work")
	r.write("notes.txt", "notes\n")
	r.commit("notes")
	t.Setenv("HEPHAESTUS_MESSAGE_KIND", "bookmark-ready")

	b := r.depCheck("github-actions")

	equal(t, "yes", b.Vote)
	contains(t, b.Reasons, "The change touches no GitHub Actions dependency file: nothing to check.")
}
