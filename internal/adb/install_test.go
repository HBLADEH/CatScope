package adb

import (
	"reflect"
	"strings"
	"testing"

	"catscope/internal/logcat"
)

func TestBuildInstallArgsIncludesOptions(t *testing.T) {
	args := BuildInstallArgs("device-1", "C:\\tmp\\app.apk", InstallOptions{
		AllowDowngrade:   true,
		GrantPermissions: true,
		AllowTestOnly:    true,
	})
	expected := []string{"-s", "device-1", "install", "-r", "-d", "-g", "-t", "C:\\tmp\\app.apk"}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected install args:\nwant %+v\n got %+v", expected, args)
	}
}

func TestBuildInstallArgsDefaultReplaceOnly(t *testing.T) {
	args := BuildInstallArgs("serial", "app.apk", InstallOptions{})
	expected := []string{"-s", "serial", "install", "-r", "app.apk"}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected default install args: %+v", args)
	}
}

func TestAnalyzeInstallOutputFromADBFailure(t *testing.T) {
	results := logcat.AnalyzeInstallOutput("adb: failed to install app.apk: Failure [INSTALL_FAILED_TEST_ONLY]")
	if len(results) != 1 {
		t.Fatalf("expected one result, got %+v", results)
	}
	result := results[0]
	if result.Type != logcat.AnalysisTypeInstallError {
		t.Fatalf("unexpected result: %+v", result)
	}
	if !strings.Contains(strings.Join(result.Suggestions, "\n"), "adb install -t") {
		t.Fatalf("expected test-only suggestion: %+v", result.Suggestions)
	}
}
