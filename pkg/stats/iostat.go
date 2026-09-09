// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"math"
	"strings"
)

//======================================================================

// IOStatRow is one time bucket of `tshark -z io,stat,<interval>`.
//
// Label is the bucket as tshark wrote it - "0 <> 5", and "35 <> Dur" for the
// last one, which runs to the end of the capture rather than to a round
// number.
type IOStatRow struct {
	Label  string
	Frames int
	Bytes  int
}

//======================================================================

// IOStatArg is the -z argument for a capture of the given length in seconds.
//
// The interval is chosen so that the answer is a shape rather than a list: too
// long and a burst is averaged away, too short and an hour of traffic arrives
// as thousands of rows. About twenty buckets reads well on one line, rounded
// up to a duration a person would have chosen - 1, 2, 5, 10 seconds, and so on
// - because "every 7 seconds" is harder to hold in the head than "every 10".
//
// Seconds, because that is tshark's unit here. A capture shorter than the
// smallest interval gets one bucket, which is the truthful answer for it.
func IOStatArg(seconds float64, displayFilter string) string {
	return withFilter(fmt.Sprintf("io,stat,%d", IOStatInterval(seconds)), displayFilter)
}

// IOStatInterval is the bucket size in seconds.
func IOStatInterval(seconds float64) int {
	const buckets = 20

	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return 1
	}

	want := seconds / buckets

	for _, nice := range []int{1, 2, 5, 10, 15, 30, 60, 120, 300, 600, 900, 1800, 3600} {
		if float64(nice) >= want {
			return nice
		}
	}

	// Longer than twenty hours: whole hours, rounded up, so the number stays
	// one a person can read.
	return int(math.Ceil(want/3600)) * 3600
}

//======================================================================

// ParseIOStat reads the buckets out of `tshark -z io,stat,<interval>`.
//
// The table is drawn with pipes:
//
//	| Interval | Frames | Bytes |
//	|---------------------------|
//	|  0 <> 5  |     44 |  3291 |
//	| 35 <> Dur|     12 |   876 |
//
// A data row is recognised by holding tshark's "<>" between its first pair of
// pipes, which the header and the rules do not.
func ParseIOStat(out string) []IOStatRow {
	var res []IOStatRow

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		row, ok := ioStatRow(line)
		if ok {
			res = append(res, row)
		}
	}

	return res
}

func ioStatRow(line string) (IOStatRow, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") {
		return IOStatRow{}, false
	}

	cells := strings.Split(strings.Trim(line, "|"), "|")
	if len(cells) < 3 {
		return IOStatRow{}, false
	}

	label := strings.TrimSpace(cells[0])
	if !strings.Contains(label, "<>") {
		return IOStatRow{}, false
	}

	frames, okf := parseCount(strings.TrimSpace(cells[1]))
	bytes, okb := parseCount(strings.TrimSpace(cells[2]))
	if !okf || !okb {
		return IOStatRow{}, false
	}

	return IOStatRow{
		Label:  strings.Join(strings.Fields(label), " "),
		Frames: frames,
		Bytes:  bytes,
	}, true
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
