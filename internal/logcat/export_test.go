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
