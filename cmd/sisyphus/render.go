package main

import (
	"slices"
	"strings"
)

// An arrow leaves the right side of its issue and enters the left side of the next. Between
// two columns is a gap, and each arrow that turns in a gap does so in a lane of its own. Lanes
// belong to the issue that the arrows leave or enter, so arrows from one issue share a trunk.
//
// Gap g is right of column g. Gap -1 is left of column 0, for an arrow back into column 0.

type lane struct {
	gap int
	key string
}

// route is the path of one arrow. Lanes a and b are where it turns. through is the line it
// follows from lane a to lane b, if it uses both.
type route struct {
	edge     graphEdge
	from, to *placed
	a, b     *lane
	through  int
}

// stub is a part of a line on the middle line of an issue box, inside one gap. An out stub runs
// from the gap's left side to a lane; an in stub runs from a lane to the gap's right side.
type stub struct {
	lane     lane
	row      int
	from, to string
	out      bool
}

// covers tells whether an issue box in one of the columns from first to last covers line y.
func (l *layout) covers(first, last, y int) bool {
	for _, p := range l.issues {
		if p.column >= first && p.column <= last && y >= p.y && y <= p.y+2 {
			return true
		}
	}
	return false
}

type span struct {
	first, last int
	from, to    string
}

func routeEdges(l *layout, graph issueGraph) ([]route, map[lane]int) {
	var routes []route
	var stubs []stub
	var lanes []lane
	use := func(gap int, key string) *lane {
		ln := lane{gap: gap, key: key}
		if !slices.Contains(lanes, ln) {
			lanes = append(lanes, ln)
		}
		return &ln
	}
	along := map[int][]span{}
	free := func(y int, s span) bool {
		for _, other := range along[y] {
			if other.first <= s.last && s.first <= other.last && other.from != s.from && other.to != s.to {
				return false
			}
		}
		return true
	}

	for _, edge := range graph.Edges {
		u, v := l.byName[edge.From], l.byName[edge.To]
		if u == v {
			continue
		}
		r := route{edge: edge, from: u, to: v}
		src, dst := "from:"+u.node.Name, "to:"+v.node.Name
		gapA, gapB := u.column, v.column-1
		addStub := func(ln *lane, row int, out bool) {
			stubs = append(stubs, stub{lane: *ln, row: row, from: u.node.Name, to: v.node.Name, out: out})
		}
		switch {
		case v.column == u.column+1:
			r.a = use(gapA, src)
			addStub(r.a, u.mid(), true)
			addStub(r.a, v.mid(), false)
		case v.column > u.column+1 && !l.covers(u.column+1, v.column-1, u.mid()):
			r.b = use(gapB, dst)
			addStub(r.b, u.mid(), true)
			addStub(r.b, v.mid(), false)
		case v.column > u.column+1 && !l.covers(u.column+1, v.column-1, v.mid()):
			r.a = use(gapA, src)
			addStub(r.a, u.mid(), true)
			addStub(r.a, v.mid(), false)
		default:
			r.a, r.b = use(gapA, src), use(gapB, dst)
			addStub(r.a, u.mid(), true)
			addStub(r.b, v.mid(), false)
			s := span{first: min(gapA, gapB), last: max(gapA, gapB), from: u.node.Name, to: v.node.Name}
			best := -1
			for _, y := range l.blank {
				if !free(y, s) {
					continue
				}
				distance := func(y int) int { return abs(y-u.mid()) + abs(y-v.mid()) }
				if best < 0 || distance(y) < distance(best) {
					best = y
				}
			}
			if best < 0 {
				best = l.height
				l.blank = append(l.blank, best)
				l.height++
			}
			along[best] = append(along[best], s)
			r.through = best
		}
		routes = append(routes, r)
	}
	return routes, orderLanes(lanes, stubs)
}

