// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Real output for the telnet capture, which has two gaps in it.
const ioStatOut = `=============================
| IO Statistics             |
|                           |
| Duration: 39.571274 secs  |
| Interval:  5 secs         |
|                           |
| Col  1: Frames and bytes  |
|---------------------------|
|          |1               |
| Interval | Frames | Bytes |
|---------------------------|
|  0 <> 5  |     44 |  3291 |
|  5 <> 10 |      6 |   946 |
| 10 <> 15 |      0 |     0 |
| 15 <> 20 |      2 |   158 |
| 20 <> 25 |     14 |  1321 |
| 25 <> 30 |     14 |  1156 |
| 30 <> 35 |      0 |     0 |
| 35 <> Dur|     12 |   876 |
=============================
`

func TestTheBucketsAreRead(t *testing.T) {
	rows := ParseIOStat(ioStatOut)

	require.Len(t, rows, 8)
	assert.Equal(t, "0 <> 5", rows[0].Label)
	assert.Equal(t, 44, rows[0].Frames)
	assert.Equal(t, 3291, rows[0].Bytes)
}

// A gap is the point of the whole thing, so an empty bucket is a bucket.
func TestAnEmptyBucketIsKept(t *testing.T) {
	rows := ParseIOStat(ioStatOut)

	require.Len(t, rows, 8)
	assert.Zero(t, rows[2].Frames)
	assert.Equal(t, "10 <> 15", rows[2].Label)
}

// The last bucket runs to the end of the capture rather than to a round
// number, and tshark writes it without a space before the pipe.
func TestTheLastBucketRunsToTheEnd(t *testing.T) {
	rows := ParseIOStat(ioStatOut)

	require.NotEmpty(t, rows)
	assert.Equal(t, "35 <> Dur", rows[len(rows)-1].Label)
	assert.Equal(t, 12, rows[len(rows)-1].Frames)
}

// The header, the rules and the preamble are drawn with the same pipes.
func TestOnlyTheBucketsAreRows(t *testing.T) {
	for _, r := range ParseIOStat(ioStatOut) {
		assert.Contains(t, r.Label, "<>")
		assert.NotContains(t, r.Label, "Interval")
	}
}

func TestOutputThatIsNotATableYieldsNothing(t *testing.T) {
	assert.Empty(t, ParseIOStat(""))
	assert.Empty(t, ParseIOStat("tshark: no such tap\n"))
}

//======================================================================

// About twenty buckets, rounded up to an interval a person would have chosen.
func TestTheIntervalIsARoundNumber(t *testing.T) {
	// Twenty buckets of the capture's length, rounded up to the next interval
	// a person would have chosen: an hour wants 180 s and gets 300.
	assert.Equal(t, 1, IOStatInterval(5), "shorter than the smallest interval")
	assert.Equal(t, 2, IOStatInterval(39.571274))
	assert.Equal(t, 5, IOStatInterval(100))
	assert.Equal(t, 30, IOStatInterval(600))
	assert.Equal(t, 300, IOStatInterval(3600))
}

// A capture with no duration at all still has to produce an argument tshark
// accepts rather than "io,stat,0", which means one bucket for everything.
func TestANonsenseDurationStillGivesAnInterval(t *testing.T) {
	for _, s := range []float64{0, -1} {
		assert.Equal(t, 1, IOStatInterval(s))
	}
}

func TestTheArgumentCarriesTheFilter(t *testing.T) {
	assert.Equal(t, "io,stat,5", IOStatArg(100, ""))
	assert.Equal(t, "io,stat,5,tcp.port == 80", IOStatArg(100, "tcp.port == 80"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
