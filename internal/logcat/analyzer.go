package logcat

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

type AnalysisType string
type AnalysisSeverity string

const (
	AnalysisTypeJavaCrash    AnalysisType = "java_crash"
	AnalysisTypeNativeCrash  AnalysisType = "native_crash"
	AnalysisTypeANR          AnalysisType = "anr"
	AnalysisTypeJNIError     AnalysisType = "jni_error"
	AnalysisTypeInstallError AnalysisType = "install_error"
	AnalysisTypeUnknown      AnalysisType = "unknown"

	SeverityInfo    AnalysisSeverity = "info"
	SeverityWarning AnalysisSeverity = "warning"
	SeverityError   AnalysisSeverity = "error"
	SeverityFatal   AnalysisSeverity = "fatal"
)

type AnalysisResult struct {
	ID              string           `json:"id"`
	Type            AnalysisType     `json:"type"`
	Severity        AnalysisSeverity `json:"severity"`
	Title           string           `json:"title"`
	Summary         string           `json:"summary"`
	PackageName     string           `json:"packageName,omitempty"`
	PID             int              `json:"pid,omitempty"`
	TID             int              `json:"tid,omitempty"`
	Timestamp       string           `json:"timestamp,omitempty"`
	PrimaryTag      string           `json:"primaryTag,omitempty"`
	PrimaryMessage  string           `json:"primaryMessage,omitempty"`
	ExceptionType   string           `json:"exceptionType,omitempty"`
	ThreadName      string           `json:"threadName,omitempty"`
	Signal          string           `json:"signal,omitempty"`
	LibraryName     string           `json:"libraryName,omitempty"`
	Reason          string           `json:"reason,omitempty"`
	KeyFrames       []string         `json:"keyFrames,omitempty"`
	RelatedEntryIDs []int64          `json:"relatedEntryIds,omitempty"`
	RawText         string           `json:"rawText,omitempty"`
	Suggestions     []string         `json:"suggestions,omitempty"`
}

var (
	exceptionPattern = regexp.MustCompile(`(?m)((?:java|kotlin|android|dalvik|libcore)\.[A-Za-z0-9_.$]+(?:Exception|Error)|[A-Za-z0-9_.$]+(?:Exception|Error))(?::|\s|$)`)
	threadPattern    = regexp.MustCompile(`FATAL EXCEPTION:\s*([^\n\r]+)`)
	processPattern   = regexp.MustCompile(`Process:\s*([A-Za-z0-9_.]+)\s*,\s*PID:\s*(\d+)`)
	signalPattern    = regexp.MustCompile(`(?i)\b(SIGSEGV|SIGABRT|signal\s+\d+)\b`)
	libraryPattern   = regexp.MustCompile(`\b([A-Za-z0-9_./-]*lib[A-Za-z0-9_.+-]+\.so)\b`)
)

func AnalyzeEntries(entries []LogEntry) []AnalysisResult {
	results := make([]AnalysisResult, 0)
	seen := map[string]bool{}
	covered := map[int64]bool{}

	for i, entry := range entries {
		if entry.ID > 0 && covered[entry.ID] {
			continue
		}
		context := analysisContext(entries, i)
		if result, ok := analyzeOne(entry, context); ok {
			if seen[result.ID] {
				continue
			}
			seen[result.ID] = true
			for _, id := range result.RelatedEntryIDs {
				covered[id] = true
			}
			results = append(results, result)
		}
	}

	return results
}

func analyzeOne(entry LogEntry, context []LogEntry) (AnalysisResult, bool) {
	rawText := joinEntries(context)
	lower := strings.ToLower(rawText)

	switch {
	case looksLikeJNIError(lower):
		return analyzeJNIError(entry, context, rawText), true
	case looksLikeJavaCrash(entry, lower):
		return analyzeJavaCrash(entry, context, rawText), true
	case looksLikeNativeCrash(entry, lower):
		return analyzeNativeCrash(entry, context, rawText), true
	case looksLikeANR(lower):
		return analyzeANR(entry, context, rawText), true
	default:
		return AnalysisResult{}, false
	}
}

