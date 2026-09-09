// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Real lines, from this machine's tshark.
const (
	convTCP  = `192.168.0.2:1550           <-> 192.168.0.1:23                   44 4283 bytes      48 3465 bytes      92 7748 bytes     0,000000000        39,5713`
	convIP   = `192.168.0.2          <-> 192.168.0.1            45056 4385 kB     49152 3548 kB     94208 7933 kB       0,000000000        39,5713`
	convIPv6 = `2001:db8::1:50000          <-> 2001:db8::2:80                   0 0 bytes         1 79 bytes        1 79 bytes      0,000000000         0,0000`
)

func TestAConversationWithPortsIsSplit(t *testing.T) {
	row, ok := parseConvLine(convTCP, true)

	require.True(t, ok)
	assert.Equal(t, "192.168.0.2", row.AddrA)
	assert.Equal(t, "1550", row.PortA)
	assert.Equal(t, "192.168.0.1", row.AddrB)
	assert.Equal(t, "23", row.PortB)
}

// This is the one that was silently dropped: tshark does not bracket IPv6, so
// the address has colons of its own and the port is after the last of them.
func TestAnIPv6ConversationIsRead(t *testing.T) {
	row, ok := parseConvLine(convIPv6, true)

	require.True(t, ok, "every IPv6 conversation used to be dropped here")
	assert.Equal(t, "2001:db8::1", row.AddrA)
	assert.Equal(t, "50000", row.PortA)
	assert.Equal(t, "2001:db8::2", row.AddrB)
	assert.Equal(t, "80", row.PortB)
}

// The IP and Ethernet tables have no ports, and the address must not be split
// at whatever colon it happens to contain.
func TestWithoutPortsTheAddressIsWhole(t *testing.T) {
	row, ok := parseConvLine(convIP, false)

	require.True(t, ok)
	assert.Equal(t, "192.168.0.2", row.AddrA)
	assert.Equal(t, "", row.PortA)
	assert.Equal(t, "192.168.0.1", row.AddrB)
}

// tshark prints "<-" first, then "->", then the total. Reading the first pair
// as the total is how a conversation's whole traffic came to be shown in the
// column headed B→A.
func TestTheTotalIsTheThirdPairNotTheFirst(t *testing.T) {
	row, ok := parseConvLine(convTCP, true)

	require.True(t, ok)
	assert.Equal(t, "92", row.Frames, "the total")
	assert.Equal(t, "48", row.FramesAB, "the -> pair")
	assert.Equal(t, "44", row.FramesBA, "the <- pair")
}

// The columns are declared total first, then each direction, and the cells
// have to arrive in that order or the table lies with a straight face.
func TestTheCellsFollowTheColumns(t *testing.T) {
	row, ok := parseConvLine(convTCP, true)
	require.True(t, ok)

	assert.Equal(t,
		[]string{"192.168.0.2", "1550", "192.168.0.1", "23",
			"92", "7748", "48", "3465", "44", "4283",
			"0,000000000", "39,5713"},
		row.cells(true))

	row, ok = parseConvLine(convIP, false)
	require.True(t, ok)

	assert.Equal(t,
		[]string{"192.168.0.2", "192.168.0.1",
			"94208", "7933 kB", "49152", "3548 kB", "45056", "4385 kB",
			"0,000000000", "39,5713"},
		row.cells(false))
}

// The unit is a space away from its number, which would make each byte count
// two fields. It is closed up to read the line and opened out again after,
// because the column is sorted on the number and the unit together.
func TestAByteCountKeepsItsUnit(t *testing.T) {
	row, ok := parseConvLine(convIP, false)

	require.True(t, ok)
	assert.Equal(t, "7933 kB", row.Bytes)
	assert.Equal(t, "3548 kB", row.BytesAB)
}

func TestPlainBytesLoseTheWordAndKeepTheNumber(t *testing.T) {
	row, ok := parseConvLine(convTCP, true)

	require.True(t, ok)
	assert.Equal(t, "7748", row.Bytes)
}

// Start and Duration come through as tshark wrote them, decimal comma and all,
// because that is what the comparator is written to read.
func TestTheTimesArePassedThroughUntouched(t *testing.T) {
	row, ok := parseConvLine(convTCP, true)

	require.True(t, ok)
	assert.Equal(t, "0,000000000", row.Start)
	assert.Equal(t, "39,5713", row.Durn)
}

// Everything else in the output is drawn with the same characters.
func TestWhatIsNotAConversationIsNotARow(t *testing.T) {
	for _, line := range []string{
		"",
		"================================================================================",
		"TCP Conversations",
		"Filter:<No Filter>",
		"                       | Packets | |  Bytes  | | Tx Packets |",
		"192.168.0.2 <-> 192.168.0.1",
	} {
		_, ok := parseConvLine(line, true)
		assert.False(t, ok, "%q", line)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
