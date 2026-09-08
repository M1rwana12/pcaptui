// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"strconv"
	"strings"
)

//======================================================================

// HierarchyRow is one protocol in `tshark -z io,phs`.
//
// Depth is how deep the protocol sits under the ones carrying it - frame at 0,
// eth at 1, ip at 2 - taken from the indentation, which is the only place
// tshark states it.
type HierarchyRow struct {
	Depth    int
	Protocol string
	Frames   int
	Bytes    int
}

// DisplayFilter narrows the packet list to the packets containing this
// protocol. The name tshark prints in this table is the protocol's display
// filter name, which is why the answer to "what is in this file" can be turned
// straight into "show me that".
func (r HierarchyRow) DisplayFilter() string {
	return r.Protocol
}

//======================================================================

// ParseHierarchy reads the output of `tshark -z io,phs`.
//
// The format is a rule, a title, the filter in force, a blank line, then one
// indented line per protocol:
//
//	frame                                    frames:7 bytes:1027
//	  eth                                    frames:7 bytes:1027
//
// Everything before the first `frames:` line is a preamble, and the rule
// repeats at the end, so lines are recognised by containing the counts rather
// than by their position.
func ParseHierarchy(out string) []HierarchyRow {
	var res []HierarchyRow

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		row, ok := hierarchyRow(line)
		if ok {
			res = append(res, row)
		}
	}

	return res
}

func hierarchyRow(line string) (HierarchyRow, bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return HierarchyRow{}, false
	}

	frames, okf := countField(fields, "frames:")
	bytes, okb := countField(fields, "bytes:")
	if !okf || !okb {
		return HierarchyRow{}, false
	}

	name := fields[0]
	if name == "" || strings.HasPrefix(name, "frames:") {
		return HierarchyRow{}, false
	}

	// Two spaces per level, and the protocol name is the first thing on the
	// line, so the indentation is what is in front of it.
	indent := len(line) - len(strings.TrimLeft(line, " "))

	return HierarchyRow{
		Depth:    indent / 2,
		Protocol: name,
		Frames:   frames,
		Bytes:    bytes,
	}, true
}

func countField(fields []string, prefix string) (int, bool) {
	for _, f := range fields {
		if !strings.HasPrefix(f, prefix) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(f, prefix))
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
