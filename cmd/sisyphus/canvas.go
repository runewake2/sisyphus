package main

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// The directions that a line leaves a canvas cell in.
const (
	up uint8 = 1 << iota
	down
	left
	right
)

// weightRunes is the box-drawing character for each set of directions, and the subset of them
// that is heavy. Unicode has a character for every such mix of light and heavy lines.
var weightRunes = map[[2]uint8]rune{
	{up | down, 0}:                                       '│',
	{up | down, up}:                                      '╿',
	{up | down, down}:                                    '╽',
	{up | down, up | down}:                               '┃',
	{up | left, 0}:                                       '┘',
	{up | left, up}:                                      '┚',
	{up | left, left}:                                    '┙',
	{up | left, up | left}:                               '┛',
	{down | left, 0}:                                     '┐',
	{down | left, down}:                                  '┒',
	{down | left, left}:                                  '┑',
	{down | left, down | left}:                           '┓',
	{up | down | left, 0}:                                '┤',
	{up | down | left, up}:                               '┦',
	{up | down | left, down}:                             '┧',
	{up | down | left, up | down}:                        '┨',
	{up | down | left, left}:                             '┥',
	{up | down | left, up | left}:                        '┩',
	{up | down | left, down | left}:                      '┪',
	{up | down | left, up | down | left}:                 '┫',
	{up | right, 0}:                                      '└',
	{up | right, up}:                                     '┖',
	{up | right, right}:                                  '┕',
	{up | right, up | right}:                             '┗',
	{down | right, 0}:                                    '┌',
	{down | right, down}:                                 '┎',
	{down | right, right}:                                '┍',
	{down | right, down | right}:                         '┏',
	{up | down | right, 0}:                               '├',
	{up | down | right, up}:                              '┞',
	{up | down | right, down}:                            '┟',
	{up | down | right, up | down}:                       '┠',
	{up | down | right, right}:                           '┝',
	{up | down | right, up | right}:                      '┡',
	{up | down | right, down | right}:                    '┢',
	{up | down | right, up | down | right}:               '┣',
	{left | right, 0}:                                    '─',
	{left | right, left}:                                 '╾',
	{left | right, right}:                                '╼',
	{left | right, left | right}:                         '━',
	{up | left | right, 0}:                               '┴',
	{up | left | right, up}:                              '┸',
	{up | left | right, left}:                            '┵',
	{up | left | right, up | left}:                       '┹',
	{up | left | right, right}:                           '┶',
	{up | left | right, up | right}:                      '┺',
	{up | left | right, left | right}:                    '┷',
	{up | left | right, up | left | right}:               '┻',
	{down | left | right, 0}:                             '┬',
	{down | left | right, down}:                          '┰',
	{down | left | right, left}:                          '┭',
	{down | left | right, down | left}:                   '┱',
	{down | left | right, right}:                         '┮',
	{down | left | right, down | right}:                  '┲',
	{down | left | right, left | right}:                  '┯',
	{down | left | right, down | left | right}:           '┳',
	{up | down | left | right, 0}:                        '┼',
	{up | down | left | right, up}:                       '╀',
	{up | down | left | right, down}:                     '╁',
	{up | down | left | right, up | down}:                '╂',
	{up | down | left | right, left}:                     '┽',
	{up | down | left | right, up | left}:                '╃',
	{up | down | left | right, down | left}:              '╅',
	{up | down | left | right, up | down | left}:         '╉',
	{up | down | left | right, right}:                    '┾',
	{up | down | left | right, up | right}:               '╄',
	{up | down | left | right, down | right}:             '╆',
	{up | down | left | right, up | down | right}:        '╊',
	{up | down | left | right, left | right}:             '┿',
	{up | down | left | right, up | left | right}:        '╇',
	{up | down | left | right, down | left | right}:      '╈',
	{up | down | left | right, up | down | left | right}: '╋',
}

// boxStyle is the line style of a box border.
type boxStyle int

const (
	lightBox boxStyle = iota
	heavyBox
	doubleBox
	roundedBox
	dashedBox
	heavyDashedBox
)

// styleRunes replaces weightRunes on the border of a double or rounded box, where Unicode has no
// mix with heavy lines. A dashed border uses them only on its straight parts, because Unicode has
// no dashed corners or joins. A heavy box border is a heavy line, so weightRunes draws it.
var styleRunes = map[boxStyle]map[uint8]rune{
	doubleBox: {
		up | down: '║', left | right: '═',
		down | right: '╔', down | left: '╗', up | right: '╚', up | left: '╝',
		up | down | right: '╟',
	},
	roundedBox: {
		down | right: '╭', down | left: '╮', up | right: '╰', up | left: '╯',
	},
	dashedBox: {
		up | down: '╎', left | right: '╌',
	},
	heavyDashedBox: {
		up | down: '╏', left | right: '╍',
	},
}

// cell is one column of one line. A line through the cell adds to lines. A cell with text shows its
// text instead. The column after a wide character is a continuation and draws nothing.
type cell struct {
	lines        uint8
	heavy        uint8
	solid        int
	dotted       int
	text         rune
	continuation bool
	style        boxStyle
	color        string
}

// canvas is a grid of cells that grows when something is drawn outside it. While pen is set, each
// cell that is drawn takes the pen's color. A drawing with no pen leaves a cell's color as it is,
// so a gray line that crosses a colored one does not hide it.
type canvas struct {
	cells [][]cell
	pen   string
	heavy bool
}

