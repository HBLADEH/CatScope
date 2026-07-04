package ai

import (
	"fmt"
	"sort"
	"strings"

	"catscope/internal/adb"
	"catscope/internal/logcat"
)

type AIContextOptions struct {
	IncludeDeviceInfo         bool     `json:"includeDeviceInfo"`
	IncludePackageInfo        bool     `json:"includePackageInfo"`
	IncludeAnalysisSummary    bool     `json:"includeAnalysisSummary"`
	IncludeRelatedLogs        bool     `json:"includeRelatedLogs"`
	IncludeBeforeContextLines int      `json:"includeBeforeContextLines"`
	IncludeAfterContextLines  int      `json:"includeAfterContextLines"`
	IncludeRawText            bool     `json:"includeRawText"`
	IncludeSuggestions        bool     `json:"includeSuggestions"`
	Language                  string   `json:"language"`
	PackageFilter             string   `json:"packageFilter,omitempty"`
	LevelFilter               []string `json:"levelFilter,omitempty"`
	SearchKeyword             string   `json:"searchKeyword,omitempty"`
}

type ContextInput struct {
	Device         *adb.AndroidDevice
	Analysis       logcat.AnalysisResult
	Logs           []logcat.LogEntry
	PIDState       logcat.PackagePIDState
	SelectedLog    *logcat.LogEntry
	CurrentPackage string
	Options        AIContextOptions
}

func DefaultOptions() AIContextOptions {
	return AIContextOptions{
		IncludeDeviceInfo:         true,
		IncludePackageInfo:        true,
		IncludeAnalysisSummary:    true,
		IncludeRelatedLogs:        true,
		IncludeBeforeContextLines: 50,
		IncludeAfterContextLines:  50,
		IncludeRawText:            true,
		IncludeSuggestions:        true,
		Language:                  "zh-CN",
	}
}

func GenerateMarkdown(input ContextInput) string {
	options := normalizeOptions(input.Options)
	relatedLogs, contextLogs := collectLogs(input.Logs, input.Analysis.RelatedEntryIDs, input.SelectedLog, options)

	var builder strings.Builder
	if options.Language == "en-US" {
		writeEnglish(&builder, input, relatedLogs, contextLogs, options)
	} else {
		writeChinese(&builder, input, relatedLogs, contextLogs, options)
	}
	return strings.TrimSpace(builder.String()) + "\n"
}

func normalizeOptions(options AIContextOptions) AIContextOptions {
	defaults := DefaultOptions()
	if options.Language == "" {
		options.Language = defaults.Language
	}
	if options.IncludeBeforeContextLines < 0 {
		options.IncludeBeforeContextLines = 0
	}
	if options.IncludeAfterContextLines < 0 {
		options.IncludeAfterContextLines = 0
	}
	if options.IncludeBeforeContextLines == 0 && options.IncludeAfterContextLines == 0 {
		options.IncludeBeforeContextLines = defaults.IncludeBeforeContextLines
		options.IncludeAfterContextLines = defaults.IncludeAfterContextLines
	}
	return options
}

func collectLogs(entries []logcat.LogEntry, relatedIDs []int64, selectedLog *logcat.LogEntry, options AIContextOptions) ([]logcat.LogEntry, []logcat.LogEntry) {
	if len(entries) == 0 {
		return nil, nil
	}

	indexByID := make(map[int64]int, len(entries))
	for i, entry := range entries {
		if entry.ID > 0 {
			indexByID[entry.ID] = i
		}
	}

	indices := make([]int, 0, len(relatedIDs)+1)
	for _, id := range relatedIDs {
		if index, ok := indexByID[id]; ok {
			indices = append(indices, index)
		}
	}
	if len(indices) == 0 && selectedLog != nil && selectedLog.ID > 0 {
		if index, ok := indexByID[selectedLog.ID]; ok {
			indices = append(indices, index)
		}
	}
	if len(indices) == 0 {
		return nil, nil
	}
	sort.Ints(indices)

	related := entriesByIndices(entries, indices)
	seenContext := map[int]bool{}
	for _, index := range indices {
		start := index - options.IncludeBeforeContextLines
		if start < 0 {
			start = 0
		}
		end := index + options.IncludeAfterContextLines
		if end >= len(entries) {
			end = len(entries) - 1
		}
		for i := start; i <= end; i++ {
			seenContext[i] = true
		}
	}

	contextIndices := make([]int, 0, len(seenContext))
	for index := range seenContext {
		contextIndices = append(contextIndices, index)
	}
	sort.Ints(contextIndices)
	return related, entriesByIndices(entries, contextIndices)
}

