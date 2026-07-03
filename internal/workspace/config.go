package workspace

type Config struct {
	ADBPath          string   `json:"adbPath"`
	AndroidSDKPath   string   `json:"androidSdkPath"`
	Theme            string   `json:"theme"`
	DefaultBuffer    []string `json:"defaultBuffer"`
	DefaultLogFormat string   `json:"defaultLogFormat"`
	MaxLogLines      int      `json:"maxLogLines"`
	AutoReconnect    bool     `json:"autoReconnect"`
}

func DefaultConfig() Config {
	return Config{
		Theme:            "system",
		DefaultBuffer:    []string{"main", "system", "crash"},
		DefaultLogFormat: "threadtime",
		MaxLogLines:      100000,
		AutoReconnect:    true,
	}
}
