package mcu

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"smart-cabinet/internal/config"
	"smart-cabinet/internal/state"
)

// Server MCU TCP 服务端
// 后端作为 TCP Server，单片机作为 Client 连接上来
type Server struct {
	cfg      config.MCUConfig
	listener net.Listener
	conn     net.Conn    // 当前连接的 MCU
	mu       sync.Mutex
	stopCh   chan struct{}

	// 回调：MCU 上报关门 (0)
	OnDoorClosed func()
	// 回调：MCU 上报开门
	OnDoorOpened func()
}

// NewServer 创建 MCU 通信服务
func NewServer(cfg config.MCUConfig) *Server {
	return &Server{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// Start 启动 TCP 监听
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.cfg.Addr())
	if err != nil {
		return fmt.Errorf("MCU TCP 监听失败 %s: %w", s.cfg.Addr(), err)
	}
	log.Printf("[MCU] TCP 服务已启动，监听 %s", s.cfg.Addr())

	go s.acceptLoop()
	return nil
}

// acceptLoop 等待 MCU 连接
func (s *Server) acceptLoop() {
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				log.Printf("[MCU] 接受连接失败: %v", err)
				continue
			}
		}

		log.Printf("[MCU] 单片机已连接: %s", conn.RemoteAddr())

		s.mu.Lock()
		// 如果有旧连接，先关闭
		if s.conn != nil {
			s.conn.Close()
		}
		s.conn = conn
		s.mu.Unlock()

		go s.handleConn(conn)
	}
}

// handleConn 处理 MCU 连接
func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		if s.conn == conn {
			s.conn = nil
		}
		s.mu.Unlock()
		log.Printf("[MCU] 单片机已断开: %s", conn.RemoteAddr())
	}()

	reader := bufio.NewReader(conn)
	// 心跳超时定时器
	timeout := time.Duration(s.cfg.HeartbeatTimeout) * time.Second
	readTimeout := timeout

	for {
		// 设置读超时
		conn.SetReadDeadline(time.Now().Add(readTimeout))

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("[MCU] 读取数据失败（可能心跳超时或断开）: %v", err)
			return
		}

		line = strings.TrimSpace(line)
		log.Printf("[MCU] 收到: %q", line)

		switch line {
		case "0":
			// 单片机上报关门
			log.Println("[MCU] 柜门已关闭")
			if s.OnDoorClosed != nil {
				s.OnDoorClosed()
			}
			readTimeout = timeout // 恢复正常超时

		case "HEARTBEAT":
			// 心跳
			s.sendLine("PONG")
			readTimeout = timeout

		default:
			log.Printf("[MCU] 未知消息: %s", line)
			readTimeout = timeout
		}
	}
}

// SendOpen 发送开门指令 "1" 给 MCU
func (s *Server) SendOpen() error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("单片机未连接")
	}

	if err := s.sendLine("1"); err != nil {
		return fmt.Errorf("发送开门指令失败: %w", err)
	}

	// 通知状态机门已打开
	if s.OnDoorOpened != nil {
		s.OnDoorOpened()
	}

	log.Println("[MCU] 已发送开门指令")
	return nil
}

// sendLine 发送一行数据
func (s *Server) sendLine(msg string) error {
	s.mu.Lock()
	conn := s.conn
	s.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("单片机未连接")
	}

	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := fmt.Fprintf(conn, "%s\n", msg)
	return err
}

// IsConnected MCU 是否已连接
func (s *Server) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn != nil
}

// Stop 停止服务
func (s *Server) Stop() {
	close(s.stopCh)
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
	}
	s.mu.Unlock()
	log.Println("[MCU] 服务已停止")
}

// 确保 state 包的引用
var _ = state.Get
