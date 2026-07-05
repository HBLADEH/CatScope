package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"catscope/internal/adb"
	"catscope/internal/ai"
	"catscope/internal/build"
	"catscope/internal/logcat"
	"catscope/internal/workspace"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context

	mu        sync.Mutex
	adbPath   string
	stream    *adb.LogcatProcess
	lastErr   string
	serial    string
	running   bool
	streamID  int64
	parser    *logcat.Parser
	logStore  *logcat.RingBuffer
	pidMap    *logcat.PIDMapper
	pidCancel context.CancelFunc

	analysisMu sync.RWMutex
	analysis   map[string]logcat.AnalysisResult
}

type BuildInstallLaunchResult struct {
	Build           build.BuildResult       `json:"build"`
	Install         adb.InstallResult       `json:"install"`
	Launch          adb.LaunchResult        `json:"launch"`
	PackageName     string                  `json:"packageName"`
	APK             *build.APKInfo          `json:"apk,omitempty"`
	AnalysisResults []logcat.AnalysisResult `json:"analysisResults,omitempty"`
}

func NewApp() *App {
	return &App{
		parser:   logcat.NewParser(),
		logStore: logcat.NewRingBuffer(100000),
		pidMap:   logcat.NewPIDMapper(),
		analysis: map[string]logcat.AnalysisResult{},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) FindADB(configuredPath string) (string, error) {
	path, err := adb.FindADB(configuredPath)
	if err != nil {
		a.setLastError(err.Error())
		return "", err
	}

	a.mu.Lock()
	a.adbPath = path
	a.mu.Unlock()

	return path, nil
}

func (a *App) ListDevices() ([]adb.AndroidDevice, error) {
	adbPath, err := a.ensureADB("")
	if err != nil {
		return nil, err
	}

	devices, err := adb.ListDevices(a.context(), adbPath)
	if err != nil {
		a.setLastError(err.Error())
		return nil, err
	}
	a.setLastError("")
	return devices, nil
}

func (a *App) GetDeviceInfo(serial string) (adb.AndroidDevice, error) {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return adb.AndroidDevice{}, errors.New("device serial is required")
	}

	adbPath, err := a.ensureADB("")
	if err != nil {
		return adb.AndroidDevice{}, err
	}

	device, err := adb.GetDeviceInfo(a.context(), adbPath, serial)
	if err != nil {
		a.setLastError(err.Error())
		return adb.AndroidDevice{}, err
	}
	a.setLastError("")
	return device, nil
}

func (a *App) SetActiveDevice(serial string) {
	a.mu.Lock()
	a.serial = strings.TrimSpace(serial)
	a.mu.Unlock()
	a.updateActiveWorkspace(func(config workspace.WorkspaceConfig) workspace.WorkspaceConfig {
		config.SelectedDeviceSerial = strings.TrimSpace(serial)
		return config
	})
}

func (a *App) ListPackages(serial string, mode string) ([]adb.InstalledPackage, error) {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return nil, errors.New("device serial is required")
	}

	adbPath, err := a.ensureADB("")
	if err != nil {
		return nil, err
	}

	packages, err := adb.ListPackages(a.context(), adbPath, serial, adb.PackageListMode(mode))
	if err != nil {
		a.setLastError(err.Error())
		return nil, err
	}
	a.setLastError("")
	return packages, nil
}

func (a *App) SetTrackedPackage(serial string, packageName string) error {
	serial = strings.TrimSpace(serial)
	packageName = strings.TrimSpace(packageName)

	a.stopPIDTracker()
	if packageName == "" {
		a.pidMap.Clear()
		return nil
	}
	if serial == "" {
		return errors.New("device serial is required")
	}

	a.pidMap.SetPackage(packageName)

	a.mu.Lock()
	running := a.running
	activeSerial := a.serial
	a.mu.Unlock()

	if running && activeSerial == serial {
		return a.startPIDTracker(serial, packageName)
	}
	return nil
}

func (a *App) GetPackagePIDState() logcat.PackagePIDState {
	return a.pidMap.State()
}

