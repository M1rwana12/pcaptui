// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	"fmt"
	"runtime"
	"testing"
)

//======================================================================

// The colours of a million packets, measured rather than reasoned about.
//
// Every packet carries a PacketColors, and every PacketColors carries two
// gowid.IColor interfaces. An interface is a word for the type and a word for
// the value; whether that costs an allocation as well depends on what is
// stored in it, which is the thing worth measuring.
//
//	go test -run XXX -bench Colors -benchmem ./pkg/pcap/
func BenchmarkPsmlColorsPerMillionPackets(b *testing.B) {
	const packets = 1000000

	// tshark's colour filters give a handful of distinct pairs across a whole
	// capture - the default profile has about twenty rules - so a million
	// packets draw from a very small set.
	pairs := [][2]string{
		{"#000000", "#d7ffd7"},
		{"#000000", "#ffffff"},
		{"#ffffff", "#800000"},
		{"#000000", "#ffff5f"},
		{"#00005f", "#e0e0e0"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		b.StartTimer()

		cols := make([]PacketColors, 0, packets)
		for p := 0; p < packets; p++ {
			pr := pairs[p%len(pairs)]
			cols = append(cols, PacketColors{
				FG: psmlColorToIColor(pr[0]),
				BG: psmlColorToIColor(pr[1]),
			})
		}

		b.StopTimer()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc)/(1024*1024), "MB/million")
		runtime.KeepAlive(cols)
		b.StartTimer()
	}
}

// The same million packets through the palette, which is what the loader does
// now. This is the comparison that justified the change.
func BenchmarkPsmlColorsAsIndexPerMillionPackets(b *testing.B) {
	const packets = 1000000

	pairs := [][2]string{
		{"#000000", "#d7ffd7"},
		{"#000000", "#ffffff"},
		{"#ffffff", "#800000"},
		{"#000000", "#ffff5f"},
		{"#00005f", "#e0e0e0"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		b.StartTimer()

		pal := NewColorPalette()
		idx := make([]uint16, 0, packets)
		for p := 0; p < packets; p++ {
			pr := pairs[p%len(pairs)]
			idx = append(idx, pal.Add(pr[0], pr[1]))
		}

		b.StopTimer()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc)/(1024*1024), "MB/million")
		runtime.KeepAlive(idx)
		runtime.KeepAlive(pal)
		b.StartTimer()
	}
}

// How many distinct colour pairs a capture really has is the number that
// decides whether an index is even possible. If it were thousands, a palette
// would just move the memory rather than remove it.
func BenchmarkDistinctColourPairs(b *testing.B) {
	b.Skip("not a benchmark - run TestDistinctColourPairsAreFew instead")
}

func TestDistinctColourPairsAreFew(t *testing.T) {
	seen := map[string]int{}

	// Wireshark's default colour filter set is about twenty rules, and each
	// gives one foreground/background pair.
	for i := 0; i < 1000; i++ {
		fg := fmt.Sprintf("#%06x", i%20)
		bg := fmt.Sprintf("#%06x", (i*7)%20)
		seen[fg+bg]++
	}

	if len(seen) > 255 {
		t.Fatalf("a uint8 index cannot hold %d distinct pairs", len(seen))
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
