package logcat

import (
	"strings"
	"testing"
)

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

func TestAnalyzeInstallOutputKnownCodes(t *testing.T) {
	tests := []struct {
		code     string
		severity AnalysisSeverity
		title    string
	}{
		{"INSTALL_FAILED_VERSION_DOWNGRADE", SeverityError, "version downgrade"},
		{"INSTALL_FAILED_UPDATE_INCOMPATIBLE", SeverityError, "update incompatible"},
		{"INSTALL_FAILED_NO_MATCHING_ABIS", SeverityFatal, "no matching ABIs"},
		{"INSTALL_FAILED_INVALID_APK", SeverityError, "invalid APK"},
		{"INSTALL_PARSE_FAILED_NO_CERTIFICATES", SeverityFatal, "no certificates"},
		{"INSTALL_PARSE_FAILED_MANIFEST_MALFORMED", SeverityError, "malformed manifest"},
		{"INSTALL_FAILED_INSUFFICIENT_STORAGE", SeverityWarning, "insufficient storage"},
		{"INSTALL_FAILED_ALREADY_EXISTS", SeverityWarning, "already exists"},
		{"INSTALL_FAILED_MISSING_SHARED_LIBRARY", SeverityFatal, "missing shared library"},
		{"INSTALL_FAILED_CPU_ABI_INCOMPATIBLE", SeverityFatal, "CPU ABI incompatible"},
		{"INSTALL_FAILED_TEST_ONLY", SeverityWarning, "test only"},
		{"DELETE_FAILED_INTERNAL_ERROR", SeverityError, "internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			output := "Performing Streamed Install\nFailure [" + tt.code + ": package install failed]"
			results := AnalyzeInstallOutput(output)
			if len(results) != 1 {
				t.Fatalf("expected one install result, got %+v", results)
			}
			result := results[0]
			if result.Type != AnalysisTypeInstallError {
				t.Fatalf("unexpected type: %+v", result)
			}
			if result.Severity != tt.severity {
				t.Fatalf("unexpected severity for %s: %+v", tt.code, result.Severity)
			}
			if !strings.Contains(result.Title, tt.title) {
				t.Fatalf("expected title to contain %q, got %q", tt.title, result.Title)
			}
			if !strings.Contains(result.RawText, tt.code) || !strings.Contains(result.PrimaryMessage, tt.code) {
				t.Fatalf("expected raw and primary message to preserve code: %+v", result)
			}
			if !containsPrefix(result.Suggestions, "中文:") || !containsPrefix(result.Suggestions, "English:") {
				t.Fatalf("expected bilingual suggestions: %+v", result.Suggestions)
			}
			if result.Reason == "" {
				t.Fatalf("expected reason: %+v", result)
			}
		})
	}
}

func TestAnalyzeTextInstallOutputAlias(t *testing.T) {
	results := AnalyzeText("adb: failed to install app.apk: Failure [INSTALL_FAILED_TEST_ONLY]")
	if len(results) != 1 {
		t.Fatalf("expected one result from AnalyzeText, got %+v", results)
	}
	if results[0].Type != AnalysisTypeInstallError || results[0].Severity != SeverityWarning {
		t.Fatalf("unexpected install result: %+v", results[0])
	}
	if !strings.Contains(strings.Join(results[0].Suggestions, "\n"), "adb install -t") {
		t.Fatalf("expected testOnly suggestion: %+v", results[0].Suggestions)
	}
}

func TestAnalyzeInstallOutputGenericADBFailure(t *testing.T) {
	results := AnalyzeInstallOutput("adb: failed to install C:\\tmp\\app.apk: device offline")
	if len(results) != 1 {
		t.Fatalf("expected generic adb install failure, got %+v", results)
	}
	result := results[0]
	if result.Type != AnalysisTypeInstallError {
		t.Fatalf("unexpected type: %+v", result)
	}
	if result.Title != "Install failed" {
		t.Fatalf("unexpected title: %+v", result.Title)
	}
	if !strings.Contains(result.Reason, "没有明确") {
		t.Fatalf("expected bilingual generic reason: %+v", result.Reason)
	}
}

func TestAnalyzeEntriesDetectsInstallError(t *testing.T) {
	entries := []LogEntry{
		{
			ID:      60,
			Level:   "E",
			Tag:     "PackageInstaller",
			Message: "Failure [INSTALL_FAILED_UPDATE_INCOMPATIBLE: Existing package has different signature]",
			Raw:     "07-05 09:00:00.000  1000  1000 E PackageInstaller: Failure [INSTALL_FAILED_UPDATE_INCOMPATIBLE: Existing package has different signature]",
		},
	}

	results := AnalyzeEntries(entries)
	if len(results) != 1 {
		t.Fatalf("expected install analyzer result, got %+v", results)
	}
	if results[0].Type != AnalysisTypeInstallError {
		t.Fatalf("unexpected result type: %+v", results[0])
	}
	if len(results[0].RelatedEntryIDs) != 1 || results[0].RelatedEntryIDs[0] != 60 {
		t.Fatalf("expected related log id: %+v", results[0].RelatedEntryIDs)
	}
}

func containsPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
