package logcat

import "testing"

func TestRingBufferDropsOldEntries(t *testing.T) {
	buffer := NewRingBuffer(3)

	for i := 0; i < 5; i++ {
		buffer.Add(LogEntry{Message: "line"})
	}

	count, discarded, lastID := buffer.Stats()
	if count != 3 || discarded != 2 || lastID != 5 {
		t.Fatalf("unexpected stats: count=%d discarded=%d lastID=%d", count, discarded, lastID)
	}

	batch := buffer.GetAfter(0, 10)
	if len(batch.Entries) != 3 || batch.Entries[0].ID != 3 {
		t.Fatalf("unexpected entries after drop: %+v", batch.Entries)
	}
}

func TestRingBufferGetAfterAndClear(t *testing.T) {
	buffer := NewRingBuffer(10)
	buffer.Add(LogEntry{Message: "one"})
	buffer.Add(LogEntry{Message: "two"})
	buffer.Add(LogEntry{Message: "three"})

	batch := buffer.GetAfter(1, 10)
	if len(batch.Entries) != 2 || batch.Entries[0].Message != "two" {
		t.Fatalf("unexpected batch: %+v", batch.Entries)
	}

	buffer.Clear()
	count, discarded, lastID := buffer.Stats()
	if count != 0 || discarded != 0 || lastID != 0 {
		t.Fatalf("unexpected stats after clear: count=%d discarded=%d lastID=%d", count, discarded, lastID)
	}
}

func TestRingBufferAppendMultiline(t *testing.T) {
	buffer := NewRingBuffer(10)
	if buffer.AppendMultiline("orphan") {
		t.Fatal("expected orphan multiline append to fail")
	}

	buffer.Add(LogEntry{Message: "root"})
	if !buffer.AppendMultiline("    at com.example.Main.main(Main.kt:1)") {
		t.Fatal("expected multiline append to succeed")
	}

	batch := buffer.GetAfter(0, 10)
	if len(batch.Entries) != 1 || len(batch.Entries[0].Multiline) != 1 {
		t.Fatalf("unexpected multiline batch: %+v", batch.Entries)
	}
}
