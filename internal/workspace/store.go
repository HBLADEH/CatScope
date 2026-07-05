package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"catscope/internal/adb"
	"catscope/internal/ai"
)

type AppConfig struct {
	ActiveWorkspaceID string            `json:"activeWorkspaceId"`
	ADBPath           string            `json:"adbPath,omitempty"`
	Workspaces        []WorkspaceConfig `json:"workspaces"`
	FilterPresets     []FilterPreset    `json:"filterPresets"`
}

type WorkspaceConfig struct {
	ID                   string              `json:"id"`
	WorkspaceName        string              `json:"workspaceName"`
	ProjectPath          string              `json:"projectPath"`
	PackageName          string              `json:"packageName"`
	LastAPKPath          string              `json:"lastApkPath"`
	DefaultBuildTask     string              `json:"defaultBuildTask"`
	InstallOptions       adb.InstallOptions  `json:"installOptions"`
	SelectedDeviceSerial string              `json:"selectedDeviceSerial"`
	SelectedLogLevel     []string            `json:"selectedLogLevel"`
	SearchKeyword        string              `json:"searchKeyword"`
	SelectedPackageMode  string              `json:"selectedPackageMode"`
	MaxLogLines          int                 `json:"maxLogLines"`
	AutoStartLogcat      bool                `json:"autoStartLogcat"`
	AutoClearOnLaunch    bool                `json:"autoClearOnLaunch"`
	AIContextOptions     ai.AIContextOptions `json:"aiContextOptions"`
}

type FilterPreset struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Level          []string `json:"level"`
	PackageName    string   `json:"packageName"`
	Keyword        string   `json:"keyword"`
	RegexEnabled   bool     `json:"regexEnabled"`
	Tags           []string `json:"tags"`
	ExcludeKeyword string   `json:"excludeKeyword"`
	BuiltIn        bool     `json:"builtIn,omitempty"`
}

func DefaultAppConfig() AppConfig {
	workspace := DefaultWorkspaceConfig()
	return AppConfig{
		ActiveWorkspaceID: workspace.ID,
		Workspaces:        []WorkspaceConfig{workspace},
		FilterPresets:     DefaultFilterPresets(),
	}
}

func DefaultWorkspaceConfig() WorkspaceConfig {
	return NormalizeWorkspace(WorkspaceConfig{
		ID:                  "default",
		WorkspaceName:       "Default Workspace",
		DefaultBuildTask:    "assembleDebug",
		SelectedLogLevel:    []string{"V", "D", "I", "W", "E", "F"},
		SelectedPackageMode: "thirdParty",
		MaxLogLines:         100000,
		InstallOptions: adb.InstallOptions{
			AllowDowngrade:   false,
			GrantPermissions: true,
			AllowTestOnly:    false,
		},
		AIContextOptions: ai.DefaultOptions(),
	})
}

func DefaultFilterPresets() []FilterPreset {
	return []FilterPreset{
		{ID: "all-logs", Name: "All Logs", Level: []string{"V", "D", "I", "W", "E", "F"}, BuiltIn: true},
		{ID: "errors-only", Name: "Errors Only", Level: []string{"E", "F"}, BuiltIn: true},
		{ID: "android-runtime", Name: "AndroidRuntime", Level: []string{"E", "F"}, Tags: []string{"AndroidRuntime"}, BuiltIn: true},
		{ID: "native-crash", Name: "Native Crash", Level: []string{"E", "F"}, Tags: []string{"DEBUG", "libc"}, Keyword: "backtrace", BuiltIn: true},
		{ID: "install-errors", Name: "Install Errors", Level: []string{"V", "D", "I", "W", "E", "F"}, Keyword: "INSTALL_FAILED", BuiltIn: true},
		{ID: "current-package", Name: "Current Package", Level: []string{"V", "D", "I", "W", "E", "F"}, PackageName: "$current", BuiltIn: true},
	}
}

func DefaultConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(dir) == "" {
		return ""
	}
	return filepath.Join(dir, "CatScope", "config.json")
}

