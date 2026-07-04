package adb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"catscope/internal/logcat"
)

type InstallOptions struct {
	AllowDowngrade   bool `json:"allowDowngrade"`
	GrantPermissions bool `json:"grantPermissions"`
	AllowTestOnly    bool `json:"allowTestOnly"`
}

type InstallResult struct {
	Success         bool                    `json:"success"`
	APKPath         string                  `json:"apkPath"`
	DurationMillis  int64                   `json:"durationMillis"`
	Output          string                  `json:"output"`
	Error           string                  `json:"error,omitempty"`
	AnalysisResults []logcat.AnalysisResult `json:"analysisResults,omitempty"`
}

func InstallAPK(ctx context.Context, adbPath, serial, apkPath string, options InstallOptions) (InstallResult, error) {
	serial = strings.TrimSpace(serial)
	apkPath = strings.TrimSpace(apkPath)
	if serial == "" {
		return InstallResult{}, errors.New("device serial is required")
	}
	if apkPath == "" {
		return InstallResult{}, errors.New("apk path is required")
	}
	info, err := os.Stat(apkPath)
	if err != nil {
		return InstallResult{}, fmt.Errorf("apk path is not accessible: %w", err)
	}
	if info.IsDir() {
		return InstallResult{}, fmt.Errorf("apk path is a directory: %s", apkPath)
	}

	state, err := GetDeviceState(ctx, adbPath, serial)
	if err != nil {
		return InstallResult{}, err
	}
	if state != "device" {
		return InstallResult{}, fmt.Errorf("cannot install APK: device %s is %s", serial, state)
	}

	args := BuildInstallArgs(serial, apkPath, options)
	started := time.Now()
	cmd := exec.CommandContext(ctx, adbPath, args...)
	output, runErr := cmd.CombinedOutput()
	outputText := strings.TrimSpace(strings.ReplaceAll(string(output), "\r", ""))
	result := InstallResult{
		Success:        runErr == nil && looksLikeInstallSuccess(outputText),
		APKPath:        apkPath,
		DurationMillis: time.Since(started).Milliseconds(),
		Output:         outputText,
	}
	if runErr != nil || !result.Success {
		if runErr != nil {
			result.Error = fmt.Sprintf("adb install failed: %v", runErr)
		} else {
			result.Error = "adb install did not report success"
		}
		if outputText != "" {
			result.Error += ": " + lastADBLine(outputText)
		}
		result.AnalysisResults = logcat.AnalyzeInstallOutput(outputText)
	}
	return result, nil
}

func BuildInstallArgs(serial, apkPath string, options InstallOptions) []string {
	args := []string{"-s", serial, "install", "-r"}
	if options.AllowDowngrade {
		args = append(args, "-d")
	}
	if options.GrantPermissions {
		args = append(args, "-g")
	}
	if options.AllowTestOnly {
		args = append(args, "-t")
	}
	args = append(args, apkPath)
	return args
}

func looksLikeInstallSuccess(output string) bool {
	output = strings.ToLower(output)
	return strings.Contains(output, "success")
}

func lastADBLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
