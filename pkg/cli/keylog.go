// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"os"
	"strings"
)

//======================================================================

// KeylogFlag is pcaptui's own spelling of the option; tshark has no such
// flag and needs the preference below instead.
const KeylogFlag = "--tls-keylog"

// KeylogPref is the Wireshark preference that points at a TLS key log. It was
// called ssl.keylog_file before Wireshark 3.0, and current releases print a
// conversion warning if given the old name.
const KeylogPref = "tls.keylog_file"

// KeylogArg renders a keylog path as the argument pair tshark expects.
func KeylogArg(path string) []string {
	return []string{"-o", fmt.Sprintf("%s:%s", KeylogPref, path)}
}

// CheckKeylog reports why path cannot serve as a TLS key log file.
//
// tshark does not do this itself. Given a path that does not exist, an empty
// path, or a directory, it starts normally, exits zero, decrypts nothing and
// says nothing - so a mistyped path is indistinguishable from traffic whose
// keys are genuinely absent. That silence is the reason this option exists as
// something more than a documented config key.
//
// An existing but empty file is fine: a browser creates the log when it starts
// and appends to it as sessions are negotiated.
func CheckKeylog(path string) error {
	if path == "" {
		return fmt.Errorf("no TLS key log file given")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("TLS key log file %s does not exist", path)
		}
		return fmt.Errorf("cannot use TLS key log file %s: %v", path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("TLS key log %s is not a file", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read TLS key log file %s: %v", path, err)
	}
	f.Close()

	return nil
}

//======================================================================

// TsharkArgs converts pcaptui's own command line into one tshark accepts,
// for the pass-thru case where pcaptui simply becomes tshark.
//
// Pcaptui-only flags are dropped, and --tls-keylog becomes the tshark
// preference, so decryption keeps working when output is piped.
func TsharkArgs(args []string) []string {
	res := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if path, ok := keylogValue(args, &i); ok {
			if path != "" {
				res = append(res, KeylogArg(path)...)
			}
			continue
		}

		if pcaptuiOnly(arg) {
			continue
		}

		res = append(res, arg)
	}

	return res
}

// pcaptuiOnly reports whether arg is one of pcaptui's own flags, in either
// the bare or the joined form, and so has no meaning to tshark.
func pcaptuiOnly(arg string) bool {
	for _, flag := range PcaptuiOnly {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}

// keylogValue recognises both spellings of the flag - "--tls-keylog path" and
// "--tls-keylog=path" - and advances i past the value it consumed.
func keylogValue(args []string, i *int) (string, bool) {
	arg := args[*i]

	if arg == KeylogFlag {
		if *i+1 < len(args) {
			*i++
			return args[*i], true
		}
		// Trailing flag with nothing after it; go-flags will complain, and
		// there is nothing to hand tshark.
		return "", true
	}

	prefix := KeylogFlag + "="
	if len(arg) > len(prefix) && arg[:len(prefix)] == prefix {
		return arg[len(prefix):], true
	}

	return "", false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
