// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

func lines(n int) string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("line %d", i)
	}
	return strings.Join(out, "\n")
}

func TestAShortLogIsShownWhole(t *testing.T) {
	s := "one\ntwo\nthree"

	assert.Equal(t, s, tail(s, 400))
}

func TestALongLogKeepsItsEnd(t *testing.T) {
	s := lines(1000)
	got := tail(s, 400)

	all := strings.Split(s, "\n")
	assert.True(t, strings.HasSuffix(got, all[len(all)-1]))
	assert.NotContains(t, got, "\n"+all[0]+"\n")
}

// Cutting silently would be the same fault the packet list had: what is on
// screen has to say when it is not everything.
func TestACutLogSaysHowMuchIsMissing(t *testing.T) {
	got := tail(lines(1000), 400)

	assert.Contains(t, got, "600")
	assert.Contains(t, got, "in the file itself")
}

func TestATrailingNewlineIsNotCountedAsALine(t *testing.T) {
	s := lines(400) + "\n"

	assert.Equal(t, s, tail(s, 400))
}

//======================================================================

func TestTheDialogNamesTheFileItIsShowing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pcaptui.log")
	require.NoError(t, os.WriteFile(path, []byte("hello\n"), 0o644))

	got := fileForDialog("Log", path, 400)

	assert.Contains(t, got, path)
	assert.Contains(t, got, "hello")
}

// A missing log is what someone with a broken install has, and it is the most
// useful thing the dialog can tell them.
func TestAMissingFileIsReportedRatherThanShownAsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-there.log")

	got := fileForDialog("Log", path, 400)

	assert.Contains(t, got, path)
	assert.Contains(t, got, "Could not read it")
}

func TestAnEmptyFileSaysSo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pcaptui.log")
	require.NoError(t, os.WriteFile(path, []byte("\n  \n"), 0o644))

	assert.Contains(t, fileForDialog("Log", path, 400), "empty")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
