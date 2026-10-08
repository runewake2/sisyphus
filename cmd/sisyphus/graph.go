package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	mermaid "github.com/AlexanderGrooff/mermaid-ascii/pkg/graph"
)

// issueGraph is "sisyphus graph": the focus issue, every issue below it, and the path above it, or
// with full, every issue connected to it through parent or depends-on in either direction. Edges are the
// links between the drawn issues. Hidden counts the connected issues that are not drawn.
type issueGraph struct {
	Focus  string      `json:"focus"`
	Nodes  []graphNode `json:"nodes"`
	Edges  []graphEdge `json:"edges"`
	Hidden int         `json:"hidden"`
}

// graphNode is one issue in the graph. Leaf is true for an issue that is not closed and has no
// dependency or sub-issue left open, so work on it can start now.
type graphNode struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	State string `json:"state"`
	Leaf  bool   `json:"leaf,omitempty"`
}

// graphEdge points from the issue that comes first to the one that comes after: kind "parent" from
// a parent to its sub-issue, and kind "depends-on" from a dependency to the issue that depends on it.
type graphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

// The marks are emoji that always draw two columns wide, so the boxes around them line up.
const (
	focusMark = "📍"
	leafMark  = "🍃"
)

const (
	parentEdge    = "parent"
	dependsOnEdge = "depends-on"
)

// buildGraph finds every issue reachable from name by following parent and depends-on links both
// ways. A link to an issue that does not exist becomes a node with state "?" and is not followed.
// Unless full is set, only name, everything below it, and the path above it are kept, and the
// rest are counted as hidden.
func buildGraph(root, name string, full bool) issueGraph {
	all := map[string]graphNode{}
	var edges []graphEdge
	for _, m := range matchingIssues(root, listOptions{states: states}) {
		issue := m.name
		all[issue] = graphNode{Name: issue, Title: m.doc.get("title"), State: m.doc.get("state")}
		if parent := canonicalName(root, m.doc.get("parent")); parent != "" {
			edges = append(edges, graphEdge{From: parent, To: issue, Kind: parentEdge})
		}
		for _, dependency := range parseDependsOn(m.doc.get("depends-on")) {
			edges = append(edges, graphEdge{From: canonicalName(root, dependency), To: issue, Kind: dependsOnEdge})
		}
	}

	// An issue waits on its open dependencies and its open sub-issues. A dependency that does not
	// exist does not block, as in "sisyphus list --blocked".
	blocked := map[string]bool{}
	for _, edge := range edges {
		waiting, on := edge.To, edge.From
		if edge.Kind == parentEdge {
			waiting, on = edge.From, edge.To
		}
		if node, exists := all[on]; exists && node.State != "closed" {
			blocked[waiting] = true
		}
	}
	for issue, node := range all {
		node.Leaf = node.State != "closed" && !blocked[issue]
		all[issue] = node
	}

	below, above := map[string][]string{}, map[string][]string{}
	for _, edge := range edges {
		below[edge.From] = append(below[edge.From], edge.To)
		above[edge.To] = append(above[edge.To], edge.From)
	}
	// reach walks from name along the given links. A missing issue is reached but not walked from.
	reach := func(links ...map[string][]string) map[string]bool {
		found := map[string]bool{name: true}
		queue := []string{name}
		for len(queue) > 0 {
			next := queue[0]
			queue = queue[1:]
			if _, exists := all[next]; !exists {
				continue
			}
			for _, l := range links {
				for _, neighbor := range l[next] {
					if !found[neighbor] {
						found[neighbor] = true
						queue = append(queue, neighbor)
					}
				}
			}
		}
		return found
	}
	connected := reach(below, above)
	drawn := connected
	if !full {
		drawn = reach(below)
		maps.Copy(drawn, reach(above))
	}

	graph := issueGraph{Focus: name, Hidden: len(connected) - len(drawn)}
	for _, edge := range edges {
		if drawn[edge.From] && drawn[edge.To] {
			graph.Edges = append(graph.Edges, edge)
		}
	}
	slices.SortFunc(graph.Edges, func(a, b graphEdge) int {
		return cmp.Or(strings.Compare(a.From, b.From), strings.Compare(a.To, b.To), strings.Compare(a.Kind, b.Kind))
	})

	// The drawing places nodes in the order they are declared, so declare them in link order: each
	// issue after every issue that links to it, by name among equals. Issues in a cycle go last.
	waiting := map[string]int{}
	for _, edge := range graph.Edges {
		waiting[edge.To]++
	}
	var ready []string
	for issue := range drawn {
		if waiting[issue] == 0 {
			ready = append(ready, issue)
		}
	}
	placed := map[string]bool{}
	for len(ready) > 0 {
		slices.Sort(ready)
		issue := ready[0]
		ready = ready[1:]
		placed[issue] = true
		graph.Nodes = append(graph.Nodes, nodeFor(all, issue))
		for _, edge := range graph.Edges {
			if edge.From == issue {
				if waiting[edge.To]--; waiting[edge.To] == 0 {
					ready = append(ready, edge.To)
				}
			}
		}
	}
	rest := slices.Sorted(maps.Keys(drawn))
	for _, issue := range rest {
		if !placed[issue] {
			graph.Nodes = append(graph.Nodes, nodeFor(all, issue))
		}
	}
	return graph
}

