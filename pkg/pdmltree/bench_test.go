// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pdmltree

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/antchfx/xmlquery"
)

//======================================================================

// This runs once for every press of an arrow key in the packet list: the
// selected packet's PDML is decoded into a tree from scratch, with nothing
// remembering the tree that was just thrown away. Whatever it costs is what
// scrolling costs.
//
//	go test -run XXX -bench DecodePacket -benchmem ./pkg/pdmltree/
func BenchmarkDecodePacket(b *testing.B) {
	data := []byte(p1)

	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if DecodePacket(data) == nil {
			b.Fatal("packet did not decode")
		}
	}
}

// The two halves of DecodePacket, measured apart, because they are not needed
// at the same times: the tree is what the middle pane draws on every keypress,
// and the query document is only read when somebody follows a stream.
func BenchmarkDecodePacketTreeOnly(b *testing.B) {
	data := []byte(p1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d := xml.NewDecoder(bytes.NewReader(data))
		var n Model
		if err := d.Decode(&n); err != nil {
			b.Fatal(err)
		}
		n.removeUnneeded()
	}
}

func BenchmarkDecodePacketQueryModelOnly(b *testing.B) {
	data := []byte(p1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := xmlquery.Parse(strings.NewReader(string(data))); err != nil {
			b.Fatal(err)
		}
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
