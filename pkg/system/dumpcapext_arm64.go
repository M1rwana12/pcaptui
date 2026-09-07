// Copyright 2026 m1rwana12. All rights reserved.
// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// The filename already restricts this to arm64. Windows has its own
// implementation and must be excluded here, or both files are compiled for
// windows/arm64 and DumpcapExt is declared twice.
//go:build !darwin && !linux && !windows
// +build !darwin,!linux,!windows

package system

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

//======================================================================

// DumpcapExt will run dumpcap first, but if it fails, run tshark. Intended as
// a special case to allow pcaptui -i <iface> to use dumpcap if possible,
// but if it fails (e.g. iface==randpkt), fall back to tshark. dumpcap is more
// efficient than tshark at just capturing, and will drop fewer packets, but
// tshark supports extcap interfaces.
func DumpcapExt(dumpcapBin string, tsharkBin string, args ...string) error {
	var err error

	dumpcapCmd := exec.Command(dumpcapBin, args...)
	fmt.Fprintf(os.Stderr, "Starting pcaptui's custom live capture procedure.\n")
	fmt.Fprintf(os.Stderr, "Trying dumpcap command %v\n", dumpcapCmd)
	dumpcapCmd.Stdin = os.Stdin
	dumpcapCmd.Stdout = os.Stdout
	dumpcapCmd.Stderr = os.Stderr
	if dumpcapCmd.Run() != nil {
		var tshark string
		tshark, err = exec.LookPath(tsharkBin)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Retrying with dumpcap command %v\n", append([]string{tshark}, args...))
			err = syscall.Exec(tshark, append([]string{tshark}, args...), os.Environ())
		}
	}

	return err
}
