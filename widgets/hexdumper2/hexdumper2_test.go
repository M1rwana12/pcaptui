// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package hexdumper2

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// This is the pane that shows a packet's bytes, and every way it can be wrong
// is quiet: the wrong byte highlighted, the ascii column describing a
// different byte than the hex beside it, a cursor that has walked off the end.
// None of that crashes and none of it looks wrong on the screen.
//
// The column arithmetic below is written out from the layout comment in
// hexdumper2.go rather than taken from the widget's own variables, so that a
// change to one of them has to be a deliberate change to both.

const (
	hexCol0   = 1 + 4 + 3 // where the first "41" goes
	asciiCol0 = 1 + 4 + 3 + ((8 * 3) - 1) + 2 + ((8 * 3) - 1) + 3
)

// hexAt is the x of the first of the two hex digits for byte k of a row. The
// extra column is the wider gap tshark-style dumps leave in the middle.
func hexAt(k int) int {
	x := hexCol0 + (k * 3)
	if k >= 8 {
		x++
	}
	return x
}

// asciiAt is the x of the character standing for byte k of a row.
func asciiAt(k int) int {
	x := asciiCol0 + k
	if k >= 8 {
		x++
	}
	return x
}

func lineOf(t *testing.T, c gowid.ICanvas, row int) string {
	lines := strings.Split(c.String(), "\n")
	require.Greater(t, len(lines), row)
	return lines[row]
}

func charAt(c gowid.ICanvas, x, y int) string {
	return string(c.CellAt(x, y).Rune())
}

func fgAt(c gowid.ICanvas, x, y int) gowid.TCellColor {
	return c.CellAt(x, y).ForegroundColor()
}

// key builds the event a terminal sends for a key that is not a rune. The
// rune has to be zero: gowid matches a key press on both the key and the
// rune, so an event carrying a space alongside KeyRight matches nothing and
// the widget looks as though it ignored the arrow.
func key(k tcell.Key) *tcell.EventKey {
	return tcell.NewEventKey(k, 0, tcell.ModNone)
}

//======================================================================

// The hex and the ascii are two readings of the same byte. If the column
// arithmetic drifts for one of them, the pane still looks like a hexdump -
// it just describes a byte that isn't there.
func TestTheAsciiColumnDescribesTheByteBesideIt(t *testing.T) {
	// Deliberately mixed: printable, control, high bytes, and a space.
	data := []byte{0x48, 0x69, 0x00, 0x20, 0x7e, 0x7f, 0x80, 0xff,
		0x41, 0x0a, 0x39, 0x1f, 0x5c, 0xc3, 0x2e, 0x21}
	w := New(data)

	c := w.Render(gowid.RenderBox{C: 76, R: 1}, gowid.NotSelected, gwtest.D)

	for k, b := range data {
		hex := charAt(c, hexAt(k), 0) + charAt(c, hexAt(k)+1, 0)
		assert.Equal(t, strings.ToLower(hexOf(b)), hex, "byte %d printed as hex", k)

		want := "."
		if b >= 32 && b <= 126 {
			want = string(rune(b))
		}
		assert.Equal(t, want, charAt(c, asciiAt(k), 0),
			"byte %d (%#02x) in the ascii column", k, b)
	}
}

func hexOf(b byte) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[b>>4], digits[b&0xf]})
}

// One exact line, to pin the shape of the thing: the offset, the two halves of
// eight, and the ascii. Everything else here tests a rule; this tests the
// format a reader actually sees.
func TestOneLineLooksLikeAHexdump(t *testing.T) {
	w := New([]byte("HELLO WORLD 0123"))

	c := w.Render(gowid.RenderBox{C: 76, R: 1}, gowid.NotSelected, gwtest.D)

	assert.Equal(t,
		" 0000   48 45 4c 4c 4f 20 57 4f  52 4c 44 20 30 31 32 33   HELLO WO RLD 0123",
		strings.TrimRight(lineOf(t, c, 0), " "))
}

// Sixteen bytes to a row, and the offset in the margin is the offset of the
// row's first byte - not the row number.
func TestEachRowIsSixteenBytesAndSaysWhereItStarts(t *testing.T) {
	data := make([]byte, 33)
	for i := range data {
		data[i] = byte(i)
	}
	w := New(data)

	c := w.Render(gowid.RenderBox{C: 76, R: 3}, gowid.NotSelected, gwtest.D)

	assert.Equal(t, "0000", offsetOn(c, 0))
	assert.Equal(t, "0010", offsetOn(c, 1))
	assert.Equal(t, "0020", offsetOn(c, 2))

	// The third row holds the one byte that is left over, and nothing else.
	assert.Equal(t, "20", charAt(c, hexAt(0), 2)+charAt(c, hexAt(0)+1, 2))
	assert.Equal(t, " ", charAt(c, hexAt(1), 2))
}

