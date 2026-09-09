// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Real output for the fixture, plus two rows in the shape tshark writes for
// the other protocols it can read.
const credsOut = `===================================================================
Packet     Protocol         Username         Info            
------     --------         --------         --------
1          HTTP basic auth  user                             
14         FTP              anonymous        Username in packet: 12
203        POP              mary jane        
===================================================================
`

func TestTheLoginsAreRead(t *testing.T) {
	rows := ParseCredentials(credsOut)

	require.Len(t, rows, 3)
	assert.Equal(t, 1, rows[0].Packet)
	assert.Equal(t, "HTTP basic auth", rows[0].Protocol)
	assert.Equal(t, "user", rows[0].Username)
	assert.Equal(t, "", rows[0].Info)
}

// The protocol is several words, and so can a username be. Splitting on
// whitespace would put half of one column into the next.
func TestSeveralWordsStayInTheirColumn(t *testing.T) {
	rows := ParseCredentials(credsOut)

	require.Len(t, rows, 3)
	assert.Equal(t, "mary jane", rows[2].Username)
	assert.Equal(t, "anonymous", rows[1].Username)
	assert.Equal(t, "Username in packet: 12", rows[1].Info)
}

// This is the one statistic whose rows name a packet, so its filter is exact
// rather than a search for the row's own text.
func TestARowFiltersOnItsPacket(t *testing.T) {
	rows := ParseCredentials(credsOut)

	require.NotEmpty(t, rows)
	assert.Equal(t, "frame.number == 1", rows[0].DisplayFilter())
	assert.Equal(t, "frame.number == 203", rows[2].DisplayFilter())
}

// The rule under the header is drawn in the same columns as the rows.
func TestTheRuleIsNotALogin(t *testing.T) {
	for _, r := range ParseCredentials(credsOut) {
		assert.NotContains(t, r.Protocol, "---")
		assert.Positive(t, r.Packet)
	}
}

// A capture with nothing to report prints the header and no rows at all, which
// is a real answer and not a parse failure.
func TestACaptureWithNoLoginsHasNoRows(t *testing.T) {
	empty := `===================================================================
Packet     Protocol         Username         Info            
------     --------         --------         --------
===================================================================
`
	assert.Empty(t, ParseCredentials(empty))
}

func TestOutputWithNoHeaderYieldsNoLogins(t *testing.T) {
	assert.Empty(t, ParseCredentials(""))
	assert.Empty(t, ParseCredentials("tshark: no such tap\n"))
}

//======================================================================

// Measured, not assumed: tshark accepts a display filter for this tap, exits
// zero, and reports the same rows. So it is never given one, and the dialog
// does not claim a narrowing that did not happen.
func TestTheCredentialsTapIsNeverGivenAFilter(t *testing.T) {
	assert.True(t, Credentials.IgnoresFilter())
	assert.Equal(t, "credentials", Credentials.ZArg("tcp.port == 80"))
	assert.Equal(t, []string{"credentials"}, Credentials.ZArgs("tcp.port == 80"))
}

func TestEveryOtherStatStillTakesTheFilter(t *testing.T) {
	for _, s := range All {
		if s.Command == Credentials.Command {
			continue
		}
		assert.False(t, s.IgnoresFilter(), s.Command)
		assert.Contains(t, s.ZArg("tcp"), ",tcp", s.Command)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
