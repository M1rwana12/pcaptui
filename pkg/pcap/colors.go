// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	log "github.com/sirupsen/logrus"
)

//======================================================================

// ColorPalette stores each distinct foreground/background pair once and gives
// every packet an index into it.
//
// A packet used to carry its colours directly, as two gowid.IColor interfaces
// built by parsing the hex strings tshark printed. Measured over a million
// packets: 124 MB and eight million allocations, because the parse happens per
// packet and allocates. The same million packets as indices are under 1 MB and
// one allocation.
//
// The pairs are few. tshark's colours come from the colour filter rules in the
// Wireshark profile, and the default profile has about twenty; a capture with
// a million packets still draws from that handful.
//
// uint16 rather than uint8 because 255 is close enough to a plausible number of
// custom rules to be worth avoiding, and two bytes a packet is still 2 MB per
// million.
type ColorPalette struct {
	pairs []PacketColors
	index map[colorKey]uint16
}

// colorKey is a struct rather than the two strings joined, because joining
// them allocates - once per packet, which is a million allocations on a
// million-packet capture for a key that is thrown away immediately. A struct
// of two strings is a valid map key and copies no bytes.
type colorKey struct {
	fg, bg string
}

// maxColorPairs is what a uint16 index can address. A profile with more rules
// than this does not exist, but the code says what happens if one does rather
// than wrapping around to somebody else's colour.
const maxColorPairs = 1 << 16

func NewColorPalette() *ColorPalette {
	return &ColorPalette{
		index: make(map[colorKey]uint16),
	}
}

// Add returns the index for a pair of colours as tshark wrote them, parsing
// them only the first time each pair is seen.
func (p *ColorPalette) Add(fg string, bg string) uint16 {
	key := colorKey{fg: fg, bg: bg}

	if i, ok := p.index[key]; ok {
		return i
	}

	if len(p.pairs) >= maxColorPairs {
		// Beyond this the index cannot say which pair it means. Reusing the
		// first pair is wrong, but visibly and uniformly so - and it is said
		// out loud, which a silent wrap would not be.
		log.Warnf("More than %d distinct packet colours; further packets reuse the first", maxColorPairs)
		return 0
	}

	i := uint16(len(p.pairs))
	p.pairs = append(p.pairs, PacketColors{
		FG: psmlColorToIColor(fg),
		BG: psmlColorToIColor(bg),
	})
	p.index[key] = i

	return i
}

// At returns the colours an index stands for.
func (p *ColorPalette) At(i uint16) PacketColors {
	if int(i) >= len(p.pairs) {
		return PacketColors{}
	}
	return p.pairs[i]
}

// Len is how many distinct pairs the capture turned out to have. Useful for
// saying whether the palette is doing anything.
func (p *ColorPalette) Len() int {
	return len(p.pairs)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
