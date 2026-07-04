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
	exceptionPattern   = regexp.MustCompile(`(?m)((?:java|kotlin|android|dalvik|libcore)\.[A-Za-z0-9_.$]+(?:Exception|Error)|[A-Za-z0-9_.$]+(?:Exception|Error))(?::|\s|$)`)
	threadPattern      = regexp.MustCompile(`FATAL EXCEPTION:\s*([^\n\r]+)`)
	processPattern     = regexp.MustCompile(`Process:\s*([A-Za-z0-9_.]+)\s*,\s*PID:\s*(\d+)`)
	signalPattern      = regexp.MustCompile(`(?i)\b(SIGSEGV|SIGABRT|signal\s+\d+)\b`)
	libraryPattern     = regexp.MustCompile(`\b([A-Za-z0-9_./-]*lib[A-Za-z0-9_.+-]+\.so)\b`)
	installCodePattern = regexp.MustCompile(`\b((?:INSTALL_FAILED|INSTALL_PARSE_FAILED|DELETE_FAILED)_[A-Z0-9_]+)\b`)
)

type installErrorRule struct {
	Code        string
	Severity    AnalysisSeverity
	Title       string
	Summary     string
	Reason      string
	Suggestions []string
}

var installErrorRules = map[string]installErrorRule{
	"INSTALL_FAILED_VERSION_DOWNGRADE": {
		Code:     "INSTALL_FAILED_VERSION_DOWNGRADE",
		Severity: SeverityError,
		Title:    "Install failed: version downgrade",
		Summary:  "The device already has the app installed with a higher versionCode.",
		Reason:   "Device already has a higher versionCode installed. / 设备上已有更高 versionCode 的应用。",
		Suggestions: []string{
			"中文: 卸载设备上的旧包，或在确认安全时使用 adb install -d 允许降级安装。",
			"English: Uninstall the existing package, or use adb install -d when a downgrade is intentional.",
		},
	},
	"INSTALL_FAILED_UPDATE_INCOMPATIBLE": {
		Code:     "INSTALL_FAILED_UPDATE_INCOMPATIBLE",
		Severity: SeverityError,
		Title:    "Install failed: update incompatible",
		Summary:  "An existing package with the same name cannot be updated by this APK.",
		Reason:   "The installed app likely has a different signature or came from a different build/source. / 已安装同包名应用可能签名不一致或来源不同。",
		Suggestions: []string{
			"中文: 卸载旧包后重装，并确认 debug/release 签名、applicationId 和渠道来源一致。",
			"English: Uninstall the old package, then verify debug/release signing, applicationId, and install source consistency.",
		},
	},
	"INSTALL_FAILED_NO_MATCHING_ABIS": {
		Code:     "INSTALL_FAILED_NO_MATCHING_ABIS",
		Severity: SeverityFatal,
		Title:    "Install failed: no matching ABIs",
		Summary:  "The APK does not contain native libraries compatible with the device ABI.",
		Reason:   "APK does not include a native ABI supported by the device. / APK 不包含当前设备支持的 native ABI。",
		Suggestions: []string{
			"中文: 检查 arm64-v8a、armeabi-v7a、x86_64 等 ABI 构建配置和 splits 配置。",
			"English: Check arm64-v8a, armeabi-v7a, x86_64 ABI outputs and split APK configuration.",
		},
	},
	"INSTALL_FAILED_INVALID_APK": {
		Code:     "INSTALL_FAILED_INVALID_APK",
		Severity: SeverityError,
		Title:    "Install failed: invalid APK",
		Summary:  "The APK file is malformed, incomplete, or not installable.",
		Reason:   "The APK is invalid or corrupted. / APK 文件无效、损坏或不符合安装要求。",
		Suggestions: []string{
			"中文: 重新构建 APK，确认文件未损坏，并用 apksigner / aapt 检查包结构。",
			"English: Rebuild the APK, verify the file is not corrupted, and inspect it with apksigner or aapt.",
		},
	},
	"INSTALL_PARSE_FAILED_NO_CERTIFICATES": {
		Code:     "INSTALL_PARSE_FAILED_NO_CERTIFICATES",
		Severity: SeverityFatal,
		Title:    "Install parse failed: no certificates",
		Summary:  "Android could not find valid signing certificates in the APK.",
		Reason:   "The APK is unsigned or its certificates cannot be parsed. / APK 未签名或签名证书无法解析。",
		Suggestions: []string{
			"中文: 使用正确 keystore 重新签名，并用 apksigner verify 检查签名。",
			"English: Sign the APK with the correct keystore and verify it with apksigner verify.",
		},
	},
	"INSTALL_PARSE_FAILED_MANIFEST_MALFORMED": {
		Code:     "INSTALL_PARSE_FAILED_MANIFEST_MALFORMED",
		Severity: SeverityError,
		Title:    "Install parse failed: malformed manifest",
		Summary:  "The APK manifest is malformed or contains unsupported values.",
		Reason:   "AndroidManifest.xml cannot be parsed correctly. / AndroidManifest.xml 格式错误或包含不兼容配置。",
		Suggestions: []string{
			"中文: 检查 manifest 合并结果、组件声明、权限、provider authorities 和 min/target SDK 配置。",
			"English: Inspect the merged manifest, component declarations, permissions, provider authorities, and SDK settings.",
		},
	},
	"INSTALL_FAILED_INSUFFICIENT_STORAGE": {
		Code:     "INSTALL_FAILED_INSUFFICIENT_STORAGE",
		Severity: SeverityWarning,
		Title:    "Install failed: insufficient storage",
		Summary:  "The device does not have enough available storage for installation.",
		Reason:   "Device storage is insufficient. / 设备可用存储空间不足。",
		Suggestions: []string{
			"中文: 清理设备空间、卸载无关应用，或检查工作资料/多用户空间占用。",
			"English: Free device storage, uninstall unused apps, or inspect work profile and multi-user storage usage.",
		},
	},
	"INSTALL_FAILED_ALREADY_EXISTS": {
		Code:     "INSTALL_FAILED_ALREADY_EXISTS",
		Severity: SeverityWarning,
		Title:    "Install failed: already exists",
		Summary:  "The package already exists and the install command did not allow replacement.",
		Reason:   "The package is already installed. / 设备上已存在该包，当前安装命令未允许覆盖。",
		Suggestions: []string{
			"中文: 使用 adb install -r 覆盖安装，或先卸载已有包。",
			"English: Use adb install -r to replace the existing package, or uninstall it first.",
		},
	},
	"INSTALL_FAILED_MISSING_SHARED_LIBRARY": {
		Code:     "INSTALL_FAILED_MISSING_SHARED_LIBRARY",
		Severity: SeverityFatal,
		Title:    "Install failed: missing shared library",
		Summary:  "The APK requires a shared library that is missing on the device.",
		Reason:   "Required shared library is not available on the target device. / 目标设备缺少应用声明依赖的共享库。",
		Suggestions: []string{
			"中文: 检查 uses-library 声明，确认设备系统镜像包含所需库，或将非必需库标记为 required=false。",
			"English: Check uses-library declarations, verify the system image provides the library, or mark optional libraries required=false.",
		},
	},
	"INSTALL_FAILED_CPU_ABI_INCOMPATIBLE": {
		Code:     "INSTALL_FAILED_CPU_ABI_INCOMPATIBLE",
		Severity: SeverityFatal,
		Title:    "Install failed: CPU ABI incompatible",
		Summary:  "The APK native code is incompatible with the device CPU ABI.",
		Reason:   "APK native binaries do not match the device CPU ABI. / APK native 二进制与设备 CPU ABI 不兼容。",
		Suggestions: []string{
			"中文: 检查 ndk.abiFilters、splits.abi 和产物中 lib/ 目录，确保包含设备 ABI。",
			"English: Check ndk.abiFilters, splits.abi, and the APK lib/ directories to include the device ABI.",
		},
	},
	"INSTALL_FAILED_TEST_ONLY": {
		Code:     "INSTALL_FAILED_TEST_ONLY",
		Severity: SeverityWarning,
		Title:    "Install failed: test only",
		Summary:  "The APK is marked as testOnly and adb install did not allow test packages.",
		Reason:   "The manifest has testOnly=true. / APK manifest 带有 testOnly=true。",
		Suggestions: []string{
			"中文: 使用 adb install -t 安装测试包，或调整构建配置去掉 testOnly。",
			"English: Use adb install -t for test-only APKs, or adjust the build configuration to remove testOnly.",
		},
	},
	"DELETE_FAILED_INTERNAL_ERROR": {
		Code:     "DELETE_FAILED_INTERNAL_ERROR",
		Severity: SeverityError,
		Title:    "Delete failed: internal error",
		Summary:  "Package deletion failed because the package manager reported an internal error.",
		Reason:   "Package manager failed internally while deleting the package. / PackageManager 删除应用时发生内部错误。",
		Suggestions: []string{
			"中文: 重试卸载，必要时重启设备或 adb server，并检查多用户/工作资料中是否仍安装该包。",
			"English: Retry uninstall, restart the device or adb server if needed, and check other users or work profiles for the package.",
		},
	},
}

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