func analyzeJavaCrash(entry LogEntry, context []LogEntry, rawText string) AnalysisResult {
	exceptionType := firstSubmatch(exceptionPattern, rawText, 1)
	threadName := strings.TrimSpace(firstSubmatch(threadPattern, rawText, 1))
	packageName := entry.PackageName
	if packageName == "" {
		packageName = firstSubmatch(processPattern, rawText, 1)
	}
	keyFrames := extractKeyFrames(rawText, 5)
	title := "Java crash"
	if exceptionType != "" {
		title = "Java crash: " + exceptionType
	}

	return resultBase(AnalysisTypeJavaCrash, SeverityFatal, entry, context, rawText, title, fmt.Sprintf(
		"%s crashed%s.",
		valueOrUnknown(packageName, "The app"),
		threadSuffix(threadName),
	), func(result *AnalysisResult) {
		result.ExceptionType = exceptionType
		result.ThreadName = threadName
		result.PackageName = packageName
		result.KeyFrames = keyFrames
		result.Reason = firstReasonLine(rawText, []string{"Caused by:", exceptionType})
		result.Suggestions = []string{
			"Read the first business stack frame and inspect the surrounding code path.",
			"Check the Caused by section before fixing symptoms in later frames.",
			"Reproduce with the same package, version, and user action if possible.",
		}
	})
}

func analyzeNativeCrash(entry LogEntry, context []LogEntry, rawText string) AnalysisResult {
	signal := normalizeSignal(firstMatch(signalPattern, rawText))
	libraryName := basename(firstMatch(libraryPattern, rawText))
	keyFrames := extractNativeFrames(rawText, 6)
	title := "Native crash"
	if signal != "" {
		title += ": " + signal
	}
	if libraryName != "" {
		title += " in " + libraryName
	}

	return resultBase(AnalysisTypeNativeCrash, SeverityFatal, entry, context, rawText, title, "Native crash indicators were found in logcat.", func(result *AnalysisResult) {
		result.Signal = signal
		result.LibraryName = libraryName
		result.KeyFrames = keyFrames
		result.Reason = firstReasonLine(rawText, []string{"Abort message", "fault addr", "signal"})
		result.Suggestions = []string{
			"Use ndk-stack or addr2line with matching unstripped symbols.",
			"Confirm the ABI and build variant match the device and crashing artifact.",
			"Inspect the first app-owned native frame before system library frames.",
		}
	})
}

func analyzeANR(entry LogEntry, context []LogEntry, rawText string) AnalysisResult {
	packageName := entry.PackageName
	if packageName == "" {
		packageName = extractAfterPrefix(rawText, "ANR in")
	}
	reason := firstReasonLine(rawText, []string{"Input dispatching timed out", "BroadcastQueue", "executing service", "Timeout executing service", "Application Not Responding"})
	title := "ANR detected"
	if packageName != "" {
		title = "ANR: " + packageName
	}

	return resultBase(AnalysisTypeANR, SeverityFatal, entry, context, rawText, title, "Application Not Responding indicators were found.", func(result *AnalysisResult) {
		result.PackageName = packageName
		result.Reason = reason
		result.KeyFrames = nonEmpty([]string{reason})
		result.Suggestions = []string{
			"Inspect main-thread work around the reported timeout.",
			"Check broadcast receiver or service execution time if mentioned.",
			"Collect traces.txt or an Android bugreport for thread state details.",
		}
	})
}

func analyzeJNIError(entry LogEntry, context []LogEntry, rawText string) AnalysisResult {
	reason := firstReasonLine(rawText, []string{"JNI DETECTED ERROR IN APPLICATION", "use of deleted local reference", "accessed stale", "pending exception", "thread exiting with uncaught exception", "CheckJNI"})

	return resultBase(AnalysisTypeJNIError, SeverityFatal, entry, context, rawText, "JNI error detected", "CheckJNI/JNI runtime error indicators were found.", func(result *AnalysisResult) {
		result.Reason = reason
		result.KeyFrames = extractKeyFrames(rawText, 5)
		if len(result.KeyFrames) == 0 {
			result.KeyFrames = nonEmpty([]string{reason})
		}
		result.Suggestions = []string{
			"Inspect JNI reference lifetime and delete/use ordering.",
			"Check for pending Java exceptions before returning to native code.",
			"Enable CheckJNI in a debug build and symbolicate native frames if present.",
		}
	})
}

