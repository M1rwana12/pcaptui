// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// A display filter can arrive two ways, and the whole point of resolving them
// in one place is that one value comes out to be checked. The bug this
// replaces was that the flag was checked and the positional argument was not:
// a typo in the positional form started the program with an empty filter box
// and empty panes, and exited zero.
func TestAFilterArrivesEitherWayAndComesOutOnce(t *testing.T) {
	got, err := displayFilterFor("tcp.port == 80", "", true)
	require.NoError(t, err)
	assert.Equal(t, "tcp.port == 80", got, "the -Y flag")

	got, err = displayFilterFor("", "tcp.port == 80", true)
	require.NoError(t, err)
	assert.Equal(t, "tcp.port == 80", got, "the positional argument")

	got, err = displayFilterFor("", "", true)
	require.NoError(t, err)
	assert.Empty(t, got)
}

// Preferring one of two given filters silently would apply something the user
// did not ask for and drop what they did.
func TestTwoFiltersAreRefusedAndBothNamed(t *testing.T) {
	_, err := displayFilterFor("tcp.port == 80", "udp.port == 53", true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "tcp.port == 80")
	assert.Contains(t, err.Error(), "udp.port == 53")
	assert.Contains(t, err.Error(), "one only")
}

// On an interface the positional argument is a capture filter - a different
// language, taken by the caller before this is reached. Treating it as a
// display filter here would hand tshark a BPF expression to parse as one.
func TestOnAnInterfaceThePositionalArgumentIsNotADisplayFilter(t *testing.T) {
	got, err := displayFilterFor("", "port 443", false)

	require.NoError(t, err)
	assert.Empty(t, got, "the capture filter must not become a display filter")

	// The flag still applies while capturing: -Y narrows what is shown, the
	// capture filter narrows what is recorded.
	got, err = displayFilterFor("http", "port 443", false)
	require.NoError(t, err)
	assert.Equal(t, "http", got)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
