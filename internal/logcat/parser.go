package logcat

import (
	"regexp"
	"strconv"
	"strings"
)

var threadtimePattern = regexp.MustCompile(`^\s*(\d{2}-\d{2})\s+(\d{2}:\d{2}:\d{2}\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEF])\s+([^:]+):\s?(.*)$`)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Reset() {}

func (p *Parser) Parse(line string) (LogEntry, bool) {
	match := threadtimePattern.FindStringSubmatch(line)
	if match == nil {
		return LogEntry{}, false
	}

	pid, err := strconv.Atoi(match[3])
	if err != nil {
		return LogEntry{}, false
	}
	tid, err := strconv.Atoi(match[4])
	if err != nil {
		return LogEntry{}, false
	}

	return LogEntry{
		Timestamp: match[1] + " " + match[2],
		PID:       pid,
		TID:       tid,
		Level:     match[5],
		Tag:       strings.TrimSpace(match[6]),
		Message:   match[7],
		Raw:       line,
	}, true
}

func IsContinuation(previous, current LogEntry) bool {
	if previous.ID == 0 && previous.Raw == "" {
		return false
	}

	message := strings.TrimSpace(current.Message)
	if message == "" {
		return false
	}

	if current.PID != 0 && previous.PID != 0 && current.PID != previous.PID {
		return false
	}

	if strings.EqualFold(current.Tag, "AndroidRuntime") && (current.Level == "E" || current.Level == "F") {
		if strings.EqualFold(previous.Tag, "AndroidRuntime") &&
			(strings.Contains(previous.Message, "FATAL EXCEPTION") || len(previous.Multiline) > 0) {
			return true
		}
	}

	return isJavaStackMessage(message) && (previous.Level == "E" || previous.Level == "F" || previous.Tag == current.Tag)
}

func isJavaStackMessage(message string) bool {
	stackPrefixes := []string{
		"at ",
		"... ",
		"Caused by:",
		"Suppressed:",
		"java.",
		"javax.",
		"kotlin.",
		"android.",
	}
	for _, prefix := range stackPrefixes {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}
