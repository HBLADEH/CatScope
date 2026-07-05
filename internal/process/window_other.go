//go:build !windows

package process

import "os/exec"

func HideConsoleWindow(cmd *exec.Cmd) {
}
