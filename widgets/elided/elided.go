// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package elided renders text that says when it has been cut short.
//
// A packet list column that is too narrow simply stops drawing, so
// 192.0.2.10 in an eighty-column terminal appears as 192.0.2.1 - a different
// address, equally valid, with nothing to say a digit is missing. Somebody
// reading that writes down the wrong host. Every value in the list can be cut
// this way; addresses are only where it does the most damage.
package elided

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/text"
	"github.com/mattn/go-runewidth"
)

//======================================================================

// Marker ends a value that did not fit. One cell wide, and distinct from
// anything that appears in an address, a protocol name or a length.
const Marker = '…'

// Clip shortens s to fit width columns, ending it with Marker when anything
// was removed.
//
// Width is counted in terminal cells rather than runes: a CJK hostname or an
// emoji in a payload occupies two, and counting runes would overflow the
// column and push the rest of the row sideways.
func Clip(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= width {
		return s
	}
	if width == 1 {
		return string(Marker)
	}

	// Leave one cell for the marker.
	room := width - 1
	used := 0
	out := make([]rune, 0, width)

	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if used+w > room {
			break
		}
		out = append(out, r)
		used += w
	}

	return string(out) + string(Marker)
}

//======================================================================

// Widget is a text widget that clips to the width it is given, when it is
// given only one row to draw in.
type Widget struct {
	txt string
}

var _ gowid.IWidget = (*Widget)(nil)

func New(txt string) *Widget {
	return &Widget{txt: txt}
}

func (w *Widget) Text() string {
	return w.txt
}

// widthOf reports the width to clip to, and whether clipping applies at all.
//
// Only a render that pins the height to a single row is clipped. The packet
// list wraps every cell in expander, which draws the focused row at its
// natural height and each other row inside a one-line box - so the row under
// the cursor still shows its value in full, wrapped onto as many lines as it
// needs, and only the rows with nowhere to put the rest of the text get a
// marker. A flow or fixed render is left alone.
func widthOf(size gowid.IRenderSize) (int, bool) {
	if box, ok := size.(gowid.IRenderBox); ok && box.BoxRows() == 1 {
		return box.BoxColumns(), true
	}
	return 0, false
}

func (w *Widget) inner(size gowid.IRenderSize) *text.Widget {
	if width, ok := widthOf(size); ok {
		return text.New(Clip(w.txt, width))
	}
	return text.New(w.txt)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return w.inner(size).Render(size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return w.inner(size).RenderSize(size, focus, app)
}

func (w *Widget) Selectable() bool {
	return false
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
