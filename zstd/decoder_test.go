package zstd

import (
	"runtime"
	"testing"
)

func TestDecoderMemoryLeak(t *testing.T) {
	runtime.GC()
	var ms1 runtime.MemStats
	runtime.ReadMemStats(&ms1)

	dict := make([]byte, 10<<20) // 10MB
	for i := range dict {
		dict[i] = byte(i)
	}

	dec, err := NewReader(nil, WithDecoderDict(dict))
	if err != nil {
		t.Fatal(err)
	}

	dec.Close()

	dict = nil
	dec = nil

	runtime.GC()
	runtime.GC()

	var ms2 runtime.MemStats
	runtime.ReadMemStats(&ms2)

	if diff := int64(ms2.HeapAlloc) - int64(ms1.HeapAlloc); diff > 2<<20 {
		t.Errorf("Memory leak detected: heap allocation increased by %d bytes", diff)
	}
}

func TestDecoderMemoryLeakReset(t *testing.T) {
	runtime.GC()
	var ms1 runtime.MemStats
	runtime.ReadMemStats(&ms1)

	dict := make([]byte, 10<<20) // 10MB
	for i := range dict {
		dict[i] = byte(i)
	}

	dec, err := NewReader(nil, WithDecoderDict(dict))
	if err != nil {
		t.Fatal(err)
	}

	dec.Reset(nil)

	dict = nil
	dec = nil

	runtime.GC()
	runtime.GC()

	var ms2 runtime.MemStats
	runtime.ReadMemStats(&ms2)

	if diff := int64(ms2.HeapAlloc) - int64(ms1.HeapAlloc); diff > 2<<20 {
		t.Errorf("Memory leak detected after Reset(nil): heap allocation increased by %d bytes", diff)
	}
}
