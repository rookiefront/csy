package csy_ip_util

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// IPInfo IP信息结构体
type IPInfo struct {
	IP        string
	Interface string
	IsIPv4    bool
	IsIPv6    bool
}

// GetLocalIPList 获取所有本地IP（接口名 -> IP地址）
func GetLocalIPList() (map[string]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口失败: %w", err)
	}

	ipList := make(map[string]string)

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue // 跳过无法获取地址的接口
		}

		for _, addr := range addrs {
			ip := extractIP(addr)
			if ip.String() == "" {
				continue
			}

			// 过滤环回地址
			if ip.IsLoopback() {
				continue
			}

			// 优先保存 IPv4
			if ip.To4() != nil {
				ipList[iface.Name] = ip.String()
				break // 找到 IPv4 就停止查找该接口
			}

			// 如果没有 IPv4，保存第一个非链路本地 IPv6
			if _, exists := ipList[iface.Name]; !exists && !ip.IsLinkLocalUnicast() {
				ipList[iface.Name] = ip.String()
			}
		}
	}

	if len(ipList) == 0 {
		return nil, fmt.Errorf("未找到任何有效的本地IP地址")
	}

	return ipList, nil
}

// GetLocalIPv4List 获取所有本地IPv4地址
func GetLocalIPv4List() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口失败: %w", err)
	}

	var ipList []string

	for _, iface := range interfaces {
		// 跳过未启用或回环接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip := extractIP(addr)
			if ip == nil {
				continue
			}

			// 只保留 IPv4 且非环回
			if ip.To4() != nil && !ip.IsLoopback() {
				ipList = append(ipList, ip.String())
			}
		}
	}

	if len(ipList) == 0 {
		return nil, fmt.Errorf("无法获取本机IPv4地址")
	}

	return ipList, nil
}

// GetLocalIPv6List 获取所有本地IPv6地址
func GetLocalIPv6List() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口失败: %w", err)
	}

	var ipList []string

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip := extractIP(addr)
			if ip == nil {
				continue
			}

			// 过滤 IPv4、环回和链路本地地址
			if ip.To4() == nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
				ipList = append(ipList, ip.String())
			}
		}
	}

	return ipList, nil
}

// GetAllLocalIPs 获取所有本地IP（包括详细信息）
func GetAllLocalIPs() ([]IPInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("获取网络接口失败: %w", err)
	}

	var ipList []IPInfo

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip := extractIP(addr)
			if ip == nil {
				continue
			}

			// 跳过环回地址
			if ip.IsLoopback() {
				continue
			}

			info := IPInfo{
				IP:        ip.String(),
				Interface: iface.Name,
				IsIPv4:    ip.To4() != nil,
				IsIPv6:    ip.To4() == nil,
			}
			ipList = append(ipList, info)
		}
	}

	if len(ipList) == 0 {
		return nil, fmt.Errorf("未找到任何有效的本地IP地址")
	}

	return ipList, nil
}

// GetFirstIPv4 获取第一个非环回IPv4地址
func GetFirstIPv4() (string, error) {
	ips, err := GetLocalIPv4List()
	if err != nil {
		return "", err
	}

	if len(ips) > 0 {
		return ips[0], nil
	}

	return "", fmt.Errorf("未找到IPv4地址")
}

// GetPublicIP 获取公网IP地址（带超时和重试）
func GetPublicIP() (string, error) {
	return GetPublicIPWithTimeout(5 * time.Second)
}

// GetPublicIPWithTimeout 获取公网IP地址（自定义超时）
func GetPublicIPWithTimeout(timeout time.Duration) (string, error) {
	services := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ipinfo.io/ip",
		"https://checkip.amazonaws.com",
		"https://api.my-ip.io/ip",
	}

	// 随机打乱顺序
	rand.Shuffle(len(services), func(i, j int) {
		services[i], services[j] = services[j], services[i]
	})

	client := &http.Client{
		Timeout: timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for _, service := range services {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("获取公网IP超时")
		default:
		}

		ip, err := fetchPublicIP(client, service)
		if err == nil && ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("所有公网IP服务都失败了")
}

// fetchPublicIP 从指定服务获取公网IP
func fetchPublicIP(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(body))

	// 验证返回的是否为有效IP
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("无效的IP地址: %s", ip)
	}

	return ip, nil
}

// IsPrivateIP 判断是否为私有IP地址
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 私有IP范围
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// GetIPByInterface 根据接口名获取IP地址
func GetIPByInterface(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", fmt.Errorf("接口不存在: %w", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ip := extractIP(addr)
		if ip != nil && !ip.IsLoopback() {
			// 优先返回IPv4
			if ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("接口 %s 没有有效的IP地址", interfaceName)
}

// extractIP 从net.Addr中提取IP地址
func extractIP(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPNet:
		return v.IP
	case *net.IPAddr:
		return v.IP
	default:
		return nil
	}
}

// 初始化随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}