// nodeFor is issue's node, or a "?" node if the issue does not exist.
func nodeFor(all map[string]graphNode, issue string) graphNode {
	if node, exists := all[issue]; exists {
		return node
	}
	return graphNode{Name: issue, Title: "(missing)", State: "?"}
}

// mermaidSource writes the graph as a Mermaid flowchart. A solid arrow points from a parent to a
// sub-issue; a dotted arrow points from a dependency to the issue that depends on it. The focus issue is
// marked with focusMark, and each leaf with leafMark.
func mermaidSource(graph issueGraph) string {
	ids := map[string]string{}
	for i, node := range graph.Nodes {
		ids[node.Name] = fmt.Sprintf("n%d", i)
	}
	var b strings.Builder
	b.WriteString("graph LR\n")
	for _, node := range graph.Nodes {
		marks := ""
		if node.Name == graph.Focus {
			marks += focusMark
		}
		if node.Leaf {
			marks += leafMark
		}
		label := node.Name + " [" + node.State + "]"
		if marks != "" {
			label = marks + " " + label
		}
		fmt.Fprintf(&b, "    %s[\"%s\"]\n", ids[node.Name], label)
	}
	for _, edge := range graph.Edges {
		arrow := "-->"
		if edge.Kind == dependsOnEdge {
			arrow = "-.->"
		}
		fmt.Fprintf(&b, "    %s %s %s\n", ids[edge.From], arrow, ids[edge.To])
	}
	return b.String()
}

// drawGraph renders the graph as boxes and arrows for the terminal, followed by a legend.
func drawGraph(graph issueGraph) (string, error) {
	properties, err := mermaid.Parse(mermaidSource(graph), "cli")
	if err != nil {
		return "", err
	}
	config := diagram.DefaultConfig()
	config.BoxBorderPadding = 0
	config.PaddingBetweenY = 3
	properties.Apply(config)

	var lines []string
	for _, line := range strings.Split(mermaid.Draw(properties), "\n") {
		lines = append(lines, strings.TrimRight(line, " "))
	}
	drawing := strings.TrimRight(strings.Join(lines, "\n"), "\n")

	legend := focusMark + " " + graph.Focus
	if slices.ContainsFunc(graph.Nodes, func(node graphNode) bool { return node.Leaf }) {
		legend += "   " + leafMark + " leaf (ready to start)"
	}
	if len(graph.Edges) > 0 {
		legend += "   ──► sub-issue   ┄┄► needed by"
	}
	if note := hiddenNote(graph); note != "" {
		legend += "\n" + note
	}
	return drawing + "\n\n" + legend + "\n", nil
}

// hiddenNote tells how many connected issues the graph leaves out, if any.
func hiddenNote(graph issueGraph) string {
	switch graph.Hidden {
	case 0:
		return ""
	case 1:
		return "1 more linked issue is not drawn. Use --full to draw it."
	}
	return fmt.Sprintf("%d more linked issues are not drawn. Use --full to draw them.", graph.Hidden)
}
