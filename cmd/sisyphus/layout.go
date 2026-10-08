package main

import (
	"cmp"
	"math"
	"slices"
)

// The drawing flows left to right. Each issue is in a column, and each directory owns a band of
// lines that no other directory's issues enter. So the box of a directory holds exactly its own
// issues, and the boxes of two sibling directories cannot overlap.

// An issue box shows the file name of its issue on one line and the state on the next, so a long
// state such as in-progress does not widen the box.
const (
	boxHeight = 4
	rowPitch  = boxHeight + 1
)

// placed is one issue box. x and y are its top left corner.
type placed struct {
	node   graphNode
	index  int
	name   string
	column int
	x, y   int
}

// mid is the line where arrows leave and enter the box: the line of the name.
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

// columnsOf puts each issue one column after the farthest issue that points to it, so a chain of
// dependencies reads left to right and most arrows reach only the next column. graph.Nodes is in
// link order, with issues in a cycle last; an arrow back to an earlier issue does not move it.
func columnsOf(graph issueGraph) map[string]int {
	index := map[string]int{}
	for i, node := range graph.Nodes {
		index[node.Name] = i
	}
	outgoing := map[string][]graphEdge{}
	for _, edge := range graph.Edges {
		outgoing[edge.From] = append(outgoing[edge.From], edge)
	}
	column := map[string]int{}
	for _, node := range graph.Nodes {
		for _, edge := range outgoing[node.Name] {
			if index[edge.To] > index[edge.From] {
				column[edge.To] = max(column[edge.To], column[node.Name]+1)
			}
		}
	}
	return column
}

func layOut(graph issueGraph) *layout {
	l := &layout{byName: map[string]*placed{}}
	column := columnsOf(graph)
	for i, node := range graph.Nodes {
		p := &placed{node: node, index: i, name: baseName(node.Name), column: column[node.Name]}
		l.issues = append(l.issues, p)
		l.byName[node.Name] = p
		l.columns = max(l.columns, p.column+1)
	}
	l.columnWidth = make([]int, l.columns)
	for _, p := range l.issues {
		width := max(textWidth(p.name), textWidth(p.node.State)) + 2
		l.columnWidth[p.column] = max(l.columnWidth[p.column], width)
	}

	incoming := map[string][]graphEdge{}
	for _, edge := range graph.Edges {
		incoming[edge.To] = append(incoming[edge.To], edge)
	}

	// Lines go down the directory tree in order of each part's first issue in link order, so
	// the focus and its family come first.
	firstOf := func(nodes []graphNode) int {
		i := len(l.issues)
		for _, node := range nodes {
			i = min(i, l.byName[node.Name].index)
		}
		return i
	}
	var firstIssue func(d *directory) int
	firstIssue = func(d *directory) int {
		i := firstOf(d.issues)
		for _, child := range d.children {
			i = min(i, firstIssue(child))
		}
		return i
	}

	y := 0
	var placeDirectory func(d *directory, depth int)
	placeIssues := func(d *directory) {
		// Column by column, each issue wants the mean row of the issues in this directory that
		// point to it, so an arrow can run straight. Issues keep that order down the column, one
		// row apart at least. An arrow from two or more columns back runs along the row of the
		// issue it enters, so that row must be free in each column between.
		rowOf := map[string]int{}
		taken := map[[2]int]bool{}
		blocked := func(p *placed, row int) bool {
			for _, edge := range incoming[p.node.Name] {
				from := l.byName[edge.From]
				if from.column >= p.column-1 {
					continue
				}
				for c := from.column + 1; c < p.column; c++ {
					if taken[[2]int{c, row}] {
						return true
					}
				}
			}
			return false
		}
		rows := 0
		for column := range l.columns {
			type want struct {
				p   *placed
				row float64
				has bool
			}
			var wants []want
			for _, node := range d.issues {
				p := l.byName[node.Name]
				if p.column != column {
					continue
				}
				w := want{p: p}
				sum, n := 0, 0
				for _, edge := range incoming[node.Name] {
					if r, placed := rowOf[edge.From]; placed {
						sum, n = sum+r, n+1
					}
				}
				if n > 0 {
					w.row, w.has = float64(sum)/float64(n), true
				}
				wants = append(wants, w)
			}
			slices.SortStableFunc(wants, func(a, b want) int {
				switch {
				case a.has && b.has && a.row != b.row:
					return cmp.Compare(a.row, b.row)
				case a.has != b.has:
					if a.has {
						return -1
					}
					return 1
				}
				return a.p.index - b.p.index
			})
			next := 0
			for _, w := range wants {
				row := next
				if w.has {
					row = max(row, int(math.Round(w.row)))
				}
				for blocked(w.p, row) {
					row++
				}
				taken[[2]int{column, row}] = true
				w.p.y = y + rowPitch*row
				rowOf[w.p.node.Name] = row
				next = row + 1
				rows = max(rows, next)
			}
		}
		for r := 1; r < rows; r++ {
			l.blank = append(l.blank, y+rowPitch*r-1)
		}
		y += rowPitch*rows - 1
	}
	placeDirectory = func(d *directory, depth int) {
		type part struct {
			first int
			dir   *directory
		}
		var parts []part
		if len(d.issues) > 0 {
			parts = append(parts, part{first: firstOf(d.issues)})
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

	var columnsIn func(d *directory) (first, last int)
	columnsIn = func(d *directory) (first, last int) {
		first, last = l.columns, 0
		for _, node := range d.issues {
			first, last = min(first, l.byName[node.Name].column), max(last, l.byName[node.Name].column)
		}
		for _, child := range d.children {
			f, la := columnsIn(child)
			first, last = min(first, f), max(last, la)
		}
		return first, last
	}
	for _, f := range l.frames {
		f.firstColumn, f.lastColumn = columnsIn(f.dir)
		// The name sits in the top border above the first column, where no arrow crosses it.
		need := textWidth(f.dir.name) + 4 - l.margin(f)
		l.columnWidth[f.firstColumn] = max(l.columnWidth[f.firstColumn], need)
	}
	return l
}

// margin is how far a frame's border is from the issues in it. Frames that are not nested
// as deep sit farther out, so the borders of nested frames do not meet.
func (l *layout) margin(f *frame) int {
	return 2 * (l.depth - f.depth + 1)
}
