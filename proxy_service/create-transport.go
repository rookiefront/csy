package proxy_service

import (
	"context"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/proxy"
)

func (s *ProxyService) CreateSocks5Transport(u string) *http.Transport {
	// 创建 SOCKS5 代理 Dialer
	//socksDialer, err := proxy.SOCKS5("tcp", "192.168.3.210:10808", nil, proxy.Direct)
	socksDialer, err := proxy.SOCKS5("tcp", u, nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Failed to create SOCKS5 proxy dialer: %v", err)
	}

	// 将 socksDialer 的 Dial 适配为 DialContext
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return socksDialer.Dial(network, addr)
	}

	// 自定义 transport
	transport := &http.Transport{
		DialContext: dialContext,
	}
	return transport
}
