package csy_runtime_util

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func KillProcessByPort(port int) error {
	// 查找占用端口的进程 PID
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("执行 netstat 失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var pid int

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pid, _ = strconv.Atoi(fields[4])
				break
			}
		}
	}

	if pid == 0 {
		fmt.Printf("未找到占用端口 %d 的进程", port)
		return nil
	}

	// 杀死进程
	killCmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	if err := killCmd.Run(); err != nil {
		return fmt.Errorf("杀死进程失败: %w", err)
	}

	fmt.Printf("成功杀死占用端口 %d 的进程 (PID: %d)\n", port, pid)
	return nil
}

func PortIsUse(port int) (bool, error) {
	// 查找占用端口的进程 PID
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("执行 netstat 失败: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var pid int

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pid, _ = strconv.Atoi(fields[4])
				break
			}
		}
	}
	return pid != 0, nil

}

func KillProcessByName(exeName string) error {
	// /F 强制终止进程
	// /IM 指定图像名称 (Image Name)
	// /T 终止子进程
	cmd := exec.Command("taskkill", "/F", "/IM", exeName, "/T")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果进程本身就不存在，taskkill 会返回错误码 128
		// 我们可以根据需求决定是否忽略这个错误
		return fmt.Errorf("执行 taskkill 失败 (可能是进程未运行): %w, 输出: %s", err, string(output))
	}

	fmt.Printf("成功终止所有名为 %s 的进程\n", exeName)
	return nil
}

func IsProcessRunning(exeName string) bool {
	// /NH 参数表示不显示标题行 (No Header)
	// /FI "IMAGENAME eq ..." 是按映像名称过滤
	cmd := exec.Command("tasklist", "/NH", "/FI", "IMAGENAME eq "+exeName)

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// 如果进程存在，输出中会包含该 exeName
	// 如果不存在，tasklist 会输出 "信息: 没有运行带有指定标准的任务。"
	return strings.Contains(string(output), exeName)
}
