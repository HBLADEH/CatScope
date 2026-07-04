package ai

import (
	"strings"
	"testing"

	"catscope/internal/adb"
	"catscope/internal/logcat"
)

func TestGenerateMarkdownJavaCrashZhCN(t *testing.T) {
	options := DefaultOptions()
	analysis := logcat.AnalysisResult{
		ID:              "java-1",
		Type:            logcat.AnalysisTypeJavaCrash,
		Severity:        logcat.SeverityFatal,
		Title:           "Java crash: java.lang.NullPointerException",
		Summary:         "com.example.app crashed on thread main.",
		PackageName:     "com.example.app",
		PID:             1234,
		TID:             1234,
		Timestamp:       "07-04 12:00:00.000",
		PrimaryTag:      "AndroidRuntime",
		PrimaryMessage:  "FATAL EXCEPTION: main",
		ExceptionType:   "java.lang.NullPointerException",
		ThreadName:      "main",
		KeyFrames:       []string{"at com.example.app.MainActivity.onCreate(MainActivity.kt:42)"},
		RelatedEntryIDs: []int64{2},
		RawText:         "raw crash text",
		Suggestions:     []string{"Check MainActivity.onCreate."},
	}
	logs := []logcat.LogEntry{
		{ID: 1, Timestamp: "07-04 11:59:59.999", PID: 1234, TID: 1234, Level: "D", Tag: "App", Message: "before"},
		{ID: 2, Timestamp: "07-04 12:00:00.000", PID: 1234, TID: 1234, Level: "E", Tag: "AndroidRuntime", Message: "FATAL EXCEPTION: main", Raw: "07-04 12:00:00.000  1234  1234 E AndroidRuntime: FATAL EXCEPTION: main", Multiline: []string{"    at com.example.app.MainActivity.onCreate(MainActivity.kt:42)"}},
		{ID: 3, Timestamp: "07-04 12:00:00.001", PID: 1234, TID: 1234, Level: "D", Tag: "App", Message: "after"},
	}

	markdown := GenerateMarkdown(ContextInput{
		Device:         &adb.AndroidDevice{Serial: "ABC", Model: "Pixel", AndroidVersion: "15", SDKVersion: "35", ABI: "arm64-v8a"},
		Analysis:       analysis,
		Logs:           logs,
		PIDState:       logcat.PackagePIDState{PackageName: "com.example.app", PIDs: []int{1234}, KnownPIDs: []int{1111, 1234}},
		CurrentPackage: "com.example.app",
		Options:        options,
	})

	assertContains(t, markdown, "# Android Logcat 问题分析请求")
	assertContains(t, markdown, "- Device: Pixel")
	assertContains(t, markdown, "- Package: com.example.app")
	assertContains(t, markdown, "- Exception: java.lang.NullPointerException")
	assertContains(t, markdown, "at com.example.app.MainActivity.onCreate")
	assertContains(t, markdown, "Check MainActivity.onCreate")
}

func TestGenerateMarkdownNativeCrashEnUS(t *testing.T) {
	options := DefaultOptions()
	options.Language = "en-US"
	analysis := logcat.AnalysisResult{
		ID:              "native-1",
		Type:            logcat.AnalysisTypeNativeCrash,
		Severity:        logcat.SeverityFatal,
		Title:           "Native crash: SIGSEGV in libfoo.so",
		Summary:         "Native crash indicators were found.",
		Signal:          "SIGSEGV",
		LibraryName:     "libfoo.so",
		KeyFrames:       []string{"#00 pc 123 /data/app/libfoo.so"},
		RelatedEntryIDs: []int64{1},
		Suggestions:     []string{"Use ndk-stack."},
	}
	logs := []logcat.LogEntry{{ID: 1, Raw: "07-04 12:00:00.000  2222  2222 F DEBUG: signal 11 (SIGSEGV)"}}

	markdown := GenerateMarkdown(ContextInput{Analysis: analysis, Logs: logs, Options: options})

	assertContains(t, markdown, "# Android Logcat Analysis Request")
	assertContains(t, markdown, "- Signal: SIGSEGV")
	assertContains(t, markdown, "- Library: libfoo.so")
	assertContains(t, markdown, "Use ndk-stack.")
}

