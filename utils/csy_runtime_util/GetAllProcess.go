package csy_runtime_util

import "github.com/shirou/gopsutil/process"

type ProcessInfo struct {
	Exe    string
	Name   string
	Pid    int
	Origin *process.Process
}

func GetAllProcess() []ProcessInfo {
	var result []ProcessInfo
	pids, err := process.Pids()
	if err != nil {
		return result
	}
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
		result = append(result, info)
	}
	return result
}
