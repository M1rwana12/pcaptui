// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func TestTheSamePairGetsTheSameIndex(t *testing.T) {
	p := NewColorPalette()

	a := p.Add("#000000", "#d7ffd7")
	b := p.Add("#000000", "#d7ffd7")

	assert.Equal(t, a, b)
	assert.Equal(t, 1, p.Len(), "the same pair should be stored once")
}

func TestDifferentPairsGetDifferentIndices(t *testing.T) {
	p := NewColorPalette()

	a := p.Add("#000000", "#d7ffd7")
	b := p.Add("#000000", "#ffffff")
	c := p.Add("#ffffff", "#000000")

	assert.NotEqual(t, a, b)
	assert.NotEqual(t, b, c)
	assert.Equal(t, 3, p.Len())
}

// Foreground and background are not interchangeable: black on white and white
// on black are different rows on screen.
func TestTheOrderOfThePairMatters(t *testing.T) {
	p := NewColorPalette()

	assert.NotEqual(t,
		p.Add("#000000", "#ffffff"),
		p.Add("#ffffff", "#000000"))
}

func TestAnIndexGivesBackItsColours(t *testing.T) {
	p := NewColorPalette()

	i := p.Add("#000000", "#d7ffd7")
	got := p.At(i)

	require.NotNil(t, got.FG)
	require.NotNil(t, got.BG)
	assert.NotEqual(t, got.FG, got.BG)
}

// A row asked about before its colour has arrived, which happens while a
// capture is still loading.
func TestAnIndexNobodyAddedIsBlankNotAPanic(t *testing.T) {
	p := NewColorPalette()

	assert.Equal(t, PacketColors{}, p.At(0))
	assert.Equal(t, PacketColors{}, p.At(9999))
}

// tshark writes an empty attribute when a packet matches no colour rule.
func TestAnEmptyColourIsStillAPair(t *testing.T) {
	p := NewColorPalette()

	i := p.Add("", "")

	assert.Equal(t, 1, p.Len())
	assert.Equal(t, PacketColors{}, p.At(i), "unparseable colours are nil, not a crash")
}

// A million packets drawing from a handful of rules is the case this exists
// for: the palette must not grow with the capture.
func TestThePaletteGrowsWithRulesNotWithPackets(t *testing.T) {
	p := NewColorPalette()

	for i := 0; i < 100000; i++ {
		p.Add(fmt.Sprintf("#%06x", i%7), fmt.Sprintf("#%06x", i%3))
	}

	assert.LessOrEqual(t, p.Len(), 21,
		"seven foregrounds and three backgrounds cannot make more than 21 pairs")
}

//======================================================================

func TestLoaderColoursComeBackByRow(t *testing.T) {
	p := &PsmlLoader{colorPalette: NewColorPalette()}

	p.packetPsmlColorIdx = append(p.packetPsmlColorIdx,
		p.colorPalette.Add("#000000", "#d7ffd7"),
		p.colorPalette.Add("#ffffff", "#800000"))

	assert.Equal(t, 2, p.NumColors())
	assert.NotEqual(t, p.ColorAt(0), p.ColorAt(1))
}

// The packet list asks for a row's colour while the capture is still loading,
// so a row past the end has to be an answer rather than a panic.
func TestALoaderRowWithNoColourYetIsBlank(t *testing.T) {
	p := &PsmlLoader{colorPalette: NewColorPalette()}

	assert.Equal(t, 0, p.NumColors())
	assert.Equal(t, PacketColors{}, p.ColorAt(0))
	assert.Equal(t, PacketColors{}, p.ColorAt(-1))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
