// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestKeylogArgIsTsharkPreferenceSyntax(t *testing.T) {
	assert.Equal(t,
		[]string{"-o", "tls.keylog_file:/home/me/keys.log"},
		KeylogArg("/home/me/keys.log"))
}

// Wireshark renamed the preference from ssl.keylog_file in 3.0 and now warns
// that the old spelling is converted. Emitting the current name keeps that
// warning off the user's terminal.
func TestKeylogArgUsesTheCurrentPreferenceName(t *testing.T) {
	assert.Contains(t, KeylogArg("x")[1], "tls.keylog_file:")
	assert.NotContains(t, KeylogArg("x")[1], "ssl.keylog_file")
}

func TestCheckKeylogAcceptsAReadableFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "keys.log")
	assert.NoError(t, os.WriteFile(p, []byte("CLIENT_RANDOM abc def\n"), 0644))

	assert.NoError(t, CheckKeylog(p))
}

// An empty keylog is normal: the browser creates the file and writes to it as
// sessions are negotiated. Rejecting it would be wrong.
func TestCheckKeylogAcceptsAnEmptyFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "keys.log")
	assert.NoError(t, os.WriteFile(p, nil, 0644))

	assert.NoError(t, CheckKeylog(p))
}

// This is the whole point of the check. tshark accepts a keylog path that does
// not exist, exits 0, and decrypts nothing - so a typo looks exactly like
// traffic it cannot decrypt.
func TestCheckKeylogRejectsAMissingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nope.log")

	err := CheckKeylog(p)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), p, "the message has to name the path")
}

func TestCheckKeylogRejectsADirectory(t *testing.T) {
	err := CheckKeylog(t.TempDir())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a file")
}

func TestCheckKeylogRejectsAnEmptyPath(t *testing.T) {
	assert.Error(t, CheckKeylog(""))
}

//======================================================================

func TestTsharkArgsPassesOrdinaryArgumentsThrough(t *testing.T) {
	in := []string{"-r", "foo.pcap", "-Y", "tcp.port==80"}

	assert.Equal(t, in, TsharkArgs(in))
}

func TestTsharkArgsDropsPcaptuiOnlyFlags(t *testing.T) {
	in := []string{"--pass-thru=true", "-r", "foo.pcap", "--log-tty"}

	assert.Equal(t, []string{"-r", "foo.pcap"}, TsharkArgs(in))
}

// The joined form is the one the existing filter already handled.
func TestTsharkArgsTranslatesJoinedKeylogFlag(t *testing.T) {
	in := []string{"--tls-keylog=/tmp/keys.log", "-r", "foo.pcap"}

	assert.Equal(t,
		[]string{"-o", "tls.keylog_file:/tmp/keys.log", "-r", "foo.pcap"},
		TsharkArgs(in))
}

// The separated form is the one people actually type, and the one the old
// filter would have left a stray path behind for.
func TestTsharkArgsTranslatesSeparatedKeylogFlag(t *testing.T) {
	in := []string{"--tls-keylog", "/tmp/keys.log", "-r", "foo.pcap"}

	assert.Equal(t,
		[]string{"-o", "tls.keylog_file:/tmp/keys.log", "-r", "foo.pcap"},
		TsharkArgs(in))
}

func TestTsharkArgsIgnoresATrailingKeylogFlagWithNoValue(t *testing.T) {
	in := []string{"-r", "foo.pcap", "--tls-keylog"}

	assert.Equal(t, []string{"-r", "foo.pcap"}, TsharkArgs(in))
}

//======================================================================

// --two-pass is pcaptui's spelling of a flag tshark has, so in pass-thru it is
// translated rather than dropped: a pipeline should behave the way the
// terminal does.
func TestTwoPassBecomesTsharksOwnFlag(t *testing.T) {
	got := TsharkArgs([]string{"--two-pass", "-r", "foo.pcap"})

	assert.Equal(t, []string{"-2", "-r", "foo.pcap"}, got)
}

func TestWithoutTwoPassNothingIsAdded(t *testing.T) {
	got := TsharkArgs([]string{"-r", "foo.pcap"})

	assert.NotContains(t, got, "-2")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