func offsetOn(c gowid.ICanvas, row int) string {
	var s string
	for x := 1; x < 5; x++ {
		s += charAt(c, x, row)
	}
	return s
}

//======================================================================

// The cursor marks one byte, in both columns, and nothing else. This is
// compared cell against cell in the same render - reading the colour out of
// the theme and grepping for it does not survive 256-colour quantisation.
func TestTheCursorMarksItsOwnByteInBothColumns(t *testing.T) {
	w := New([]byte("0123456789abcdef"), Options{
		CursorUnselected: "test1notfocus",
		CursorSelected:   "test1focus",
	})
	w.SetPosition(3, gwtest.D)

	c := w.Render(gowid.RenderBox{C: 76, R: 1}, gowid.NotSelected, gwtest.D)

	cursor := fgAt(c, hexAt(3), 0)
	plain := fgAt(c, hexAt(5), 0)
	require.NotEqual(t, plain, cursor, "the cursor byte is not coloured at all")

	// Both hex digits and the ascii character, and not the neighbours.
	assert.Equal(t, cursor, fgAt(c, hexAt(3)+1, 0))
	assert.Equal(t, cursor, fgAt(c, asciiAt(3), 0))
	assert.Equal(t, plain, fgAt(c, hexAt(2), 0))
	assert.Equal(t, plain, fgAt(c, hexAt(4), 0))
	assert.Equal(t, plain, fgAt(c, asciiAt(2), 0))
	assert.Equal(t, plain, fgAt(c, asciiAt(4), 0))
}

// A layer is the field or protocol the tree pane has selected, and its edges
// are where an off-by-one hides best: one byte too many belongs to the field
// before it, one too few and a byte of the field looks like padding.
func TestALayerColoursItsOwnBytesAndStopsThere(t *testing.T) {
	w := New([]byte("0123456789abcdef"), Options{
		CursorUnselected: "test1focus",
		StyledLayers: []LayerStyler{
			{Start: 2, End: 5, ColUnselected: "test1notfocus", ColSelected: "test1notfocus"},
		},
	})
	w.SetPosition(0, gwtest.D)

	c := w.Render(gowid.RenderBox{C: 76, R: 1}, gowid.NotSelected, gwtest.D)

	layer := fgAt(c, hexAt(3), 0)
	plain := fgAt(c, hexAt(9), 0)
	cursor := fgAt(c, hexAt(0), 0)
	require.NotEqual(t, plain, layer, "the layer is not coloured at all")
	require.NotEqual(t, layer, cursor, "the cursor and the layer share a colour")

	// End is the first byte after the layer, so 2, 3, 4 are in and 5 is out.
	for _, k := range []int{2, 3, 4} {
		assert.Equal(t, layer, fgAt(c, hexAt(k), 0), "byte %d should be in the layer", k)
		assert.Equal(t, layer, fgAt(c, asciiAt(k), 0), "byte %d ascii should be in the layer", k)
	}
	assert.Equal(t, plain, fgAt(c, hexAt(1), 0), "the byte before the layer")
	assert.Equal(t, plain, fgAt(c, hexAt(5), 0), "End is exclusive")
	assert.Equal(t, plain, fgAt(c, asciiAt(5), 0), "End is exclusive in ascii too")
}

//======================================================================

// Right at the last byte used to move the cursor one past the end of the
// packet. Nothing crashed: Render draws no cursor at all for a position that
// is not a byte, and the tree pane, asked which field owns byte len(data),
// answers with nothing - so the highlight simply vanished and the two panes
// stopped agreeing. Compare Down, which has always clamped to len-1.
func TestRightAtTheLastByteStaysOnTheLastByte(t *testing.T) {
	data := []byte("0123456789")
	w := New(data)
	w.SetPosition(len(data)-1, gwtest.D)

	handled := w.UserInput(key(tcell.KeyRight),
		gowid.RenderBox{C: 76, R: 4}, gowid.Focused, gwtest.D)

	assert.Equal(t, len(data)-1, w.Position())
	assert.False(t, handled, "there was nowhere to go")
}

// The other half of that guard: the last byte is still reachable.
func TestRightFromTheByteBeforeReachesTheLastOne(t *testing.T) {
	data := []byte("0123456789")
	w := New(data)
	w.SetPosition(len(data)-2, gwtest.D)

	handled := w.UserInput(key(tcell.KeyRight),
		gowid.RenderBox{C: 76, R: 4}, gowid.Focused, gwtest.D)

	assert.Equal(t, len(data)-1, w.Position())
	assert.True(t, handled)
}

