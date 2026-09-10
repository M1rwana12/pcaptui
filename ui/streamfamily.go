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

// knownField answers whether this tshark has a field of that name, for
// Family.Resolve. Before the field list has loaded there is nothing to ask, and
// nil means "assume the field is there".
func knownField() func(string) bool {
	if FieldCompleter == nil {
		return nil
	}
	return func(name string) bool {
		ok, _ := FieldCompleter.LookupField(name)
		return ok
	}
}

// pickStreamFamily answers which stream to follow for this packet: the family,
// and the one or two indexes that name it.
//
// With want == nil this is the `s` key: the transport stream, TCP first, then
// UDP - unchanged behaviour. With a want it is :streams <family>, and the
// packet has to actually carry that family; a stream index alone does not
// prove it, because a TLS packet has a tcp.stream like any other.
func pickStreamFamily(want *streams.Family, pkt streamPacket) (streams.Ref, error) {
	return pickStreamFamilyWith(want, pkt, knownField())
}

func pickStreamFamilyWith(want *streams.Family, pkt streamPacket, known func(string) bool) (streams.Ref, error) {
	if want != nil {
		resolved := want.Resolve(known)
		want = &resolved

		if !pkt.HasLayer(want.Layer) {
			return streams.Ref{}, fmt.Errorf(
				"This packet carries no %s.%s", want.Label, offerFor(pkt, known))
		}

		ref, err := refFor(*want, pkt)
		if err != nil {
			return streams.Ref{}, err
		}

		return ref, nil
	}

	for _, f := range streams.Families() {
		if f.Explicit {
			continue
		}
		if !pkt.HasLayer(f.Layer) {
			continue
		}
		if ref, err := refFor(f, pkt); err == nil {
			return ref, nil
		}
	}

	return streams.Ref{}, fmt.Errorf("Please select a TCP or UDP packet.")
}

// refFor reads the one or two indexes that name this packet's stream of that
// family. Naming the field that is missing is the only useful thing to say
// when tshark dissected the layer but did not number it - that field is the
// thing that would have to be there.
func refFor(f streams.Family, pkt streamPacket) (streams.Ref, error) {
	idx := pkt.FieldIndex(f.IndexField)
	if idx.IsNone() {
		return streams.Ref{}, fmt.Errorf(
			"This %s packet has no %s, so there is no stream to follow.", f.Label, f.IndexField)
	}

	ref := streams.Ref{Family: f, Index: idx.Val()}

	if f.SubIndexField == "" {
		return ref, nil
	}

	// A two-index family. Handing tshark one index is not an empty answer but
	// an error, so a packet whose second index is missing has to be refused
	// here rather than followed.
	sub := pkt.FieldIndex(f.SubIndexField)
	if sub.IsNone() {
		return streams.Ref{}, fmt.Errorf(
			"This %s packet has no %s. %s multiplexes streams over one connection, "+
				"so a stream of it needs both numbers; this packet carries only the connection.",
			f.Label, f.SubIndexField, f.Label)
	}

	ref.Sub = sub.Val()
	return ref, nil
}

// offerFor names the families this packet does have, so that a refusal is not
// merely a refusal. It returns the empty string when there is nothing to
// offer, which is the one case where there is nothing useful to add.
func offerFor(pkt streamPacket, known func(string) bool) string {
	got := make([]string, 0, 2)
	for _, f := range streams.Families() {
		f = f.Resolve(known)
		if !pkt.HasLayer(f.Layer) {
			continue
		}
		if _, err := refFor(f, pkt); err == nil {
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
