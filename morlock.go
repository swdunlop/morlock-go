package morlock

import (
	"github.com/swdunlop/termbox-go"
	"unicode/utf8"
)

// Draw draws the root widget using the screen as a pen.
func Draw(root Widget) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()
	if root == nil {
		return
	}
	p := &Pen{fg: termbox.ColorDefault, bg: termbox.ColorDefault}
	p.w, p.h = termbox.Size()
	root.Draw(p)
}

// a Grid widget organizes Rows and bypasses their Draw implementation to ensure that each widget is vertically aligned
type Grid []Row

var _ Widget = Grid{}

// Grids combine the ReqWidth of their rows, but cheat to add a space between elements
func (t Grid) ReqWidth() (min, max int) {
	for _, row := range t {
		n, x := row.ReqWidth()
		if n > min {
			min = n
		}
		if x > max {
			max = x
		}
	}
	min += len(t)
	max += len(t)
	return
}

// The required height of a grid is the combined required height of its rows.
func (t Grid) ReqHeight() (min, max int) {
	for _, row := range t {
		n, x := row.ReqHeight()
		min += n
		max += x
	}
	return
}

// Grids draw themselves by drawing each row.
func (t Grid) Draw(p *Pen) {
	// println("grid", p.x, p.y, p.w, p.h)
	cols := 0
	for _, row := range t {
		if len(row) > cols {
			cols = len(row)
		}
	}

	w := make([]int, cols)
	h := make([]int, len(t))

	for i, row := range t {
		for j, t := range row {
			n, _ := t.ReqWidth()
			if n > w[j] {
				w[j] = n
			}
			n, _ = t.ReqHeight()
			if n > h[i] {
				h[i] = n
			}
		}
	}

	y := 0
	for i, row := range t {
		x := 0
		for j, t := range row {
			t.Draw(p.Clip(x, y, w[j], h[i]))
			x += w[j] + 1
		}
		y += h[i]
	}
}

// Row is a Widget that divides space horizontally
type Row []Widget

func (t Row) ReqHeight() (min, max int) {
	for _, t := range t {
		m, n := t.ReqHeight()
		if m > min {
			min = m
		}
		if n > max {
			max = n
		}
	}
	return
}

func (t Row) ReqWidth() (min, max int) {
	for _, w := range t {
		m, n := w.ReqWidth()
		min += m
		max += n
	}
	return
}

func (t Row) Draw(p *Pen) {
	w, h := p.Width(), p.Height()
	sz := fitSize(w, t, func(t Widget) (int, int) {
		return t.ReqWidth()
	})

	x := 0
	for i, t := range t {
		tw := sz[i]
		t.Draw(p.Clip(x, 0, tw, h))
		x += tw
	}
}

// a Tint colors the pen before drawing the enclosed Widget; the default Fg and Bg matches the terminal default.
type Tint struct {
	Fg termbox.Attribute
	Bg termbox.Attribute
	Widget
}

var _ Widget = Tint{}

func (t Tint) ReqWidth() (int, int) {
	return t.Widget.ReqWidth()
}

func (t Tint) ReqHeight() (int, int) {
	return t.Widget.ReqHeight()
}

func (t Tint) Draw(p *Pen) {
	if t.Fg != 0 {
		p.SetFg(t.Fg)
	}
	if t.Bg != 0 {
		p.SetBg(t.Bg)
	}
	t.Widget.Draw(p)
}

// a Blank is empty space used for padding out Column and Row containers
type Blank struct {
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
}

var _ Widget = Blank{}

func (t Blank) ReqWidth() (min, max int) {
	return t.MinWidth, t.MaxWidth
}
func (t Blank) ReqHeight() (min, max int) {
	return t.MinHeight, t.MaxHeight
}
func (t Blank) Draw(p *Pen) {}

// a Label is a constant string
type Label string

var _ Widget = Label("")

func (t Label) ReqWidth() (min, max int) {
	return len(t), len(t)
}

func (t Label) ReqHeight() (min, max int) {
	return 1, 1
}

func (t Label) Draw(p *Pen) {
	// println("label", p.x, p.y, p.w, p.h)
	p.Print(string(t))
}

// Widgets are areas in morlock that can describe their size requirements, and draw themselves using a morlock.Pen.
type Widget interface {
	ReqWidth() (min, max int)
	ReqHeight() (min, max int)
	Draw(p *Pen)
}

