package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAppConfig(t *testing.T) {
	config := DefaultAppConfig()
	if config.ActiveWorkspaceID == "" {
		t.Fatal("expected active workspace id")
	}
	if len(config.Workspaces) != 1 {
		t.Fatalf("expected one default workspace, got %+v", config.Workspaces)
	}
	workspace := config.Workspaces[0]
	if workspace.DefaultBuildTask != "assembleDebug" {
		t.Fatalf("unexpected build task: %s", workspace.DefaultBuildTask)
	}
	if workspace.MaxLogLines != 100000 {
		t.Fatalf("unexpected max log lines: %d", workspace.MaxLogLines)
	}
	if len(config.FilterPresets) < 6 {
		t.Fatalf("expected built-in presets, got %+v", config.FilterPresets)
	}
}

func TestSaveLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	config := DefaultAppConfig()
	config.Workspaces[0].WorkspaceName = "Demo"
	config.Workspaces[0].ProjectPath = filepath.Join("C:\\", "src", "demo")
	config.Workspaces[0].PackageName = "com.example.demo"
	config.Workspaces[0].SearchKeyword = "boom"
	config.ADBPath = filepath.Join("C:\\", "Android", "platform-tools", "adb.exe")

	if err := SaveConfig(path, config); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	active := ActiveWorkspace(loaded)
	if active.WorkspaceName != "Demo" || active.PackageName != "com.example.demo" || active.SearchKeyword != "boom" {
		t.Fatalf("unexpected loaded workspace: %+v", active)
	}
	if loaded.ADBPath == "" {
		t.Fatalf("expected adb path to be persisted: %+v", loaded)
	}
}

func TestLoadConfigCorruptJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil {
		t.Fatal("expected corrupt config error")
	}
}

func TestWorkspaceCRUD(t *testing.T) {
	config := DefaultAppConfig()
	workspace := DefaultWorkspaceConfig()
	workspace.ID = "project-a"
	workspace.WorkspaceName = "Project A"
	workspace.ProjectPath = "D:\\projects\\a"
	workspace.PackageName = "com.example.a"

	config = SaveWorkspace(config, workspace)
	if config.ActiveWorkspaceID != "project-a" {
		t.Fatalf("expected active workspace to be project-a: %+v", config)
	}
	if ActiveWorkspace(config).PackageName != "com.example.a" {
		t.Fatalf("unexpected active workspace: %+v", ActiveWorkspace(config))
	}

	workspace.PackageName = "com.example.changed"
	config = SaveWorkspace(config, workspace)
	if ActiveWorkspace(config).PackageName != "com.example.changed" {
		t.Fatalf("workspace update failed: %+v", ActiveWorkspace(config))
	}

	config = SetActiveWorkspace(config, "default")
	if config.ActiveWorkspaceID != "default" {
		t.Fatalf("set active workspace failed: %+v", config)
	}
	config = DeleteWorkspace(config, "project-a")
	if len(config.Workspaces) != 1 || config.Workspaces[0].ID != "default" {
		t.Fatalf("delete workspace failed: %+v", config.Workspaces)
	}
}

func TestFilterPresetCRUD(t *testing.T) {
	config := DefaultAppConfig()
	initial := len(config.FilterPresets)
	preset := FilterPreset{
		ID:             "my-errors",
		Name:           "My Errors",
		Level:          []string{"E"},
		PackageName:    "com.example",
		Keyword:        "crash",
		RegexEnabled:   true,
		Tags:           []string{"AndroidRuntime"},
		ExcludeKeyword: "noise",
	}

	config = SaveFilterPreset(config, preset)
	if len(config.FilterPresets) != initial+1 {
		t.Fatalf("expected preset to be added: %+v", config.FilterPresets)
	}

	preset.Name = "Renamed Errors"
	config = SaveFilterPreset(config, preset)
	found := findPreset(config, "my-errors")
	if found.Name != "Renamed Errors" || !found.RegexEnabled {
		t.Fatalf("preset update failed: %+v", found)
	}

	config = DeleteFilterPreset(config, "my-errors")
	if len(config.FilterPresets) != initial {
		t.Fatalf("expected preset to be deleted: %+v", config.FilterPresets)
	}
}

func TestBuiltInPresetsCannotBeDeleted(t *testing.T) {
	config := DefaultAppConfig()
	config = DeleteFilterPreset(config, "errors-only")
	if findPreset(config, "errors-only").ID == "" {
		t.Fatal("expected built-in preset to remain")
	}
}

func findPreset(config AppConfig, id string) FilterPreset {
	for _, preset := range config.FilterPresets {
		if preset.ID == id {
			return preset
		}
	}
	return FilterPreset{}
}
