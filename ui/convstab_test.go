// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// The tabs run Ethernet, IPv4, IPv6, TCP, UDP.
const (
	tabIP   = 1
	tabIPv6 = 2
	tabTCP  = 3
	tabUDP  = 4
)

func TestTheBusiestTabWins(t *testing.T) {
	assert.True(t, preferTab(12, tabIP, 3, tabTCP))
	assert.False(t, preferTab(3, tabTCP, 12, tabIP))
}

// A capture with one IPv4 conversation and one TCP conversation is better
// answered by the tab that names ports.
func TestATieGoesToTheMoreSpecificTab(t *testing.T) {
	assert.True(t, preferTab(1, tabTCP, 1, tabIP))
	assert.False(t, preferTab(1, tabIP, 1, tabTCP))
}

// An empty tab is never worth opening on, whatever is showing.
func TestAnEmptyTabIsNeverChosen(t *testing.T) {
	assert.False(t, preferTab(0, tabUDP, 0, tabIP))
	assert.False(t, preferTab(0, tabUDP, 5, tabTCP))
}

// The first tab with anything in it takes the screen from the Ethernet tab the
// view starts on, which is what nothing having been chosen looks like.
func TestTheFirstTabWithRowsTakesTheScreen(t *testing.T) {
	assert.True(t, preferTab(1, tabIPv6, 0, 0))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
