//go:build !windows

package update

import "errors"

func LaunchInstaller(Download) error {
	return errors.New("当前平台暂不支持应用内自动安装")
}

func HandleCommandLine([]string) (bool, error) {
	return false, nil
}

func ScheduleCleanup([]string) {}
