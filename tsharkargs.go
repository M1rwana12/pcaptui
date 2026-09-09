// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcaptui

//======================================================================

// tsharkExtras are the arguments that belong on every tshark command line
// pcaptui builds, whichever part of the program builds it.
//
// They were reaching only the commands that read packets - the packet list,
// the protocol tree and the bytes - because those are built by pkg/pcap and
// the four other loaders each assemble their own argument list from scratch.
// The consequences were all of the same shape, and all silent:
//
//   - `--tls-keylog` decrypted the packet list and not the reassembled
//     stream, while the User Guide says in as many words that with the
//     session keys "stream reassembly shows the plaintext". It showed
//     ciphertext.
//   - `-d tcp.port==23,http` re-dissected the packet list and not the
//     protocol hierarchy, so the same capture answered "telnet" and "http" to
//     two questions about what was in it.
//   - `main.tshark-args` is documented as "added to every invocation pcaptui
//     makes" and was added to three of seven.
var tsharkExtras []string

// SetTsharkExtras records the arguments, once, at startup. decodeAs is the
// list of -d values; args is everything else already in tshark's own spelling.
func SetTsharkExtras(decodeAs []string, args []string) {
	res := make([]string, 0, len(decodeAs)*2+len(args))

	for _, d := range decodeAs {
		res = append(res, "-d", d)
	}
	res = append(res, args...)

	tsharkExtras = res
}

// TsharkExtras returns a copy, because callers append their own arguments to
// what they get back and the slice is shared by every loader in the program.
func TsharkExtras() []string {
	res := make([]string, len(tsharkExtras))
	copy(res, tsharkExtras)
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
