// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package ui

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid/gwutil"
	"github.com/m1rwana12/pcaptui/pkg/streams"
)

//======================================================================

// Deciding which stream to follow is arithmetic over one packet's fields, and
// it is kept out of the widget building around it so that it can be called
// from a test. The conversations parser was fused to its widgets in exactly
// that way, and an IPv6 bug lived in it for years because nothing could reach
// it.
type streamPacket interface {
	// FieldIndex reads a numeric display-filter field, e.g. tcp.stream.
	FieldIndex(field string) gwutil.IntOption
	// HasLayer says whether the packet carries this protocol at all.
	HasLayer(name string) bool
}

// pickStreamFamily answers which family to follow for this packet and which
// stream of it.
//
// With want == nil this is the `s` key: the transport stream, TCP first, then
// UDP - unchanged behaviour. With a want it is :streams <family>, and the
// packet has to actually carry that family; a stream index alone does not
// prove it, because a TLS packet has a tcp.stream like any other.
func pickStreamFamily(want *streams.Family, pkt streamPacket) (streams.Family, int, error) {
	if want != nil {
		if !pkt.HasLayer(want.Layer) {
			return streams.Family{}, 0, fmt.Errorf(
				"This packet carries no %s.%s", want.Label, offerFor(pkt))
		}

		idx := pkt.FieldIndex(want.IndexField)
		if idx.IsNone() {
			// tshark dissected the layer but did not number the stream. Naming
			// the field is the only useful thing to say: it is the thing that
			// would have to be there.
			return streams.Family{}, 0, fmt.Errorf(
				"This %s packet has no %s, so there is no stream to follow.",
				want.Label, want.IndexField)
		}

		return *want, idx.Val(), nil
	}

	for _, f := range streams.Families() {
		if f.Explicit {
			continue
		}
		if !pkt.HasLayer(f.Layer) {
			continue
		}
		if idx := pkt.FieldIndex(f.IndexField); !idx.IsNone() {
			return f, idx.Val(), nil
		}
	}

	return streams.Family{}, 0, fmt.Errorf("Please select a TCP or UDP packet.")
}

// offerFor names the families this packet does have, so that a refusal is not
// merely a refusal. It returns the empty string when there is nothing to
// offer, which is the one case where there is nothing useful to add.
func offerFor(pkt streamPacket) string {
	got := make([]string, 0, 2)
	for _, f := range streams.Families() {
		if pkt.HasLayer(f.Layer) && !pkt.FieldIndex(f.IndexField).IsNone() {
			got = append(got, ":streams "+f.Token)
		}
	}

	if len(got) == 0 {
		return ""
	}

	return fmt.Sprintf(" Here you can follow %s.", strings.Join(got, ", "))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
