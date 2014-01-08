package main

import (
	morlock "../.."
	"github.com/swdunlop/termbox-go"
)

func main() {
	root := morlock.Grid{
		morlock.Row{
			morlock.Label("wind speed:"),
			morlock.Label("40 knots/s"),
		},
		morlock.Row{
			morlock.Label("species:"),
			morlock.Label("african swallow"),
			tint(termbox.ColorYellow, morlock.Column{
				morlock.Label("// multiple line"),
				morlock.Label("// comment"),
			}),
		},
		morlock.Row{
			morlock.Label("laden:"),
			tint(termbox.ColorRed, morlock.Label("true")),
			tint(termbox.ColorYellow, morlock.Label("// weight of coconut required!")),
		},
	}

	termbox.Init()
	defer termbox.Close()
	for {
		morlock.Draw(root)
		e := termbox.PollEvent()
		if e.Type != termbox.EventKey {
			continue
		}
		break
	}
}

func tint(fg termbox.Attribute, w morlock.Widget) morlock.Widget {
	return morlock.Tint{Fg: fg, Widget: w}
}
