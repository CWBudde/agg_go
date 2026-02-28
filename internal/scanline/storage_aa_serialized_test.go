package scanline

import (
	"encoding/binary"
	"reflect"
	"testing"

	"agg_go/internal/basics"
)

func buildSerializedAAData(t *testing.T) []byte {
	t.Helper()

	storage := NewScanlineStorageAA[basics.Int8u]()
	storage.Prepare()

	sl := NewMockScanline(30)
	sl.spans = append(sl.spans,
		MockSpan{
			X:      10,
			Len:    3,
			Covers: []basics.Int8u{10, 20, 30},
		},
		MockSpan{
			X:      20,
			Len:    -2, // Solid span.
			Covers: []basics.Int8u{200},
		},
	)
	storage.Render(sl)

	size := storage.ByteSize()
	data := make([]byte, size)
	storage.Serialize(data)
	return data
}

func TestSerializedScanlinesAdaptorAARewind(t *testing.T) {
	data := buildSerializedAAData(t)
	adaptor := NewSerializedScanlinesAdaptorAA[uint8](data, len(data), 5.0, 3.0)

	if !adaptor.RewindScanlines() {
		t.Fatal("expected rewind to succeed")
	}

	if adaptor.MinX() != 15 || adaptor.MaxX() != 26 {
		t.Fatalf("unexpected X bounds: got [%d, %d], want [15, 26]", adaptor.MinX(), adaptor.MaxX())
	}
	if adaptor.MinY() != 33 || adaptor.MaxY() != 33 {
		t.Fatalf("unexpected Y bounds: got [%d, %d], want [33, 33]", adaptor.MinY(), adaptor.MaxY())
	}
}

func TestSerializedScanlinesAdaptorAASweepScanline(t *testing.T) {
	data := buildSerializedAAData(t)
	adaptor := NewSerializedScanlinesAdaptorAA[uint8](data, len(data), 2.0, -1.0)

	if !adaptor.RewindScanlines() {
		t.Fatal("expected rewind to succeed")
	}

	out := NewMockScanline(0)
	if !adaptor.SweepScanline(out) {
		t.Fatal("expected first sweep to succeed")
	}

	if out.Y() != 29 {
		t.Fatalf("unexpected scanline Y: got %d, want 29", out.Y())
	}
	if out.NumSpans() != 2 {
		t.Fatalf("unexpected span count: got %d, want 2", out.NumSpans())
	}

	s0 := out.spans[0]
	if s0.X != 12 || s0.Len != 3 || !reflect.DeepEqual(s0.Covers, []basics.Int8u{10, 20, 30}) {
		t.Fatalf("unexpected first span: %+v", s0)
	}

	s1 := out.spans[1]
	if s1.X != 22 || s1.Len != -2 || !reflect.DeepEqual(s1.Covers, []basics.Int8u{200}) {
		t.Fatalf("unexpected second span: %+v", s1)
	}

	if adaptor.SweepScanline(out) {
		t.Fatal("expected second sweep to return false")
	}
}

func TestSerializedScanlinesAdaptorAASweepEmbedded(t *testing.T) {
	data := buildSerializedAAData(t)
	adaptor := NewSerializedScanlinesAdaptorAA[uint8](data, len(data), 1.0, 2.0)

	if !adaptor.RewindScanlines() {
		t.Fatal("expected rewind to succeed")
	}

	embedded := NewSerializedEmbeddedScanline[uint8]()
	if !adaptor.SweepSerializedEmbeddedScanline(embedded) {
		t.Fatal("expected embedded sweep to succeed")
	}

	if embedded.Y() != 32 {
		t.Fatalf("unexpected embedded Y: got %d, want 32", embedded.Y())
	}
	if embedded.NumSpans() != 2 {
		t.Fatalf("unexpected embedded span count: got %d, want 2", embedded.NumSpans())
	}

	iter := embedded.Begin()
	if !iter.IsValid() {
		t.Fatal("iterator should be valid at first span")
	}
	if iter.X() != 11 || iter.Len() != 3 || !reflect.DeepEqual(iter.Covers(), []uint8{10, 20, 30}) {
		t.Fatalf("unexpected first embedded span: x=%d len=%d covers=%v", iter.X(), iter.Len(), iter.Covers())
	}

	if !iter.Next() {
		t.Fatal("expected second span to exist")
	}
	if !iter.IsValid() {
		t.Fatal("iterator should be valid on second span")
	}
	if iter.X() != 21 || iter.Len() != -2 || !reflect.DeepEqual(iter.Covers(), []uint8{200}) {
		t.Fatalf("unexpected second embedded span: x=%d len=%d covers=%v", iter.X(), iter.Len(), iter.Covers())
	}

	if iter.Next() {
		t.Fatal("expected iterator to end after second span")
	}
	if iter.IsValid() {
		t.Fatal("iterator should be invalid after exhausting spans")
	}
}

func TestSerializedScanlinesAdaptorAAClampsSize(t *testing.T) {
	data := buildSerializedAAData(t)

	// Size larger than data length should be clamped, not panic.
	adaptor := NewSerializedScanlinesAdaptorAA[uint8](data, len(data)+1024, 0.0, 0.0)
	if !adaptor.RewindScanlines() {
		t.Fatal("expected rewind to succeed with oversized size parameter")
	}

	// Truncate the stream by one byte within the first serialized scanline payload.
	// Layout: [16 bytes bounds][4 bytes scanline_size][...scanline payload...].
	if len(data) < 20 {
		t.Fatalf("serialized stream too short for test: %d", len(data))
	}
	scanlineSize := int(int32(binary.LittleEndian.Uint32(data[16:20])))
	truncatedSize := 16 + scanlineSize - 1
	if truncatedSize < 20 {
		t.Fatalf("invalid computed truncation size: %d", truncatedSize)
	}

	truncated := NewSerializedScanlinesAdaptorAA[uint8](data, truncatedSize, 0.0, 0.0)
	if !truncated.RewindScanlines() {
		t.Fatal("expected rewind to succeed for partially truncated stream header")
	}
	out := NewMockScanline(0)
	if truncated.SweepScanline(out) {
		t.Fatal("expected sweep to fail for truncated serialized stream")
	}
}
