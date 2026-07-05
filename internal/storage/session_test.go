package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"catscope/internal/ai"
	"catscope/internal/logcat"
)

func sampleSession() Session {
	return Session{
		SessionID:      "session-test",
		Name:           "Crash Room",
		SourceMode:     logcat.LogSourceOffline,
		SourceName:     "crash.log",
		SourcePath:     "C:/tmp/crash.log",
		WorkspaceID:    "workspace-test",
		WorkspaceName:  "App Workspace",
		ProjectPath:    "C:/src/app",
		PackageName:    "com.example.app",
		SelectedDevice: "device-1",
		KnownPIDs:      []int{1234},
		Filters: SessionFilters{
			Level:          []string{"E", "F"},
			PackageName:    "com.example.app",
			Keyword:        "FATAL",
			RegexEnabled:   true,
			Tags:           []string{"AndroidRuntime"},
			ExcludeKeyword: "noise",
		},
		AIContextOptions: ai.DefaultOptions(),
		LogEntries: []logcat.LogEntry{
			{
				ID:          42,
				Timestamp:   "07-04 12:34:56.789",
				PID:         1234,
				TID:         1234,
				Level:       "E",
				Tag:         "AndroidRuntime",
				Message:     "FATAL EXCEPTION: main",
				PackageName: "com.example.app",
				Raw:         "07-04 12:34:56.789  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main",
				Multiline:   []string{"java.lang.RuntimeException: boom", "\tat com.example.MainActivity.onCreate(MainActivity.kt:42)"},
			},
		},
		AnalysisResults: []logcat.AnalysisResult{
			{
				ID:              "java_crash-test",
				Type:            logcat.AnalysisTypeJavaCrash,
				Severity:        logcat.SeverityFatal,
				Title:           "Java crash",
				RelatedEntryIDs: []int64{42},
				RawText:         "FATAL EXCEPTION: main\njava.lang.RuntimeException: boom",
			},
		},
		Notes: "repro after login",
	}
}

func TestSaveAndOpenSessionPreservesContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "debug.catscope-session")
	saved, err := SaveSession(path, sampleSession())
	if err != nil {
		t.Fatalf("SaveSession returned error: %v", err)
	}
	if saved.Version != SessionVersion || strings.TrimSpace(saved.CreatedAt) == "" || strings.TrimSpace(saved.UpdatedAt) == "" {
		t.Fatalf("session timestamps/version were not normalized: %+v", saved)
	}

	opened, err := OpenSession(path)
	if err != nil {
		t.Fatalf("OpenSession returned error: %v", err)
	}
	if opened.SessionID != "session-test" || opened.Name != "Crash Room" {
		t.Fatalf("unexpected metadata: %+v", opened)
	}
	if opened.Filters.Keyword != "FATAL" || !opened.Filters.RegexEnabled || opened.Filters.ExcludeKeyword != "noise" {
		t.Fatalf("filters were not preserved: %+v", opened.Filters)
	}
	if len(opened.LogEntries) != 1 || opened.LogEntries[0].ID != 42 || opened.LogEntries[0].PackageName != "com.example.app" {
		t.Fatalf("log entry was not preserved: %+v", opened.LogEntries)
	}
	if got := strings.Join(opened.LogEntries[0].Multiline, "\n"); !strings.Contains(got, "MainActivity.kt:42") {
		t.Fatalf("multiline was not preserved: %q", got)
	}
	if len(opened.AnalysisResults) != 1 || opened.AnalysisResults[0].RelatedEntryIDs[0] != 42 {
		t.Fatalf("analysis result was not preserved: %+v", opened.AnalysisResults)
	}
}

func TestSaveSessionRejectsEmptyLogs(t *testing.T) {
	session := sampleSession()
	session.LogEntries = nil
	if _, err := SaveSession(filepath.Join(t.TempDir(), "empty.catscope-session"), session); err == nil {
		t.Fatal("SaveSession returned nil error for empty logs")
	}
}

func TestOpenSessionRejectsInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.catscope-session")
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatalf("write invalid fixture: %v", err)
	}
	if _, err := OpenSession(path); err == nil {
		t.Fatal("OpenSession returned nil error for invalid JSON")
	}
}
