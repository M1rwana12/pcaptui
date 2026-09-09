// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

// Package export pulls the files carried by a capture out of it, using
// tshark's --export-objects.
//
// The objects are what was transferred rather than what was sent: a file
// downloaded over HTTP, a mail body, a block copied over SMB. Wireshark shows
// them under File > Export Objects.
package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/m1rwana12/pcaptui"
)

//======================================================================

// Types asks tshark which kinds of object this build can export.
//
// Asked rather than listed here, because the answer is a property of the
// tshark on this machine: the list has grown between Wireshark releases, and a
// hard-coded one would offer types that fail or hide types that work.
func Types() ([]string, error) {
	// tshark prints the list and exits successfully; it goes to stderr,
	// because to tshark this is a usage message.
	cmd := exec.Command(pcaptui.TSharkBin(), "--export-objects", "help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not ask %s which objects it can export: %v",
			pcaptui.TSharkBin(), err)
	}

	types := ParseTypes(string(out))
	if len(types) == 0 {
		return nil, fmt.Errorf("%s listed no object types it can export. It said:\n%s",
			pcaptui.TSharkBin(), strings.TrimSpace(string(out)))
	}

	return types, nil
}

// ParseTypes reads the list out of tshark's answer:
//
//	tshark: The available export object types for the "--export-objects" ...
//	     dicom
//	     ftp-data
//	     http
//
// A sentence and then one type per line. The sentence is recognised by not
// being a type: a type is one word of lower-case letters, digits and dashes.
func ParseTypes(out string) []string {
	var res []string

	for _, line := range strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n") {
		word := strings.TrimSpace(line)
		if typeName.MatchString(word) {
			res = append(res, word)
		}
	}

	return res
}

var typeName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

//======================================================================

// Dir is where the objects of one type from one capture go.
//
// Under pcaptui's own directory rather than beside the capture: a capture is
// often somewhere the user cannot write - handed over read-only, or on a
// mounted share - and a failed export is a worse answer than a known place to
// find the files. The name says which capture and which type, and carries a
// timestamp so that exporting twice does not mix the two.
func Dir(pcapPath string, typ string) string {
	base := strings.TrimSuffix(filepath.Base(pcapPath), filepath.Ext(pcapPath))

	return filepath.Join(pcaptui.PcapDir(), "objects",
		fmt.Sprintf("%s-%s-%s", clean(base), clean(typ), pcaptui.DateStringForFilename()))
}

var unsafeInName = regexp.MustCompile(`[^a-zA-Z0-9.-]`)

func clean(s string) string {
	return unsafeInName.ReplaceAllString(s, "_")
}

//======================================================================

// Objects writes every object of one type in the capture into dir, and returns
// the names of the files that appeared.
//
// The whole capture, never the display filter. An object is reassembled from
// the packets carrying it, and a filter that hides some of them would write a
// file that is quietly short - which is a worse outcome than exporting more
// than was asked for.
//
// tshark says nothing at all about what it wrote, and exits successfully
// whether it wrote a hundred files or none, so the directory is read before
// and after and the difference is the answer.
func Objects(pcapPath string, typ string, dir string) ([]string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("could not make %s: %v", dir, err)
	}

	before, err := names(dir)
	if err != nil {
		return nil, err
	}

	cmd := objectsCommand(pcapPath, typ, dir)

	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%s could not export %s objects: %v\n%s",
			pcaptui.TSharkBin(), typ, err, strings.TrimSpace(string(out)))
	}

	after, err := names(dir)
	if err != nil {
		return nil, err
	}

	return added(before, after), nil
}

// objectsCommand is the tshark that does the writing.
//
// It carries the shared arguments - the decode-as rules, the TLS key log and
// main.tshark-args - because the objects inside a decrypted session do not
// exist without them, and a decode-as rule decides what a stream even is.
func objectsCommand(pcapPath string, typ string, dir string) *exec.Cmd {
	args := []string{"-q", "-r", pcapPath, "--export-objects", typ + "," + dir}
	args = append(args, pcaptui.TsharkExtras()...)

	return exec.Command(pcaptui.TSharkBin(), args...)
}

// names is the set of file names directly in dir.
func names(dir string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %v", dir, err)
	}

	res := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			res[e.Name()] = struct{}{}
		}
	}

	return res, nil
}

// added is what is in after and was not in before, in order.
func added(before map[string]struct{}, after map[string]struct{}) []string {
	res := make([]string, 0, len(after))
	for name := range after {
		if _, had := before[name]; !had {
			res = append(res, name)
		}
	}
	sort.Strings(res)

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