// A Column is a vertical assortment of widgets.
type Column []Widget

var _ Widget = Column{}

// The required width of a column is the maximum requirements of its contents
func (t Column) ReqWidth() (min, max int) {
	for _, t := range t {
		m, n := t.ReqWidth()
		if m > min {
			min = m
		}
		if n > max {
			max = n
		}
	}
	return
}

// The required height of a column is the total requirements of its contents
func (t Column) ReqHeight() (min, max int) {
	for _, w := range t {
		m, n := w.ReqHeight()
		min += m
		max += n
	}
	return
}

func (t Column) Draw(p *Pen) {
	w, h := p.Width(), p.Height()
	// println("column", p.x, p.y, p.w, p.h)
	sz := fitSize(h, t, func(t Widget) (int, int) {
		return t.ReqHeight()
	})

	y := 0
	for i, t := range t {
		th := sz[i]
		t.Draw(p.Clip(0, y, w, th))
		y += th
	}
}

// fitSize iterates through sz, adjusting it upwards until either all sz == max or n is exhausted
func fitSize(n int, t []Widget, sel func(Widget) (int, int)) []int {
	max := make([]int, len(t))
	sz := make([]int, len(t))
	for i, t := range t {
		sz[i], max[i] = sel(t)
		n -= sz[i]
	}

	// this is really baroque, but it fits
	more := true
	for n > 0 && more {
		more = false
		for i, max := range max {
			if sz[i] == max {
				continue
			}
			more = true
			sz[i]++
			n--
		}
	}

	return sz
}

type Pen struct {
	x, y, w, h int               // described the area the pen can draw in (the clip)
	dx, dy     int               // where we are printing in the clipped area
	fg, bg     termbox.Attribute // how we should be printing
}

func (p *Pen) Clip(x, y, w, h int) *Pen {
	// println("clip[", p.x, p.y, p.w, p.h, "] -> [", x, y, w, h, "]")
	if p == nil {
		return nil
	}

	o := new(Pen)

	o.x = p.x + x
	o.y = p.y + y
	o.w = w
	o.h = h

	switch {
	case o.x < 0:
		return nil
	case o.y < 0:
		return nil
	case o.w < 0:
		return nil
	case o.h < 0:
		return nil
	case o.x+o.w > p.x+p.w:
		return nil
	case o.y+o.h > p.y+p.h:
		return nil
	}

	o.fg = p.fg
	o.bg = p.bg

	return o
}

func (p *Pen) Clear() {
	if p == nil {
		return
	}

	buf := termbox.CellBuffer()
	sw, _ := termbox.Size()
	cell := termbox.Cell{Ch: ' ', Bg: p.bg, Fg: p.fg}
	for y := 0; y < p.h; y++ {
		b := buf[y*sw+p.x:]
		for x := 0; x < p.w; x++ {
			b[x] = cell
		}
	}
	p.dx, p.dy = 0, 0
}

func (p *Pen) Width() int {
	if p == nil {
		return 0
	}
	return p.w
}

func (p *Pen) Height() int {
	if p == nil {
		return 0
	}
	return p.h
}

func (p *Pen) SetFg(fg termbox.Attribute) {
	if p == nil {
		return
	}
	p.fg = fg
}

func (p *Pen) SetBg(bg termbox.Attribute) {
	if p == nil {
		return
	}
	p.bg = bg
}

func (p *Pen) Println(s string) {
	if p == nil {
		return
	}
	p.Print(s)
	p.dx = 0
	p.dy++
}

func (p *Pen) Print(s string) {
	if p == nil {
		return
	}

	n := 0
	for n < len(s) && p.dy < p.h {
		n += p.printRow(s[n:])
		p.dx = 0
		p.dy++
	}
}

func (p *Pen) printRow(s string) int {
	if p == nil {
		return 0
	}

	stride, _ := termbox.Size()
	x := p.dx + p.x
	y := p.dy + p.y

	row := termbox.CellBuffer()[x+y*stride:]
	row = row[:p.w-p.dx]

	ct := utf8.RuneCountInString(s)
	if len(row) < ct {
		s = s[:ct]
	}

	for i, r := range s {
		row[i] = termbox.Cell{Fg: p.fg, Bg: p.bg, Ch: r}
	}
	p.dx += len(row)
	return len(row)
}
