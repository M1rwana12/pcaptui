// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package shark

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// Verbatim from `tshark -G column-formats`, Wireshark 4.6.8. The third
// tab-separated field arrived in Wireshark 4.x; older releases printed two.
const columnFormats = "%q\t802.1Q VLAN id\tvlan.id\n" +
	"%Yt\tAbsolute date, as YYYY-MM-DD, and time\tframe.time\n" +
	"%m\tNumber\tframe.number\n" +
	"%i\tInformation\t\n"

func TestEveryFormatLineBecomesASpec(t *testing.T) {
	specs, err := ParseColumnFormats(strings.NewReader(columnFormats))

	require.NoError(t, err)
	require.Len(t, specs, 4)
	assert.Equal(t, "%q", specs[0].Field.Token)
	assert.Equal(t, "802.1Q VLAN id", specs[0].Name)
	assert.Equal(t, "%i", specs[3].Field.Token)
}

// Wireshark 4.x added a third field. Splitting on the wrong thing was a real
// bug here once: every column title carried its padding and its filter name.
func TestTwoFieldsAndThreeFieldsBothWork(t *testing.T) {
	specs, err := ParseColumnFormats(strings.NewReader("%m\tNumber\n%t\tTime\tframe.time\n"))

	require.NoError(t, err)
	require.Len(t, specs, 2)
	assert.Equal(t, "Number", specs[0].Name)
	assert.Equal(t, "Time", specs[1].Name)
}

func TestLinesThatAreNotFormatsAreSkipped(t *testing.T) {
	specs, err := ParseColumnFormats(strings.NewReader(
		"some preamble\n%m\tNumber\tframe.number\n\n"))

	require.NoError(t, err)
	require.Len(t, specs, 1)
}

// tshark on Windows ends its lines with CRLF, and the stray CR used to end up
// inside the column's field name.
func TestWindowsLineEndingsAreStripped(t *testing.T) {
	specs, err := ParseColumnFormats(strings.NewReader("%m\tNumber\tframe.number\r\n"))

	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, "Number", specs[0].Name)
}

//======================================================================

// The reason this function exists. No columns at all means tshark did not run,
// or ran and printed something else - every build since 1.10 knows dozens.
// Treating that as success wrote an empty list to the cache, and since the
// cache is only rebuilt when the tshark binary is newer than it, one failed
// run left the install permanently unable to name a column: Edit Columns
// offered nothing, and every configured column was discarded as unrecognised.
func TestNoColumnsIsAnErrorNotAnEmptyAnswer(t *testing.T) {
	for _, in := range []string{"", "\n\n", "tshark: unrecognized option\n"} {
		_, err := ParseColumnFormats(strings.NewReader(in))
		assert.Error(t, err, "should have refused %q", in)
	}
}

func TestTheErrorSaysWhatWasWrong(t *testing.T) {
	_, err := ParseColumnFormats(strings.NewReader(""))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "column formats")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
