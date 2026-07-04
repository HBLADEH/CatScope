package adb

import (
	"reflect"
	"testing"
)

func TestBuildLaunchArgsUsesMonkeyPackage(t *testing.T) {
	args := BuildLaunchArgs("device-1", "com.example.app")
	expected := []string{"-s", "device-1", "shell", "monkey", "-p", "com.example.app", "1"}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected launch args:\nwant %+v\n got %+v", expected, args)
	}
}

func TestLooksLikeMonkeyFailure(t *testing.T) {
	if !looksLikeMonkeyFailure("** No activities found to run, monkey aborted.") {
		t.Fatal("expected monkey failure")
	}
	if looksLikeMonkeyFailure("Events injected: 1") {
		t.Fatal("did not expect monkey failure")
	}
}
