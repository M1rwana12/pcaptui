// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcaptui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func TestABareNumberIsRead(t *testing.T) {
	n, ok := ParseCount("7748")

	require.True(t, ok)
	assert.Equal(t, 7748, n)
}

// The macOS runner printed this where this machine printed 7748.
func TestAGroupedNumberIsRead(t *testing.T) {
	for _, s := range []string{"7,748", "7'748", "1,234,567"} {
		n, ok := ParseCount(s)
		require.True(t, ok, "%q", s)
		assert.Greater(t, n, 999)
	}
}

// "1,5" is one and a half where this machine's tshark writes numbers. Reading
// it as fifteen is the wrong number, which is worse than no number.
func TestADecimalCommaIsNotAGroup(t *testing.T) {
	_, ok := ParseCount("1,5")

	assert.False(t, ok, "one and a half is not a count")
}

func TestWhatIsNotANumberIsNotRead(t *testing.T) {
	for _, s := range []string{"", "-", "kB", "1.5", "abc"} {
		_, ok := ParseCount(s)
		assert.False(t, ok, "%q", s)
	}
}

//======================================================================

// What tshark writes in the Start and Duration columns on this machine.
func TestADecimalCommaIsADecimalPoint(t *testing.T) {
	for _, c := range []struct {
		in   string
		want float64
	}{
		{"0,000000000", 0},
		{"39,5713", 39.5713},
		{"1,5", 1.5},
	} {
		got, ok := ParseDecimal(c.in)
		require.True(t, ok, "%q", c.in)
		assert.InDelta(t, c.want, got, 0.00001, c.in)
	}
}

func TestADecimalPointIsStillADecimalPoint(t *testing.T) {
	got, ok := ParseDecimal("39.5713")

	require.True(t, ok)
	assert.InDelta(t, 39.5713, got, 0.00001)
}

// A number carrying both separators has its groups first, so the later one is
// the decimal.
func TestBothSeparatorsMeanTheLaterOneIsTheDecimal(t *testing.T) {
	got, ok := ParseDecimal("1.234,56")
	require.True(t, ok)
	assert.InDelta(t, 1234.56, got, 0.001)

	got, ok = ParseDecimal("1,234.56")
	require.True(t, ok)
	assert.InDelta(t, 1234.56, got, 0.001)
}

// The one genuinely ambiguous shape, decided the same way ParseCount decides
// it and for the same reason.
func TestALoneGroupOfThreeIsAGroup(t *testing.T) {
	got, ok := ParseDecimal("1,234")

	require.True(t, ok)
	assert.InDelta(t, 1234, got, 0.001)
}

func TestWhatIsNotADecimalIsNotRead(t *testing.T) {
	for _, s := range []string{"", "-", "kB", "1,2,3"} {
		_, ok := ParseDecimal(s)
		assert.False(t, ok, "%q", s)
	}
}

//======================================================================

// Both columns were unsortable on this machine: gowid's FloatCompare is
// ParseFloat, which rejects "0,000000000" outright, so clicking the header did
// nothing and said nothing.
func TestStartTimesSortByTime(t *testing.T) {
	c := ConvFloatCompare{}

	assert.True(t, c.Less("0,000000000", "1,500000000"))
	assert.False(t, c.Less("1,500000000", "0,000000000"))
	assert.True(t, c.Less("2,0000", "39,5713"))
}

// "1,5 kB" is one and a half kilobytes here. Stripping the comma sorted it as
// fifteen.
func TestBytesSortByHowManyBytesTheyAre(t *testing.T) {
	c := ConvPktsCompare{}

	assert.True(t, c.Less("900 bytes", "2 kB"))
	assert.True(t, c.Less("2 kB", "1 MB"))
	assert.True(t, c.Less("999 bytes", "7,748 bytes"), "a grouped count is still a count")
	assert.False(t, c.Less("2 kB", "1,5 kB"),
		"one and a half kilobytes is not fifteen")
}

func TestAnUnreadableCellClaimsNoOrder(t *testing.T) {
	c := ConvPktsCompare{}

	assert.False(t, c.Less("nonsense", "2 kB"))
	assert.False(t, c.Less("2 kB", "nonsense"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
