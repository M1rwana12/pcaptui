// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package tsharktest decides whether a test that drives a real tshark can
// run at all.
//
// Thirty-one of this project's tests hand arguments to tshark and read what
// comes back, because that is the only half of the job a unit test cannot do:
// the arguments can be spelled the way this program intends and still not be
// the spelling tshark accepts. On a machine with no Wireshark installed those
// tests used to fail - thirty-one red results that say nothing about the code,
// and three minutes spent producing them, since each one waits for a process
// that never starts.
//
// Skipping them is the obvious answer and it has a trap in it: a CI runner
// whose Wireshark failed to install would then go green with a third of the
// suite quietly not run, which is exactly the kind of silence this project
// keeps finding bugs behind. So the skip is conditional on nobody having said
// tshark must be there. CI says so, by setting PCAPTUI_REQUIRE_TSHARK.
package tsharktest

import (
	"os"
	"os/exec"
	"sync"
	"testing"
)

// RequireEnv, when set to anything non-empty, turns a missing tshark from a
// reason to skip into a failure. The workflows set it.
const RequireEnv = "PCAPTUI_REQUIRE_TSHARK"

var (
	once       sync.Once
	runnable   bool
	runErr     error
	binary     = "tshark"
	versionArg = "-v"
)

// Available reports whether tshark can actually be run - not whether a file
// of that name exists somewhere on PATH. It is asked once per test binary.
func Available() bool {
	once.Do(func() {
		runErr = exec.Command(binary, versionArg).Run()
		runnable = runErr == nil
	})
	return runnable
}

// Need takes the calling test out of the run when there is no tshark to
// drive, and fails it instead if the environment says there must be one.
func Need(t testing.TB) {
	t.Helper()

	if Available() {
		return
	}

	if os.Getenv(RequireEnv) != "" {
		t.Fatalf("%s is set, so this test may not be skipped, but tshark would not run: %v",
			RequireEnv, runErr)
	}

	t.Skipf("no tshark to drive (%v); this test needs a real one. Set %s to make its absence an error.",
		runErr, RequireEnv)
}
