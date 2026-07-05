package logcat

type LogEntry struct {
	ID          int64    `json:"id"`
	Timestamp   string   `json:"timestamp"`
	PID         int      `json:"pid"`
	TID         int      `json:"tid"`
	Level       string   `json:"level"`
	Tag         string   `json:"tag"`
	Message     string   `json:"message"`
	PackageName string   `json:"packageName,omitempty"`
	Raw         string   `json:"raw"`
	Multiline   []string `json:"multiline,omitempty"`
}

type LogBatch struct {
	Entries        []LogEntry `json:"entries"`
	Count          int        `json:"count"`
	DiscardedCount int64      `json:"discardedCount"`
	LastID         int64      `json:"lastID"`
}

type LogStatus struct {
	Running                 bool   `json:"running"`
	Serial                  string `json:"serial"`
	LastError               string `json:"lastError,omitempty"`
	Count                   int    `json:"count"`
	DiscardedCount          int64  `json:"discardedCount"`
	LastID                  int64  `json:"lastID"`
	ADBPath                 string `json:"adbPath,omitempty"`
	Source                  string `json:"source"`
	OfflineFilePath         string `json:"offlineFilePath,omitempty"`
	OfflineFileName         string `json:"offlineFileName,omitempty"`
	OfflineParseFailedCount int    `json:"offlineParseFailedCount,omitempty"`
	SessionFilePath         string `json:"sessionFilePath,omitempty"`
	SessionName             string `json:"sessionName,omitempty"`
}
