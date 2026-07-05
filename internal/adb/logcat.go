package adb

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"catscope/internal/process"
)

type LogcatProcess struct {
	cancel context.CancelFunc
	cmd    *exec.Cmd
	done   chan struct{}
	once   sync.Once
}

func StartLogcat(
	parent context.Context,
	adbPath string,
	serial string,
	onLine func(string),
	onError func(string),
	onExit func(error),
) (*LogcatProcess, error) {
	ctx, cancel := context.WithCancel(parent)
	cmd := exec.CommandContext(ctx, adbPath, "-s", serial, "logcat", "-v", "threadtime", "-b", "main,system,crash")
	process.HideConsoleWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start logcat failed: %w", err)
	}

	proc := &LogcatProcess{
		cancel: cancel,
		cmd:    cmd,
		done:   make(chan struct{}),
	}

	var scanWG sync.WaitGroup
	var stderrMu sync.Mutex
	var stderrLines []string

	scanWG.Add(2)
	go func() {
		defer scanWG.Done()
		scanLines(stdout, onLine)
	}()
	go func() {
		defer scanWG.Done()
		scanLines(stderr, func(line string) {
			if strings.TrimSpace(line) == "" {
				return
			}
			stderrMu.Lock()
			stderrLines = append(stderrLines, line)
			if len(stderrLines) > 20 {
				stderrLines = stderrLines[len(stderrLines)-20:]
			}
			stderrMu.Unlock()
			onError(line)
		})
	}()
	go func() {
		err := cmd.Wait()
		scanWG.Wait()
		close(proc.done)
		if ctx.Err() == context.Canceled {
			onExit(nil)
			return
		}
		stderrMu.Lock()
		stderrText := strings.TrimSpace(strings.Join(stderrLines, "\n"))
		stderrMu.Unlock()
		if err != nil {
			if stderrText != "" {
				onExit(fmt.Errorf("logcat stopped unexpectedly: %w: %s", err, stderrText))
				return
			}
			onExit(fmt.Errorf("logcat stopped unexpectedly: %w", err))
			return
		}
		if stderrText != "" {
			onExit(fmt.Errorf("logcat stopped unexpectedly: %s", stderrText))
			return
		}
		onExit(nil)
	}()

	return proc, nil
}

func (p *LogcatProcess) Stop() error {
	var err error
	p.once.Do(func() {
		p.cancel()
		<-p.done
	})
	return err
}

func scanLines(reader io.Reader, fn func(string)) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		fn(scanner.Text())
	}
}
