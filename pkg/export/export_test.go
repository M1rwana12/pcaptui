// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/m1rwana12/pcaptui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//======================================================================

// What tshark actually prints. It is a usage message, so it goes to stderr and
// begins with a sentence.
const typesHelp = `tshark: The available export object types for the "--export-objects" option are:
     dicom
     ftp-data
     http
     imf
     smb
     tftp
     x509af
`

func TestTheTypesAreReadOutOfTsharksAnswer(t *testing.T) {
	got := ParseTypes(typesHelp)

	assert.Equal(t,
		[]string{"dicom", "ftp-data", "http", "imf", "smb", "tftp", "x509af"},
		got)
}

// The sentence above the list is not a type, and neither is anything else
// tshark might print around it.
func TestTheSentenceIsNotAType(t *testing.T) {
	for _, line := range ParseTypes(typesHelp) {
		assert.NotContains(t, line, " ")
		assert.NotContains(t, line, ":")
	}
}

func TestAnAnswerWithNoListYieldsNothing(t *testing.T) {
	assert.Empty(t, ParseTypes(""))
	assert.Empty(t, ParseTypes("tshark: unknown option\n"))
	assert.Empty(t, ParseTypes("Usage: tshark [options]\n"))
}

//======================================================================

// The name says which capture and which type, so that two exports do not land
// in one directory and look like one.
func TestTheDirectoryNamesTheCaptureAndTheType(t *testing.T) {
	dir := Dir("/home/me/captures/demo.pcap", "http")

	base := filepath.Base(dir)
	assert.True(t, strings.HasPrefix(base, "demo-http-"), "got %q", base)
	assert.Equal(t, "objects", filepath.Base(filepath.Dir(dir)))
}

// A capture called "../../etc/passwd.pcap" must not put anything in /etc.
func TestACaptureNameCannotEscapeTheDirectory(t *testing.T) {
	dir := Dir("../../etc/passwd.pcap", "http")

	assert.NotContains(t, filepath.Base(dir), "..")
	assert.NotContains(t, filepath.Base(dir), string(filepath.Separator))
	assert.True(t, strings.HasPrefix(filepath.Base(dir), "passwd-http-"))
}

func TestTwoExportsOfTheSameThingDoNotShareADirectory(t *testing.T) {
	// The timestamp is to the second, so this only says the name carries one.
	dir := Dir("/tmp/demo.pcap", "http")

	assert.Regexp(t, `demo-http-\d{4}-\d{2}-\d{2}`, filepath.Base(dir))
}

//======================================================================

// added is what turns "tshark said nothing and exited zero" into an answer.
func TestOnlyTheNewFilesAreReported(t *testing.T) {
	before := map[string]struct{}{"old": {}}
	after := map[string]struct{}{"old": {}, "b": {}, "a": {}}

	assert.Equal(t, []string{"a", "b"}, added(before, after))
}

func TestNothingNewIsAnEmptyList(t *testing.T) {
	same := map[string]struct{}{"old": {}}

	assert.Empty(t, added(same, same))
}

//======================================================================

// These drive a real tshark over a real capture, because the argument this
// package builds is only right if tshark says it is.

const demoPcap = "../../scripts/pcaps/demo.pcap"

func TestTheObjectsInTheCaptureAreWritten(t *testing.T) {
	dir := t.TempDir()

	written, err := Objects(demoPcap, "http", dir)
	require.NoError(t, err)

	assert.Len(t, written, 3, "the demo capture carries three HTTP bodies")

	for _, name := range written {
		info, err := os.Stat(filepath.Join(dir, name))
		require.NoError(t, err, "%s was reported but is not there", name)
		assert.NotZero(t, info.Size(), "%s was written empty", name)
	}
}

// tshark exits successfully whether it wrote a hundred files or none, so this
// is the case that would otherwise look like success and leave the user
// looking for files that were never written.
func TestACaptureWithNoObjectsOfThatTypeWritesNothing(t *testing.T) {
	dir := t.TempDir()

	written, err := Objects(demoPcap, "tftp", dir)

	require.NoError(t, err, "no objects is not an error")
	assert.Empty(t, written)
}

// The directory is made if it is not there, including its parents - the name
// carries a timestamp, so it never is.
func TestTheDirectoryIsMadeIfItIsNotThere(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "not", "there", "yet")

	_, err := Objects(demoPcap, "http", dir)

	require.NoError(t, err)
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// Files already in the directory are not claimed as newly written.
func TestFilesThatWereAlreadyThereAreNotClaimed(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mine.txt"), []byte("hi"), 0644))

	written, err := Objects(demoPcap, "http", dir)
	require.NoError(t, err)

	assert.NotContains(t, written, "mine.txt")
	assert.Len(t, written, 3)
}

func TestAnUnknownTypeIsAnError(t *testing.T) {
	_, err := Objects(demoPcap, "nosuchthing", t.TempDir())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nosuchthing")
}

// The list this offers has to be the list tshark accepts.
func TestEveryTypeTsharkOffersIsOneItAccepts(t *testing.T) {
	types, err := Types()
	require.NoError(t, err)
	require.NotEmpty(t, types)

	for _, typ := range types {
		_, err := Objects(demoPcap, typ, t.TempDir())
		assert.NoError(t, err, "tshark offered %q and then refused it", typ)
	}
}

// The objects inside a decrypted session do not exist without the key log, and
// what counts as a stream at all depends on the decode-as rules.
func TestTheExportCommandCarriesTheSharedArguments(t *testing.T) {
	defer pcaptui.SetTsharkExtras(nil, nil)
	pcaptui.SetTsharkExtras([]string{"tcp.port==8080,http"},
		[]string{"-o", "tls.keylog_file:/keys"})

	got := strings.Join(objectsCommand("capture.pcap", "http", "/tmp/out").Args[1:], " ")

	assert.Contains(t, got, "-d tcp.port==8080,http")
	assert.Contains(t, got, "-o tls.keylog_file:/keys")
	assert.Contains(t, got, "--export-objects http,/tmp/out")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
