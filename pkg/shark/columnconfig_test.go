// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package shark

import (
	"testing"

	"github.com/m1rwana12/pcaptui/configs/profiles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// This is what the packet list shows, read on every capture load from a file
// the user can edit by hand. Its branches are the ones that keep a broken
// configuration from taking the program down with it - and hiding every column
// really did, before AllHidden: gowid renders each column as Max, finds no
// maximum height and panics before the interface appears.
//
// The key is one of this test's own so that nothing here depends on, or
// disturbs, the configuration the machine is actually using.
const testColKey = "main.test-column-format"

func withColumns(t *testing.T, entries ...string) []PsmlColumnSpec {
	t.Helper()

	// The tokens are cross-referenced against what tshark reported; seed just
	// the two these tests use.
	saved := AllowedColumnFormats
	AllowedColumnFormats = map[string]PsmlColumnInfo{
		"%m": {Short: "No."},
		"%t": {Short: "Time"},
	}
	t.Cleanup(func() { AllowedColumnFormats = saved })

	profiles.SetConf(testColKey, entries)
	t.Cleanup(func() { profiles.DeleteConf(testColKey) })

	return GetPsmlColumnFormatFrom(testColKey)
}

// Nobody has a column-format until they edit their columns, so this is the
// normal state and not a problem to warn about.
func TestNoConfigurationMeansTheDefaults(t *testing.T) {
	got := withColumns(t)

	assert.Equal(t, DefaultPsmlColumnSpec, got)
}

// The entries come in threes: token, name, visible.
func TestAListThatIsNotAMultipleOfThreeFallsBack(t *testing.T) {
	got := withColumns(t, "%m", "No.")

	assert.Equal(t, DefaultPsmlColumnSpec, got)
}

func TestAWellFormedListIsUsed(t *testing.T) {
	got := withColumns(t, "%m", "No.", "true", "%t", "Time", "false")

	require.Len(t, got, 2)
	assert.Equal(t, "No.", got[0].Name)
	assert.False(t, got[0].Hidden)
	assert.Equal(t, "Time", got[1].Name)
	assert.True(t, got[1].Hidden, "visible false means hidden")
}

// A token this tshark does not have is dropped, and the columns that are
// understood still arrive - one unknown entry is not a reason to lose the lot.
func TestAnUnknownTokenIsSkippedAndTheRestSurvive(t *testing.T) {
	got := withColumns(t, "%zz", "Nonsense", "true", "%m", "No.", "true")

	require.Len(t, got, 1)
	assert.Equal(t, "No.", got[0].Name)
}

// Nothing understood at all is a different case: fall back rather than show a
// packet list with no columns in it.
func TestNothingUnderstoodFallsBack(t *testing.T) {
	got := withColumns(t, "%zz", "Nonsense", "true")

	assert.Equal(t, DefaultPsmlColumnSpec, got)
}

// The visible flag is a boolean; anything else is not a reason to guess.
func TestAVisibleFlagThatIsNotABooleanSkipsTheColumn(t *testing.T) {
	got := withColumns(t, "%m", "No.", "perhaps", "%t", "Time", "true")

	require.Len(t, got, 1)
	assert.Equal(t, "Time", got[0].Name)
}

// The one that stopped the program starting. gowid renders each column as Max,
// finds no maximum height among none of them, and panics before the interface
// is drawn: "All columns widgets were rendered Max, so there is no max height
// to use."
func TestEveryColumnHiddenFallsBackToTheDefaults(t *testing.T) {
	got := withColumns(t, "%m", "No.", "false", "%t", "Time", "false")

	assert.Equal(t, DefaultPsmlColumnSpec, got,
		"a packet list with every column hidden crashed before it was drawn")
	assert.False(t, AllHidden(got))
}

// An empty name is not an error: tshark's own short name for the column is
// used instead.
func TestAColumnWithNoNameTakesTsharksOwn(t *testing.T) {
	got := withColumns(t, "%m", "", "true")

	require.Len(t, got, 1)
	assert.Equal(t, "No.", got[0].Name)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
