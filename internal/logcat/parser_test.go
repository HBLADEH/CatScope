package logcat

import "testing"

func TestParserThreadtime(t *testing.T) {
	parser := NewParser()

	entry, ok := parser.Parse("07-04 01:02:03.456  1234  5678 E AndroidRuntime: FATAL EXCEPTION: main")
	if !ok {
		t.Fatal("expected line to parse")
	}

	if entry.Timestamp != "07-04 01:02:03.456" || entry.PID != 1234 || entry.TID != 5678 || entry.Level != "E" {
		t.Fatalf("unexpected parsed fields: %+v", entry)
	}
	if entry.Tag != "AndroidRuntime" || entry.Message != "FATAL EXCEPTION: main" {
		t.Fatalf("unexpected tag/message: %+v", entry)
	}
}

func TestParserTagWithSpaces(t *testing.T) {
	parser := NewParser()

	entry, ok := parser.Parse("07-04 01:02:03.456  1234  5678 W System UI: message")
	if !ok {
		t.Fatal("expected line to parse")
	}
	if entry.Tag != "System UI" {
		t.Fatalf("expected tag with spaces, got %q", entry.Tag)
	}
}

func TestParserRejectsNonThreadtime(t *testing.T) {
	parser := NewParser()

	if _, ok := parser.Parse("    at com.example.MainActivity.onCreate(MainActivity.kt:12)"); ok {
		t.Fatal("expected stacktrace line to be treated as multiline")
	}
}

func TestAndroidRuntimeFatalExceptionContinuation(t *testing.T) {
	parser := NewParser()
	buffer := NewRingBuffer(100)

	lines := []string{
		"07-04 01:02:03.456  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main",
		"07-04 01:02:03.457  1234  1234 E AndroidRuntime: Process: com.example.app, PID: 1234",
		"07-04 01:02:03.458  1234  1234 E AndroidRuntime: java.lang.RuntimeException: boom",
		"07-04 01:02:03.459  1234  1234 E AndroidRuntime: \tat com.example.MainActivity.onCreate(MainActivity.kt:12)",
	}

	for _, line := range lines {
		entry, ok := parser.Parse(line)
		if !ok {
			t.Fatalf("expected line to parse: %s", line)
		}
		if !buffer.AppendContinuation(entry) {
			buffer.Add(entry)
		}
	}

	batch := buffer.GetAfter(0, 10)
	if len(batch.Entries) != 1 {
		t.Fatalf("expected fatal exception to merge into one entry, got %+v", batch.Entries)
	}
	if len(batch.Entries[0].Multiline) != 3 {
		t.Fatalf("expected 3 multiline rows, got %+v", batch.Entries[0].Multiline)
	}
}

func TestUnparsedLineAppendsToPreviousEntry(t *testing.T) {
	parser := NewParser()
	buffer := NewRingBuffer(100)

	entry, ok := parser.Parse("07-04 01:02:03.456  1234  5678 E AppTag: failed")
	if !ok {
		t.Fatal("expected root entry to parse")
	}
	buffer.Add(entry)

	if _, ok := parser.Parse("    at com.example.MainActivity.onCreate(MainActivity.kt:12)"); ok {
		t.Fatal("expected stacktrace line not to parse as threadtime")
	}
	if !buffer.AppendMultiline("    at com.example.MainActivity.onCreate(MainActivity.kt:12)") {
		t.Fatal("expected raw stacktrace to append to previous entry")
	}

	batch := buffer.GetAfter(0, 10)
	if len(batch.Entries) != 1 || len(batch.Entries[0].Multiline) != 1 {
		t.Fatalf("unexpected batch: %+v", batch.Entries)
	}
}
