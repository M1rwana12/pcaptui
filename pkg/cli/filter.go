// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

//======================================================================

// CheckDisplayFilter reports why tshark will not accept filter.
//
// A display filter given on the command line is not checked anywhere else.
// tshark rejects it, writes to stderr and stops - and because errors from
// tshark are not shown by default, what the user sees is a program with no
// filename in the title, an empty filter box, three empty panes and an exit
// status of zero. The expression they mistyped is not even echoed back for
// them to correct.
//
// The filter box in the interface has always validated as you type, using
// exactly this method. This applies the same check to the flag.
//
// emptyPcap is a capture with no packets, which pcaptui writes into its cache
// directory at startup; tshark needs a file to read before it will parse a
// filter.
func CheckDisplayFilter(tsharkBin, emptyPcap, filter string) error {
	if strings.TrimSpace(filter) == "" {
		return nil
	}

	cmd := exec.Command(tsharkBin, "-Y", filter, "-r", emptyPcap)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("display filter %s\n%s", quoteFilter(filter), tsharkComplaint(stderr.String(), err))
	}

	return nil
}

// quoteFilter shows the expression as the user typed it, so the message can be
// read next to the command they ran.
func quoteFilter(filter string) string {
	return fmt.Sprintf("%q is not valid.", filter)
}

// tsharkComplaint prefers what tshark said over what the exit status implies.
// tshark's own message names the field or the syntax it choked on, and is far
// more use than "exit status 2".
func tsharkComplaint(stderr string, err error) string {
	msg := strings.TrimSpace(stderr)
	if msg == "" {
		return err.Error()
	}
	return msg
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