// orderLanes puts each lane in a position within its gap. Where an out stub and an in stub of
// different arrows meet on one line, the out stub's lane comes first, so the two do not overlap.
// A cycle of such needs is broken in the order that the lanes were first used.
func orderLanes(lanes []lane, stubs []stub) map[lane]int {
	before := map[lane]map[lane]bool{}
	for _, out := range stubs {
		for _, in := range stubs {
			if out.out && !in.out && out.lane.gap == in.lane.gap && out.row == in.row &&
				out.lane != in.lane && out.from != in.from && out.to != in.to {
				if before[in.lane] == nil {
					before[in.lane] = map[lane]bool{}
				}
				before[in.lane][out.lane] = true
			}
		}
	}
	position := map[lane]int{}
	count := map[int]int{}
	remaining := slices.Clone(lanes)
	for len(remaining) > 0 {
		pick := 0
		for i, ln := range remaining {
			ready := true
			for need := range before[ln] {
				if _, done := position[need]; !done {
					ready = false
				}
			}
			if ready {
				pick = i
				break
			}
		}
		ln := remaining[pick]
		position[ln] = count[ln.gap]
		count[ln.gap]++
		remaining = slices.Delete(remaining, pick, pick+1)
	}
	return position
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// render draws the layout and its arrows.
func render(graph issueGraph, l *layout, routes []route, position map[lane]int) string {
	lanesIn := map[int]int{}
	for ln, i := range position {
		lanesIn[ln.gap] = max(lanesIn[ln.gap], i+1)
	}
	borders := 2*l.depth + 1
	laneX := map[int]int{}
	columnX := make([]int, l.columns)
	x := 0
	laneX[-1] = x
	x += 2 * lanesIn[-1]
	if lanesIn[-1] > 0 {
		x += borders
	} else {
		x += borders - 1
	}
	for column := range l.columns {
		columnX[column] = x
		x += l.columnWidth[column] + borders
		laneX[column] = x
		x += 2 * lanesIn[column]
		if column < l.columns-1 {
			x += borders
		}
	}
	laneAt := func(ln *lane) int { return laneX[ln.gap] + 2*position[*ln] }

	var c canvas
	for _, f := range l.frames {
		x1 := columnX[f.firstColumn] - l.margin(f)
		x2 := columnX[f.lastColumn] + l.columnWidth[f.lastColumn] - 1 + l.margin(f)
		c.box(x1, f.top, x2, f.bottom, roundedBox)
		c.text(x1+2, f.top, " "+f.dir.name+" ")
	}
	for _, p := range l.issues {
		p.x = columnX[p.column]
		width := l.columnWidth[p.column]
		c.box(p.x, p.y, p.x+width-1, p.y+2, issueStyle(graph, p.node))
		c.text(p.x+1+(width-2-textWidth(p.label))/2, p.y+1, p.label)
	}
	for _, r := range routes {
		start := [2]int{r.from.x + l.columnWidth[r.from.column] - 1, r.from.mid()}
		end := [2]int{r.to.x - 1, r.to.mid()}
		points := [][2]int{start}
		switch {
		case r.a != nil && r.b != nil:
			a, b := laneAt(r.a), laneAt(r.b)
			points = append(points, [2]int{a, start[1]}, [2]int{a, r.through}, [2]int{b, r.through}, [2]int{b, end[1]})
		case r.a != nil:
			points = append(points, [2]int{laneAt(r.a), start[1]}, [2]int{laneAt(r.a), end[1]})
		default:
			points = append(points, [2]int{laneAt(r.b), start[1]}, [2]int{laneAt(r.b), end[1]})
		}
		points = append(points, end)
		points = slices.CompactFunc(points, func(p, q [2]int) bool { return p == q })
		c.path(r.edge.Kind == dependsOnEdge, points...)
	}
	for _, r := range routes {
		c.text(r.to.x-1, r.to.mid(), "►")
	}
	return c.String()
}

// issueStyle is the border of an issue box: double for the focus issue, heavy for a leaf, and
// light for the rest. No box is both double and heavy, so a focus issue that is a leaf is double,
// and the legend tells that it is a leaf.
func issueStyle(graph issueGraph, node graphNode) boxStyle {
	switch {
	case node.Name == graph.Focus:
		return doubleBox
	case node.Leaf:
		return heavyBox
	}
	return lightBox
}

func textWidth(s string) int {
	var c canvas
	return c.text(0, 0, s)
}

// drawGraph renders the graph as boxes and arrows for the terminal, followed by a legend.
func drawGraph(graph issueGraph) (string, error) {
	l := layOut(graph)
	routes, position := routeEdges(l, graph)
	drawing := render(graph, l, routes, position)

	legend := "╔═╗ " + graph.Focus
	if slices.ContainsFunc(graph.Nodes, func(node graphNode) bool { return node.Leaf && node.Name == graph.Focus }) {
		legend += " (a leaf)"
	}
	if slices.ContainsFunc(graph.Nodes, func(node graphNode) bool { return node.Leaf && node.Name != graph.Focus }) {
		legend += "   ┏━┓ leaf (ready to start)"
	}
	if len(graph.Edges) > 0 {
		legend += "   ──► sub-issue   ┄┄► needed by"
	}
	if note := hiddenNote(graph); note != "" {
		legend += "\n" + note
	}
	return strings.TrimRight(drawing, "\n") + "\n\n" + legend + "\n", nil
}
