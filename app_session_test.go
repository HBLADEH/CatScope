package main

import (
	"path/filepath"
	"strings"
	"testing"

	"catscope/internal/ai"
	"catscope/internal/logcat"
	"catscope/internal/storage"
)

func appSessionEntry() logcat.LogEntry {
	return logcat.LogEntry{
		ID:          10,
		Timestamp:   "07-04 12:34:56.789",
		PID:         1234,
		TID:         1234,
		Level:       "E",
		Tag:         "AndroidRuntime",
		Message:     "FATAL EXCEPTION: main",
		PackageName: "com.example.app",
		Raw:         "07-04 12:34:56.789  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main",
		Multiline: []string{
			"07-04 12:34:56.790  1234  1234 E AndroidRuntime: Process: com.example.app, PID: 1234",
			"07-04 12:34:56.791  1234  1234 E AndroidRuntime: java.lang.RuntimeException: boom",
			"07-04 12:34:56.792  1234  1234 E AndroidRuntime: \tat com.example.MainActivity.onCreate(MainActivity.kt:42)",
		},
	}
}

func TestAppSaveOpenSessionAndContinueAnalysis(t *testing.T) {
	app := NewApp()
	entry := appSessionEntry()
	app.logStore.Replace([]logcat.LogEntry{entry})
	results := app.AnalyzeLogs(app.logStore.Snapshot())
	if len(results) == 0 {
		t.Fatal("expected fixture to produce analysis result")
	}

	path := filepath.Join(t.TempDir(), "capture.catscope-session")
	summary, err := app.SaveSession(path, SessionSaveOptions{
		Name: "Saved Crash",
		Filters: storage.SessionFilters{
			Level:          []string{"E", "F"},
			PackageName:    "com.example.app",
			Keyword:        "RuntimeException",
			RegexEnabled:   true,
			Tags:           []string{"AndroidRuntime"},
			ExcludeKeyword: "ignore",
		},
		AIContextOptions: ai.DefaultOptions(),
		Notes:            "session test",
	})
	if err != nil {
		t.Fatalf("SaveSession returned error: %v", err)
	}
	if summary.LogCount != 1 || summary.AnalysisCount != len(results) {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	opened, err := app.OpenSession(path)
	if err != nil {
		t.Fatalf("OpenSession returned error: %v", err)
	}
	if opened.Summary.Name != "Saved Crash" || opened.Session.Filters.ExcludeKeyword != "ignore" {
		t.Fatalf("session metadata was not restored: %+v", opened)
	}
	if opened.Entries[0].ID != entry.ID || len(opened.Entries[0].Multiline) != len(entry.Multiline) {
		t.Fatalf("entry was not restored: %+v", opened.Entries[0])
	}
	status := app.GetLogStatus()
	if status.Source != logcat.LogSourceSession || status.SessionName != "Saved Crash" {
		t.Fatalf("app did not enter session mode: %+v", status)
	}

	nextResults := app.AnalyzeLogs(app.logStore.Snapshot())
	if len(nextResults) == 0 {
		t.Fatal("analyzer did not work after opening session")
	}
	context, err := app.GenerateAIContext(opened.AnalysisResults[0].ID, ai.DefaultOptions())
	if err != nil {
		t.Fatalf("GenerateAIContext returned error after session open: %v", err)
	}
	if !strings.Contains(context, "RuntimeException") {
		t.Fatalf("AI context does not include restored crash: %s", context)
	}
}

func TestAppSaveSessionRejectsEmptyBuffer(t *testing.T) {
	app := NewApp()
	if _, err := app.SaveSession(filepath.Join(t.TempDir(), "empty.catscope-session"), SessionSaveOptions{}); err == nil {
		t.Fatal("SaveSession returned nil error for empty buffer")
	}
}
