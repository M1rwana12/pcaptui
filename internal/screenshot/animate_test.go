// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package screenshot

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func castFrames(t *testing.T, texts ...string) [][][]Cell {
	t.Helper()

	var res [][][]Cell
	for _, s := range texts {
		rows, _, _ := drawn(len(s)+2, 1, func(sc tcell.SimulationScreen) {
			put(sc, 0, 0, s, tcell.StyleDefault)
		})
		res = append(res, rows)
	}
	return res
}

func TestEveryFrameIsInTheAnimation(t *testing.T) {
	out := AnimatedSVG(castFrames(t, "one", "two", "three"), 8, 1, 1.5)

	assert.Contains(t, out, "one")
	assert.Contains(t, out, "two")
	assert.Contains(t, out, "three")
	assert.Equal(t, 3, strings.Count(out, "<animate "))
}

// Each frame gets its own slice of the loop, in order and without gaps or
// overlaps - two frames visible at once is a smear, and a gap is a flash of
// empty terminal.
func TestTheFramesShareTheLoopInOrder(t *testing.T) {
	out := AnimatedSVG(castFrames(t, "a", "b", "c", "d"), 6, 1, 1.0)

	assert.Contains(t, out, `dur="4.00s"`)
	assert.Contains(t, out, `keyTimes="0;0.2500"`)
	assert.Contains(t, out, `keyTimes="0;0.2500;0.5000"`)
	assert.Contains(t, out, `keyTimes="0;0.5000;0.7500"`)
	assert.Contains(t, out, `keyTimes="0;0.7500"`)
}

// A renderer that ignores SMIL shows what is painted, so the first frame has
// to be the one that stands on its own.
func TestTheFirstFrameStartsVisible(t *testing.T) {
	out := AnimatedSVG(castFrames(t, "first", "second"), 10, 1, 1.0)

	first := strings.Index(out, `<g opacity="1">`)
	second := strings.Index(out, `<g opacity="0">`)

	require.Positive(t, first)
	require.Positive(t, second)
	assert.Less(t, first, second, "the visible group should be the first one")
}

// A terminal does not fade between states.
func TestTheAnimationIsDiscrete(t *testing.T) {
	out := AnimatedSVG(castFrames(t, "a", "b"), 6, 1, 1.0)

	assert.Equal(t, 2, strings.Count(out, `calcMode="discrete"`))
	assert.NotContains(t, out, "calcMode=\"linear\"")
}

func TestOneFrameNeedsNoAnimation(t *testing.T) {
	out := AnimatedSVG(castFrames(t, "only"), 6, 1, 1.0)

	assert.NotContains(t, out, "<animate")
	assert.Contains(t, out, "only")
}

func TestNoFramesIsNoPicture(t *testing.T) {
	assert.Empty(t, AnimatedSVG(nil, 10, 10, 1.0))
}

// The still and the animation come out of the same renderer, so a change to
// one cannot silently leave the other behind.
func TestAFrameIsDrawnTheSameWayAStillIs(t *testing.T) {
	frames := castFrames(t, "shared")

	still := SVG(frames[0], 8, 1)
	anim := AnimatedSVG(append(frames, frames[0]), 8, 1, 1.0)

	assert.Contains(t, still, ">shared<")
	assert.Contains(t, anim, ">shared<")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
