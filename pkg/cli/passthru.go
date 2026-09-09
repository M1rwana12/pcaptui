// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package cli

import (
	"strings"
)

//======================================================================

// TsharkArgs converts pcaptui's own command line into one tshark accepts,
// for the pass-thru case where pcaptui simply becomes tshark.
//
// Pcaptui-only flags are dropped. The two that mean something to tshark are
// translated rather than dropped, so that a pipeline behaves the way the
// terminal does: --tls-keylog becomes the preference that turns on
// decryption, and --two-pass becomes -2.
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

		if arg == TwoPassFlag {
			res = append(res, TwoPassArg)
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

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
