package referendum

import (
	"testing"

	"github.com/runewake2/sisyphus/hephaestus/internal/roles"
	"github.com/runewake2/sisyphus/hephaestus/internal/toolagent"
)

func TestAToolRejectsOnlyWhatTheModificationMakesWorse(t *testing.T) {
	role := roles.Role{Name: "formatter/yaml"}
	yes := toolagent.Ballot{Vote: Yes}
	no := toolagent.Ballot{Vote: No}
	modify := func(files ...string) toolagent.Ballot { return toolagent.Ballot{Vote: Modify, Files: files} }

	for _, c := range []struct {
		name          string
		before, after toolagent.Ballot
		accept        bool
	}{
		{"clean after", modify("a.yaml"), yes, true},
		{"the same files before and after", modify("a.yaml"), modify("a.yaml"), true},
		{"a check that failed before too", no, no, true},
		{"a new file to change", modify("a.yaml"), modify("a.yaml", "b.yaml"), false},
		{"clean before, not after", yes, modify("b.yaml"), false},
		{"a check that fails only after", modify("a.yaml"), no, false},
	} {
		accept, reasons := toolGate(role, c.before, c.after)
		if accept != c.accept {
			t.Errorf("%s: want accept %v, got %v (%s)", c.name, c.accept, accept, reasons)
		}
	}
}