// The ANSI colors of a colored drawing. Each line resets at its end, so a line that is cut or
// pasted alone does not color what follows it.
const (
	strongColor = "\x1b[1;96m"
	// The color of the work path and the borders of the issues to start now.
	pathColor = "\x1b[1;95m"
	// The colors of the issues that the focus issue links directly, by state.
	openColor       = "\x1b[1;94m"
	inProgressColor = "\x1b[1;93m"
	completedColor  = "\x1b[1;92m"
	abandonedColor  = "\x1b[1;91m"
	missingColor    = "\x1b[1;97m"
	weakColor       = "\x1b[90m"
	resetColor      = "\x1b[0m"
)

func (c *canvas) at(x, y int) *cell {
	for len(c.cells) <= y {
		c.cells = append(c.cells, nil)
	}
	for len(c.cells[y]) <= x {
		c.cells[y] = append(c.cells[y], cell{})
	}
	cl := &c.cells[y][x]
	if c.pen != "" {
		cl.color = c.pen
	}
	return cl
}

// line draws a horizontal or vertical line from (x1, y1) to (x2, y2). A cell shows a dotted
// character only if every line through it is dotted and the lines do not turn there.
func (c *canvas) line(x1, y1, x2, y2 int, dotted bool) {
	mark := func(x, y int, directions uint8) {
		cl := c.at(x, y)
		cl.lines |= directions
		if c.heavy {
			cl.heavy |= directions
		}
		if dotted {
			cl.dotted++
		} else {
			cl.solid++
		}
	}
	switch {
	case y1 == y2:
		lo, hi := min(x1, x2), max(x1, x2)
		for x := lo; x <= hi; x++ {
			var directions uint8
			if x > lo {
				directions |= left
			}
			if x < hi {
				directions |= right
			}
			mark(x, y1, directions)
		}
	case x1 == x2:
		lo, hi := min(y1, y2), max(y1, y2)
		for y := lo; y <= hi; y++ {
			var directions uint8
			if y > lo {
				directions |= up
			}
			if y < hi {
				directions |= down
			}
			mark(x1, y, directions)
		}
	}
}

// path draws lines through each point in turn.
func (c *canvas) path(dotted bool, points ...[2]int) {
	for i := 1; i < len(points); i++ {
		c.line(points[i-1][0], points[i-1][1], points[i][0], points[i][1], dotted)
	}
}

// box draws a rectangle in style. Lines that end on its border join it.
func (c *canvas) box(x1, y1, x2, y2 int, style boxStyle) {
	heavy := c.heavy
	c.heavy = style == heavyBox || style == heavyDashedBox
	defer func() { c.heavy = heavy }()
	c.path(false, [2]int{x1, y1}, [2]int{x2, y1}, [2]int{x2, y2}, [2]int{x1, y2}, [2]int{x1, y1})
	for x := x1; x <= x2; x++ {
		c.at(x, y1).style, c.at(x, y2).style = style, style
	}
	for y := y1; y <= y2; y++ {
		c.at(x1, y).style, c.at(x2, y).style = style, style
	}
}

// text writes s from (x, y) and returns the column after it.
func (c *canvas) text(x, y int, s string) int {
	for _, r := range s {
		c.at(x, y).text = r
		width := runewidth.RuneWidth(r)
		for i := 1; i < width; i++ {
			c.at(x+i, y).continuation = true
		}
		x += max(width, 1)
	}
	return x
}

func (cl cell) rune() rune {
	if cl.text != 0 {
		return cl.text
	}
	// A line that ends in a cell draws through the whole cell.
	lines, heavy := cl.lines, cl.heavy
	for _, pair := range [][2]uint8{{up, down}, {left, right}} {
		if lines == pair[0] || lines == pair[1] {
			lines = pair[0] | pair[1]
			if heavy != 0 {
				heavy = lines
			}
		}
	}
	horizontal, vertical := lines == left|right, lines == up|down
	switch {
	case lines == 0:
		return ' '
	case cl.solid == 0 && horizontal:
		return map[bool]rune{false: '┄', true: '┅'}[heavy != 0]
	case cl.solid == 0 && vertical:
		return map[bool]rune{false: '┆', true: '┇'}[heavy != 0]
	}
	if r, styled := styleRunes[cl.style][lines]; styled && (horizontal || vertical || cl.style == doubleBox || cl.style == roundedBox) {
		return r
	}
	return weightRunes[[2]uint8{lines, heavy}]
}

// render writes the canvas as text. With color, each cell has its own color or weakColor.
func (c *canvas) render(color bool) string {
	var lines []string
	for _, row := range c.cells {
		last := len(row) - 1
		for last >= 0 && (row[last].continuation || row[last].rune() == ' ') {
			last--
		}
		var b strings.Builder
		current := ""
		for _, cl := range row[:last+1] {
			if cl.continuation {
				continue
			}
			r := cl.rune()
			if color && r != ' ' {
				want := cl.color
				if want == "" {
					want = weakColor
				}
				if want != current {
					if current != "" {
						b.WriteString(resetColor)
					}
					b.WriteString(want)
					current = want
				}
			}
			b.WriteRune(r)
		}
		if current != "" {
			b.WriteString(resetColor)
		}
		lines = append(lines, b.String())
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

func (c *canvas) String() string { return c.render(false) }
