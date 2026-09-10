// Copyright 2026 m1rwana12. All rights reserved.
// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package resizable

import (
	"encoding/json"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// This package is what happens when the user drags the divider between two
// panes. Every failure it can have is quiet: a column one cell wider than the
// screen, a pane that has been given a negative width, or a divider that
// swallows a neighbour it was only supposed to borrow from.

func threeColumns(widths ...int) *ColumnsWidget {
	cols := make([]gowid.IContainerWidget, 0, len(widths))
	for _, w := range widths {
		cols = append(cols, &gowid.ContainerWidget{
			IWidget: fill.New(' '),
			D:       gowid.RenderWithUnits{U: w},
		})
	}
	return NewColumns(cols)
}

// Dragging a divider moves width from one column to the other. What it must
// never do is change how much width there is - the columns have to keep adding
// up to the screen they were given.
func TestAnAdjustmentMovesWidthAndDoesNotCreateIt(t *testing.T) {
	w := threeColumns(10, 10, 10)
	size := gowid.RenderBox{C: 30, R: 1}

	before := w.WidgetWidths(size, gowid.NotSelected, 0, gwtest.D)
	require.Equal(t, []int{10, 10, 10}, before)

	w.SetOffsets([]Offset{{Col1: 0, Col2: 1, Adjust: 3}}, gwtest.D)
	after := w.WidgetWidths(size, gowid.NotSelected, 0, gwtest.D)

	assert.Equal(t, []int{13, 7, 10}, after, "three cells moved from the second column to the first")
	assert.Equal(t, sum(before), sum(after), "the total width changed")
}

// A drag can ask for more than the neighbour has. The neighbour goes to zero;
// it does not go negative, and it does not take the difference out of a column
// nobody was dragging.
func TestAColumnIsNeverGivenANegativeWidth(t *testing.T) {
	w := threeColumns(10, 4, 10)
	size := gowid.RenderBox{C: 24, R: 1}

	w.SetOffsets([]Offset{{Col1: 0, Col2: 1, Adjust: 99}}, gwtest.D)
	widths := w.WidgetWidths(size, gowid.NotSelected, 0, gwtest.D)

	assert.Equal(t, []int{14, 0, 10}, widths)
	assert.Equal(t, 24, sum(widths), "the total width changed")
}

// And the same in the other direction: dragging a divider left cannot make the
// column being shrunk narrower than nothing.
func TestTheColumnBeingShrunkStopsAtZeroToo(t *testing.T) {
	w := threeColumns(4, 10, 10)
	size := gowid.RenderBox{C: 24, R: 1}

	w.SetOffsets([]Offset{{Col1: 0, Col2: 1, Adjust: -99}}, gwtest.D)
	widths := w.WidgetWidths(size, gowid.NotSelected, 0, gwtest.D)

	assert.Equal(t, []int{0, 14, 10}, widths)
	assert.Equal(t, 24, sum(widths), "the total width changed")
}

func sum(xs []int) int {
	var t int
	for _, x := range xs {
		t += x
	}
	return t
}

//======================================================================

// The first drag of a pair of columns has no offset to adjust, so one is made.
// Every drag after that adds to the same one - a second entry for the same pair
// would apply twice on every render.
func TestTheFirstDragMakesAnOffsetAndTheRestAddToIt(t *testing.T) {
	w := threeColumns(10, 10, 10)

	w.AdjustOffset(0, 1, Add1, gwtest.D)
	require.Len(t, w.GetOffsets(), 1)
	assert.Equal(t, Offset{Col1: 0, Col2: 1, Adjust: 1}, w.GetOffsets()[0])

	w.AdjustOffset(0, 1, Add1, gwtest.D)
	w.AdjustOffset(0, 1, Add1, gwtest.D)
	require.Len(t, w.GetOffsets(), 1, "one entry per pair of columns")
	assert.Equal(t, 3, w.GetOffsets()[0].Adjust)

	w.AdjustOffset(0, 1, Subtract1, gwtest.D)
	assert.Equal(t, 2, w.GetOffsets()[0].Adjust)

	// A different pair of columns is a different offset.
	w.AdjustOffset(1, 2, Add1, gwtest.D)
	require.Len(t, w.GetOffsets(), 2)
	assert.Equal(t, Offset{Col1: 1, Col2: 2, Adjust: 1}, w.GetOffsets()[1])
}

// The callback is how the rest of the program learns the panes were resized -
// it is what saves the new layout.
//
// The first drag of a pair announces itself twice: AdjustOffset makes the new
// entry through SetOffsets, which notifies, and then notifies again itself.
// Written down rather than corrected, because the listener saves the layout and
// saving it twice costs nothing - but a listener that counted drags, or moved
// something by a step per call, would be wrong here and would have no way of
// knowing.
func TestAdjustingTellsWhoeverAsked(t *testing.T) {
	w := threeColumns(10, 10, 10)
	fired := 0
	w.OnOffsetsSet(gowid.MakeWidgetCallback("test",
		func(app gowid.IApp, target gowid.IWidget) {
			fired++
		}))

	w.AdjustOffset(0, 1, Add1, gwtest.D)
	assert.Equal(t, 2, fired, "once for the new entry, once for the adjustment")

	// Once the pair has an entry, a drag is one notification.
	w.AdjustOffset(0, 1, Add1, gwtest.D)
	assert.Equal(t, 3, fired)

	w.SetOffsets([]Offset{{Col1: 0, Col2: 1, Adjust: 5}}, gwtest.D)
	assert.Equal(t, 4, fired, "the layout was replaced wholesale")
}

// A pile is the same idea rotated: dragging a horizontal divider between the
// packet list and the panes below it.
func TestAPileAdjustsTheSameWay(t *testing.T) {
	rows := make([]gowid.IContainerWidget, 0, 2)
	for i := 0; i < 2; i++ {
		rows = append(rows, &gowid.ContainerWidget{
			IWidget: fill.New(' '),
			D:       gowid.RenderWithUnits{U: 5},
		})
	}
	w := NewPile(rows)

	w.AdjustOffset(0, 1, Add1, gwtest.D)

	require.Len(t, w.GetOffsets(), 1)
	assert.Equal(t, Offset{Col1: 0, Col2: 1, Adjust: 1}, w.GetOffsets()[0])
}

//======================================================================

// These offsets are written into the user's configuration file, so the three
// json names are a saved format and not an implementation detail: renaming a
// field silently loses the layout of everyone who already has one saved.
func TestTheSavedNamesAreTheOnesAlreadyOnDisk(t *testing.T) {
	out, err := json.Marshal(Offset{Col1: 2, Col2: 4, Adjust: 7})
	require.NoError(t, err)
	assert.JSONEq(t, `{"col1":2,"col2":4,"adjust":7}`, string(out))

	var back []Offset
	require.NoError(t, json.Unmarshal([]byte(`[{"col1":3,"col2":1,"adjust":15}]`), &back))
	assert.Equal(t, []Offset{{Col1: 3, Col2: 1, Adjust: 15}}, back)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