func TestLeftAtTheFirstByteStaysThere(t *testing.T) {
	w := New([]byte("0123456789"))
	w.SetPosition(0, gwtest.D)

	handled := w.UserInput(key(tcell.KeyLeft),
		gowid.RenderBox{C: 76, R: 4}, gowid.Focused, gwtest.D)

	assert.Equal(t, 0, w.Position())
	assert.False(t, handled, "there was nowhere to go")
}

func TestTheArrowsMoveOneByteAndOneRow(t *testing.T) {
	w := New([]byte("0123456789abcdefghijklmnopqrstuv"))
	w.SetPosition(0, gwtest.D)
	size := gowid.RenderBox{C: 76, R: 2}

	w.UserInput(key(tcell.KeyRight), size, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, w.Position())

	w.UserInput(key(tcell.KeyDown), size, gowid.Focused, gwtest.D)
	assert.Equal(t, 17, w.Position(), "down is the same column, one row on")

	w.UserInput(key(tcell.KeyUp), size, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, w.Position())
}

// Down on the last row has nowhere to go, and the byte it settles on has to be
// a byte that exists - the last row is usually a short one.
func TestDownOnTheLastRowStopsAtTheLastByte(t *testing.T) {
	data := []byte("0123456789abcdefghij") // 20 bytes: a full row and four
	w := New(data)
	w.SetPosition(15, gwtest.D)

	w.UserInput(key(tcell.KeyDown),
		gowid.RenderBox{C: 76, R: 2}, gowid.Focused, gwtest.D)

	assert.Equal(t, len(data)-1, w.Position())
}

// A click has to land on the byte under the pointer in all four places a byte
// is drawn: either half of the hex, either half of the ascii.
func TestAClickLandsOnTheByteUnderThePointer(t *testing.T) {
	w := New([]byte("0123456789abcdefghijklmnopqrstuv"))
	size := gowid.RenderBox{C: 76, R: 2}

	for _, k := range []int{0, 7, 8, 15} {
		for _, place := range []struct {
			what string
			x    int
		}{
			{"hex", hexAt(k)},
			{"hex second digit", hexAt(k) + 1},
			{"ascii", asciiAt(k)},
		} {
			w.SetPosition(0, gwtest.D)
			ev := tcell.NewEventMouse(place.x, 1, tcell.Button1, tcell.ModNone)

			w.UserInput(ev, size, gowid.Focused, gwtest.D)

			assert.Equal(t, 16+k, w.Position(),
				"clicking the %s of byte %d on the second row", place.what, k)
		}
	}
}

//======================================================================

// The scrollbar beside the pane reads these two numbers. If they disagree with
// the data the bar is simply a lie - it does not fail.
func TestTheScrollbarNumbersMatchTheData(t *testing.T) {
	for _, tc := range []struct {
		bytes int
		rows  int
	}{{0, 0}, {1, 1}, {16, 1}, {17, 2}, {32, 2}, {33, 3}} {
		w := New(make([]byte, tc.bytes))
		assert.Equal(t, tc.rows, w.ScrollLength(), "%d bytes", tc.bytes)
	}

	w := New(make([]byte, 64))
	w.SetPosition(33, gwtest.D)
	assert.Equal(t, 2, w.ScrollPosition(), "byte 33 is on row 2")
}

// SetPosition is how another pane moves this one - the tree pane selects a
// field and the bytes follow. The cursor has to be on screen afterwards,
// otherwise the pane looks like it ignored the request.
func TestMovingTheCursorOffScreenScrollsItIntoView(t *testing.T) {
	data := make([]byte, 16*8)
	w := New(data, Options{CursorUnselected: "test1notfocus"})

	// Two rows visible, cursor sent to row 5.
	w.SetPosition(16*5, gwtest.D)
	c := w.Render(gowid.RenderBox{C: 76, R: 2}, gowid.NotSelected, gwtest.D)

	assert.Equal(t, "0040", offsetOn(c, 0), "the last visible row is the cursor's")
	assert.Equal(t, "0050", offsetOn(c, 1))

	// And back up again.
	w.SetPosition(0, gwtest.D)
	c = w.Render(gowid.RenderBox{C: 76, R: 2}, gowid.NotSelected, gwtest.D)

	assert.Equal(t, "0000", offsetOn(c, 0))
}

func TestGoHomeAndGoToEndLandOnBytesThatExist(t *testing.T) {
	data := make([]byte, 100)
	w := New(data)
	size := gowid.RenderBox{C: 76, R: 3}

	w.GoToEnd(size, gwtest.D)
	assert.Equal(t, len(data)-1, w.Position())

	w.GoHome(size, gwtest.D)
	assert.Equal(t, 0, w.Position())
}

