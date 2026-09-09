// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pdmltree

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// A packet cut down to what the byte-to-field mapping needs: two layers, the
// second with two fields of its own, at known offsets.
//
// Ethernet is the fourteen bytes at 0, IP the twenty at 14, and inside IP the
// source and destination addresses at 26 and 30.
const layered = `<packet>
  <proto name="geninfo" pos="0" showname="General information" size="54"/>
  <proto name="frame" showname="Frame 1" size="54" pos="0"/>
  <proto name="eth" showname="Ethernet II" size="14" pos="0">
    <field name="eth.dst" showname="Destination" size="6" pos="0"/>
    <field name="eth.src" showname="Source" size="6" pos="6"/>
  </proto>
  <proto name="ip" showname="Internet Protocol Version 4" size="20" pos="14">
    <field name="ip.src" showname="Source Address" size="4" pos="26"/>
    <field name="ip.dst" showname="Destination Address" size="4" pos="30"/>
  </proto>
</packet>`

func layeredTree(t *testing.T) *Model {
	t.Helper()

	tree := DecodePacket([]byte(layered))
	require.NotNil(t, tree)

	return tree
}

// This is what tells the hex pane which bytes to mark. An off-byte here
// highlights the wrong bytes and nothing complains - the sibling of the bug
// that had the hex reader splicing bytes that were not in the packet.
func TestTheLayerAndTheFieldUnderTheCursorAreBothMarked(t *testing.T) {
	got := layeredTree(t).HexLayers(27, false)

	require.Len(t, got, 2, "the layer, then the field inside it")

	assert.Equal(t, 14, got[0].Start, "the IP layer")
	assert.Equal(t, 34, got[0].End)

	assert.Equal(t, 26, got[1].Start, "the source address inside it")
	assert.Equal(t, 30, got[1].End)
}

// A byte inside a layer but outside every one of its fields still marks the
// layer: the pane should say which protocol owns the byte even where nothing
// finer does.
func TestAByteInNoFieldStillMarksItsLayer(t *testing.T) {
	got := layeredTree(t).HexLayers(20, false)

	require.Len(t, got, 1)
	assert.Equal(t, 14, got[0].Start)
}

// The end of a field is the first byte after it. Reading that as inclusive
// marks one byte too many, and the byte belongs to whatever comes next.
func TestAFieldEndsBeforeItsLastOffset(t *testing.T) {
	inside := layeredTree(t).HexLayers(29, false)
	require.Len(t, inside, 2)
	assert.Equal(t, 26, inside[1].Start, "29 is the last byte of the source address")

	next := layeredTree(t).HexLayers(30, false)
	require.Len(t, next, 2)
	assert.Equal(t, 30, next[1].Start, "30 is the first byte of the destination")
}

// The first child is the geninfo block, which covers the whole packet and is
// not a layer of it - marking it would put a highlight over every byte.
func TestTheFirstChildIsSkippedUnlessAskedFor(t *testing.T) {
	without := layeredTree(t).HexLayers(0, false)
	with := layeredTree(t).HexLayers(0, true)

	assert.Greater(t, len(with), len(without),
		"including the first child adds the block that covers everything")
}

// A byte past the end of the packet belongs to nothing, which is an answer
// rather than a panic - the hex pane asks about the cursor, and the cursor can
// be anywhere the last render put it.
func TestAByteBeyondThePacketMarksNothing(t *testing.T) {
	assert.Empty(t, layeredTree(t).HexLayers(9999, false))
	assert.Empty(t, layeredTree(t).HexLayers(-1, false))
}

// The layer comes before the field it contains, because the pane draws them in
// order and the narrower mark has to win.
func TestTheLayerComesBeforeItsField(t *testing.T) {
	got := layeredTree(t).HexLayers(27, false)

	require.Len(t, got, 2)
	assert.Less(t, got[0].End-got[0].Start, 21)
	assert.Greater(t, got[0].End-got[0].Start, got[1].End-got[1].Start,
		"the layer is wider than the field inside it")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
