package proxy_service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ProxyService struct {
	Address           string       `json:"address"`
	Port              int          `json:"port"`
	ProxyURL          string       `json:"proxy_url"`
	ProxySocket       string       `json:"proxy_socket"`
	ServerEngine      *gin.Engine  `json:"-"`
	HttpServer        *http.Server `json:"-"`
	isRunning         bool         // 添加运行状态
	startTime         time.Time    // 添加启动时间
	BeforeStartEngine func(*gin.Engine)
	InterceptResponse func(res *http.Response) error
}

func NewProxyService(address string, port int) *ProxyService {
	return &ProxyService{
		Address:           address,
		Port:              port,
		BeforeStartEngine: nil,
		InterceptResponse: nil,
	}
}

func (s *ProxyService) SetProxyURL(href string) {
	s.ProxyURL = href
}
func (s *ProxyService) SetInterceptResponse(f func(res *http.Response) error) {
	s.InterceptResponse = f
}

func (s *ProxyService) SetProxySocket(href string) {
	s.ProxySocket = href
}

func (s *ProxyService) SetBeforeStartEngine(f func(*gin.Engine)) {
	s.BeforeStartEngine = f
}

// initServer 初始化服务器
func (s *ProxyService) initServer() error {
	engine := gin.New()
	s.ServerEngine = engine

	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	s.HttpServer = &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	s.isRunning = true
	s.startTime = time.Now()

	// 配置反向代理
	targetProxyURL, err := url.Parse(strings.TrimSpace(s.ProxyURL)) //
	if err != nil {
		return err
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(targetProxyURL)

	originalDirector := reverseProxy.Director

	reverseProxy.Director = func(req *http.Request) {
		// 保留原始的转发逻辑
		originalDirector(req)
		req.Host = targetProxyURL.Host
	}

	engine.NoRoute(func(c *gin.Context) {
		reverseProxy.ServeHTTP(c.Writer, c.Request)
	})

	// 如果存在 socket 代理
	if s.ProxySocket != "" {
		transport := s.CreateSocks5Transport(s.ProxySocket)
		reverseProxy.Transport = transport
	}
	if s.InterceptResponse != nil {
		reverseProxy.ModifyResponse = s.InterceptResponse
	}

	return nil
}

// Start - 同步启动（会阻塞）
func (s *ProxyService) Start() error {
	err := s.initServer()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	fmt.Printf("Proxy service started on %s\n", addr)
	if s.BeforeStartEngine != nil {
		s.BeforeStartEngine(s.ServerEngine)
	}

	// 直接运行，会阻塞
	return s.HttpServer.ListenAndServe()
}

// StartAsync - 异步启动（不阻塞）
func (s *ProxyService) StartAsync() error {
	err := s.initServer()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	if s.BeforeStartEngine != nil {
		s.BeforeStartEngine(s.ServerEngine)
	}
	// 在协程中启动
	go func() {
		fmt.Printf("Proxy service started asynchronously on %s\n", addr)
		if err := s.HttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Proxy service error: %v\n", err)
			s.isRunning = false
		}
	}()

	return nil
}

// Stop - 优雅停止
func (s *ProxyService) Stop() error {
	if s.HttpServer == nil {
		return errors.New("server not started")
	}

	if !s.isRunning {
		return errors.New("server is not running")
	}

	fmt.Println("Proxy service stopping gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.HttpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %v", err)
	}

	s.isRunning = false
	fmt.Printf("Proxy service stopped (ran for %v)\n", time.Since(s.startTime))
	return nil
}

// ForceStop - 强制停止
func (s *ProxyService) ForceStop() error {
	if s.HttpServer == nil {
		return errors.New("server not started")
	}

	if !s.isRunning {
		return errors.New("server is not running")
	}

	fmt.Println("Proxy service forcing stop...")

	if err := s.HttpServer.Close(); err != nil {
		return fmt.Errorf("server close error: %v", err)
	}

	s.isRunning = false
	fmt.Printf("Proxy service force stopped (ran for %v)\n", time.Since(s.startTime))
	return nil
}

// IsRunning - 检查服务是否在运行
func (s *ProxyService) IsRunning() bool {
	return s.isRunning
}

// GetUptime - 获取服务运行时间
func (s *ProxyService) GetUptime() time.Duration {
	if !s.isRunning {
		return 0
	}
	return time.Since(s.startTime)
}