func resultBase(kind AnalysisType, severity AnalysisSeverity, entry LogEntry, context []LogEntry, rawText string, title string, summary string, fill func(*AnalysisResult)) AnalysisResult {
	result := AnalysisResult{
		Type:            kind,
		Severity:        severity,
		Title:           title,
		Summary:         summary,
		PackageName:     entry.PackageName,
		PID:             entry.PID,
		TID:             entry.TID,
		Timestamp:       entry.Timestamp,
		PrimaryTag:      entry.Tag,
		PrimaryMessage:  entry.Message,
		RelatedEntryIDs: relatedIDs(context),
		RawText:         rawText,
	}
	fill(&result)
	result.ID = analysisID(result)
	return result
}

func analysisContext(entries []LogEntry, index int) []LogEntry {
	if index < 0 || index >= len(entries) {
		return nil
	}
	start := index
	end := index + 1
	root := entries[index]

	for end < len(entries) && end < index+12 {
		next := entries[end]
		if root.ID > 0 && next.ID == root.ID {
			break
		}
		if root.PID != 0 && next.PID != 0 && root.PID != next.PID && !contextTag(next.Tag) {
			break
		}
		if likelyRelatedText(entryText(next)) || next.PID == root.PID || contextTag(next.Tag) {
			end++
			continue
		}
		break
	}

	return entries[start:end]
}

func looksLikeJavaCrash(entry LogEntry, lower string) bool {
	return strings.EqualFold(entry.Tag, "AndroidRuntime") && strings.Contains(lower, "fatal exception") ||
		strings.Contains(lower, "fatal exception") ||
		strings.Contains(lower, "caused by:") && strings.Contains(lower, "exception") ||
		strings.Contains(lower, "java.lang.") && strings.Contains(lower, "exception")
}

func looksLikeNativeCrash(entry LogEntry, lower string) bool {
	return strings.EqualFold(entry.Tag, "DEBUG") && strings.Contains(lower, "backtrace") ||
		strings.Contains(lower, "sigsegv") ||
		strings.Contains(lower, "sigabrt") ||
		strings.Contains(lower, "signal 11") ||
		strings.Contains(lower, "tombstone") ||
		strings.Contains(lower, "abort message") ||
		strings.Contains(lower, "fault addr") && strings.Contains(lower, ".so")
}

func looksLikeANR(lower string) bool {
	return strings.Contains(lower, "anr in") ||
		strings.Contains(lower, "application not responding") ||
		strings.Contains(lower, "input dispatching timed out") ||
		strings.Contains(lower, "broadcastqueue") ||
		strings.Contains(lower, "timeout executing service") ||
		strings.Contains(lower, "executing service") && strings.Contains(lower, "timeout")
}

func looksLikeJNIError(lower string) bool {
	return strings.Contains(lower, "jni detected error in application") ||
		strings.Contains(lower, "java_vm_ext.cc") ||
		strings.Contains(lower, "checkjni") ||
		strings.Contains(lower, "use of deleted local reference") ||
		strings.Contains(lower, "accessed stale") ||
		strings.Contains(lower, "pending exception") ||
		strings.Contains(lower, "thread exiting with uncaught exception")
}