func entriesByIndices(entries []logcat.LogEntry, indices []int) []logcat.LogEntry {
	seen := map[int]bool{}
	result := make([]logcat.LogEntry, 0, len(indices))
	for _, index := range indices {
		if index < 0 || index >= len(entries) || seen[index] {
			continue
		}
		seen[index] = true
		result = append(result, entries[index])
	}
	return result
}

func writeChinese(builder *strings.Builder, input ContextInput, relatedLogs []logcat.LogEntry, contextLogs []logcat.LogEntry, options AIContextOptions) {
	builder.WriteString("# Android Logcat 问题分析请求\n\n")
	writeBasicInfo(builder, input, "## 1. 基本信息", options)
	writeAnalysisSummary(builder, input.Analysis, "## 2. 问题摘要", options)
	writeFrames(builder, input.Analysis.KeyFrames, "## 3. 关键帧")
	writeLogs(builder, relatedLogs, input.Analysis.RawText, "## 4. 关联日志", options.IncludeRelatedLogs, options.IncludeRawText)
	writeLogs(builder, contextLogs, "", "## 5. 上下文日志", true, false)
	writeSuggestions(builder, input.Analysis.Suggestions, "## 6. CatScope 初步建议", options)
	builder.WriteString("## 7. 希望 AI 帮助分析的问题\n\n")
	builder.WriteString("请基于以上 Android Logcat 日志，分析：\n\n")
	builder.WriteString("1. 最可能的根因是什么？\n")
	builder.WriteString("2. 应该优先检查哪些代码位置？\n")
	builder.WriteString("3. 如果是 Native/JNI 问题，下一步应该如何定位？\n")
	builder.WriteString("4. 有哪些可执行的修复建议？\n")
}

func writeEnglish(builder *strings.Builder, input ContextInput, relatedLogs []logcat.LogEntry, contextLogs []logcat.LogEntry, options AIContextOptions) {
	builder.WriteString("# Android Logcat Analysis Request\n\n")
	writeBasicInfo(builder, input, "## 1. Basic Information", options)
	writeAnalysisSummary(builder, input.Analysis, "## 2. Issue Summary", options)
	writeFrames(builder, input.Analysis.KeyFrames, "## 3. Key Frames")
	writeLogs(builder, relatedLogs, input.Analysis.RawText, "## 4. Related Logs", options.IncludeRelatedLogs, options.IncludeRawText)
	writeLogs(builder, contextLogs, "", "## 5. Context Logs", true, false)
	writeSuggestions(builder, input.Analysis.Suggestions, "## 6. CatScope Initial Suggestions", options)
	builder.WriteString("## 7. Questions For AI Analysis\n\n")
	builder.WriteString("Please analyze the Android Logcat logs above and answer:\n\n")
	builder.WriteString("1. What is the most likely root cause?\n")
	builder.WriteString("2. Which code locations should be checked first?\n")
	builder.WriteString("3. If this is a Native/JNI issue, what should the next debugging step be?\n")
	builder.WriteString("4. What concrete fixes or mitigations are recommended?\n")
}

func writeBasicInfo(builder *strings.Builder, input ContextInput, title string, options AIContextOptions) {
	builder.WriteString(title + "\n\n")
	if options.IncludeDeviceInfo {
		if input.Device != nil {
			builder.WriteString(fmt.Sprintf("- Device: %s\n", dash(input.Device.Model)))
			builder.WriteString(fmt.Sprintf("- Android: %s\n", dash(input.Device.AndroidVersion)))
			builder.WriteString(fmt.Sprintf("- SDK: %s\n", dash(input.Device.SDKVersion)))
			builder.WriteString(fmt.Sprintf("- ABI: %s\n", dash(input.Device.ABI)))
			builder.WriteString(fmt.Sprintf("- Serial: %s\n", dash(input.Device.Serial)))
		} else {
			builder.WriteString("- Device: -\n- Android: -\n- SDK: -\n- ABI: -\n- Serial: -\n")
		}
	}
	if options.IncludePackageInfo {
		packageName := firstNonEmpty(input.Analysis.PackageName, input.CurrentPackage, input.PIDState.PackageName)
		builder.WriteString(fmt.Sprintf("- Package: %s\n", dash(packageName)))
		builder.WriteString(fmt.Sprintf("- Current PID: %s\n", formatInts(input.PIDState.PIDs)))
		builder.WriteString(fmt.Sprintf("- Known PIDs: %s\n", formatInts(input.PIDState.KnownPIDs)))
		builder.WriteString(fmt.Sprintf("- Package Filter: %s\n", dash(options.PackageFilter)))
		builder.WriteString(fmt.Sprintf("- Level Filter: %s\n", formatStrings(options.LevelFilter)))
		builder.WriteString(fmt.Sprintf("- Search: %s\n", dash(options.SearchKeyword)))
	}
	builder.WriteString("\n")
}

