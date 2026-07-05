package logcat

import (
	"encoding/json"
	"fmt"
	"strings"
)

func FormatEntriesText(entries []LogEntry) string {
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString(formatEntryLine(entry))
		builder.WriteByte('\n')
		for _, line := range entry.Multiline {
			builder.WriteString(line)
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

func FormatEntriesJSONL(entries []LogEntry) (string, error) {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			return "", fmt.Errorf("encode log entry failed: %w", err)
		}
	}
	return builder.String(), nil
}

func formatEntryLine(entry LogEntry) string {
	return fmt.Sprintf(
		"%s %5d %5d %s %-24s %s",
		valueOrDash(entry.Timestamp),
		entry.PID,
		entry.TID,
		valueOrDash(entry.Level),
		valueOrDash(entry.Tag)+":",
		entry.Message,
	)
}

func valueOrDash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}
