package adb

import "testing"

func TestParsePackageList(t *testing.T) {
	output := "package:com.example.one\r\npackage:com.android.settings\npackage:/data/app/base.apk=com.example.two\n\n"

	packages := ParsePackageList(output)
	if len(packages) != 3 {
		t.Fatalf("expected 3 packages, got %+v", packages)
	}
	if packages[0].PackageName != "com.example.one" {
		t.Fatalf("unexpected first package: %+v", packages[0])
	}
	if packages[2].PackageName != "com.example.two" {
		t.Fatalf("unexpected apk-path package parse: %+v", packages[2])
	}
}

func TestParsePackageListDeduplicates(t *testing.T) {
	packages := ParsePackageList("package:com.example.app\npackage:com.example.app\n")
	if len(packages) != 1 {
		t.Fatalf("expected duplicate packages to collapse, got %+v", packages)
	}
}

func TestParsePidOf(t *testing.T) {
	pids := ParsePidOf("1234 5678\r\nbad 1234 0")
	if len(pids) != 2 || pids[0] != 1234 || pids[1] != 5678 {
		t.Fatalf("unexpected pids: %+v", pids)
	}
}

func TestParsePidOfEmptyWhenAppNotRunning(t *testing.T) {
	pids := ParsePidOf("\n")
	if len(pids) != 0 {
		t.Fatalf("expected empty pid list, got %+v", pids)
	}
}
