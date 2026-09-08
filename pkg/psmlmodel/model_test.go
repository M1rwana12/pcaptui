// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package psmlmodel

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/table"
	"github.com/m1rwana12/pcaptui/widgets/elided"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// findElided digs through the styling, button and expander wrappers a cell is
// built from and returns the widget that actually holds the text.
func findElided(w gowid.IWidget) *elided.Widget {
	for i := 0; i < 10; i++ {
		if e, ok := w.(*elided.Widget); ok {
			return e
		}
		c, ok := w.(gowid.IComposite)
		if !ok {
			return nil
		}
		w = c.SubWidget()
	}
	return nil
}

func testModel() *Model {
	return New(table.NewSimpleModel(
		[]string{"No.", "Source", "Dest"},
		[][]string{{"1", "192.0.2.10", "198.51.100.20"}},
		table.SimpleOptions{},
	), nil)
}

// The packet list has to build its cells from elided rather than gowid's text
// widget. Without it a column too narrow for an address shows a shorter
// address instead of saying it ran out of room - see widgets/elided.
func TestPacketListCellsCanSayTheyWereCutShort(t *testing.T) {
	w := testModel().CellWidget(1, "192.0.2.10")
	require.NotNil(t, w)

	e := findElided(w)
	require.NotNil(t, e, "the cell is not built from an elided widget")
	assert.Equal(t, "192.0.2.10", e.Text())
}

func TestEveryCellOfARowIsBuiltTheSameWay(t *testing.T) {
	ws := testModel().CellWidgets(table.RowId(0))
	require.Len(t, ws, 3)

	for i, w := range ws {
		assert.NotNil(t, findElided(w), "column %d is not elided", i)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