func joinEntries(entries []LogEntry) string {
	lines := make([]string, 0, len(entries)*2)
	for _, entry := range entries {
		lines = append(lines, entryText(entry))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func entryText(entry LogEntry) string {
	lines := []string{entry.Raw}
	if entry.Raw == "" {
		lines[0] = strings.TrimSpace(entry.Message)
	}
	lines = append(lines, entry.Multiline...)
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func relatedIDs(entries []LogEntry) []int64 {
	result := make([]int64, 0, len(entries))
	for _, entry := range entries {
		if entry.ID > 0 {
			result = append(result, entry.ID)
		}
	}
	return result
}

func contextTag(tag string) bool {
	switch tag {
	case "AndroidRuntime", "DEBUG", "ActivityManager", "libc":
		return true
	default:
		return false
	}
}

func likelyRelatedText(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "caused by:") ||
		strings.Contains(lower, "at ") ||
		strings.Contains(lower, "backtrace") ||
		strings.Contains(lower, "abort message") ||
		strings.Contains(lower, "fault addr") ||
		strings.Contains(lower, ".so") ||
		strings.Contains(lower, "anr") ||
		strings.Contains(lower, "jni")
}

func extractKeyFrames(text string, limit int) []string {
	var frames []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := normalizeStackFrameLine(line)
		if strings.HasPrefix(trimmed, "at ") && !isFrameworkFrame(trimmed) {
			frames = append(frames, trimmed)
		}
		if len(frames) >= limit {
			return frames
		}
	}
	for _, line := range strings.Split(text, "\n") {
		trimmed := normalizeStackFrameLine(line)
		if strings.HasPrefix(trimmed, "at ") {
			frames = append(frames, trimmed)
		}
		if len(frames) >= limit {
			break
		}
	}
	return frames
}

func normalizeStackFrameLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if idx := strings.Index(trimmed, "\tat "); idx >= 0 {
		return strings.TrimSpace(trimmed[idx+1:])
	}
	if idx := strings.Index(trimmed, " at "); idx >= 0 {
		return strings.TrimSpace(trimmed[idx+1:])
	}
	return trimmed
}

func extractNativeFrames(text string, limit int) []string {
	var frames []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "backtrace:") || strings.Contains(lower, ".so") || strings.Contains(lower, " pc ") {
			frames = append(frames, trimmed)
		}
		if len(frames) >= limit {
			break
		}
	}
	return frames
}

func isFrameworkFrame(frame string) bool {
	return strings.Contains(frame, "android.") ||
		strings.Contains(frame, "java.") ||
		strings.Contains(frame, "kotlin.") ||
		strings.Contains(frame, "com.android.")
}

func firstMatch(pattern *regexp.Regexp, text string) string {
	match := pattern.FindString(text)
	return strings.TrimSpace(match)
}

func firstSubmatch(pattern *regexp.Regexp, text string, index int) string {
	match := pattern.FindStringSubmatch(text)
	if len(match) <= index {
		return ""
	}
	return strings.TrimSpace(match[index])
}

func firstReasonLine(text string, needles []string) string {
	lowerNeedles := make([]string, 0, len(needles))
	for _, needle := range needles {
		if needle != "" {
			lowerNeedles = append(lowerNeedles, strings.ToLower(needle))
		}
	}
	for _, line := range strings.Split(text, "\n") {
		lower := strings.ToLower(line)
		for _, needle := range lowerNeedles {
			if strings.Contains(lower, needle) {
				return strings.TrimSpace(line)
			}
		}
	}
	return ""
}

func extractAfterPrefix(text string, prefix string) string {
	lowerPrefix := strings.ToLower(prefix)
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		idx := strings.Index(strings.ToLower(trimmed), lowerPrefix)
		if idx < 0 {
			continue
		}
		value := strings.TrimSpace(trimmed[idx+len(prefix):])
		fields := strings.Fields(value)
		if len(fields) > 0 {
			return strings.Trim(fields[0], ":,")
		}
		return value
	}
	return ""
}

func normalizeSignal(signal string) string {
	signal = strings.TrimSpace(signal)
	if signal == "" {
		return ""
	}
	return strings.ToUpper(signal)
}

func basename(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func valueOrUnknown(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func threadSuffix(threadName string) string {
	if strings.TrimSpace(threadName) == "" {
		return ""
	}
	return " on thread " + strings.TrimSpace(threadName)
}

func nonEmpty(values []string) []string {
	var result []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, strings.TrimSpace(value))
		}
	}
	return result
}

func analysisID(result AnalysisResult) string {
	fingerprint := strings.Join([]string{
		string(result.Type),
		result.PackageName,
		fmt.Sprint(result.PID),
		result.ExceptionType,
		result.Signal,
		result.LibraryName,
		result.Reason,
		result.PrimaryMessage,
		result.RawText,
	}, "\x00")
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(fingerprint))
	return fmt.Sprintf("%s-%x", result.Type, hash.Sum64())
}
