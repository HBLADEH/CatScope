package logcat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	LogSourceLive    = "live"
	LogSourceOffline = "offline"
)

type OfflineLogFileResult struct {
	FilePath         string           `json:"filePath"`
	FileName         string           `json:"fileName"`
	Entries          []LogEntry       `json:"entries"`
	Count            int              `json:"count"`
	ParseFailedCount int              `json:"parseFailedCount"`
	AnalysisResults  []AnalysisResult `json:"analysisResults,omitempty"`
}

func LoadOfflineLogFile(path string) (OfflineLogFileResult, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return OfflineLogFileResult{}, fmt.Errorf("log file path is required")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return OfflineLogFileResult{}, fmt.Errorf("read log file failed: %w", err)
	}

	entries, parseFailed, err := ParseOfflineContent(string(content), filepath.Ext(path))
	if err != nil {
		return OfflineLogFileResult{}, err
	}

	return OfflineLogFileResult{
		FilePath:         path,
		FileName:         filepath.Base(path),
		Entries:          entries,
		Count:            len(entries),
		ParseFailedCount: parseFailed,
	}, nil
}

func ParseOfflineContent(content string, extension string) ([]LogEntry, int, error) {
	if strings.EqualFold(extension, ".jsonl") {
		entries, err := ParseJSONL(content)
		if err == nil {
			return entries, 0, nil
		}
	}

	entries, parseFailed := ParseOfflineText(content)
	return entries, parseFailed, nil
}

func ParseOfflineText(content string) ([]LogEntry, int) {
	parser := NewParser()
	buffer := NewRingBuffer(100000)
	parseFailed := 0

	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if entry, ok := parser.Parse(line); ok {
			if buffer.AppendContinuation(entry) {
				continue
			}
			buffer.Add(entry)
			continue
		}

		parseFailed++
		if buffer.AppendMultiline(line) {
			continue
		}
		buffer.Add(LogEntry{
			Level:   "I",
			Message: line,
			Raw:     line,
		})
	}

	if err := scanner.Err(); err != nil {
		parseFailed++
		buffer.Add(LogEntry{
			Level:   "E",
			Tag:     "CatScope",
			Message: err.Error(),
			Raw:     err.Error(),
		})
	}

	return buffer.Snapshot(), parseFailed
}

func ParseJSONL(content string) ([]LogEntry, error) {
	buffer := NewRingBuffer(100000)
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("parse jsonl line %d failed: %w", lineNumber, err)
		}
		if strings.TrimSpace(entry.Raw) == "" {
			entry.Raw = formatEntryLine(entry)
		}
		buffer.Add(entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return buffer.Snapshot(), nil
}
