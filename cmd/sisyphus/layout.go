package main

import (
	"slices"

	"github.com/mattn/go-runewidth"
)

// The drawing flows left to right. Each issue is in a column, and each directory owns a band of
// lines that no other directory's issues enter. So the box of a directory holds exactly its own
// issues, and the boxes of two sibling directories cannot overlap.

// placed is one issue box. x and y are its top left corner, and the box is 3 lines high.
type placed struct {
	node   graphNode
	index  int
	label  string
	column int
	row    int
	x, y   int
}

func (p *placed) mid() int { return p.y + 1 }

// frame is the box of one directory. depth is 1 for a directory directly below a state directory.
type frame struct {
	dir                     *directory
	depth                   int
	top, bottom             int
	firstColumn, lastColumn int
}

// layout is where everything goes. A blank line has no issue box and no frame border, so an
// arrow can run along it.
type layout struct {
	issues      []*placed
	byName      map[string]*placed
	frames      []*frame
	depth       int
	columns     int
	columnWidth []int
	blank       []int
	height      int
}

// columnsOf puts each issue in the column after the first issue in link order that points to it,
// so the sub-issues of an epic share one column even when they depend on each other.
func columnsOf(graph issueGraph) map[string]int {
	column := map[string]int{}
	for _, node := range graph.Nodes {
		for _, edge := range graph.Edges {
			if _, done := column[edge.To]; edge.From == node.Name && !done && edge.To != node.Name {
				column[edge.To] = column[node.Name] + 1
			}
		}
		if _, done := column[node.Name]; !done {
			column[node.Name] = 0
		}
	}
	return column
}

func nodeLabel(node graphNode) string {
	_, file := splitIssueName(node.Name)
	return file + " [" + node.State + "]"
}

func layOut(graph issueGraph) *layout {
	l := &layout{byName: map[string]*placed{}}
	column := columnsOf(graph)
	for i, node := range graph.Nodes {
		p := &placed{node: node, index: i, label: nodeLabel(node), column: column[node.Name]}
		l.issues = append(l.issues, p)
		l.byName[node.Name] = p
		l.columns = max(l.columns, p.column+1)
	}
	l.columnWidth = make([]int, l.columns)
	for _, p := range l.issues {
		l.columnWidth[p.column] = max(l.columnWidth[p.column], runewidth.StringWidth(p.label)+2)
	}

	// Lines go down the directory tree in order of each part's first issue in link order, so
	// the focus and its family come first.
	first := map[*directory]int{}
	var firstIssue func(d *directory) int
	firstIssue = func(d *directory) int {
		if i, done := first[d]; done {
			return i
		}
		i := len(graph.Nodes)
		for _, node := range d.issues {
			i = min(i, l.byName[node.Name].index)
		}
		for _, child := range d.children {
			i = min(i, firstIssue(child))
		}
		first[d] = i
		return i
	}

	y := 0
	var placeDirectory func(d *directory, depth int)
	placeIssues := func(d *directory) {
		// An issue takes the row of the first issue that points to it in the same directory, or
		// the next free row in its column.
		taken := map[[2]int]bool{}
		rowOf := map[string]int{}
		rows := 0
		for _, node := range d.issues {
			p := l.byName[node.Name]
			row := 0
			for _, edge := range graph.Edges {
				if r, same := rowOf[edge.From]; same && edge.To == node.Name {
					row = r
					break
				}
			}
			for taken[[2]int{p.column, row}] {
				row++
			}
			taken[[2]int{p.column, row}] = true
			rowOf[node.Name] = row
			p.row = row
			rows = max(rows, row+1)
		}
		for _, node := range d.issues {
			p := l.byName[node.Name]
			p.y = y + 4*p.row
		}
		for r := 1; r < rows; r++ {
			l.blank = append(l.blank, y+4*r-1)
		}
		y += 4*rows - 1
	}
	placeDirectory = func(d *directory, depth int) {
		type part struct {
			first int
			dir   *directory
		}
		var parts []part
		if len(d.issues) > 0 {
			parts = append(parts, part{first: firstIssueOf(l, d.issues)})
		}
		for _, child := range d.children {
			parts = append(parts, part{first: firstIssue(child), dir: child})
		}
		slices.SortStableFunc(parts, func(a, b part) int { return a.first - b.first })
		for i, pt := range parts {
			if i > 0 {
				l.blank = append(l.blank, y)
				y++
			}
			if pt.dir == nil {
				placeIssues(d)
				continue
			}
			f := &frame{dir: pt.dir, depth: depth + 1, top: y}
			l.frames = append(l.frames, f)
			l.depth = max(l.depth, f.depth)
			l.blank = append(l.blank, y+1)
			y += 2
			placeDirectory(pt.dir, depth+1)
			l.blank = append(l.blank, y)
			f.bottom = y + 1
			y += 2
		}
	}
	placeDirectory(directoryTree(graph.Nodes), 0)
	l.height = y

	for _, f := range l.frames {
		f.firstColumn, f.lastColumn = l.columns, 0
		for _, p := range l.issues {
			if dir, _ := splitIssueName(p.node.Name); dir == f.dir.path || len(dir) > len(f.dir.path) && dir[:len(f.dir.path)+1] == f.dir.path+"/" {
				f.firstColumn = min(f.firstColumn, p.column)
				f.lastColumn = max(f.lastColumn, p.column)
			}
		}
		// The name sits in the top border above the first column, where no arrow crosses it.
		need := runewidth.StringWidth(f.dir.name) + 4 - l.margin(f)
		l.columnWidth[f.firstColumn] = max(l.columnWidth[f.firstColumn], need)
	}
	return l
}

func firstIssueOf(l *layout, nodes []graphNode) int {
	i := len(l.issues)
	for _, node := range nodes {
		i = min(i, l.byName[node.Name].index)
	}
	return i
}

// margin is how far a frame's border is from the issues in it. Frames that are not nested
// as deep sit farther out, so the borders of nested frames do not meet.
func (l *layout) margin(f *frame) int {
	return 2 * (l.depth - f.depth + 1)
}
