// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestTheCountIsJustTheCountWhenNothingElseIsTrue(t *testing.T) {
	assert.Equal(t, "7 packets", packetCountText(7, false, false))
}

func TestOnePacketIsNotOnePackets(t *testing.T) {
	assert.Equal(t, "1 packet", packetCountText(1, false, false))
}

// The whole point of the counter: a short list has to say whether that is the
// capture or the filter.
func TestAFilteredCountSaysItIsFiltered(t *testing.T) {
	assert.Equal(t, "3 packets · filtered", packetCountText(3, true, false))
}

func TestAnEmptyResultDistinguishesNoCaptureFromNoMatch(t *testing.T) {
	assert.Equal(t, "no packets", packetCountText(0, false, false))
	assert.Equal(t, "no packets match", packetCountText(0, true, false))
}

// A number that is still growing must not read as a final answer, or someone
// reports "only 40,000 of my packets loaded" about a capture still being read.
func TestAPartialCountSaysItIsStillLoading(t *testing.T) {
	assert.Equal(t, "40,000 packets · loading", packetCountText(40000, false, true))
	assert.Equal(t, "loading", packetCountText(0, false, true))
}

func TestEverythingAtOnce(t *testing.T) {
	assert.Equal(t, "3 packets · filtered · loading", packetCountText(3, true, true))
}

// "loading" already says the count is not final; "no packets match" would be a
// wrong answer while the filter is still being applied.
func TestLoadingWinsOverAnEmptyFilterResult(t *testing.T) {
	assert.Equal(t, "loading · filtered", packetCountText(0, true, true))
}

//======================================================================

func TestDigitsAreGroupedSoAMillionIsReadable(t *testing.T) {
	assert.Equal(t, "0", groupDigits(0))
	assert.Equal(t, "7", groupDigits(7))
	assert.Equal(t, "999", groupDigits(999))
	assert.Equal(t, "1,000", groupDigits(1000))
	assert.Equal(t, "40,000", groupDigits(40000))
	assert.Equal(t, "1,048,576", groupDigits(1048576))
	assert.Equal(t, "-1,234", groupDigits(-1234))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
