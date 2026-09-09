// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"strings"
)

//======================================================================

// convRow is one line of a `tshark -z conv,...` table.
//
// The names say which direction each pair belongs to, which the ones in the
// widget code did not: tshark prints the "<-" pair first, and reading that as
// the total is how the table came to show a conversation's whole traffic in
// the column headed B→A.
type convRow struct {
	AddrA, PortA string
	AddrB, PortB string

	Frames, Bytes     string // the totals
	FramesAB, BytesAB string
	FramesBA, BytesBA string

	Start, Durn string
}

// cells is the row in the order the table's columns are declared: the total
// first, then each direction.
func (r convRow) cells(ports bool) []string {
	if ports {
		return []string{
			r.AddrA, r.PortA, r.AddrB, r.PortB,
			r.Frames, r.Bytes,
			r.FramesAB, r.BytesAB,
			r.FramesBA, r.BytesBA,
			r.Start, r.Durn,
		}
	}

	return []string{
		r.AddrA, r.AddrB,
		r.Frames, r.Bytes,
		r.FramesAB, r.BytesAB,
		r.FramesBA, r.BytesBA,
		r.Start, r.Durn,
	}
}

//======================================================================

// parseConvLine reads one conversation:
//
//	192.168.0.2:1550  <-> 192.168.0.1:23   44 4283 bytes  48 3465 bytes  92 7748 bytes  0,000000000  39,5713
//
// tshark writes the columns as "<-", then "->", then the total - so the first
// pair is B→A and the second A→B, whatever the variable names in older code
// suggested.
//
// The unit is separated from its number by a space, which would make each byte
// count two fields, so " bytes" is dropped and " kB" and " MB" are closed up
// before the line is read and opened out again afterwards. It is the same
// trick the widget code used; it is here so that it can be tested.
//
// ports says whether this table has them - the TCP and UDP tables do. tshark
// does not bracket IPv6, so the port is taken from the last colon rather than
// the only one; see splitHostPort, which is where every IPv6 conversation used
// to be dropped.
func parseConvLine(line string, ports bool) (convRow, bool) {
	line = strings.Replace(line, " bytes", "", -1)
	line = strings.Replace(line, "bytes", "", -1)
	line = strings.Replace(line, " kB", "kB", -1)
	line = strings.Replace(line, " MB", "MB", -1)

	var (
		addra, addrb      string
		framesBA, bytesBA string
		framesAB, bytesAB string
		frames, bytes     string
		start, durn       string
	)

	n, err := fmt.Fscanf(strings.NewReader(line), "%s <-> %s %s %s %s %s %s %s %s %s",
		&addra,
		&addrb,
		&framesBA, // the "<-" pair
		&bytesBA,
		&framesAB, // the "->" pair
		&bytesAB,
		&frames, // the total
		&bytes,
		&start,
		&durn,
	)
	if err != nil || n != 10 {
		return convRow{}, false
	}

	row := convRow{
		AddrA:    addra,
		AddrB:    addrb,
		Frames:   frames,
		Bytes:    openUnit(bytes),
		FramesAB: framesAB,
		BytesAB:  openUnit(bytesAB),
		FramesBA: framesBA,
		BytesBA:  openUnit(bytesBA),
		Start:    start,
		Durn:     durn,
	}

	if ports {
		row.AddrA, row.PortA = splitHostPort(addra)
		row.AddrB, row.PortB = splitHostPort(addrb)
	}

	return row, true
}

// openUnit puts back the space that was taken out to keep a byte count in one
// field.
func openUnit(s string) string {
	s = strings.Replace(s, "kB", " kB", -1)
	return strings.Replace(s, "MB", " MB", -1)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
