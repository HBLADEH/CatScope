package logcat

import "testing"

func TestPIDMapperAppliesPackageName(t *testing.T) {
	mapper := NewPIDMapper()
	mapper.SetPackage("com.example.app")
	mapper.Update([]int{1234})

	entry := LogEntry{PID: 1234, Message: "hello"}
	mapper.Apply(&entry)

	if entry.PackageName != "com.example.app" {
		t.Fatalf("expected package attribution, got %+v", entry)
	}
}

func TestPIDMapperKeepsOldPIDMapping(t *testing.T) {
	mapper := NewPIDMapper()
	mapper.SetPackage("com.example.app")
	mapper.Update([]int{1111})
	mapper.Update([]int{2222})

	oldEntry := LogEntry{PID: 1111}
	newEntry := LogEntry{PID: 2222}
	mapper.Apply(&oldEntry)
	mapper.Apply(&newEntry)

	if oldEntry.PackageName != "com.example.app" || newEntry.PackageName != "com.example.app" {
		t.Fatalf("expected old and new pid attribution, got old=%+v new=%+v", oldEntry, newEntry)
	}

	state := mapper.State()
	if len(state.PIDs) != 1 || state.PIDs[0] != 2222 {
		t.Fatalf("unexpected current pids: %+v", state.PIDs)
	}
	if len(state.KnownPIDs) != 2 || state.KnownPIDs[0] != 1111 || state.KnownPIDs[1] != 2222 {
		t.Fatalf("unexpected known pids: %+v", state.KnownPIDs)
	}
}

func TestPIDMapperClearDisablesAttribution(t *testing.T) {
	mapper := NewPIDMapper()
	mapper.SetPackage("com.example.app")
	mapper.Update([]int{1234})
	mapper.Clear()

	entry := LogEntry{PID: 1234}
	mapper.Apply(&entry)

	if entry.PackageName != "" {
		t.Fatalf("expected no attribution after clear, got %+v", entry)
	}
}
