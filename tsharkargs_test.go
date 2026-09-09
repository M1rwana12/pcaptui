// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcaptui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestDecodeAsRulesBecomeFlagPairs(t *testing.T) {
	defer SetTsharkExtras(nil, nil)
	SetTsharkExtras([]string{"tcp.port==23,http", "udp.port==2055,cflow"}, nil)

	assert.Equal(t,
		[]string{"-d", "tcp.port==23,http", "-d", "udp.port==2055,cflow"},
		TsharkExtras())
}

func TestTheDecodeRulesComeBeforeEverythingElse(t *testing.T) {
	defer SetTsharkExtras(nil, nil)
	SetTsharkExtras([]string{"tcp.port==23,http"},
		[]string{"-o", "tls.keylog_file:/home/me/keys.log"})

	assert.Equal(t,
		[]string{"-d", "tcp.port==23,http", "-o", "tls.keylog_file:/home/me/keys.log"},
		TsharkExtras())
}

func TestNothingSetIsNoArguments(t *testing.T) {
	SetTsharkExtras(nil, nil)

	assert.Empty(t, TsharkExtras())
}

// Every loader appends its own arguments to what it gets back, so handing out
// the shared slice would let one of them write into another's.
func TestACallerCannotChangeWhatTheNextOneSees(t *testing.T) {
	defer SetTsharkExtras(nil, nil)
	SetTsharkExtras(nil, []string{"-o", "tls.keylog_file:/keys"})

	mine := TsharkExtras()
	mine[0] = "-X"
	mine = append(mine, "-q")

	assert.Equal(t, []string{"-o", "tls.keylog_file:/keys"}, TsharkExtras())
	assert.Len(t, mine, 3)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
