package build

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type APKInfo struct {
	APKPath      string `json:"apkPath"`
	FileName     string `json:"fileName"`
	ModifiedTime string `json:"modifiedTime"`
	Size         int64  `json:"size"`
}

type BuildRequest struct {
	ProjectPath string `json:"projectPath"`
	Task        string `json:"task"`
}

type BuildResult struct {
	Success        bool     `json:"success"`
	ProjectPath    string   `json:"projectPath"`
	Task           string   `json:"task"`
	DurationMillis int64    `json:"durationMillis"`
	Output         string   `json:"output"`
	Error          string   `json:"error,omitempty"`
	APK            *APKInfo `json:"apk,omitempty"`
}

func RunDebugBuild(ctx context.Context, projectPath string) (BuildResult, error) {
	return Run(ctx, BuildRequest{
		ProjectPath: projectPath,
		Task:        "assembleDebug",
	})
}

func Run(ctx context.Context, request BuildRequest) (BuildResult, error) {
	projectPath, task, err := normalizeRequest(request)
	if err != nil {
		return BuildResult{}, err
	}

	executable, args, dir, err := BuildCommand(projectPath, task, runtime.GOOS)
	if err != nil {
		return BuildResult{}, err
	}

	started := time.Now()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	output, runErr := cmd.CombinedOutput()
	duration := time.Since(started)
	outputText := strings.TrimSpace(strings.ReplaceAll(string(output), "\r", ""))

	result := BuildResult{
		Success:        runErr == nil,
		ProjectPath:    projectPath,
		Task:           task,
		DurationMillis: duration.Milliseconds(),
		Output:         outputText,
	}
	if runErr != nil {
		result.Error = fmt.Sprintf("gradle %s failed: %v", task, runErr)
		if outputText != "" {
			result.Error += ": " + lastNonEmptyLine(outputText)
		}
		return result, nil
	}

	apk, findErr := FindLatestAPK(projectPath)
	if findErr != nil {
		result.Success = false
		result.Error = findErr.Error()
		return result, nil
	}
	result.APK = &apk
	return result, nil
}

func BuildCommand(projectPath string, task string, goos string) (string, []string, string, error) {
	projectPath = strings.TrimSpace(projectPath)
	task = normalizeTask(task)
	if projectPath == "" {
		return "", nil, "", errors.New("project path is required")
	}
	if err := validateAndroidProject(projectPath); err != nil {
		return "", nil, "", err
	}

	wrapper, err := GradleWrapper(projectPath, goos)
	if err != nil {
		return "", nil, "", err
	}
	return wrapper, []string{task}, projectPath, nil
}

func GradleWrapper(projectPath string, goos string) (string, error) {
	var candidates []string
	if goos == "windows" {
		candidates = []string{
			filepath.Join(projectPath, "gradlew.bat"),
			filepath.Join(projectPath, "gradlew"),
		}
	} else {
		candidates = []string{
			filepath.Join(projectPath, "gradlew"),
			filepath.Join(projectPath, "gradlew.bat"),
		}
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("gradle wrapper not found in %s; expected gradlew or gradlew.bat", projectPath)
}

func FindLatestAPK(projectPath string) (APKInfo, error) {
	projectPath = strings.TrimSpace(projectPath)
	if projectPath == "" {
		return APKInfo{}, errors.New("project path is required")
	}
	info, err := os.Stat(projectPath)
	if err != nil {
		return APKInfo{}, fmt.Errorf("project path is not accessible: %w", err)
	}
	if !info.IsDir() {
		return APKInfo{}, fmt.Errorf("project path is not a directory: %s", projectPath)
	}

	var candidates []apkCandidate
	err = filepath.WalkDir(projectPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == ".gradle" || entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".apk") {
			return nil
		}
		normalized := filepath.ToSlash(path)
		if !strings.Contains(normalized, "/build/outputs/apk/") {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		candidates = append(candidates, apkCandidate{
			path:  path,
			info:  info,
			debug: looksLikeDebugAPK(path),
		})
		return nil
	})
	if err != nil {
		return APKInfo{}, fmt.Errorf("search APK failed: %w", err)
	}
	if len(candidates) == 0 {
		return APKInfo{}, fmt.Errorf("no APK found under %s/build/outputs/apk", projectPath)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].debug != candidates[j].debug {
			return candidates[i].debug
		}
		return candidates[i].info.ModTime().After(candidates[j].info.ModTime())
	})
	return apkInfo(candidates[0].path, candidates[0].info), nil
}

type apkCandidate struct {
	path  string
	info  os.FileInfo
	debug bool
}

func normalizeRequest(request BuildRequest) (string, string, error) {
	projectPath := strings.TrimSpace(request.ProjectPath)
	task := normalizeTask(request.Task)
	if projectPath == "" {
		return "", "", errors.New("project path is required")
	}
	return filepath.Clean(projectPath), task, nil
}

func normalizeTask(task string) string {
	task = strings.TrimSpace(task)
	if task == "" {
		return "assembleDebug"
	}
	return task
}

func validateAndroidProject(projectPath string) error {
	info, err := os.Stat(projectPath)
	if err != nil {
		return fmt.Errorf("project path is not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory: %s", projectPath)
	}
	for _, name := range []string{"settings.gradle", "settings.gradle.kts"} {
		path := filepath.Join(projectPath, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return nil
		}
	}
	return fmt.Errorf("settings.gradle or settings.gradle.kts not found in %s", projectPath)
}

func looksLikeDebugAPK(path string) bool {
	normalized := strings.ToLower(filepath.ToSlash(path))
	return strings.Contains(normalized, "/debug/") || strings.Contains(filepath.Base(normalized), "debug")
}

func apkInfo(path string, info os.FileInfo) APKInfo {
	return APKInfo{
		APKPath:      path,
		FileName:     filepath.Base(path),
		ModifiedTime: info.ModTime().Format(time.RFC3339),
		Size:         info.Size(),
	}
}

func lastNonEmptyLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
