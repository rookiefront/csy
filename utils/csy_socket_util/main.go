package csy_socket_util

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ---------- 消息结构（前后端统一） ----------
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ---------- SocketClient 封装单个连接 ----------
type SocketClient struct {
	ID       string          // 用户标识（可为空）
	Origin   string          // 客户端请求来源 Origin
	Conn     *websocket.Conn // 底层 WebSocket 连接
	sendChan chan []byte     // 内部使用：待发送消息队列（私有化）
	Mgr      *SocketManager  // 反向引用，用于操作
}

// Send 向当前客户端直接发送消息（面向对象调用，隐藏 channel 操作）
func (c *SocketClient) Send(msgType string, data interface{}) bool {
	var rawData json.RawMessage
	if d, err := json.Marshal(data); err == nil {
		rawData = d
	}

	msg := Message{Type: msgType, Data: rawData}
	bytes, _ := json.Marshal(msg)

	// 非阻塞写入发送队列
	select {
	case c.sendChan <- bytes:
		return true
	default:
		// 缓冲区满，发送失败
		return false
	}
}

// ---------- SocketManager 管理所有连接 ----------
type SocketManager struct {
	// 配置
	WriteWait      time.Duration // 每次写入的超时时间
	MessageBufSize int           // 每个客户端的发送缓冲大小

	// 连接池
	clients map[*SocketClient]bool   // 所有在线客户端
	userMap map[string]*SocketClient // 按用户ID快速索引（可选）
	mu      sync.RWMutex             // 保护 clients 和 userMap

	// 内部控制通道
	register   chan *SocketClient // 注册请求
	unregister chan *SocketClient // 注销请求
	broadcast  chan []byte        // 广播消息队列

	// 生命周期控制
	stopCh chan struct{}
	wg     sync.WaitGroup

	// 用户自定义消息处理器（由上层注入）
	Handler func(client *SocketClient, msg Message) // 处理收到的消息
}

// ---------- 创建新的 SocketManager ----------
func NewSocketManager(writeWait time.Duration, bufSize int) *SocketManager {
	if writeWait <= 0 {
		writeWait = 10 * time.Second
	}
	if bufSize <= 0 {
		bufSize = 256
	}
	mgr := &SocketManager{
		WriteWait:      writeWait,
		MessageBufSize: bufSize,
		clients:        make(map[*SocketClient]bool),
		userMap:        make(map[string]*SocketClient),
		register:       make(chan *SocketClient),
		unregister:     make(chan *SocketClient),
		broadcast:      make(chan []byte, bufSize*10),
		stopCh:         make(chan struct{}),
	}
	go mgr.run()
	return mgr
}

// ---------- 设置消息处理器（由上层调用） ----------
func (m *SocketManager) SetHandler(handler func(client *SocketClient, msg Message)) {
	m.Handler = handler
}

// ---------- 核心事件循环 ----------
func (m *SocketManager) run() {
	for {
		select {
		case <-m.stopCh:
			m.mu.Lock()
			for c := range m.clients {
				close(c.sendChan)
				c.Conn.Close()
			}
			m.clients = nil
			m.userMap = nil
			m.mu.Unlock()
			return

		case c := <-m.register:
			m.mu.Lock()
			m.clients[c] = true
			if c.ID != "" {
				m.userMap[c.ID] = c
			}
			m.mu.Unlock()

		case c := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[c]; ok {
				delete(m.clients, c)
				if c.ID != "" && m.userMap[c.ID] == c {
					delete(m.userMap, c.ID)
				}
				close(c.sendChan)
				c.Conn.Close()
			}
			m.mu.Unlock()

		case msg := <-m.broadcast:
			m.mu.RLock()
			for c := range m.clients {
				select {
				case c.sendChan <- msg:
				default:
					// 缓冲区满，丢弃
				}
			}
			m.mu.RUnlock()
		}
	}
}

// ---------- 优雅关闭 ----------
func (m *SocketManager) Close() {
	close(m.stopCh)
	m.wg.Wait()
}

// ---------- WebSocket 升级处理器 ----------
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ServeWS 升级 HTTP 并注册客户端
func (m *SocketManager) ServeWS(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &SocketClient{
		ID:       userID,
		Origin:   r.Header.Get("Origin"), // 获取并保存来源 Origin
		Conn:     conn,
		sendChan: make(chan []byte, m.MessageBufSize),
		Mgr:      m,
	}

	m.register <- client

	m.wg.Add(1)
	go m.writer(client)

	defer func() {
		m.unregister <- client
		m.wg.Done()
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		go m.handleClientMessage(client, raw) // 异步处理，不阻塞读循环
	}
}

// writer 负责从 sendChan 取数据并写入 WebSocket
func (m *SocketManager) writer(c *SocketClient) {
	for msg := range c.sendChan {
		c.Conn.SetWriteDeadline(time.Now().Add(m.WriteWait))
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

// ---------- 处理客户端消息 ----------
func (m *SocketManager) handleClientMessage(c *SocketClient, raw []byte) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}

	// 自动响应 ping（心跳）
	if msg.Type == "ping" {
		c.Send("pong", "pong")
		return
	}

	// 调用用户自定义处理器
	if m.Handler != nil {
		m.Handler(c, msg)
	}
}

// ---------- 广播 ----------
func (m *SocketManager) Broadcast(msgType string, data interface{}) {
	var rawData json.RawMessage
	if d, err := json.Marshal(data); err == nil {
		rawData = d
	}
	msg := Message{Type: msgType, Data: rawData}
	bytes, _ := json.Marshal(msg)
	m.broadcast <- bytes
}

// ---------- 单播 ----------
func (m *SocketManager) SendToUser(userID string, msgType string, data interface{}) bool {
	m.mu.RLock()
	c, ok := m.userMap[userID]
	m.mu.RUnlock()
	if !ok {
		return false
	}
	return c.Send(msgType, data)
}

// ---------- 获取在线用户数 ----------
func (m *SocketManager) UserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// ---------- 获取所有在线客户端 ----------
func (m *SocketManager) Users() []*SocketClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	users := make([]*SocketClient, 0, len(m.clients))
	for client, b := range m.clients {
		if b {
			users = append(users, client)
		}
	}
	return users
}