func LoadConfig(path string) (AppConfig, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return DefaultAppConfig(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultAppConfig(), nil
		}
		return DefaultAppConfig(), err
	}
	if len(data) == 0 {
		return DefaultAppConfig(), nil
	}
	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultAppConfig(), fmt.Errorf("load workspace config failed: %w", err)
	}
	return NormalizeConfig(config), nil
}

func SaveConfig(path string, config AppConfig) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	config = NormalizeConfig(config)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func NormalizeConfig(config AppConfig) AppConfig {
	config.ADBPath = cleanPath(config.ADBPath)
	if len(config.Workspaces) == 0 {
		config.Workspaces = []WorkspaceConfig{DefaultWorkspaceConfig()}
	}
	seenWorkspaces := map[string]bool{}
	for index := range config.Workspaces {
		config.Workspaces[index] = NormalizeWorkspace(config.Workspaces[index])
		if seenWorkspaces[config.Workspaces[index].ID] {
			config.Workspaces[index].ID = newID("workspace")
		}
		seenWorkspaces[config.Workspaces[index].ID] = true
	}
	if strings.TrimSpace(config.ActiveWorkspaceID) == "" || !hasWorkspace(config.Workspaces, config.ActiveWorkspaceID) {
		config.ActiveWorkspaceID = config.Workspaces[0].ID
	}
	config.FilterPresets = MergeDefaultPresets(config.FilterPresets)
	return config
}

func NormalizeWorkspace(workspace WorkspaceConfig) WorkspaceConfig {
	workspace.ID = strings.TrimSpace(workspace.ID)
	if workspace.ID == "" {
		workspace.ID = newID("workspace")
	}
	workspace.WorkspaceName = strings.TrimSpace(workspace.WorkspaceName)
	if workspace.WorkspaceName == "" {
		workspace.WorkspaceName = "Untitled Workspace"
	}
	workspace.ProjectPath = cleanPath(workspace.ProjectPath)
	workspace.PackageName = strings.TrimSpace(workspace.PackageName)
	workspace.LastAPKPath = cleanPath(workspace.LastAPKPath)
	workspace.DefaultBuildTask = strings.TrimSpace(workspace.DefaultBuildTask)
	if workspace.DefaultBuildTask == "" {
		workspace.DefaultBuildTask = "assembleDebug"
	}
	if len(workspace.SelectedLogLevel) == 0 {
		workspace.SelectedLogLevel = []string{"V", "D", "I", "W", "E", "F"}
	}
	if strings.TrimSpace(workspace.SelectedPackageMode) == "" {
		workspace.SelectedPackageMode = "thirdParty"
	}
	if workspace.MaxLogLines <= 0 {
		workspace.MaxLogLines = 100000
	}
	workspace.AIContextOptions = normalizeAIOptions(workspace.AIContextOptions)
	return workspace
}

func MergeDefaultPresets(presets []FilterPreset) []FilterPreset {
	normalized := make([]FilterPreset, 0, len(DefaultFilterPresets())+len(presets))
	seen := map[string]bool{}
	for _, preset := range DefaultFilterPresets() {
		preset = NormalizeFilterPreset(preset)
		normalized = append(normalized, preset)
		seen[preset.ID] = true
	}
	for _, preset := range presets {
		preset = NormalizeFilterPreset(preset)
		if seen[preset.ID] {
			if preset.BuiltIn {
				continue
			}
			preset.ID = newID("preset")
		}
		seen[preset.ID] = true
		normalized = append(normalized, preset)
	}
	return normalized
}

func NormalizeFilterPreset(preset FilterPreset) FilterPreset {
	preset.ID = strings.TrimSpace(preset.ID)
	if preset.ID == "" {
		preset.ID = newID("preset")
	}
	preset.Name = strings.TrimSpace(preset.Name)
	if preset.Name == "" {
		preset.Name = "Untitled Preset"
	}
	if len(preset.Level) == 0 {
		preset.Level = []string{"V", "D", "I", "W", "E", "F"}
	}
	preset.PackageName = strings.TrimSpace(preset.PackageName)
	preset.Keyword = strings.TrimSpace(preset.Keyword)
	preset.ExcludeKeyword = strings.TrimSpace(preset.ExcludeKeyword)
	for index := range preset.Tags {
		preset.Tags[index] = strings.TrimSpace(preset.Tags[index])
	}
	return preset
}

