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
	"catscope/internal/logcat"

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
}

func NewApp() *App {
	return &App{
		parser:   logcat.NewParser(),
		logStore: logcat.NewRingBuffer(100000),
		pidMap:   logcat.NewPIDMapper(),
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

func (a *App) ClearLogs() {
	a.logStore.Clear()
	a.parser.Reset()
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
