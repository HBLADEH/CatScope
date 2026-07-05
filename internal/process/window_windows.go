//go:build windows

package process

import (
	"os/exec"
	"syscall"
)

func HideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
