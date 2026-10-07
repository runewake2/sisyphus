package main

// graphNode is one issue in "sisyphus graph": itself, its depends-on edges, and its children
// (sub-issues, via parent).
type graphNode struct {
	Name      string      `json:"name"`
	Title     string      `json:"title"`
	State     string      `json:"state"`
	Focus     bool        `json:"focus,omitempty"`
	DependsOn []string    `json:"depends-on,omitempty"`
	Children  []graphNode `json:"children,omitempty"`
}

// buildGraph walks up from name to its root ancestor (by parent), then builds the whole tree of
// descendants rooted there, marking name's own node as the focus.
func buildGraph(root, name string) graphNode {
	top := name
	seen := map[string]bool{top: true}
	for {
		next := parentName(root, top)
		if next == "" || seen[next] {
			break
		}
		top = next
		seen[next] = true
	}
	return buildGraphNode(root, top, name, map[string]bool{})
}

func buildGraphNode(root, name, focus string, seen map[string]bool) graphNode {
	if seen[name] {
		return graphNode{Name: name, Title: "(cycle; see parent)", State: "?"}
	}
	seen[name] = true

	node := graphNode{Name: name, Title: "(missing)", State: "?"}
	if _, _, doc, err := loadIssue(root, name); err == nil {
		node = graphNode{
			Name:      name,
			Title:     doc.get("title"),
			State:     doc.get("state"),
			DependsOn: parseDependsOn(doc.get("depends-on")),
		}
	}
	node.Focus = name == focus

	for _, sub := range subIssues(root, name) {
		node.Children = append(node.Children, buildGraphNode(root, sub.name, focus, seen))
	}
	return node
}
