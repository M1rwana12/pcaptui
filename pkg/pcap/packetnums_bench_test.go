// Copyright 2026 m1rwana12. All rights reserved.  Use of this source code is
// governed by the MIT license that can be found in the LICENSE file.

package pcap

import (
	"runtime"
	"testing"
)

//======================================================================

// Two map[int]int, one entry per packet each, against the slice they can both
// be derived from.
//
//	go test -run XXX -bench PacketNum -benchmem ./pkg/pcap/
func BenchmarkPacketNumberMapsPerMillionPackets(b *testing.B) {
	const packets = 1000000

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		b.StartTimer()

		numToRow := make(map[int]int)
		next := make(map[int]int)
		prev := 0
		for p := 0; p < packets; p++ {
			num := p + 1
			numToRow[num] = p
			next[prev] = num
			prev = num
		}

		b.StopTimer()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc)/(1024*1024), "MB/million")
		runtime.KeepAlive(numToRow)
		runtime.KeepAlive(next)
		b.StartTimer()
	}
}

func BenchmarkPacketNumberSlicePerMillionPackets(b *testing.B) {
	const packets = 1000000

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		b.StartTimer()

		nums := make([]int32, 0, packets)
		for p := 0; p < packets; p++ {
			nums = append(nums, int32(p+1))
		}

		b.StopTimer()
		runtime.ReadMemStats(&after)
		b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc)/(1024*1024), "MB/million")
		runtime.KeepAlive(nums)
		b.StartTimer()
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