func writeAnalysisSummary(builder *strings.Builder, analysis logcat.AnalysisResult, title string, options AIContextOptions) {
	if !options.IncludeAnalysisSummary {
		return
	}
	builder.WriteString(title + "\n\n")
	builder.WriteString(fmt.Sprintf("- Type: %s\n", analysis.Type))
	builder.WriteString(fmt.Sprintf("- Severity: %s\n", analysis.Severity))
	builder.WriteString(fmt.Sprintf("- Title: %s\n", dash(analysis.Title)))
	builder.WriteString(fmt.Sprintf("- Summary: %s\n", dash(analysis.Summary)))
	builder.WriteString(fmt.Sprintf("- Exception: %s\n", dash(analysis.ExceptionType)))
	builder.WriteString(fmt.Sprintf("- Thread: %s\n", dash(analysis.ThreadName)))
	builder.WriteString(fmt.Sprintf("- Signal: %s\n", dash(analysis.Signal)))
	builder.WriteString(fmt.Sprintf("- Library: %s\n", dash(analysis.LibraryName)))
	builder.WriteString(fmt.Sprintf("- Reason: %s\n\n", dash(analysis.Reason)))
}

func writeFrames(builder *strings.Builder, frames []string, title string) {
	builder.WriteString(title + "\n\n")
	builder.WriteString("```text\n")
	if len(frames) == 0 {
		builder.WriteString("-\n")
	} else {
		builder.WriteString(strings.Join(frames, "\n"))
		builder.WriteByte('\n')
	}
	builder.WriteString("```\n\n")
}

func writeLogs(builder *strings.Builder, entries []logcat.LogEntry, fallbackRaw string, title string, includeLogs bool, includeRaw bool) {
	builder.WriteString(title + "\n\n")
	builder.WriteString("```log\n")
	switch {
	case includeLogs && len(entries) > 0:
		builder.WriteString(FormatLogEntries(entries))
	case includeRaw && strings.TrimSpace(fallbackRaw) != "":
		builder.WriteString(strings.TrimSpace(fallbackRaw))
		builder.WriteByte('\n')
	default:
		builder.WriteString("-\n")
	}
	builder.WriteString("```\n\n")
}

func writeSuggestions(builder *strings.Builder, suggestions []string, title string, options AIContextOptions) {
	if !options.IncludeSuggestions {
		return
	}
	builder.WriteString(title + "\n\n")
	if len(suggestions) == 0 {
		builder.WriteString("- -\n\n")
		return
	}
	for _, suggestion := range suggestions {
		builder.WriteString("- " + suggestion + "\n")
	}
	builder.WriteString("\n")
}

func FormatLogEntries(entries []logcat.LogEntry) string {
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString(formatLogEntry(entry))
		builder.WriteByte('\n')
		for _, line := range entry.Multiline {
			builder.WriteString(line)
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

func formatLogEntry(entry logcat.LogEntry) string {
	if strings.TrimSpace(entry.Raw) != "" {
		return entry.Raw
	}
	return fmt.Sprintf("%s %5d %5d %s %s: %s",
		dash(entry.Timestamp),
		entry.PID,
		entry.TID,
		dash(entry.Level),
		dash(entry.Tag),
		entry.Message,
	)
}

func dash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func formatInts(values []int) string {
	if len(values) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprint(value))
	}
	return strings.Join(parts, ", ")
}

func formatStrings(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}
