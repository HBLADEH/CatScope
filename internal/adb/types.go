package adb

type AndroidDevice struct {
	Serial         string `json:"serial"`
	State          string `json:"state"`
	Model          string `json:"model,omitempty"`
	Brand          string `json:"brand,omitempty"`
	AndroidVersion string `json:"androidVersion,omitempty"`
	SDKVersion     string `json:"sdkVersion,omitempty"`
	ABI            string `json:"abi,omitempty"`
	IsEmulator     bool   `json:"isEmulator,omitempty"`
}

type InstalledPackage struct {
	PackageName string `json:"packageName"`
	Label       string `json:"label,omitempty"`
}
