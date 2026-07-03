package adb

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type PackageListMode string

const (
	PackageListModeAll        PackageListMode = "all"
	PackageListModeThirdParty PackageListMode = "thirdParty"
)

func ListPackages(ctx context.Context, adbPath, serial string, mode PackageListMode) ([]InstalledPackage, error) {
	state, err := GetDeviceState(ctx, adbPath, serial)
	if err != nil {
		return nil, err
	}
	if state != "device" {
		return nil, fmt.Errorf("cannot list packages: device %s is %s", serial, state)
	}

	args := []string{"-s", serial, "shell", "pm", "list", "packages"}
	if mode == PackageListModeThirdParty {
		args = append(args, "-3")
	}

	output, err := adbOutput(ctx, adbPath, args...)
	if err != nil {
		return nil, err
	}
	return ParsePackageList(output), nil
}

func ParsePackageList(output string) []InstalledPackage {
	lines := strings.Split(output, "\n")
	packages := make([]InstalledPackage, 0, len(lines))
	seen := map[string]bool{}

	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "package:") {
			line = strings.TrimPrefix(line, "package:")
		}
		if idx := strings.Index(line, "="); idx >= 0 {
			line = line[idx+1:]
		}
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		packages = append(packages, InstalledPackage{PackageName: line})
	}

	return packages
}

func PidOf(ctx context.Context, adbPath, serial, packageName string) ([]int, error) {
	output, err := adbOutput(ctx, adbPath, "-s", serial, "shell", "pidof", packageName)
	if err != nil {
		if isPidofNotRunning(output, err) {
			return nil, nil
		}
		return nil, err
	}
	return ParsePidOf(output), nil
}

func ParsePidOf(output string) []int {
	fields := strings.Fields(strings.TrimSpace(strings.ReplaceAll(output, "\r", "")))
	pids := make([]int, 0, len(fields))
	seen := map[int]bool{}

	for _, field := range fields {
		pid, err := strconv.Atoi(field)
		if err != nil || pid <= 0 || seen[pid] {
			continue
		}
		seen[pid] = true
		pids = append(pids, pid)
	}

	return pids
}

func isPidofNotRunning(output string, err error) bool {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return true
	}
	lower := strings.ToLower(trimmed + " " + err.Error())
	return strings.Contains(lower, "not found") || strings.Contains(lower, "no such process")
}
