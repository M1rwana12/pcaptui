// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package cli

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

// These drive a real tshark. A unit test can only confirm that the check calls
// something; only tshark can confirm that what it calls agrees with the
// program that will later be handed the same filter.

const emptyCapture = "../../scripts/pcaps/demo.pcap"

func tsharkOrSkip(t *testing.T) string {
	t.Helper()

	path, err := exec.LookPath("tshark")
	if err != nil {
		t.Skip("tshark is not on PATH")
	}
	return path
}

func TestValidFilterIsAccepted(t *testing.T) {
	bin := tsharkOrSkip(t)

	assert.NoError(t, CheckDisplayFilter(bin, emptyCapture, "tcp.port == 443"))
	assert.NoError(t, CheckDisplayFilter(bin, emptyCapture, "http && ip.addr == 192.0.2.10"))
}

// The typo that started this: a field that does not exist. tshark rejects it,
// and without the check the program started anyway and showed nothing.
func TestMisspeltFieldIsRejected(t *testing.T) {
	bin := tsharkOrSkip(t)

	err := CheckDisplayFilter(bin, emptyCapture, "tcp.pxrt == 1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tcp.pxrt", "the message has to quote what was typed")
}

func TestBrokenSyntaxIsRejected(t *testing.T) {
	bin := tsharkOrSkip(t)

	for _, bad := range []string{
		"tcp.port ===",
		"(((((",
		"ip.addr == 1.2.3.4 and",
	} {
		assert.Error(t, CheckDisplayFilter(bin, emptyCapture, bad), "should have rejected %q", bad)
	}
}

// tshark's own words say which field or which piece of syntax it choked on.
// "exit status 2" does not.
func TestTsharkExplanationIsCarriedThrough(t *testing.T) {
	bin := tsharkOrSkip(t)

	err := CheckDisplayFilter(bin, emptyCapture, "tcp.pxrt == 1")

	assert.Error(t, err)
	assert.NotEqual(t, "exit status 2", err.Error())
	assert.Greater(t, len(strings.Split(err.Error(), "\n")), 1,
		"the error should carry tshark's explanation on its own line")
}

func TestNoFilterIsNotAnError(t *testing.T) {
	assert.NoError(t, CheckDisplayFilter("tshark-does-not-need-to-exist", "", ""))
	assert.NoError(t, CheckDisplayFilter("tshark-does-not-need-to-exist", "", "   "))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
