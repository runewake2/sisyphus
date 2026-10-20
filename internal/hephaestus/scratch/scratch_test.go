package scratch

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// noIdentity hides every git identity from the test: the global and system config go to empty or
// nowhere, so only what the test configures counts.
func noIdentity(t *testing.T) {
	t.Helper()
	empty := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func TestIdentityFallsBackInPairOrder(t *testing.T) {
	noIdentity(t)
	args := Identity(t.TempDir(), "hephaestus test")
	want := []string{"-c", "user.email=hephaestus@localhost", "-c", "user.name=hephaestus test"}
	if !slices.Equal(args, want) {
		// A sorted flat list would put the two -c flags first and break git.
		t.Fatalf("the fallback is %v, want %v", args, want)
	}
}

func TestAConfiguredIdentityWins(t *testing.T) {
	noIdentity(t)
	config := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(config, []byte("[user]\n\tname = someone\n\temail = someone@example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", config)
	if args := Identity(t.TempDir(), "hephaestus test"); len(args) != 0 {
		t.Fatalf("a configured identity still gets a fallback: %v", args)
	}
}
