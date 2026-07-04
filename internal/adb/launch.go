package adb

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type LaunchResult struct {
	Success        bool   `json:"success"`
	PackageName    string `json:"packageName"`
	DurationMillis int64  `json:"durationMillis"`
	Output         string `json:"output"`
	Error          string `json:"error,omitempty"`
}

func LaunchApp(ctx context.Context, adbPath, serial, packageName string) (LaunchResult, error) {
	serial = strings.TrimSpace(serial)
	packageName = strings.TrimSpace(packageName)
	if serial == "" {
		return LaunchResult{}, errors.New("device serial is required")
	}
	if packageName == "" {
		return LaunchResult{}, errors.New("package name is required")
	}

	state, err := GetDeviceState(ctx, adbPath, serial)
	if err != nil {
		return LaunchResult{}, err
	}
	if state != "device" {
		return LaunchResult{}, fmt.Errorf("cannot launch app: device %s is %s", serial, state)
	}

	args := BuildLaunchArgs(serial, packageName)
	started := time.Now()
	cmd := exec.CommandContext(ctx, adbPath, args...)
	output, runErr := cmd.CombinedOutput()
	outputText := strings.TrimSpace(strings.ReplaceAll(string(output), "\r", ""))
	success := runErr == nil && !looksLikeMonkeyFailure(outputText)

	result := LaunchResult{
		Success:        success,
		PackageName:    packageName,
		DurationMillis: time.Since(started).Milliseconds(),
		Output:         outputText,
	}
	if !success {
		if runErr != nil {
			result.Error = fmt.Sprintf("adb monkey launch failed: %v", runErr)
		} else {
			result.Error = "adb monkey launch failed"
		}
		if outputText != "" {
			result.Error += ": " + lastADBLine(outputText)
		}
	}
	return result, nil
}

func BuildLaunchArgs(serial, packageName string) []string {
	return []string{"-s", serial, "shell", "monkey", "-p", packageName, "1"}
}

func looksLikeMonkeyFailure(output string) bool {
	lower := strings.ToLower(output)
	return strings.Contains(lower, "no activities found") ||
		strings.Contains(lower, "monkey aborted") ||
		strings.Contains(lower, "error")
}
