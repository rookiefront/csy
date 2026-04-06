package csy_device_util

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenUrl(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	err := cmd.Start()
	if err != nil {
		return err
	}

	return nil
}
