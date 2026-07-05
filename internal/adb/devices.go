package adb

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"catscope/internal/process"
)

func ListDevices(ctx context.Context, adbPath string) ([]AndroidDevice, error) {
	cmd := exec.CommandContext(ctx, adbPath, "devices", "-l")
	process.HideConsoleWindow(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("adb devices failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return ParseDevices(string(output)), nil
}

func ParseDevices(output string) []AndroidDevice {
	lines := strings.Split(output, "\n")
	devices := make([]AndroidDevice, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") || strings.HasPrefix(line, "* daemon") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		device := AndroidDevice{
			Serial:     fields[0],
			State:      normalizeState(fields[1]),
			IsEmulator: strings.HasPrefix(fields[0], "emulator-"),
		}

		for _, field := range fields[2:] {
			key, value, ok := strings.Cut(field, ":")
			if !ok {
				continue
			}
			switch key {
			case "model":
				device.Model = strings.ReplaceAll(value, "_", " ")
			case "device":
				if device.Model == "" {
					device.Model = strings.ReplaceAll(value, "_", " ")
				}
			case "product":
				if device.Brand == "" {
					device.Brand = value
				}
			}
		}

		devices = append(devices, device)
	}

	return devices
}

func GetDeviceInfo(ctx context.Context, adbPath, serial string) (AndroidDevice, error) {
	device := AndroidDevice{
		Serial:     serial,
		State:      "unknown",
		IsEmulator: strings.HasPrefix(serial, "emulator-"),
	}

	state, _ := adbOutput(ctx, adbPath, "-s", serial, "get-state")
	if state != "" {
		device.State = normalizeState(state)
	}

	props := map[string]*string{
		"ro.product.brand":         &device.Brand,
		"ro.product.model":         &device.Model,
		"ro.build.version.release": &device.AndroidVersion,
		"ro.build.version.sdk":     &device.SDKVersion,
		"ro.product.cpu.abi":       &device.ABI,
	}

	for prop, target := range props {
		value, err := adbOutput(ctx, adbPath, "-s", serial, "shell", "getprop", prop)
		if err == nil {
			*target = value
		}
	}

	return device, nil
}

func GetDeviceState(ctx context.Context, adbPath, serial string) (string, error) {
	state, err := adbOutput(ctx, adbPath, "-s", serial, "get-state")
	if err != nil {
		return "unknown", err
	}
	return normalizeState(state), nil
}

func adbOutput(ctx context.Context, adbPath string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, adbPath, args...)
	process.HideConsoleWindow(cmd)
	output, err := cmd.CombinedOutput()
	value := strings.TrimSpace(strings.ReplaceAll(string(output), "\r", ""))
	if err != nil {
		return value, fmt.Errorf("adb %s failed: %w: %s", strings.Join(args, " "), err, value)
	}
	return value, nil
}

func normalizeState(state string) string {
	state = strings.TrimSpace(state)
	switch state {
	case "device", "offline", "unauthorized":
		return state
	case "":
		return "unknown"
	default:
		return "unknown"
	}
}
