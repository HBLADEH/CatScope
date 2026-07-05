package adb

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"catscope/internal/process"
)

func FindADB(configuredPath string) (string, error) {
	candidates := make([]string, 0, 5)

	if configuredPath = strings.TrimSpace(configuredPath); configuredPath != "" {
		candidates = append(candidates, expandADBCandidate(configuredPath)...)
	}

	for _, envName := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if sdk := strings.TrimSpace(os.Getenv(envName)); sdk != "" {
			candidates = append(candidates, adbFromSDK(sdk))
		}
	}

	if pathCandidate, err := exec.LookPath(adbExecutableName()); err == nil {
		candidates = append(candidates, pathCandidate)
	}

	seen := map[string]bool{}
	var failures []string
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if seen[strings.ToLower(candidate)] {
			continue
		}
		seen[strings.ToLower(candidate)] = true

		if err := validateADB(candidate); err == nil {
			return candidate, nil
		} else {
			failures = append(failures, fmt.Sprintf("%s: %v", candidate, err))
		}
	}

	if len(failures) > 0 {
		return "", fmt.Errorf("adb not found or invalid: %s", strings.Join(failures, "; "))
	}
	return "", errors.New("adb not found; set adbPath, ANDROID_HOME, ANDROID_SDK_ROOT, or add adb to PATH")
}

func validateADB(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	cmd := exec.Command(path, "version")
	process.HideConsoleWindow(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	if !strings.Contains(strings.ToLower(string(output)), "android debug bridge") {
		return errors.New("adb version output did not identify Android Debug Bridge")
	}
	return nil
}

func expandADBCandidate(path string) []string {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return []string{
			filepath.Join(path, "platform-tools", adbExecutableName()),
			filepath.Join(path, adbExecutableName()),
		}
	}
	return []string{path}
}

func adbFromSDK(sdk string) string {
	return filepath.Join(sdk, "platform-tools", adbExecutableName())
}

func adbExecutableName() string {
	if runtime.GOOS == "windows" {
		return "adb.exe"
	}
	return "adb"
}
