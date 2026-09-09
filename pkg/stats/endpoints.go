// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"strings"

	"github.com/m1rwana12/pcaptui"
)

//======================================================================

// EndpointRow is one address in `tshark -z endpoints,ip`.
//
// Tx and Rx are kept apart from the totals because that is the question the
// table is usually asked: not how much an address moved, but which direction
// it moved it in. A host that received four megabytes and sent two hundred
// bytes is doing something different from one that did the reverse.
type EndpointRow struct {
	Address string
	Packets int
	Bytes   int
	TxPkts  int
	TxBytes int
	RxPkts  int
	RxBytes int
}

// DisplayFilter narrows the packet list to the traffic this address is part
// of, in either direction.
//
// Which field depends on the address: Wireshark has no one name covering both,
// so `ip.addr` selects nothing for an IPv6 host and tshark rejects the filter
// outright. An IPv6 address is the one with a colon in it.
func (r EndpointRow) DisplayFilter() string {
	if strings.Contains(r.Address, ":") {
		return fmt.Sprintf("ipv6.addr == %s", r.Address)
	}
	return fmt.Sprintf("ip.addr == %s", r.Address)
}

//======================================================================

// ParseEndpoints reads the output of `tshark -z endpoints,ip`.
//
// The layout is a rule, a title, the filter in force, a header of pipe-wrapped
// column names, then one line per address:
//
//	192.168.0.2    92   7748 bytes    48   3465 bytes    44   4283 bytes
//
// The word "bytes" follows three of the six numbers, so the fields are read by
// dropping it and taking what is left: an address and six counts. Reading by
// column position would not survive an IPv6 address, which is wider than the
// column tshark lays out for it.
func ParseEndpoints(out string) []EndpointRow {
	var res []EndpointRow

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		row, ok := endpointRow(line)
		if ok {
			res = append(res, row)
		}
	}

	return res
}

func endpointRow(line string) (EndpointRow, bool) {
	if strings.HasPrefix(line, "=") || strings.HasPrefix(line, "Filter:") {
		return EndpointRow{}, false
	}

	fields := make([]string, 0, 7)
	for _, f := range strings.Fields(line) {
		if f == "bytes" {
			continue
		}
		fields = append(fields, f)
	}
	if len(fields) != 7 {
		return EndpointRow{}, false
	}

	// The first field has to look like an address. tshark can be asked for
	// several statistics in one pass, and then the expert and hierarchy tables
	// arrive in the same stream as this one; a count of seven fields alone
	// would eventually let one of their lines through and invent an endpoint.
	if !strings.ContainsAny(fields[0], ".:") {
		return EndpointRow{}, false
	}

	// The header line has the same field count once the pipes are counted as
	// words, so the numbers are what tells a row from a heading.
	var n [6]int
	for i := 0; i < 6; i++ {
		v, ok := parseCount(fields[i+1])
		if !ok {
			return EndpointRow{}, false
		}
		n[i] = v
	}

	return EndpointRow{
		Address: fields[0],
		Packets: n[0],
		Bytes:   n[1],
		TxPkts:  n[2],
		TxBytes: n[3],
		RxPkts:  n[4],
		RxBytes: n[5],
	}, true
}

// parseCount reads a count that tshark may have grouped for readability.
//
// Measured, not assumed: the macOS runner printed "7,748" where this machine's
// tshark printed "7748" - and only the endpoints table did it, since the
// protocol hierarchy's byte counts came through ungrouped on the same run. The
// separator therefore depends on the statistic as well as the environment, and
// a parser that only understands bare digits works for the developer and
// silently reads nothing for somebody else.
//
// The rule itself lives in the root package now, because the conversations
// table needed the same one and had the opposite bug: it stripped every comma,
// so a machine that writes "1,5 kB" sorted that as fifteen kilobytes.
func parseCount(s string) (int, bool) {
	return pcaptui.ParseCount(s)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
