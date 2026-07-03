package logcat

import "testing"

func TestAnalyzeJavaCrash(t *testing.T) {
	entries := []LogEntry{
		{
			ID:        1,
			Timestamp: "07-04 12:00:00.000",
			PID:       1234,
			TID:       1234,
			Level:     "E",
			Tag:       "AndroidRuntime",
			Message:   "FATAL EXCEPTION: main",
			Raw:       "07-04 12:00:00.000  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main",
			Multiline: []string{
				"07-04 12:00:00.001  1234  1234 E AndroidRuntime: Process: com.example.app, PID: 1234",
				"07-04 12:00:00.002  1234  1234 E AndroidRuntime: java.lang.NullPointerException: boom",
				"07-04 12:00:00.003  1234  1234 E AndroidRuntime: \tat com.example.app.MainActivity.onCreate(MainActivity.kt:42)",
				"07-04 12:00:00.004  1234  1234 E AndroidRuntime: Caused by: java.lang.IllegalStateException: bad state",
			},
		},
	}

	results := AnalyzeEntries(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %+v", results)
	}
	result := results[0]
	if result.Type != AnalysisTypeJavaCrash || result.Severity != SeverityFatal {
		t.Fatalf("unexpected java crash result: %+v", result)
	}
	if result.ExceptionType != "java.lang.NullPointerException" {
		t.Fatalf("unexpected exception type: %+v", result.ExceptionType)
	}
	if result.ThreadName != "main" {
		t.Fatalf("unexpected thread name: %+v", result.ThreadName)
	}
	if result.PackageName != "com.example.app" {
		t.Fatalf("unexpected package: %+v", result.PackageName)
	}
	if len(result.KeyFrames) == 0 || result.KeyFrames[0] != "at com.example.app.MainActivity.onCreate(MainActivity.kt:42)" {
		t.Fatalf("unexpected key frames: %+v", result.KeyFrames)
	}
}

func TestAnalyzeNativeCrash(t *testing.T) {
	entries := []LogEntry{
		{
			ID:      10,
			PID:     2222,
			TID:     2222,
			Level:   "F",
			Tag:     "DEBUG",
			Message: "signal 11 (SIGSEGV), code 1 (SEGV_MAPERR), fault addr 0x0",
			Raw:     "07-04 12:00:00.000  2222  2222 F DEBUG: signal 11 (SIGSEGV), code 1 (SEGV_MAPERR), fault addr 0x0",
		},
		{
			ID:      11,
			PID:     2222,
			TID:     2222,
			Level:   "F",
			Tag:     "DEBUG",
			Message: "backtrace:",
			Raw:     "07-04 12:00:00.001  2222  2222 F DEBUG: backtrace:",
		},
		{
			ID:      12,
			PID:     2222,
			TID:     2222,
			Level:   "F",
			Tag:     "DEBUG",
			Message: "#00 pc 0000000000012345  /data/app/lib/arm64/libcatscope.so (CrashHere+12)",
			Raw:     "07-04 12:00:00.002  2222  2222 F DEBUG: #00 pc 0000000000012345  /data/app/lib/arm64/libcatscope.so (CrashHere+12)",
		},
	}

	results := AnalyzeEntries(entries)
	if len(results) != 1 {
		t.Fatalf("expected deduped native crash result, got %+v", results)
	}
	result := results[0]
	if result.Type != AnalysisTypeNativeCrash {
		t.Fatalf("unexpected result type: %+v", result)
	}
	if result.Signal != "SIGSEGV" && result.Signal != "SIGNAL 11" {
		t.Fatalf("unexpected signal: %+v", result.Signal)
	}
	if result.LibraryName != "libcatscope.so" {
		t.Fatalf("unexpected library: %+v", result.LibraryName)
	}
	if len(result.KeyFrames) == 0 {
		t.Fatalf("expected native frames: %+v", result)
	}
}

func TestAnalyzeANR(t *testing.T) {
	entries := []LogEntry{
		{
			ID:      20,
			PID:     3333,
			Level:   "E",
			Tag:     "ActivityManager",
			Message: "ANR in com.example.app",
			Raw:     "07-04 12:00:00.000  1000  1000 E ActivityManager: ANR in com.example.app",
			Multiline: []string{
				"07-04 12:00:00.001  1000  1000 E ActivityManager: Reason: Input dispatching timed out",
			},
		},
	}

	results := AnalyzeEntries(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 ANR result, got %+v", results)
	}
	if results[0].Type != AnalysisTypeANR || results[0].PackageName != "com.example.app" {
		t.Fatalf("unexpected ANR result: %+v", results[0])
	}
	if results[0].Reason == "" {
		t.Fatalf("expected ANR reason: %+v", results[0])
	}
}

func TestAnalyzeJNIError(t *testing.T) {
	entries := []LogEntry{
		{
			ID:      30,
			PID:     4444,
			Level:   "F",
			Tag:     "AndroidRuntime",
			Message: "JNI DETECTED ERROR IN APPLICATION: use of deleted local reference",
			Raw:     "07-04 12:00:00.000  4444  4444 F AndroidRuntime: JNI DETECTED ERROR IN APPLICATION: use of deleted local reference",
			Multiline: []string{
				"07-04 12:00:00.001  4444  4444 F AndroidRuntime: java_vm_ext.cc: CheckJNI failed",
			},
		},
	}

	results := AnalyzeEntries(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 JNI result, got %+v", results)
	}
	if results[0].Type != AnalysisTypeJNIError {
		t.Fatalf("unexpected JNI result: %+v", results[0])
	}
	if results[0].Reason == "" {
		t.Fatalf("expected JNI reason: %+v", results[0])
	}
}

func TestAnalyzeEntriesDeduplicatesRepeatedCrash(t *testing.T) {
	entry := LogEntry{
		ID:      40,
		PID:     5555,
		Level:   "E",
		Tag:     "AndroidRuntime",
		Message: "FATAL EXCEPTION: main",
		Raw:     "07-04 12:00:00.000  5555  5555 E AndroidRuntime: FATAL EXCEPTION: main",
		Multiline: []string{
			"07-04 12:00:00.001  5555  5555 E AndroidRuntime: java.lang.IllegalStateException: duplicate",
		},
	}

	results := AnalyzeEntries([]LogEntry{entry, entry})
	if len(results) != 1 {
		t.Fatalf("expected deduped result, got %+v", results)
	}
}
