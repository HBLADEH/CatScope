package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"catscope/internal/adb"
)

type ProjectConfig struct {
	ProjectPath      string             `json:"projectPath"`
	PackageName      string             `json:"packageName"`
	LastAPKPath      string             `json:"lastApkPath"`
	DefaultBuildTask string             `json:"defaultBuildTask"`
	InstallOptions   adb.InstallOptions `json:"installOptions"`
}

func DefaultProjectConfig() ProjectConfig {
	return ProjectConfig{
		DefaultBuildTask: "assembleDebug",
		InstallOptions: adb.InstallOptions{
			AllowDowngrade:   false,
			GrantPermissions: true,
			AllowTestOnly:    false,
		},
	}
}

func LoadProjectConfig(path string) (ProjectConfig, error) {
	config := DefaultProjectConfig()
	path = strings.TrimSpace(path)
	if path == "" {
		return config, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, err
	}
	if len(data) == 0 {
		return config, nil
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultProjectConfig(), fmt.Errorf("load project config failed: %w", err)
	}
	return NormalizeProjectConfig(config), nil
}

func SaveProjectConfig(path string, config ProjectConfig) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	config = NormalizeProjectConfig(config)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func NormalizeProjectConfig(config ProjectConfig) ProjectConfig {
	config.ProjectPath = strings.TrimSpace(config.ProjectPath)
	if config.ProjectPath != "" {
		config.ProjectPath = filepath.Clean(config.ProjectPath)
	}
	config.PackageName = strings.TrimSpace(config.PackageName)
	config.LastAPKPath = strings.TrimSpace(config.LastAPKPath)
	if config.LastAPKPath != "" {
		config.LastAPKPath = filepath.Clean(config.LastAPKPath)
	}
	config.DefaultBuildTask = strings.TrimSpace(config.DefaultBuildTask)
	if config.DefaultBuildTask == "" {
		config.DefaultBuildTask = "assembleDebug"
	}
	return config
}

func DefaultProjectConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(dir) == "" {
		return ""
	}
	return filepath.Join(dir, "CatScope", "project.json")
}