func (a *App) StartLogcat(serial string) error {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return errors.New("select a connected Android device first")
	}

	adbPath, err := a.ensureADB("")
	if err != nil {
		return err
	}

	state, err := adb.GetDeviceState(a.context(), adbPath, serial)
	if err != nil {
		a.setLastError(err.Error())
		return err
	}
	if state != "device" {
		message := deviceStateError(serial, state)
		a.setLastError(message)
		return errors.New(message)
	}

	_ = a.StopLogcat()
	a.logStore.Clear()
	a.parser.Reset()

	a.mu.Lock()
	a.streamID++
	streamID := a.streamID
	a.mu.Unlock()

	proc, err := adb.StartLogcat(a.context(), adbPath, serial, a.ingestLogLine, a.setLastError, func(err error) {
		a.markStopped(streamID, err)
	})
	if err != nil {
		a.setLastError(err.Error())
		return err
	}

	a.mu.Lock()
	a.stream = proc
	a.serial = serial
	a.running = true
	a.lastErr = ""
	trackedPackage := a.pidMap.State().PackageName
	a.mu.Unlock()

	if trackedPackage != "" {
		if err := a.startPIDTracker(serial, trackedPackage); err != nil {
			a.setLastError(err.Error())
		}
	}

	return nil
}

func (a *App) StopLogcat() error {
	a.mu.Lock()
	proc := a.stream
	a.stream = nil
	a.running = false
	a.mu.Unlock()

	a.stopPIDTracker()

	if proc != nil {
		return proc.Stop()
	}
	return nil
}

func (a *App) ExportLogs(entries []logcat.LogEntry) (string, error) {
	if len(entries) == 0 {
		return "", errors.New("no logs to export")
	}

	defaultName := "catscope-logcat.txt"
	path, err := wailsruntime.SaveFileDialog(a.context(), wailsruntime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Text Log (*.txt)",
				Pattern:     "*.txt",
			},
		},
	})
	if err != nil {
		a.setLastError(err.Error())
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", errors.New("export canceled")
	}
	if filepath.Ext(path) == "" {
		path += ".txt"
	}

	content := logcat.FormatEntriesText(entries)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		a.setLastError(err.Error())
		return "", fmt.Errorf("export logs failed: %w", err)
	}
	a.setLastError("")
	return path, nil
}

func (a *App) SelectProjectDirectory() (string, error) {
	path, err := wailsruntime.OpenDirectoryDialog(a.context(), wailsruntime.OpenDialogOptions{
		Title: "Select Android Project",
	})
	if err != nil {
		a.setLastError(err.Error())
		return "", err
	}
	return strings.TrimSpace(path), nil
}

func (a *App) GetProjectConfig() (workspace.ProjectConfig, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return workspace.ProjectConfig{}, err
	}
	return workspace.ProjectFromWorkspace(workspace.ActiveWorkspace(config)), nil
}

func (a *App) SaveProjectConfig(config workspace.ProjectConfig) error {
	return a.updateActiveWorkspace(func(current workspace.WorkspaceConfig) workspace.WorkspaceConfig {
		return workspace.WorkspaceFromProject(current, config)
	})
}

func (a *App) LoadConfig() (workspace.AppConfig, error) {
	config, err := workspace.LoadConfig(workspace.DefaultConfigPath())
	if err != nil {
		a.setLastError(err.Error())
		return workspace.AppConfig{}, err
	}
	a.setLastError("")
	return config, nil
}

func (a *App) SaveConfig(config workspace.AppConfig) error {
	if err := workspace.SaveConfig(workspace.DefaultConfigPath(), config); err != nil {
		a.setLastError(err.Error())
		return err
	}
	a.setLastError("")
	return nil
}

func (a *App) ResetConfig() (workspace.AppConfig, error) {
	config := workspace.DefaultAppConfig()
	if err := a.SaveConfig(config); err != nil {
		return workspace.AppConfig{}, err
	}
	return config, nil
}

func (a *App) ListWorkspaces() ([]workspace.WorkspaceConfig, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return nil, err
	}
	return config.Workspaces, nil
}

func (a *App) SaveWorkspace(next workspace.WorkspaceConfig) (workspace.AppConfig, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return workspace.AppConfig{}, err
	}
	config = workspace.SaveWorkspace(config, next)
	if err := a.SaveConfig(config); err != nil {
		return workspace.AppConfig{}, err
	}
	return config, nil
}

func (a *App) DeleteWorkspace(id string) (workspace.AppConfig, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return workspace.AppConfig{}, err
	}
	config = workspace.DeleteWorkspace(config, id)
	if err := a.SaveConfig(config); err != nil {
		return workspace.AppConfig{}, err
	}
	return config, nil
}

func (a *App) SetActiveWorkspace(id string) (workspace.AppConfig, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return workspace.AppConfig{}, err
	}
	config = workspace.SetActiveWorkspace(config, id)
	if err := a.SaveConfig(config); err != nil {
		return workspace.AppConfig{}, err
	}
	return config, nil
}

