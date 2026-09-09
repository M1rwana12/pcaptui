// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/stats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

const expertData = `
Notes (1)
=============
   Frequency      Group           Protocol  Summary
           1   Sequence                TCP  This frame is a (suspected) retransmission
`

func TestStatsViewCarriesTitleAndRows(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert}
	h.OnStatsData(expertData)

	v := h.view()

	assert.True(t, strings.HasPrefix(v.Heading, "Expert Information"))
	assert.NotContains(t, v.Heading, "display filter",
		"no filter was set, so none should be claimed")
	require.Len(t, v.Rows, 1)
	assert.Contains(t, v.Rows[0].Text, "retransmission")
}

func TestStatsViewNamesTheFilterInUse(t *testing.T) {
	h := &statsParseHandler{stat: stats.ProtoHierarchy, filter: "tcp.port in {23,80}"}
	h.OnStatsData("frame  frames:7 bytes:1027\n")

	v := h.view()

	assert.Contains(t, v.Heading, "Protocol Hierarchy")
	assert.Contains(t, v.Heading, "tcp.port in {23,80}")
}

// The row is not just text - it is the packets it is about, which is the whole
// difference between this and the dialog it replaced.
func TestAnExpertRowCarriesItsFilter(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert}
	h.OnStatsData(expertData)

	v := h.view()

	require.Len(t, v.Rows, 1)
	assert.Equal(t,
		`_ws.expert.message == "This frame is a (suspected) retransmission"`,
		v.Rows[0].Filter)
}

func TestAHierarchyRowCarriesItsProtocol(t *testing.T) {
	h := &statsParseHandler{stat: stats.ProtoHierarchy}
	h.OnStatsData("frame            frames:7 bytes:1027\n  eth            frames:7 bytes:1027\n")

	v := h.view()

	require.Len(t, v.Rows, 2)
	assert.Equal(t, "eth", v.Rows[1].Filter)
}

// tshark prints nothing at all - not even a header - when a statistic is
// narrowed to a filter that matches no packets. Without this the user gets an
// empty dialog and no reason for it.
func TestAnEmptyResultIsAnEmptyView(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert, filter: "udp"}
	h.OnStatsData("")

	assert.True(t, h.view().empty())
}

func TestWhitespaceOnlyOutputIsAlsoEmpty(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert}
	h.OnStatsData("   \n\t\n")

	assert.True(t, h.view().empty())
}

// tshark on Windows ends its lines with CRLF, and a stray CR renders as a
// glyph.
func TestStatsDataNormalisesWindowsLineEndings(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert}
	h.OnStatsData(strings.ReplaceAll(expertData, "\n", "\r\n"))

	v := h.view()

	require.Len(t, v.Rows, 1)
	assert.NotContains(t, v.Rows[0].Text, "\r")
}

// A statistic with no layout of its own is still worth showing; it just cannot
// offer a filter for any of its rows.
func TestAnUnknownStatisticIsShownAsPlainLines(t *testing.T) {
	h := &statsParseHandler{stat: stats.Stat{Name: "Something Else", Command: "else"}}
	h.OnStatsData("one\ntwo\n")

	v := h.view()

	require.Len(t, v.Rows, 2)
	assert.Equal(t, "one", v.Rows[0].Text)
	assert.False(t, v.Rows[0].actionable())
}

//======================================================================

// A statistic is one pass over the whole capture, and the overview starts one
// by itself when a file loads: 11.6 s on a 44 MB capture here, and the
// README's own pitch is a two-gigabyte one. The shared please-wait dialog has
// no buttons, and Escape closed it without stopping anything - tshark went on
// reading and the result opened over whatever the user had moved on to.
func TestClosingTheWaitDialogStopsTheLoad(t *testing.T) {
	stopped := 0
	w := newStatsWait("Overview", func() { stopped++ })

	// What the Cancel button and Escape both arrive at.
	w.onClose()

	assert.Equal(t, 1, stopped)
}

// The load that finished takes its own dialog down, and that must not be read
// as the user asking for it to stop.
func TestFinishingDoesNotLookLikeCancelling(t *testing.T) {
	stopped := 0
	w := newStatsWait("Overview", func() { stopped++ })

	w.closing = true
	w.onClose()

	assert.Zero(t, stopped)
}

// close() on a dialog that was never opened does nothing at all, which is what
// makes it safe to call from every ending path.
func TestClosingWhatWasNeverOpenedIsHarmless(t *testing.T) {
	stopped := 0
	w := newStatsWait("Overview", func() { stopped++ })

	w.close(nil)

	assert.Zero(t, stopped)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