func AnalyzeText(text string) []AnalysisResult {
	return AnalyzeInstallOutput(text)
}

func AnalyzeInstallOutput(output string) []AnalysisResult {
	output = strings.TrimSpace(output)
	if output == "" || !looksLikeInstallOutput(strings.ToLower(output)) {
		return nil
	}

	result := analyzeInstallText(output)
	if strings.TrimSpace(result.ID) == "" {
		return nil
	}
	return []AnalysisResult{result}
}

func analyzeOne(entry LogEntry, context []LogEntry) (AnalysisResult, bool) {
	rawText := joinEntries(context)
	lower := strings.ToLower(rawText)

	switch {
	case looksLikeInstallOutput(lower):
		return analyzeInstallError(entry, context, rawText), true
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

func analyzeInstallError(entry LogEntry, context []LogEntry, rawText string) AnalysisResult {
	result := buildInstallResult(entry, context, rawText)
	return result
}

func analyzeInstallText(rawText string) AnalysisResult {
	entry := LogEntry{
		Level:   "E",
		Tag:     "PackageInstaller",
		Message: primaryInstallLine(rawText),
		Raw:     primaryInstallLine(rawText),
	}
	return buildInstallResult(entry, nil, rawText)
}

func buildInstallResult(entry LogEntry, context []LogEntry, rawText string) AnalysisResult {
	code := extractInstallCode(rawText)
	rule := installRuleForCode(code)
	if strings.TrimSpace(entry.Message) == "" {
		entry.Message = primaryInstallLine(rawText)
	}
	if strings.TrimSpace(entry.Raw) == "" {
		entry.Raw = entry.Message
	}

	return resultBase(AnalysisTypeInstallError, rule.Severity, entry, context, rawText, rule.Title, rule.Summary, func(result *AnalysisResult) {
		result.Reason = rule.Reason
		result.KeyFrames = nonEmpty([]string{code, primaryInstallLine(rawText)})
		result.RawText = strings.TrimSpace(rawText)
		result.PrimaryMessage = primaryInstallLine(rawText)
		result.Suggestions = append([]string(nil), rule.Suggestions...)
	})
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

func looksLikeInstallOutput(lower string) bool {
	return strings.Contains(lower, "install_failed_") ||
		strings.Contains(lower, "install_parse_failed_") ||
		strings.Contains(lower, "delete_failed_") ||
		strings.Contains(lower, "failure [install_failed") ||
		strings.Contains(lower, "adb: failed to install")
}

func extractInstallCode(text string) string {
	return strings.TrimSpace(firstSubmatch(installCodePattern, text, 1))
}

func installRuleForCode(code string) installErrorRule {
	code = strings.TrimSpace(code)
	if rule, ok := installErrorRules[code]; ok {
		return rule
	}
	if strings.HasPrefix(code, "INSTALL_FAILED_") || strings.HasPrefix(code, "INSTALL_PARSE_FAILED_") || strings.HasPrefix(code, "DELETE_FAILED_") {
		return installErrorRule{
			Code:     code,
			Severity: SeverityError,
			Title:    "Install failed: " + code,
			Summary:  "Android package installation failed with a recognized package manager error code.",
			Reason:   code + " was reported by adb or PackageManager. / adb 或 PackageManager 返回了该安装错误码。",
			Suggestions: []string{
				"中文: 根据错误码检查签名、versionCode、ABI、manifest、存储空间和设备系统兼容性。",
				"English: Use the error code to inspect signing, versionCode, ABI, manifest, storage, and device compatibility.",
			},
		}
	}
	return installErrorRule{
		Code:     "ADB_FAILED_TO_INSTALL",
		Severity: SeverityError,
		Title:    "Install failed",
		Summary:  "adb reported that the APK installation failed.",
		Reason:   "adb failed to install the APK but did not expose a specific INSTALL_FAILED code. / adb 安装失败，但输出中没有明确的 INSTALL_FAILED 错误码。",
		Suggestions: []string{
			"中文: 查看完整 adb 输出和设备 logcat，确认 APK 路径、设备连接、签名、ABI、manifest 和剩余空间。",
			"English: Review the full adb output and device logcat, then check APK path, device connection, signing, ABI, manifest, and storage.",
		},
	}
}

func primaryInstallLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if looksLikeInstallOutput(lower) {
			return trimmed
		}
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) != "" {
			return strings.TrimSpace(line)
		}
	}
	return strings.TrimSpace(text)
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
	case "AndroidRuntime", "DEBUG", "ActivityManager", "PackageInstaller", "PackageManager", "PackageManagerService", "libc":
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
		strings.Contains(lower, "jni") ||
		looksLikeInstallOutput(lower)
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
