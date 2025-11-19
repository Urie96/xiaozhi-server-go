package api

import (
	"fmt"
	"net/http"

	"github.com/urie96/xiaozhi-server-go/configs"
	"github.com/urie96/xiaozhi-server-go/core"
	"github.com/urie96/xiaozhi-server-go/core/pool"
	"github.com/urie96/xiaozhi-server-go/logger"

	"github.com/gorilla/websocket"
)

// WebSocketTransport WebSocket传输层实现
type WebSocketTransport struct {
	config      *configs.Config
	poolManager *pool.PoolManager
}

// NewWebSocketTransport 创建新的WebSocket传输层
func NewWebSocketTransport(config *configs.Config) *WebSocketTransport {
	// 初始化资源池管理器
	poolManager, err := pool.NewPoolManager(config)
	if err != nil {
		logger.Error("%s", fmt.Sprintf("初始化资源池管理器失败: %v", err))
		panic(err)
	}

	return &WebSocketTransport{
		poolManager: poolManager,
		config:      config,
	}
}

// HandleWebSocket 处理WebSocket连接
func (t *WebSocketTransport) HandleWebSocket(conn *websocket.Conn, r *http.Request) {
	logger.Info("[WebSocket] [连接请求]")

	// 从资源池获取提供者集合
	providerSet, err := t.poolManager.GetProviderSet()
	if err != nil {
		logger.Error("获取提供者集合失败: %v", err)
		return
	}

	handler := core.NewConnectionHandler(t.config, providerSet, r)

	handler.Handle(conn)

	// 先关闭连接处理器
	if handler != nil {
		handler.Close()
	}

	// 关闭连接
	if conn != nil {
		conn.Close()
	}

	// 归还资源到池中
	if providerSet != nil && t.poolManager != nil {
		if err := t.poolManager.ReturnProviderSet(providerSet); err != nil {
			logger.Error("客户端归还资源失败: %v", err)
		} else {
			logger.Info("客户端资源已成功归还到池中")
		}
	}
}
