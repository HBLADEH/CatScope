package build

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildCommandWindowsPrefersGradlewBat(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "settings.gradle"), "")
	writeFile(t, filepath.Join(dir, "gradlew"), "")
	writeFile(t, filepath.Join(dir, "gradlew.bat"), "")

	executable, args, workdir, err := BuildCommand(dir, "", "windows")
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	if executable != filepath.Join(dir, "gradlew.bat") {
		t.Fatalf("expected gradlew.bat, got %s", executable)
	}
	if workdir != dir {
		t.Fatalf("unexpected workdir: %s", workdir)
	}
	if len(args) != 1 || args[0] != "assembleDebug" {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestBuildCommandUnixPrefersGradlew(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "settings.gradle.kts"), "")
	writeFile(t, filepath.Join(dir, "gradlew"), "")
	writeFile(t, filepath.Join(dir, "gradlew.bat"), "")

	executable, args, _, err := BuildCommand(dir, ":app:assembleDebug", "linux")
	if err != nil {
		t.Fatalf("BuildCommand failed: %v", err)
	}
	if executable != filepath.Join(dir, "gradlew") {
		t.Fatalf("expected gradlew, got %s", executable)
	}
	if args[0] != ":app:assembleDebug" {
		t.Fatalf("unexpected args: %+v", args)
	}
}

func TestBuildCommandRequiresSettingsGradle(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "gradlew.bat"), "")

	_, _, _, err := BuildCommand(dir, "", "windows")
	if err == nil {
		t.Fatal("expected missing settings.gradle error")
	}
}

func TestFindLatestAPKPrefersNewestDebugAPK(t *testing.T) {
	dir := t.TempDir()
	oldDebug := filepath.Join(dir, "app", "build", "outputs", "apk", "debug", "app-debug.apk")
	newRelease := filepath.Join(dir, "app", "build", "outputs", "apk", "release", "app-release.apk")
	newDebug := filepath.Join(dir, "feature", "build", "outputs", "apk", "debug", "feature-debug.apk")
	writeFile(t, oldDebug, "old")
	writeFile(t, newRelease, "release")
	writeFile(t, newDebug, "debug")

	now := time.Now()
	setModTime(t, oldDebug, now.Add(-3*time.Hour))
	setModTime(t, newRelease, now.Add(-1*time.Hour))
	setModTime(t, newDebug, now.Add(-2*time.Hour))

	apk, err := FindLatestAPK(dir)
	if err != nil {
		t.Fatalf("FindLatestAPK failed: %v", err)
	}
	if apk.APKPath != newDebug {
		t.Fatalf("expected newest debug APK, got %+v", apk)
	}
	if apk.FileName != "feature-debug.apk" || apk.Size == 0 || apk.ModifiedTime == "" {
		t.Fatalf("incomplete APK info: %+v", apk)
	}
}

func TestFindLatestAPKReturnsErrorWhenMissing(t *testing.T) {
	_, err := FindLatestAPK(t.TempDir())
	if err == nil {
		t.Fatal("expected missing APK error")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func setModTime(t *testing.T, path string, at time.Time) {
	t.Helper()
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatal(err)
	}
}