func (a *App) ListFilterPresets() ([]workspace.FilterPreset, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return nil, err
	}
	return config.FilterPresets, nil
}

func (a *App) SaveFilterPreset(preset workspace.FilterPreset) (workspace.AppConfig, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return workspace.AppConfig{}, err
	}
	config = workspace.SaveFilterPreset(config, preset)
	if err := a.SaveConfig(config); err != nil {
		return workspace.AppConfig{}, err
	}
	return config, nil
}

func (a *App) DeleteFilterPreset(id string) (workspace.AppConfig, error) {
	config, err := a.LoadConfig()
	if err != nil {
		return workspace.AppConfig{}, err
	}
	config = workspace.DeleteFilterPreset(config, id)
	if err := a.SaveConfig(config); err != nil {
		return workspace.AppConfig{}, err
	}
	return config, nil
}

func (a *App) BuildDebug(projectPath string) (build.BuildResult, error) {
	result, err := build.RunDebugBuild(a.context(), projectPath)
	if err != nil {
		a.setLastError(err.Error())
		return build.BuildResult{}, err
	}
	a.updateProjectConfig(func(config workspace.ProjectConfig) workspace.ProjectConfig {
		config.ProjectPath = result.ProjectPath
		config.DefaultBuildTask = result.Task
		if result.APK != nil {
			config.LastAPKPath = result.APK.APKPath
		}
		return config
	})
	if result.Error != "" {
		a.setLastError(result.Error)
	} else {
		a.setLastError("")
	}
	return result, nil
}

func (a *App) FindLatestAPK(projectPath string) (build.APKInfo, error) {
	apk, err := build.FindLatestAPK(projectPath)
	if err != nil {
		a.setLastError(err.Error())
		return build.APKInfo{}, err
	}
	a.updateProjectConfig(func(config workspace.ProjectConfig) workspace.ProjectConfig {
		config.ProjectPath = strings.TrimSpace(projectPath)
		config.LastAPKPath = apk.APKPath
		return config
	})
	a.setLastError("")
	return apk, nil
}

func (a *App) InstallAPK(apkPath string, options adb.InstallOptions) (adb.InstallResult, error) {
	adbPath, serial, err := a.adbAndSerial()
	if err != nil {
		return adb.InstallResult{}, err
	}
	result, err := adb.InstallAPK(a.context(), adbPath, serial, apkPath, options)
	if err != nil {
		a.setLastError(err.Error())
		return adb.InstallResult{}, err
	}
	if len(result.AnalysisResults) > 0 {
		a.storeAnalysisResults(result.AnalysisResults)
	}
	a.updateProjectConfig(func(config workspace.ProjectConfig) workspace.ProjectConfig {
		config.LastAPKPath = result.APKPath
		config.InstallOptions = options
		return config
	})
	if result.Error != "" {
		a.setLastError(result.Error)
	} else {
		a.setLastError("")
	}
	return result, nil
}

func (a *App) LaunchApp(packageName string) (adb.LaunchResult, error) {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		packageName = a.pidMap.State().PackageName
	}
	adbPath, serial, err := a.adbAndSerial()
	if err != nil {
		return adb.LaunchResult{}, err
	}
	result, err := adb.LaunchApp(a.context(), adbPath, serial, packageName)
	if err != nil {
		a.setLastError(err.Error())
		return adb.LaunchResult{}, err
	}
	if result.Success {
		_ = a.SetTrackedPackage(serial, packageName)
		a.updateProjectConfig(func(config workspace.ProjectConfig) workspace.ProjectConfig {
			config.PackageName = packageName
			return config
		})
	}
	if result.Error != "" {
		a.setLastError(result.Error)
	} else {
		a.setLastError("")
	}
	return result, nil
}

