package logcat

import (
	"strings"
	"testing"
)

func TestFormatEntriesTextIncludesMultiline(t *testing.T) {
	text := FormatEntriesText([]LogEntry{
		{
			Timestamp: "07-04 01:02:03.456",
			PID:       1234,
			TID:       5678,
			Level:     "E",
			Tag:       "AndroidRuntime",
			Message:   "FATAL EXCEPTION: main",
			Multiline: []string{
				"07-04 01:02:03.457  1234  1234 E AndroidRuntime: java.lang.RuntimeException: boom",
				"    at com.example.Main.onCreate(Main.kt:1)",
			},
		},
	})

	for _, want := range []string{"07-04 01:02:03.456", "1234", "5678", "E", "AndroidRuntime", "FATAL EXCEPTION", "RuntimeException", "Main.kt:1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("formatted export missing %q: %s", want, text)
		}
	}
}

func TestFormatEntriesJSONLRoundTrip(t *testing.T) {
	jsonl, err := FormatEntriesJSONL([]LogEntry{
		{
			Timestamp:   "07-04 01:02:03.456",
			PID:         1234,
			TID:         5678,
			Level:       "E",
			Tag:         "AndroidRuntime",
			Message:     "FATAL EXCEPTION: main",
			PackageName: "com.example.app",
			Raw:         "07-04 01:02:03.456  1234  5678 E AndroidRuntime: FATAL EXCEPTION: main",
			Multiline:   []string{"    at com.example.Main.onCreate(Main.kt:1)"},
		},
	})
	if err != nil {
		t.Fatalf("FormatEntriesJSONL returned error: %v", err)
	}

	entries, err := ParseJSONL(jsonl)
	if err != nil {
		t.Fatalf("ParseJSONL returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	if entries[0].PackageName != "com.example.app" || entries[0].Multiline[0] != "    at com.example.Main.onCreate(Main.kt:1)" {
		t.Fatalf("round-trip entry mismatch: %+v", entries[0])
	}
}
