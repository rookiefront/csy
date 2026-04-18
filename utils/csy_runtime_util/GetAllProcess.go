package csy_runtime_util

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/process"
)

type ProcessInfo struct {
	Exe    string
	Name   string
	Pid    int
	Origin *process.Process
	Ports  []int
}

func GetAllProcess() []ProcessInfo {
	var result []ProcessInfo
	pids, err := process.Pids()
	if err != nil {
		return result
	}

	// 获取端口与PID的映射关系
	portMap := getPortToPidMap()

	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		info := ProcessInfo{
			Pid:    int(pid),
			Origin: p,
		}
		info.Exe, _ = p.Exe()
		info.Name, _ = p.Name()

		// 添加端口信息
		if ports, exists := portMap[info.Pid]; exists {
			info.Ports = ports
		} else {
			info.Ports = []int{}
		}

		result = append(result, info)
	}
	return result
}

// getPortToPidMap 获取端口到PID的映射
func getPortToPidMap() map[int][]int {
	portMap := make(map[int][]int)

	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return portMap
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "LISTENING") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// 提取端口
		localAddr := fields[1]
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon == -1 {
			continue
		}

		portStr := localAddr[lastColon+1:]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		// 提取PID
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			continue
		}

		// 避免重复端口
		exists := false
		for _, existingPort := range portMap[pid] {
			if existingPort == port {
				exists = true
				break
			}
		}
		if !exists {
			portMap[pid] = append(portMap[pid], port)
		}
	}

	return portMap
}