func (a *App) BuildInstallLaunch(config workspace.ProjectConfig) (BuildInstallLaunchResult, error) {
	config = workspace.NormalizeProjectConfig(config)
	if config.PackageName == "" {
		config.PackageName = a.pidMap.State().PackageName
	}
	buildResult, err := build.Run(a.context(), build.BuildRequest{
		ProjectPath: config.ProjectPath,
		Task:        config.DefaultBuildTask,
	})
	if err != nil {
		a.setLastError(err.Error())
		return BuildInstallLaunchResult{}, err
	}
	result := BuildInstallLaunchResult{
		Build:       buildResult,
		PackageName: config.PackageName,
		APK:         buildResult.APK,
	}
	if !buildResult.Success || buildResult.APK == nil {
		if buildResult.Error != "" {
			a.setLastError(buildResult.Error)
		}
		return result, nil
	}

	installResult, err := a.InstallAPK(buildResult.APK.APKPath, config.InstallOptions)
	if err != nil {
		return result, err
	}
	result.Install = installResult
	result.AnalysisResults = installResult.AnalysisResults
	if !installResult.Success {
		return result, nil
	}

	if strings.TrimSpace(config.PackageName) == "" {
		result.Install.Error = "package name is required to launch after install"
		a.setLastError(result.Install.Error)
		return result, nil
	}
	launchResult, err := a.LaunchApp(config.PackageName)
	if err != nil {
		return result, err
	}
	result.Launch = launchResult

	config.LastAPKPath = buildResult.APK.APKPath
	_ = a.SaveProjectConfig(config)
	return result, nil
}

func (a *App) AnalyzeLogs(entries []logcat.LogEntry) []logcat.AnalysisResult {
	results := logcat.AnalyzeEntries(entries)
	a.storeAnalysisResults(results)
	return results
}

func (a *App) GenerateAIContext(resultID string, options ai.AIContextOptions) (string, error) {
	result, err := a.findAnalysisResult(resultID)
	if err != nil {
		return "", err
	}

	input := ai.ContextInput{
		Device:         a.currentDeviceInfo(),
		Analysis:       result,
		Logs:           a.logStore.Snapshot(),
		PIDState:       a.pidMap.State(),
		CurrentPackage: a.pidMap.State().PackageName,
		Options:        options,
	}

	return ai.GenerateMarkdown(input), nil
}

func (a *App) CopyAIContext(resultID string, options ai.AIContextOptions) error {
	markdown, err := a.GenerateAIContext(resultID, options)
	if err != nil {
		return err
	}
	if err := wailsruntime.ClipboardSetText(a.context(), markdown); err != nil {
		a.setLastError(err.Error())
		return err
	}
	a.setLastError("")
	return nil
}

func (a *App) ExportAIContext(resultID string, options ai.AIContextOptions) (string, error) {
	markdown, err := a.GenerateAIContext(resultID, options)
	if err != nil {
		return "", err
	}

	path, err := wailsruntime.SaveFileDialog(a.context(), wailsruntime.SaveDialogOptions{
		DefaultFilename: "catscope-ai-context.md",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "Markdown (*.md)",
				Pattern:     "*.md",
			},
		},
	})
	if err != nil {
		a.setLastError(err.Error())
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", errors.New("export canceled")
	}
	if filepath.Ext(path) == "" {
		path += ".md"
	}
	if err := os.WriteFile(path, []byte(markdown), 0644); err != nil {
		a.setLastError(err.Error())
		return "", fmt.Errorf("export AI context failed: %w", err)
	}
	a.setLastError("")
	return path, nil
}

func (a *App) ClearLogs() {
	a.logStore.Clear()
	a.parser.Reset()
	a.analysisMu.Lock()
	a.analysis = map[string]logcat.AnalysisResult{}
	a.analysisMu.Unlock()
}

func (a *App) GetLogBatch(afterID int64, limit int) logcat.LogBatch {
	return a.logStore.GetAfter(afterID, limit)
}

func (a *App) GetLogStatus() logcat.LogStatus {
	count, discarded, lastID := a.logStore.Stats()

	a.mu.Lock()
	defer a.mu.Unlock()

	return logcat.LogStatus{
		Running:        a.running,
		Serial:         a.serial,
		LastError:      a.lastErr,
		Count:          count,
		DiscardedCount: discarded,
		LastID:         lastID,
		ADBPath:        a.adbPath,
	}
}

func (a *App) ensureADB(configuredPath string) (string, error) {
	a.mu.Lock()
	cached := a.adbPath
	a.mu.Unlock()

	if configuredPath == "" && cached != "" {
		return cached, nil
	}
	return a.FindADB(configuredPath)
}

func (a *App) adbAndSerial() (string, string, error) {
	adbPath, err := a.ensureADB("")
	if err != nil {
		return "", "", err
	}
	a.mu.Lock()
	serial := strings.TrimSpace(a.serial)
	a.mu.Unlock()
	if serial == "" {
		return "", "", errors.New("select a connected Android device first")
	}
	return adbPath, serial, nil
}

