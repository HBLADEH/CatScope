package adb

import "testing"

func TestParseDevicesEmpty(t *testing.T) {
	devices := ParseDevices("List of devices attached\n\n")
	if len(devices) != 0 {
		t.Fatalf("expected no devices, got %d", len(devices))
	}
}

func TestParseDevicesStatesAndFields(t *testing.T) {
	output := `List of devices attached
emulator-5554 device product:sdk_gphone64_x86_64 model:Pixel_7_API_35 device:emu64xa transport_id:1
ABC123 offline usb:1-1 product:oriole model:Pixel_6 device:oriole
XYZ987 unauthorized usb:1-2
`

	devices := ParseDevices(output)
	if len(devices) != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devices))
	}

	if devices[0].Serial != "emulator-5554" || devices[0].State != "device" || !devices[0].IsEmulator {
		t.Fatalf("unexpected first device: %+v", devices[0])
	}
	if devices[0].Model != "Pixel 7 API 35" {
		t.Fatalf("model underscores were not normalized: %q", devices[0].Model)
	}
	if devices[1].State != "offline" || devices[1].Brand != "oriole" {
		t.Fatalf("unexpected offline device: %+v", devices[1])
	}
	if devices[2].State != "unauthorized" {
		t.Fatalf("unexpected unauthorized state: %+v", devices[2])
	}
}

func TestNormalizeStateLimitsKnownDeviceStates(t *testing.T) {
	tests := map[string]string{
		"device":       "device",
		"offline":      "offline",
		"unauthorized": "unauthorized",
		"recovery":     "unknown",
		"":             "unknown",
	}

	for input, want := range tests {
		if got := normalizeState(input); got != want {
			t.Fatalf("normalizeState(%q)=%q, want %q", input, got, want)
		}
	}
}
