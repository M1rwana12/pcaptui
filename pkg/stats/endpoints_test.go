// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Verbatim from `tshark -q -z endpoints,ip -r scripts/pcaps/telnet-cooked.pcap`,
// Wireshark 4.6.8.
const endpointsOutput = `================================================================================
IPv4 Endpoints
Filter:<No Filter>
                       | Packets | |  Bytes  | | Tx Packets | | Tx Bytes | | Rx Packets | | Rx Bytes |
192.168.0.2                    92   7748 bytes         48      3465 bytes          44      4283 bytes
192.168.0.1                    92   7748 bytes         44      4283 bytes          48      3465 bytes
================================================================================
`

func TestEveryAddressIsARow(t *testing.T) {
	rows := ParseEndpoints(endpointsOutput)

	require.Len(t, rows, 2)
	assert.Equal(t, "192.168.0.2", rows[0].Address)
	assert.Equal(t, "192.168.0.1", rows[1].Address)
}

// The direction is the question this table is usually asked: not how much an
// address moved, but which way.
func TestTheCountsAreReadInBothDirections(t *testing.T) {
	rows := ParseEndpoints(endpointsOutput)

	assert.Equal(t, 92, rows[0].Packets)
	assert.Equal(t, 7748, rows[0].Bytes)
	assert.Equal(t, 48, rows[0].TxPkts)
	assert.Equal(t, 3465, rows[0].TxBytes)
	assert.Equal(t, 44, rows[0].RxPkts)
	assert.Equal(t, 4283, rows[0].RxBytes)
}

// The rules, the title, the filter line and the pipe-wrapped column header are
// not addresses. The header survives the "bytes" filter with the right number
// of words, so it is the numbers that tell a row from a heading.
func TestTheFurnitureIsNotARow(t *testing.T) {
	for _, line := range []string{
		"================================================================================",
		"IPv4 Endpoints",
		"Filter:<No Filter>",
		"Filter:telnet",
		"                       | Packets | |  Bytes  | | Tx Packets | | Tx Bytes | | Rx Packets | | Rx Bytes |",
		"",
	} {
		assert.Empty(t, ParseEndpoints(line), "should not have read a row from %q", line)
	}
}

// An IPv6 address is wider than the column tshark lays out for it, so reading
// by column position would cut it in half.
func TestAWideAddressIsNotTruncated(t *testing.T) {
	out := "2001:0db8:85a3:0000:0000:8a2e:0370:7334      10   1000 bytes      4    400 bytes      6    600 bytes\n"

	rows := ParseEndpoints(out)

	require.Len(t, rows, 1)
	assert.Equal(t, "2001:0db8:85a3:0000:0000:8a2e:0370:7334", rows[0].Address)
	assert.Equal(t, 10, rows[0].Packets)
}

func TestNoEndpointOutputIsNoRows(t *testing.T) {
	assert.Empty(t, ParseEndpoints(""))
}

//======================================================================

// The row is the traffic that address is part of, in either direction - which
// is what makes "who is on the wire" one keypress from "show me only them".
func TestARowFiltersOnItsOwnAddress(t *testing.T) {
	r := EndpointRow{Address: "192.168.0.2"}

	assert.Equal(t, "ip.addr == 192.168.0.2", r.DisplayFilter())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
