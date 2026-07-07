package main

import "testing"

func TestADBExecutableDialogFiltersWindows(t *testing.T) {
	filters := adbExecutableDialogFilters("windows")
	if len(filters) != 2 {
		t.Fatalf("expected Windows filters, got %+v", filters)
	}
	if filters[0].Pattern != "adb.exe" {
		t.Fatalf("expected adb.exe filter, got %+v", filters[0])
	}
}

func TestADBExecutableDialogFiltersNonWindows(t *testing.T) {
	for _, goos := range []string{"darwin", "linux"} {
		t.Run(goos, func(t *testing.T) {
			if filters := adbExecutableDialogFilters(goos); filters != nil {
				t.Fatalf("expected no filters for %s, got %+v", goos, filters)
			}
		})
	}
}