// Two panes call each other's callbacks when a position changes - the hex pane
// moves the tree, which sets the layers, which can move the hex pane. A
// callback for a position that did not change is how that becomes a loop.
func TestThePositionCallbackFiresOnlyOnAChange(t *testing.T) {
	w := New([]byte("0123456789"))
	fired := 0
	w.OnPositionChanged(gowid.MakeWidgetCallback("test",
		func(app gowid.IApp, target gowid.IWidget) {
			fired++
		}))

	w.SetPosition(4, gwtest.D)
	assert.Equal(t, 1, fired)

	w.SetPosition(4, gwtest.D)
	assert.Equal(t, 1, fired, "the position did not change")

	w.SetPosition(5, gwtest.D)
	assert.Equal(t, 2, fired)
}

//======================================================================

// Copy mode offers the bytes in four forms. Each one is a different reading of
// the same range, and a range that is off by one copies a byte the user did
// not point at - which is only ever noticed somewhere else, later.
func TestCopyOffersFourFormsOfTheSameRange(t *testing.T) {
	data := []byte{0x41, 0x42, 0x00, 0x43}

	clips := clipsForBytes(data, 1, 3)
	require.Len(t, clips, 4)

	for _, clip := range clips {
		assert.Contains(t, clip.ClipName(), "1-3", "every name says which bytes")
	}

	vals := make([]string, 0, 4)
	for _, clip := range clips {
		vals = append(vals, clip.ClipValue())
	}

	assert.Contains(t, vals[0], "42 00", "hex + ascii dump of bytes 1 and 2")
	assert.Equal(t, `"\x42\x00"`, vals[1], "escaped string: every byte, quoted, ready to paste")
	assert.Equal(t, "B", vals[2], "printable string: the byte that can be read, and only it")
	assert.Equal(t, "4200", vals[3], "hex stream")
}

func TestCopyingTheWholePacketIsTheWholePacket(t *testing.T) {
	data := []byte{0x41, 0x42, 0x43}

	clips := clipsForBytes(data, 0, len(data))

	assert.Equal(t, "414243", clips[3].ClipValue())
}

// The number of copy-mode levels is the number of layers, because each level
// copies one of them. Clips indexes w.layers by level, so a mismatch here is
// an out-of-range panic rather than a wrong answer.
func TestThereIsOneCopyLevelPerLayer(t *testing.T) {
	w := New([]byte("0123456789"), Options{
		StyledLayers: []LayerStyler{{Start: 0, End: 4}, {Start: 1, End: 2}},
	})

	assert.Equal(t, 2, w.CopyModeLevels())
}

//======================================================================

// A packet with no bytes is not a reason to panic, and there is nothing to
// draw for it either. Every fixture in the repository has bytes, so this is
// the case no other test would reach.
func TestNoDataDrawsNothing(t *testing.T) {
	w := New(nil)
	size := gowid.RenderBox{C: 76, R: 4}

	// Asked to fill a box, it fills it with nothing rather than refusing.
	c := w.Render(size, gowid.NotSelected, gwtest.D)
	assert.Equal(t, 4, c.BoxRows())
	assert.Equal(t, "", strings.TrimSpace(c.String()), "there are no bytes to draw")

	// Asked how tall it wants to be, it says nothing at all.
	assert.Equal(t, 0, w.Render(gowid.RenderFlowWith{C: 76}, gowid.NotSelected, gwtest.D).BoxRows())
	assert.Equal(t, 0, w.ScrollLength())

	// None of the movements have a byte to land on; none of them may panic.
	w.Up(1, size, gwtest.D)
	w.Down(1, size, gwtest.D)
	w.UpPage(1, size, gwtest.D)
	w.DownPage(1, size, gwtest.D)
	w.GoHome(size, gwtest.D)
	assert.Equal(t, 0, w.Position())

	w.UserInput(key(tcell.KeyRight), size, gowid.Focused, gwtest.D)
	w.UserInput(key(tcell.KeyDown), size, gowid.Focused, gwtest.D)
	w.UserInput(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone), size, gowid.Focused, gwtest.D)
	w.GoToEnd(size, gwtest.D)
	assert.Equal(t, 0, w.Position(), "there is no byte to move to")
}

// SetData is how the pane is handed the next packet. The bytes have to change
// with it, and the widget keeps its identity so the callbacks stay attached.
func TestSetDataReplacesTheBytes(t *testing.T) {
	w := New([]byte("aaaa"))

	w.SetData([]byte("bbbbbbbb"), gwtest.D)

	assert.Equal(t, []byte("bbbbbbbb"), w.Data())
	c := w.Render(gowid.RenderBox{C: 76, R: 1}, gowid.NotSelected, gwtest.D)
	assert.Equal(t, "62", charAt(c, hexAt(0), 0)+charAt(c, hexAt(0)+1, 0))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
