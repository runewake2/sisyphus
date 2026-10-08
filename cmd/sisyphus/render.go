package main

import (
	"cmp"
	"maps"
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

type span struct {
	first, last int
	from, to    string
}

func routeEdges(l *layout, graph issueGraph) ([]route, map[lane]int) {
	var routes []route
	var stubs []stub
	var lanes []lane
	known := map[lane]bool{}
	use := func(gap int, key string) *lane {
		ln := lane{gap: gap, key: key}
		if !known[ln] {
			known[ln] = true
			lanes = append(lanes, ln)
		}
		return &ln
	}
	inColumn := make([][]*placed, l.columns)
	for _, p := range l.issues {
		inColumn[p.column] = append(inColumn[p.column], p)
	}
	// covers tells whether an issue box in one of the columns from first to last covers line y.
	covers := func(first, last, y int) bool {
		for c := first; c <= last; c++ {
			for _, p := range inColumn[c] {
				if y >= p.y && y < p.y+boxHeight {
					return true
				}
			}
		}
		return false
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
		case v.column == u.column+1, v.column > u.column+1 && covers(u.column+1, v.column-1, u.mid()) && !covers(u.column+1, v.column-1, v.mid()):
			r.a = use(gapA, src)
		case v.column > u.column+1 && !covers(u.column+1, v.column-1, u.mid()):
			r.a = use(gapB, dst)
		default:
			r.a, r.b = use(gapA, src), use(gapB, dst)
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
		addStub(r.a, u.mid(), true)
		addStub(cmp.Or(r.b, r.a), v.mid(), false)
		routes = append(routes, r)
	}
	return routes, orderLanes(lanes, stubs)
}

// orderLanes puts each lane in a position within its gap. Where an out stub and an in stub of
// different arrows meet on one line, the out stub's lane comes first, so the two do not overlap.
// A cycle of such needs is broken in the order that the lanes were first used.
func orderLanes(lanes []lane, stubs []stub) map[lane]int {
	type place struct{ gap, row int }
	outs := map[place][]stub{}
	for _, out := range stubs {
		if out.out {
			outs[place{out.lane.gap, out.row}] = append(outs[place{out.lane.gap, out.row}], out)
		}
	}
	before := map[lane]map[lane]bool{}
	for _, in := range stubs {
		if in.out {
			continue
		}
		for _, out := range outs[place{in.lane.gap, in.row}] {
			if out.lane != in.lane && out.from != in.from && out.to != in.to {
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

// The classes of an arrow, in the order they are drawn: a later class wins a shared line.
const (
	plainArrow = iota
	focusArrow
	pathArrow
)

var arrowPens = [...]pen{plainArrow: noPen, focusArrow: focusPen, pathArrow: pathPen}

// look is how a drawing shows each issue and arrow. drawGraph works it out once, and the drawing
// and the legend both read it.
type look struct {
	style    map[string]boxStyle
	text     map[string]pen
	border   map[string]pen
	path     map[graphEdge]bool
	startNow map[string]bool
}

func lookOf(graph issueGraph) look {
	lk := look{style: map[string]boxStyle{}, text: map[string]pen{}, border: map[string]pen{}}
	lk.path, lk.startNow = workPath(graph)
	linked := map[string]bool{}
	for _, edge := range graph.Edges {
		if touchesFocus(graph, edge) {
			linked[edge.From], linked[edge.To] = true, true
		}
	}
	for _, node := range graph.Nodes {
		lk.style[node.Name] = issueStyle(graph, node)
		switch {
		case node.Name == graph.Focus:
			lk.text[node.Name] = focusPen
		case linked[node.Name]:
			lk.text[node.Name] = stateColors[stateIndex(node)].pen
		}
		lk.border[node.Name] = lk.text[node.Name]
		if lk.startNow[node.Name] {
			lk.border[node.Name] = pathPen
		}
	}
	return lk
}

func (lk look) arrowClass(graph issueGraph, edge graphEdge) int {
	switch {
	case lk.path[edge]:
		return pathArrow
	case touchesFocus(graph, edge):
		return focusArrow
	}
	return plainArrow
}

func touchesFocus(graph issueGraph, edge graphEdge) bool {
	return edge.From == graph.Focus || edge.To == graph.Focus
}

// stateColors names the state of an issue for its color. A closed issue with no resolution is
// completed, and an issue that does not exist is missing.
var stateColors = []struct {
	name string
	pen  pen
}{
	{"open", openPen},
	{"in-progress", inProgressPen},
	{"completed", completedPen},
	{"abandoned", abandonedPen},
	{"missing", missingPen},
}

func stateIndex(node graphNode) int {
	switch {
	case node.State == "open":
		return 0
	case node.State == "in-progress":
		return 1
	case node.State == "closed" && node.Resolution == "abandoned":
		return 3
	case node.State == "closed":
		return 2
	}
	return 4
}

func render(graph issueGraph, lk look, l *layout, routes []route, position map[lane]int, color bool) string {
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

	c := newCanvas(x, l.height)
	for _, f := range l.frames {
		x1 := columnX[f.firstColumn] - l.margin(f)
		x2 := columnX[f.lastColumn] + l.columnWidth[f.lastColumn] - 1 + l.margin(f)
		c.box(x1, f.top, x2, f.bottom, roundedBox)
		c.text(x1+2, f.top, " "+f.dir.name+" ")
	}
	for _, p := range l.issues {
		p.x = columnX[p.column]
		width := l.columnWidth[p.column]
		c.pen = lk.border[p.node.Name]
		c.box(p.x, p.y, p.x+width-1, p.y+boxHeight-1, lk.style[p.node.Name])
		c.pen = lk.text[p.node.Name]
		c.text(p.x+1+(width-2-textWidth(p.name))/2, p.y+1, p.name)
		c.text(p.x+1+(width-2-textWidth(p.node.State))/2, p.y+2, p.node.State)
	}
	slices.SortStableFunc(routes, func(a, b route) int {
		return lk.arrowClass(graph, a.edge) - lk.arrowClass(graph, b.edge)
	})
	for _, r := range routes {
		start := [2]int{r.from.x + l.columnWidth[r.from.column] - 1, r.from.mid()}
		end := [2]int{r.to.x - 1, r.to.mid()}
		points := [][2]int{start, {laneAt(r.a), start[1]}}
		if r.b != nil {
			points = append(points, [2]int{laneAt(r.a), r.through}, [2]int{laneAt(r.b), r.through})
		}
		last := cmp.Or(r.b, r.a)
		points = append(points, [2]int{laneAt(last), end[1]}, end)
		points = slices.CompactFunc(points, func(p, q [2]int) bool { return p == q })
		class := lk.arrowClass(graph, r.edge)
		c.pen = arrowPens[class]
		c.path(stroke{dotted: r.edge.Kind == dependsOnEdge, heavy: class == pathArrow}, points...)
	}
	for _, r := range routes {
		c.pen = arrowPens[lk.arrowClass(graph, r.edge)]
		c.text(r.to.x-1, r.to.mid(), "►")
	}
	return c.render(color)
}

// workPath is what the focus issue waits on, all the way down: an issue waits on its dependencies
// and its sub-issues that are not closed. The chains end at available issues other than the focus,
// which are the work to start now. path holds each drawn arrow on those chains.
func workPath(graph issueGraph) (path map[graphEdge]bool, startNow map[string]bool) {
	open := map[string]graphNode{}
	for _, node := range graph.Nodes {
		if node.State != "closed" && node.State != "?" {
			open[node.Name] = node
		}
	}
	waitsOn := map[string][]graphEdge{}
	for _, edge := range graph.Edges {
		waiting := edge.To
		if edge.Kind == parentEdge {
			waiting = edge.From
		}
		waitsOn[waiting] = append(waitsOn[waiting], edge)
	}
	path, startNow = map[graphEdge]bool{}, map[string]bool{}
	seen := map[string]bool{}
	var walk func(issue string)
	walk = func(issue string) {
		if seen[issue] {
			return
		}
		seen[issue] = true
		if open[issue].Available && issue != graph.Focus {
			startNow[issue] = true
		}
		for _, edge := range waitsOn[issue] {
			on := edge.From
			if edge.Kind == parentEdge {
				on = edge.To
			}
			if _, waits := open[on]; waits {
				path[edge] = true
				walk(on)
			}
		}
	}
	if _, waits := open[graph.Focus]; waits {
		walk(graph.Focus)
	}
	return path, startNow
}

// issueStyle is the border of an issue box: double for the focus issue, heavy for an available
// issue, and light for the rest, dashed if the issue is linked only indirectly. No box is both
// double and heavy, so an available focus issue is double, and the legend tells that it is
// available.
func issueStyle(graph issueGraph, node graphNode) boxStyle {
	switch {
	case node.Name == graph.Focus:
		return doubleBox
	case node.Available && node.Indirect:
		return heavyDashedBox
	case node.Available:
		return heavyBox
	case node.Indirect:
		return dashedBox
	}
	return lightBox
}

// The legend entry of each issue box style, in legend order. The focus has its own entry.
var styleLegend = []struct {
	style boxStyle
	entry string
}{
	{heavyBox, "┏━┓ available (ready to start)"},
	{dashedBox, "┌╌┐ linked indirectly"},
	{heavyDashedBox, "┏╍┓ available, linked indirectly"},
}

// drawGraph renders the graph as boxes and arrows for the terminal, followed by a legend. With
// color, the focus issue and the arrows that leave or enter it are bright, and the rest is dim.
func drawGraph(graph issueGraph, color bool) string {
	lk := lookOf(graph)
	l := layOut(graph)
	routes, position := routeEdges(l, graph)
	drawing := render(graph, lk, l, routes, position, color)

	used := map[boxStyle]bool{}
	for _, style := range lk.style {
		used[style] = true
	}
	legend := "╔═╗ " + graph.Focus
	if slices.ContainsFunc(graph.Nodes, func(node graphNode) bool { return node.Available && node.Name == graph.Focus }) {
		legend += " (available)"
	}
	for _, entry := range styleLegend {
		if used[entry.style] {
			legend += "   " + entry.entry
		}
	}
	if len(graph.Edges) > 0 {
		legend += "   ──► sub-issue   ┄┄► needed by"
	}
	if len(lk.path) > 0 {
		legend += "   ━━► work path"
	}
	if len(lk.startNow) > 0 {
		legend += "\nstart now: " + strings.Join(slices.Sorted(maps.Keys(lk.startNow)), ", ")
	}
	if color {
		legend += stateKey(graph, lk)
	}
	if note := hiddenNote(graph); note != "" {
		legend += "\n" + note
	}
	return drawing + "\n\n" + legend + "\n"
}

// stateKey is a line of the legend that names each state color of the drawing, in its color.
func stateKey(graph issueGraph, lk look) string {
	used := map[pen]bool{}
	for _, p := range lk.text {
		used[p] = true
	}
	var key []string
	for _, state := range stateColors {
		if used[state.pen] {
			key = append(key, palette[state.pen]+state.name+resetColor)
		}
	}
	if len(key) == 0 {
		return ""
	}
	return "\nlinked to " + graph.Focus + ": " + strings.Join(key, "   ")
}
