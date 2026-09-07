// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package screenshot

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func drawn(w, h int, draw func(s tcell.SimulationScreen)) ([][]Cell, int, int) {
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		panic(err)
	}
	s.SetSize(w, h)
	draw(s)
	s.Show()
	return Grab(s)
}

func put(s tcell.SimulationScreen, x, y int, text string, st tcell.Style) {
	for i, r := range text {
		s.SetContent(x+i, y, r, nil, st)
	}
}

//======================================================================

func TestTextReadsBackWhatWasDrawn(t *testing.T) {
	rows, _, _ := drawn(20, 2, func(s tcell.SimulationScreen) {
		put(s, 0, 0, "packet 1", tcell.StyleDefault)
	})

	assert.Equal(t, "packet 1\n\n", Text(rows))
}

// A cell nobody wrote to comes back as a NUL rune, not a space. Left alone it
// would put a zero byte in every golden file.
func TestUntouchedCellsBecomeSpaces(t *testing.T) {
	rows, _, _ := drawn(6, 1, func(s tcell.SimulationScreen) {
		put(s, 0, 0, "ab", tcell.StyleDefault)
	})

	assert.NotContains(t, Text(rows), "\x00")
	assert.Equal(t, "ab\n", Text(rows))
}

func TestTextTrimsTrailingBlanksButKeepsRows(t *testing.T) {
	rows, _, _ := drawn(10, 3, func(s tcell.SimulationScreen) {
		put(s, 0, 1, "middle", tcell.StyleDefault)
	})

	assert.Equal(t, "\nmiddle\n\n", Text(rows))
}

//======================================================================

func TestSVGIsWellFormedAndCarriesTheText(t *testing.T) {
	rows, w, h := drawn(12, 1, func(s tcell.SimulationScreen) {
		put(s, 0, 0, "TCP 74 SYN", tcell.StyleDefault)
	})

	out := SVG(rows, w, h)

	assert.True(t, strings.HasPrefix(out, "<svg "))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(out), "</svg>"))
	assert.Contains(t, out, "TCP 74 SYN")
}

func TestSVGEscapesMarkupInPacketData(t *testing.T) {
	rows, w, h := drawn(20, 1, func(s tcell.SimulationScreen) {
		put(s, 0, 0, "a<b>&c", tcell.StyleDefault)
	})

	out := SVG(rows, w, h)

	assert.Contains(t, out, "a&lt;b&gt;&amp;c")
	assert.NotContains(t, out, "<b>")
}

func TestSVGPaintsBackgroundsForStyledRuns(t *testing.T) {
	sel := tcell.StyleDefault.Background(tcell.NewRGBColor(240, 136, 62))
	rows, w, h := drawn(12, 1, func(s tcell.SimulationScreen) {
		put(s, 0, 0, "selected", sel)
	})

	out := SVG(rows, w, h)

	assert.Contains(t, out, "#F0883E", "the selected row needs its own background")
}

// One element per cell makes a file nobody can review. Cells sharing a style
// have to be emitted as a single run.
func TestSVGGroupsCellsIntoRuns(t *testing.T) {
	rows, w, h := drawn(60, 1, func(s tcell.SimulationScreen) {
		put(s, 0, 0, strings.Repeat("x", 60), tcell.StyleDefault)
	})

	out := SVG(rows, w, h)

	assert.Equal(t, 1, strings.Count(out, "<text"),
		"sixty identical cells should be one text element")
}

func TestRunsSplitOnStyleChange(t *testing.T) {
	bold := tcell.StyleDefault.Bold(true)
	rows, _, _ := drawn(9, 1, func(s tcell.SimulationScreen) {
		put(s, 0, 0, "abc", tcell.StyleDefault)
		put(s, 3, 0, "def", bold)
		put(s, 6, 0, "ghi", tcell.StyleDefault)
	})

	assert.Len(t, runs(rows[0]), 3)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