func SaveWorkspace(config AppConfig, workspace WorkspaceConfig) AppConfig {
	config = NormalizeConfig(config)
	workspace = NormalizeWorkspace(workspace)
	for index := range config.Workspaces {
		if config.Workspaces[index].ID == workspace.ID {
			config.Workspaces[index] = workspace
			config.ActiveWorkspaceID = workspace.ID
			return NormalizeConfig(config)
		}
	}
	config.Workspaces = append(config.Workspaces, workspace)
	config.ActiveWorkspaceID = workspace.ID
	return NormalizeConfig(config)
}

func DeleteWorkspace(config AppConfig, id string) AppConfig {
	config = NormalizeConfig(config)
	id = strings.TrimSpace(id)
	if len(config.Workspaces) <= 1 {
		return config
	}
	next := config.Workspaces[:0]
	for _, workspace := range config.Workspaces {
		if workspace.ID != id {
			next = append(next, workspace)
		}
	}
	config.Workspaces = next
	if config.ActiveWorkspaceID == id && len(config.Workspaces) > 0 {
		config.ActiveWorkspaceID = config.Workspaces[0].ID
	}
	return NormalizeConfig(config)
}

func SetActiveWorkspace(config AppConfig, id string) AppConfig {
	config = NormalizeConfig(config)
	if hasWorkspace(config.Workspaces, id) {
		config.ActiveWorkspaceID = id
	}
	return config
}

func SaveFilterPreset(config AppConfig, preset FilterPreset) AppConfig {
	config = NormalizeConfig(config)
	preset = NormalizeFilterPreset(preset)
	for index := range config.FilterPresets {
		if config.FilterPresets[index].ID == preset.ID {
			preset.BuiltIn = config.FilterPresets[index].BuiltIn
			config.FilterPresets[index] = preset
			return NormalizeConfig(config)
		}
	}
	config.FilterPresets = append(config.FilterPresets, preset)
	return NormalizeConfig(config)
}

func DeleteFilterPreset(config AppConfig, id string) AppConfig {
	config = NormalizeConfig(config)
	id = strings.TrimSpace(id)
	next := config.FilterPresets[:0]
	for _, preset := range config.FilterPresets {
		if preset.ID == id && !preset.BuiltIn {
			continue
		}
		next = append(next, preset)
	}
	config.FilterPresets = next
	return NormalizeConfig(config)
}

func ActiveWorkspace(config AppConfig) WorkspaceConfig {
	config = NormalizeConfig(config)
	for _, workspace := range config.Workspaces {
		if workspace.ID == config.ActiveWorkspaceID {
			return workspace
		}
	}
	return config.Workspaces[0]
}

func ProjectFromWorkspace(workspace WorkspaceConfig) ProjectConfig {
	return ProjectConfig{
		ProjectPath:      workspace.ProjectPath,
		PackageName:      workspace.PackageName,
		LastAPKPath:      workspace.LastAPKPath,
		DefaultBuildTask: workspace.DefaultBuildTask,
		InstallOptions:   workspace.InstallOptions,
	}
}

func WorkspaceFromProject(base WorkspaceConfig, project ProjectConfig) WorkspaceConfig {
	base.ProjectPath = project.ProjectPath
	base.PackageName = project.PackageName
	base.LastAPKPath = project.LastAPKPath
	base.DefaultBuildTask = project.DefaultBuildTask
	base.InstallOptions = project.InstallOptions
	return NormalizeWorkspace(base)
}

func hasWorkspace(workspaces []WorkspaceConfig, id string) bool {
	id = strings.TrimSpace(id)
	for _, workspace := range workspaces {
		if workspace.ID == id {
			return true
		}
	}
	return false
}

func cleanPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func normalizeAIOptions(options ai.AIContextOptions) ai.AIContextOptions {
	defaults := ai.DefaultOptions()
	if options.Language == "" {
		options.Language = defaults.Language
	}
	if options.IncludeBeforeContextLines == 0 && options.IncludeAfterContextLines == 0 {
		options.IncludeBeforeContextLines = defaults.IncludeBeforeContextLines
		options.IncludeAfterContextLines = defaults.IncludeAfterContextLines
	}
	return options
}

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
