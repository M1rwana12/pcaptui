// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/capinfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// The Overview opens by itself when a capture finishes loading, and it is one
// tshark pass over every packet in the file. Measured with
// scripts/bench-overview.sh: 85 MB is 19 s on Linux and 47 s on macOS, and the
// rate gets worse as the file grows. So there is a size above which starting
// that pass unasked is spending somebody's minute for them.
func TestABigCapturesStatisticsAreOfferedNotStarted(t *testing.T) {
	assert.True(t, overviewAutoStart(0), "an empty file is not a reason to refuse")
	assert.True(t, overviewAutoStart(2*1024*1024))
	assert.True(t, overviewAutoStart(overviewAutoLimit), "the limit itself is still automatic")

	assert.False(t, overviewAutoStart(overviewAutoLimit+1))
	assert.False(t, overviewAutoStart(2*1024*1024*1024),
		"the size the README says this program handles")
}

// What is shown instead has to be worth opening on its own, or the dialog is
// just an interruption. capinfos answers the question the three statistics do
// not - what is this file - and it stays cheap at every size: 0.52 s on the
// same 85 MB that costs 47 s to summarise.
func TestTheOfferShowsWhatTheFileSaysAboutItself(t *testing.T) {
	info := capinfo.Info{
		Earliest: "2026-01-01 12:00:00",
		Duration: "39.571274 seconds",
		Packets:  "753,664",
		Size:     "85 MB",
		SnapLen:  "262144 (not set)",
	}

	v := overviewFactsView(info, "")

	require.False(t, v.empty(), "an empty view is reported as nothing to show")
	text := rowsText(v)

	assert.Contains(t, text, "2026-01-01 12:00:00")
	assert.Contains(t, text, "753,664")
	assert.Contains(t, text, "85 MB")

	// The offer names the key and the size, because "press o" without knowing
	// what it will cost is the thing being avoided.
	assert.Contains(t, text, "press o")
	assert.Contains(t, text, "one pass over all 85 MB")

	// A snapshot length that was never set is not a fact about this capture.
	assert.NotContains(t, text, "262144")
}

// Nothing about the offer is actionable: every row of a statistics dialog that
// carries a filter can be pressed, and none of these has one to press.
func TestNoRowOfTheOfferPretendsToBeAFilter(t *testing.T) {
	v := overviewFactsView(capinfo.Info{Packets: "10", Size: "1 kB"}, "")

	for _, row := range v.Rows {
		assert.False(t, row.actionable(), "%q offers a filter", row.Text)
	}
}

func rowsText(v statsView) string {
	var b strings.Builder
	for _, row := range v.Rows {
		b.WriteString(row.Text)
		b.WriteString("\n")
	}
	return b.String()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
