// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package elided

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/mattn/go-runewidth"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestTextThatFitsIsUntouched(t *testing.T) {
	assert.Equal(t, "192.0.2.10", Clip("192.0.2.10", 10))
	assert.Equal(t, "192.0.2.10", Clip("192.0.2.10", 40))
	assert.Equal(t, "", Clip("", 10))
}

// The whole reason this package exists. Without a marker, 192.0.2.10 becomes
// 192.0.2.1 - which is a real address, so nothing about the result looks
// wrong, and a reader copies down the wrong host.
func TestAnAddressCutShortSaysSo(t *testing.T) {
	got := Clip("192.0.2.10", 9)

	assert.Equal(t, "192.0.2.…", got)
	assert.NotEqual(t, "192.0.2.1", got, "a silent truncation is another valid address")
}

func TestClippedTextNeverExceedsTheWidth(t *testing.T) {
	for _, w := range []int{1, 2, 3, 5, 9, 12, 13} {
		got := Clip("198.51.100.20", w)
		assert.LessOrEqual(t, runewidth.StringWidth(got), w,
			"width %d produced %q", w, got)
	}
}

func TestNoRoomAtAll(t *testing.T) {
	assert.Equal(t, "", Clip("192.0.2.10", 0))
	assert.Equal(t, "", Clip("192.0.2.10", -1))
	assert.Equal(t, "…", Clip("192.0.2.10", 1))
}

// Cells, not runes. A hostname in a payload can be double width, and counting
// runes would overflow the column and shove the rest of the row sideways.
func TestDoubleWidthRunesAreCountedInCells(t *testing.T) {
	// Four ideographs: eight cells.
	s := "日本語版"

	assert.Equal(t, s, Clip(s, 8))

	got := Clip(s, 5)
	assert.LessOrEqual(t, runewidth.StringWidth(got), 5)
	assert.Contains(t, got, string(Marker))
}

func TestARuneIsNeverSplitInHalf(t *testing.T) {
	// Width 4 leaves room for one ideograph plus the marker; a second would
	// need two more cells and must not be half-drawn.
	got := Clip("日本語", 4)

	assert.Equal(t, "日…", got)
}

//======================================================================

func TestWidgetKeepsItsText(t *testing.T) {
	assert.Equal(t, "192.0.2.10", New("192.0.2.10").Text())
}

func TestWidgetIsNotSelectable(t *testing.T) {
	// A packet list cell is read, not focused; the row is what takes focus.
	assert.False(t, New("x").Selectable())
	assert.False(t, New("x").UserInput(nil, nil, gowid.Selector{}, nil))
}

// One row is the only case where text that does not fit has nowhere to go.
func TestOnlyASingleRowIsClipped(t *testing.T) {
	width, clip := widthOf(gowid.RenderBox{C: 9, R: 1})
	assert.True(t, clip)
	assert.Equal(t, 9, width)
}

// The focused row of the packet list is drawn by expander at its natural
// height, so the value is still there in full, wrapped. Clipping it would take
// away the one place the whole address can be read.
func TestARowWithSpaceToWrapIsNotClipped(t *testing.T) {
	for _, size := range []gowid.IRenderSize{
		gowid.RenderFlowWith{C: 9},
		gowid.RenderBox{C: 9, R: 2},
		gowid.RenderFixed{},
	} {
		_, clip := widthOf(size)
		assert.False(t, clip, "should not have clipped at %v", size)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
