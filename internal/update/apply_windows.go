//go:build windows

package update

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	applyFlag   = "--catscope-apply-update"
	cleanupFlag = "--catscope-cleanup-update"
)

func LaunchInstaller(download Download) error {
	target, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法确定当前程序路径: %w", err)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("无法解析当前程序路径: %w", err)
	}
	if filepath.Ext(target) != ".exe" {
		return errors.New("当前程序不是 Windows EXE，无法自动替换")
	}
	if err := verifyDirectoryWritable(filepath.Dir(target)); err != nil {
		return fmt.Errorf("CatScope 所在目录不可写，请打开 Release 页面手动更新: %w", err)
	}
	if !isUpdateDirectory(download.Directory) || filepath.Dir(download.Path) != download.Directory {
		return errors.New("升级临时目录无效")
	}

	helper := filepath.Join(download.Directory, "CatScope-updater.exe")
	if err := copyFile(target, helper); err != nil {
		return fmt.Errorf("准备升级程序失败: %w", err)
	}
	command := exec.Command(helper, applyFlag, target, download.Path, download.Directory)
	if err := command.Start(); err != nil {
		return fmt.Errorf("启动升级程序失败: %w", err)
	}
	return command.Process.Release()
}

func HandleCommandLine(args []string) (bool, error) {
	if len(args) == 0 || args[0] != applyFlag {
		return false, nil
	}
	if len(args) != 4 {
		return true, errors.New("升级参数无效")
	}
	target, source, directory := args[1], args[2], args[3]
	if !isUpdateDirectory(directory) || filepath.Dir(source) != directory || filepath.Base(source) != "CatScope.exe.new" {
		return true, errors.New("升级文件路径无效")
	}
	if filepath.Base(target) == "" || !strings.EqualFold(filepath.Ext(target), ".exe") {
		return true, errors.New("目标程序路径无效")
	}

	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if backup, err := replaceExecutable(target, source); err == nil {
			command := exec.Command(target, cleanupFlag, directory)
			if err := command.Start(); err != nil {
				_ = os.Remove(target)
				_ = os.Rename(backup, target)
				fallback := exec.Command(target)
				_ = fallback.Start()
				if fallback.Process != nil {
					_ = fallback.Process.Release()
				}
				return true, fmt.Errorf("新版本启动失败，已恢复旧版本: %w", err)
			}
			_ = command.Process.Release()
			_ = os.Remove(backup)
			return true, nil
		} else {
			lastErr = err
		}
		time.Sleep(300 * time.Millisecond)
	}
	return true, fmt.Errorf("等待 CatScope 退出并替换文件超时: %w", lastErr)
}

func ScheduleCleanup(args []string) {
	if len(args) != 2 || args[0] != cleanupFlag || !isUpdateDirectory(args[1]) {
		return
	}
	directory := args[1]
	go func() {
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(time.Second)
			if err := os.RemoveAll(directory); err == nil {
				return
			}
		}
	}()
}

func replaceExecutable(target, source string) (string, error) {
	backup := target + ".catscope-old"
	_ = os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		return "", err
	}
	if err := os.Rename(source, target); err != nil {
		_ = os.Rename(backup, target)
		return "", err
	}
	return backup, nil
}

func verifyDirectoryWritable(directory string) error {
	file, err := os.CreateTemp(directory, ".catscope-update-write-test-")
	if err != nil {
		return err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return os.Remove(path)
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func isUpdateDirectory(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	temp, err := filepath.Abs(os.TempDir())
	if err != nil || filepath.Dir(abs) != temp {
		return false
	}
	return strings.HasPrefix(filepath.Base(abs), "catscope-update-")
}
