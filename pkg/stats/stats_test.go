// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestZArgWithoutFilter(t *testing.T) {
	assert.Equal(t, "expert", Expert.ZArg(""))
	assert.Equal(t, "io,phs", ProtoHierarchy.ZArg(""))
}

func TestZArgWithFilter(t *testing.T) {
	assert.Equal(t, "expert,udp", Expert.ZArg("udp"))
	assert.Equal(t, "io,phs,udp", ProtoHierarchy.ZArg("udp"))
}

// A display filter may itself contain commas - "tcp.port in {80,443}" is
// ordinary Wireshark syntax. tshark takes everything after the statistic's own
// arguments as the filter, so the commas need no escaping; this test exists so
// nobody "fixes" that by quoting or splitting it.
func TestZArgFilterMayContainCommas(t *testing.T) {
	f := "tcp.port in {23,80}"
	assert.Equal(t, "expert,tcp.port in {23,80}", Expert.ZArg(f))
	assert.Equal(t, "io,phs,tcp.port in {23,80}", ProtoHierarchy.ZArg(f))
}

// The UI hands over whatever is in the filter box, which is blank-but-not-empty
// often enough to matter.
func TestZArgTreatsBlankFilterAsAbsent(t *testing.T) {
	assert.Equal(t, "expert", Expert.ZArg("   "))
	assert.Equal(t, "io,phs", ProtoHierarchy.ZArg("\t\n"))
}

func TestAllStatsAreDistinctAndNamed(t *testing.T) {
	seen := make(map[string]bool)
	for _, s := range All {
		assert.NotEmpty(t, s.Name, "a statistic needs a title for its dialog")
		assert.NotEmpty(t, s.Command, "a statistic needs a minibuffer command")
		assert.False(t, seen[s.Command], "duplicate command: %s", s.Command)
		seen[s.Command] = true
	}
	assert.Equal(t, len(All), len(seen))
}

func TestLookupFindsByCommand(t *testing.T) {
	s, ok := Lookup("expert")
	assert.True(t, ok)
	assert.Equal(t, Expert.Command, s.Command)

	_, ok = Lookup("nosuchstat")
	assert.False(t, ok)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