func (a *App) context() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func (a *App) ingestLogLine(line string) {
	if entry, ok := a.parser.Parse(line); ok {
		a.pidMap.Apply(&entry)
		if a.logStore.AppendContinuation(entry) {
			return
		}
		a.logStore.Add(entry)
		return
	}

	if a.logStore.AppendMultiline(line) {
		return
	}

	a.logStore.Add(logcat.LogEntry{
		Level:   "I",
		Message: line,
		Raw:     line,
	})
}

func (a *App) startPIDTracker(serial string, packageName string) error {
	a.stopPIDTracker()

	adbPath, err := a.ensureADB("")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(a.context())
	a.mu.Lock()
	a.pidCancel = cancel
	a.mu.Unlock()

	refresh := func() {
		pids, err := adb.PidOf(ctx, adbPath, serial, packageName)
		if err != nil {
			if ctx.Err() == nil {
				a.setLastError(err.Error())
			}
			return
		}
		a.pidMap.Update(pids)
	}

	refresh()
	go func() {
		ticker := time.NewTicker(1500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refresh()
			}
		}
	}()

	return nil
}

func (a *App) stopPIDTracker() {
	a.mu.Lock()
	cancel := a.pidCancel
	a.pidCancel = nil
	a.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (a *App) storeAnalysisResults(results []logcat.AnalysisResult) {
	a.analysisMu.Lock()
	defer a.analysisMu.Unlock()

	if a.analysis == nil {
		a.analysis = map[string]logcat.AnalysisResult{}
	}
	for _, result := range results {
		if strings.TrimSpace(result.ID) == "" {
			continue
		}
		a.analysis[result.ID] = result
	}
}

func (a *App) updateProjectConfig(update func(workspace.ProjectConfig) workspace.ProjectConfig) {
	_ = a.updateActiveWorkspace(func(current workspace.WorkspaceConfig) workspace.WorkspaceConfig {
		return workspace.WorkspaceFromProject(current, update(workspace.ProjectFromWorkspace(current)))
	})
}

func (a *App) updateActiveWorkspace(update func(workspace.WorkspaceConfig) workspace.WorkspaceConfig) error {
	config, err := workspace.LoadConfig(workspace.DefaultConfigPath())
	if err != nil {
		return err
	}
	active := update(workspace.ActiveWorkspace(config))
	config = workspace.SaveWorkspace(config, active)
	return workspace.SaveConfig(workspace.DefaultConfigPath(), config)
}

func (a *App) findAnalysisResult(resultID string) (logcat.AnalysisResult, error) {
	resultID = strings.TrimSpace(resultID)
	if resultID == "" {
		return logcat.AnalysisResult{}, errors.New("select an analysis result first")
	}

	a.analysisMu.RLock()
	result, ok := a.analysis[resultID]
	a.analysisMu.RUnlock()
	if ok {
		return result, nil
	}

	results := a.AnalyzeLogs(a.logStore.Snapshot())
	for _, result := range results {
		if result.ID == resultID {
			return result, nil
		}
	}

	return logcat.AnalysisResult{}, fmt.Errorf("analysis result not found: %s", resultID)
}

func (a *App) currentDeviceInfo() *adb.AndroidDevice {
	a.mu.Lock()
	serial := a.serial
	adbPath := a.adbPath
	a.mu.Unlock()

	if strings.TrimSpace(serial) == "" || strings.TrimSpace(adbPath) == "" {
		return nil
	}
	device, err := adb.GetDeviceInfo(a.context(), adbPath, serial)
	if err != nil {
		return &adb.AndroidDevice{Serial: serial}
	}
	return &device
}

func (a *App) setLastError(message string) {
	a.mu.Lock()
	a.lastErr = strings.TrimSpace(message)
	a.mu.Unlock()
}

func (a *App) markStopped(streamID int64, err error) {
	a.mu.Lock()
	if streamID != a.streamID {
		a.mu.Unlock()
		return
	}
	a.running = false
	if err != nil {
		a.lastErr = err.Error()
	}
	a.mu.Unlock()
}

func deviceStateError(serial, state string) string {
	switch state {
	case "unauthorized":
		return fmt.Sprintf("device %s is unauthorized; please allow USB debugging authorization on the phone", serial)
	case "offline":
		return fmt.Sprintf("device %s is offline; reconnect the cable or refresh devices", serial)
	default:
		return fmt.Sprintf("device %s is not ready for logcat (state: %s)", serial, state)
	}
}
