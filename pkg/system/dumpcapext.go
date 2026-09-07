// Copyright 2026 m1rwana12. All rights reserved.
// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// This is the full version, using Dup2 to hand a descriptor to tshark. The
// arm64 file next to it is the version without that, for platforms where Dup2
// is not available - the BSDs on arm64. Linux has it on every architecture, so
// linux stays here whatever the arch; anything else on arm64 goes there.
//go:build !windows && !darwin && (linux || !arm64)
// +build !windows
// +build !darwin
// +build linux !arm64

package system

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
)

//======================================================================

var fdre *regexp.Regexp = regexp.MustCompile(`/dev/fd/([[:digit:]]+)`)

// DumpcapExt will run dumpcap first, but if it fails, run tshark. Intended as
// a special case to allow pcaptui -i <iface> to use dumpcap if possible,
// but if it fails (e.g. iface==randpkt), fall back to tshark. dumpcap is more
// efficient than tshark at just capturing, and will drop fewer packets, but
// tshark supports extcap interfaces.
func DumpcapExt(dumpcapBin string, tsharkBin string, args ...string) error {
	var err error

	// If the first argument is /dev/fd/X, it means the process should have
	// descriptor X open and will expect packet data to be readable on it.
	// This /dev/fd feature does not work on tshark when run on freebsd, meaning
	// tshark will fail if you do something like
	//
	// cat foo.pcap | tshark -r /dev/fd/0
	//
	// The fix here is to replace /dev/fd/X with the arg "-", which tshark will
	// interpret as stdin, then dup descriptor X to 0 before starting dumpcap/tshark
	//
	if len(args) >= 2 {
		if os.Getenv("PCAPTUI_REPLACE_DEVFD") != "0" {
			fdnum := fdre.FindStringSubmatch(args[1])
			if len(fdnum) == 2 {
				fd, err := strconv.Atoi(fdnum[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unexpected error parsing %s: %v\n", args[1], err)
				} else {
					err = Dup2(fd, 0)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Problem duplicating fd %d to 0: %v\n", fd, err)
						fmt.Fprintf(os.Stderr, "Will not try to replace argument %s to tshark\n", args[1])
					} else {
						fmt.Fprintf(os.Stderr, "Replacing argument %s with - for tshark compatibility\n", args[1])
						args[1] = "-"
					}
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Starting pcaptui's custom live capture procedure.\n")
	dumpcapCmd := exec.Command(dumpcapBin, args...)
	fmt.Fprintf(os.Stderr, "Trying dumpcap command %v\n", dumpcapCmd)
	dumpcapCmd.Stdin = os.Stdin
	dumpcapCmd.Stdout = os.Stdout
	dumpcapCmd.Stderr = os.Stderr
	if dumpcapCmd.Run() != nil {
		var tshark string
		tshark, err = exec.LookPath(tsharkBin)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Retrying with capture command %v\n", append([]string{tshark}, args...))
			err = syscall.Exec(tshark, append([]string{tshark}, args...), os.Environ())
		}
	}

	return err
}
