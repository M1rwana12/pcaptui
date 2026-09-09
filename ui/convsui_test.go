// Copyright 2026 m1rwana12. All rights reserved.
// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package ui contains user-interface functions and helpers for pcaptui.
package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// What used to be here was a test that called no production code at all: it
// re-declared ten locals and ran its own Fscanf over a conversation line. It
// would have passed with ui/convsui.go emptied, and its format string had
// stopped matching the real one - so the only test near the conversations
// parser protected nothing, while a bug that dropped every IPv6 conversation
// lived in that parser for as long as it existed.
//
// The parser is a function of its own now, and convsparse_test.go drives it.
// What is left here is the case that test was reaching for: a line with no
// ports at all, where a real capture's loopback conversation has none of the
// separators the port-splitting relies on.
func TestALoopbackConversationIsRead(t *testing.T) {
	line := `127.0.0.1:47416            <-> 127.0.0.1:9191                   0         0   43549   9951808   43549   9951808     4.160565000         9.4522`

	row, ok := parseConvLine(line, true)

	require.True(t, ok)
	assert.Equal(t, "127.0.0.1", row.AddrA)
	assert.Equal(t, "47416", row.PortA)
	assert.Equal(t, "127.0.0.1", row.AddrB)
	assert.Equal(t, "9191", row.PortB)
	assert.Equal(t, "4.160565000", row.Start, "a decimal point, on the machine that wrote this line")
	assert.Equal(t, "43549", row.Frames)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
