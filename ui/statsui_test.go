// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui/pkg/stats"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestStatsReportCarriesTitleAndBody(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert}
	h.OnStatsData("Errors (6)\n=============\n")

	out := h.report()

	assert.True(t, strings.HasPrefix(out, "Expert Information"))
	assert.Contains(t, out, "Errors (6)")
	assert.NotContains(t, out, "Display filter:",
		"no filter was set, so none should be claimed")
}

func TestStatsReportNamesTheFilterInUse(t *testing.T) {
	h := &statsParseHandler{stat: stats.ProtoHierarchy, filter: "tcp.port in {23,80}"}
	h.OnStatsData("Protocol Hierarchy Statistics\n")

	out := h.report()

	assert.Contains(t, out, "Protocol Hierarchy")
	assert.Contains(t, out, "Display filter: tcp.port in {23,80}")
}

// tshark prints nothing at all - not even a header - when a statistic is
// narrowed to a filter that matches no packets. Without this, the user gets an
// empty dialog and no reason for it.
func TestStatsReportExplainsAnEmptyResult(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert, filter: "udp"}
	h.OnStatsData("")

	out := h.report()

	assert.Contains(t, out, "Nothing to report for this display filter")
}

func TestStatsReportExplainsAnEmptyResultWithNoFilter(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert}
	h.OnStatsData("   \n\t\n")

	out := h.report()

	assert.Contains(t, out, "Nothing to report for this capture")
}

// tshark on Windows ends its lines with CRLF; the dialog renders the stray CR
// as a glyph.
func TestStatsReportNormalisesWindowsLineEndings(t *testing.T) {
	h := &statsParseHandler{stat: stats.Expert}
	h.OnStatsData("Errors (6)\r\n=====\r\n")

	assert.NotContains(t, h.report(), "\r")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
