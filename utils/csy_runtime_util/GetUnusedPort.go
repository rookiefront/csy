package csy_runtime_util

import "net"

func GetUnusedPort() (int, error) {
	// 监听一个随机端口
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	// 获取监听的地址
	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
