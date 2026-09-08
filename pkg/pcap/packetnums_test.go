// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// A display filter leaves gaps: these are the packets that passed it, and
// their rows are 0, 1, 2, 3 whatever their numbers.
func filteredLoader() *PsmlLoader {
	return &PsmlLoader{packetNumbers: []int32{12, 44, 71, 72}}
}

func TestAMarkedPacketFindsItsRow(t *testing.T) {
	p := filteredLoader()

	for row, num := range []int{12, 44, 71, 72} {
		got, ok := p.PacketRow(num)
		assert.True(t, ok, "packet %d is in the table", num)
		assert.Equal(t, row, got)
	}
}

// A mark can outlive the filter that was in force when it was set.
func TestAPacketTheFilterRemovedHasNoRow(t *testing.T) {
	p := filteredLoader()

	for _, num := range []int{1, 11, 13, 70, 73, 1000} {
		_, ok := p.PacketRow(num)
		assert.False(t, ok, "packet %d was filtered out", num)
	}
}

func TestTheNextPacketIsTheNextOneInTheTable(t *testing.T) {
	p := filteredLoader()

	for _, c := range []struct{ from, want int }{
		{12, 44}, {44, 71}, {71, 72},
	} {
		got, ok := p.PacketAfter(c.from)
		assert.True(t, ok)
		assert.Equal(t, c.want, got)
	}
}

// Zero means before the first packet: that is how a search that has run off
// the end starts again from the top.
func TestPacketAfterZeroIsTheFirstPacket(t *testing.T) {
	p := filteredLoader()

	got, ok := p.PacketAfter(0)

	assert.True(t, ok)
	assert.Equal(t, 12, got)
}

func TestThereIsNoPacketAfterTheLastOne(t *testing.T) {
	p := filteredLoader()

	_, ok := p.PacketAfter(72)

	assert.False(t, ok, "a search at the end should be told so, and wrap")
}

// A search asks these while the load is still running, so an empty table is
// an answer rather than a panic.
func TestAnEmptyTableAnswersNo(t *testing.T) {
	p := &PsmlLoader{}

	assert.Equal(t, 0, p.NumLoadedPackets())

	_, ok := p.PacketRow(1)
	assert.False(t, ok)

	_, ok = p.PacketAfter(0)
	assert.False(t, ok)
}

func TestNumLoadedPacketsCountsRows(t *testing.T) {
	assert.Equal(t, 4, filteredLoader().NumLoadedPackets())
}

// The rows are int32. A number that cannot be one has to be answered no,
// rather than truncated onto some other packet.s row.
func TestANumberTooBigToBeAPacketIsNotTruncated(t *testing.T) {
	if math.MaxInt == math.MaxInt32 {
		t.Skip("int is 32 bits here, so no int can be too big to be a packet")
	}

	p := &PsmlLoader{packetNumbers: []int32{1, 2, 3}}

	// Built from a variable rather than written as a constant, so this file
	// still compiles for the 32-bit builds - a constant this size does not fit
	// an int there. Truncated to int32 it is 2, which is a real row here.
	var beyond int64 = math.MaxInt32
	big := int(beyond + 3)

	_, ok := p.PacketRow(big)
	assert.False(t, ok)

	_, ok = p.PacketAfter(big)
	assert.False(t, ok)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
