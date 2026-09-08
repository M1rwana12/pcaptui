// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"strconv"
	"strings"
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
func (r EndpointRow) DisplayFilter() string {
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

	// The header line has the same field count once the pipes are counted as
	// words, so the numbers are what tells a row from a heading.
	var n [6]int
	for i := 0; i < 6; i++ {
		v, err := strconv.Atoi(fields[i+1])
		if err != nil {
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

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
