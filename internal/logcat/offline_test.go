package logcat

import (
	"strings"
	"testing"
)

func TestParseOfflineTextThreadtimeFile(t *testing.T) {
	content := strings.Join([]string{
		"07-04 12:34:56.789  1234  5678 I MainActivity: hello",
		"07-04 12:34:56.790  1234  5678 W MainActivity: warning",
	}, "\n")

	entries, failed := ParseOfflineText(content)
	if failed != 0 {
		t.Fatalf("parse failed count = %d, want 0", failed)
	}
	if len(entries) != 2 {
		t.Fatalf("entry count = %d, want 2", len(entries))
	}
	if entries[0].Tag != "MainActivity" || entries[0].Message != "hello" || entries[0].Raw == "" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
}

func TestParseOfflineTextMergesStacktrace(t *testing.T) {
	content := strings.Join([]string{
		"07-04 12:34:56.789  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main",
		"07-04 12:34:56.790  1234  1234 E AndroidRuntime: java.lang.RuntimeException: boom",
		"07-04 12:34:56.791  1234  1234 E AndroidRuntime: \tat com.example.MainActivity.onCreate(MainActivity.kt:42)",
	}, "\n")

	entries, failed := ParseOfflineText(content)
	if failed != 0 {
		t.Fatalf("parse failed count = %d, want 0", failed)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	if len(entries[0].Multiline) != 2 {
		t.Fatalf("multiline count = %d, want 2: %+v", len(entries[0].Multiline), entries[0])
	}
	if !strings.Contains(strings.Join(entries[0].Multiline, "\n"), "MainActivity.kt:42") {
		t.Fatalf("stack frame was not preserved: %+v", entries[0].Multiline)
	}
}

func TestParseOfflineTextPreservesUnparsedLines(t *testing.T) {
	content := strings.Join([]string{
		"07-04 12:34:56.789  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main",
		"this line is not threadtime but belongs to the crash",
		"standalone raw line",
	}, "\n")

	entries, failed := ParseOfflineText(content)
	if failed != 2 {
		t.Fatalf("parse failed count = %d, want 2", failed)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	if got := strings.Join(entries[0].Multiline, "\n"); !strings.Contains(got, "standalone raw line") {
		t.Fatalf("unparsed raw lines were not preserved: %q", got)
	}
}

func TestParseJSONL(t *testing.T) {
	jsonl := `{"id":99,"timestamp":"07-04 12:34:56.789","pid":1234,"tid":1234,"level":"E","tag":"AndroidRuntime","message":"boom","raw":"raw line","multiline":["stack"]}` + "\n"

	entries, err := ParseJSONL(jsonl)
	if err != nil {
		t.Fatalf("ParseJSONL returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entry count = %d, want 1", len(entries))
	}
	if entries[0].ID != 1 {
		t.Fatalf("imported ID = %d, want reassigned 1", entries[0].ID)
	}
	if entries[0].PackageName != "" || entries[0].Multiline[0] != "stack" {
		t.Fatalf("unexpected imported entry: %+v", entries[0])
	}
}

func TestParseOfflineContentFallsBackWhenJSONLInvalid(t *testing.T) {
	entries, failed, err := ParseOfflineContent("not json\nstill raw", ".jsonl")
	if err != nil {
		t.Fatalf("ParseOfflineContent returned error: %v", err)
	}
	if failed != 2 {
		t.Fatalf("parse failed count = %d, want 2", failed)
	}
	if len(entries) != 1 || !strings.Contains(strings.Join(entries[0].Multiline, "\n"), "still raw") {
		t.Fatalf("expected fallback raw preservation, got %+v", entries)
	}
}

func TestOfflineAnalyzerDetectsJavaCrash(t *testing.T) {
	content := strings.Join([]string{
		"07-04 12:34:56.789  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main",
		"07-04 12:34:56.790  1234  1234 E AndroidRuntime: Process: com.example.app, PID: 1234",
		"07-04 12:34:56.791  1234  1234 E AndroidRuntime: java.lang.NullPointerException: boom",
		"07-04 12:34:56.792  1234  1234 E AndroidRuntime: \tat com.example.MainActivity.onCreate(MainActivity.kt:42)",
	}, "\n")

	entries, _ := ParseOfflineText(content)
	results := AnalyzeEntries(entries)
	if len(results) != 1 {
		t.Fatalf("analysis result count = %d, want 1: %+v", len(results), results)
	}
	if results[0].Type != AnalysisTypeJavaCrash || results[0].PackageName != "com.example.app" {
		t.Fatalf("unexpected analysis result: %+v", results[0])
	}
}