func TestGenerateMarkdownWithoutDeviceInfo(t *testing.T) {
	options := DefaultOptions()
	analysis := logcat.AnalysisResult{ID: "a", Type: logcat.AnalysisTypeANR, Severity: logcat.SeverityFatal, Title: "ANR"}

	markdown := GenerateMarkdown(ContextInput{Analysis: analysis, Options: options})

	assertContains(t, markdown, "- Device: -")
	assertContains(t, markdown, "- Serial: -")
}

func TestGenerateMarkdownWithoutRelatedIDsUsesRawText(t *testing.T) {
	options := DefaultOptions()
	analysis := logcat.AnalysisResult{ID: "a", Type: logcat.AnalysisTypeJNIError, Severity: logcat.SeverityFatal, Title: "JNI", RawText: "JNI DETECTED ERROR IN APPLICATION"}

	markdown := GenerateMarkdown(ContextInput{Analysis: analysis, Options: options})

	assertContains(t, markdown, "JNI DETECTED ERROR IN APPLICATION")
}

func TestGenerateMarkdownMultilineLogs(t *testing.T) {
	options := DefaultOptions()
	analysis := logcat.AnalysisResult{ID: "a", Type: logcat.AnalysisTypeJavaCrash, Severity: logcat.SeverityFatal, Title: "Crash", RelatedEntryIDs: []int64{1}}
	logs := []logcat.LogEntry{{ID: 1, Raw: "root", Multiline: []string{"    at com.example.Main.main(Main.kt:1)"}}}

	markdown := GenerateMarkdown(ContextInput{Analysis: analysis, Logs: logs, Options: options})

	assertContains(t, markdown, "root")
	assertContains(t, markdown, "Main.kt:1")
}

func TestGenerateMarkdownContextDeduplicates(t *testing.T) {
	options := DefaultOptions()
	options.IncludeBeforeContextLines = 1
	options.IncludeAfterContextLines = 1
	analysis := logcat.AnalysisResult{ID: "a", Type: logcat.AnalysisTypeJavaCrash, Severity: logcat.SeverityFatal, Title: "Crash", RelatedEntryIDs: []int64{2, 3}}
	logs := []logcat.LogEntry{{ID: 1, Raw: "one"}, {ID: 2, Raw: "two"}, {ID: 3, Raw: "three"}, {ID: 4, Raw: "four"}}

	markdown := GenerateMarkdown(ContextInput{Analysis: analysis, Logs: logs, Options: options})

	if strings.Count(markdown, "two") != 2 {
		t.Fatalf("expected related/context output to include 'two' exactly twice, got:\n%s", markdown)
	}
	if strings.Count(markdown, "three") != 2 {
		t.Fatalf("expected related/context output to include 'three' exactly twice, got:\n%s", markdown)
	}
	assertContains(t, markdown, "one")
	assertContains(t, markdown, "four")
}

func TestGenerateMarkdownOptionExcludesRawTextAndSuggestions(t *testing.T) {
	options := DefaultOptions()
	options.IncludeRawText = false
	options.IncludeSuggestions = false
	analysis := logcat.AnalysisResult{ID: "a", Type: logcat.AnalysisTypeJavaCrash, Severity: logcat.SeverityFatal, Title: "Crash", RawText: "raw should be hidden", Suggestions: []string{"hidden suggestion"}}

	markdown := GenerateMarkdown(ContextInput{Analysis: analysis, Options: options})

	assertNotContains(t, markdown, "raw should be hidden")
	assertNotContains(t, markdown, "hidden suggestion")
	assertNotContains(t, markdown, "CatScope 初步建议")
}

func assertContains(t *testing.T, value string, want string) {
	t.Helper()
	if !strings.Contains(value, want) {
		t.Fatalf("expected markdown to contain %q:\n%s", want, value)
	}
}

func assertNotContains(t *testing.T, value string, unwanted string) {
	t.Helper()
	if strings.Contains(value, unwanted) {
		t.Fatalf("expected markdown not to contain %q:\n%s", unwanted, value)
	}
}
