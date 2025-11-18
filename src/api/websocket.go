package api

import (
	"fmt"
	"net/http"
	"xiaozhi-server-go/src/configs"
	"xiaozhi-server-go/src/core/pool"
	"xiaozhi-server-go/src/core/transport"
	"xiaozhi-server-go/src/core/utils"
	"xiaozhi-server-go/src/logger"

	"github.com/gorilla/websocket"
)

// WebSocketTransport WebSocket传输层实现
type WebSocketTransport struct {
	config      *configs.Config
	server      *http.Server
	logger      *utils.Logger
	poolManager *pool.PoolManager
	upgrader    *websocket.Upgrader
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
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源的连接
			},
		},
	}
}

// HandleWebSocket 处理WebSocket连接
func (t *WebSocketTransport) HandleWebSocket(conn *websocket.Conn, r *http.Request) {
	logger.Info("[WebSocket] [连接请求]")

	// 从资源池获取提供者集合
	providerSet, err := t.poolManager.GetProviderSet()
	if err != nil {
		logger.Error(fmt.Sprintf("获取提供者集合失败: %v", err))
		return
	}

	// 创建连接上下文适配器
	adapter := transport.NewConnectionContextAdapter(
		conn,
		t.config,
		providerSet,
		t.poolManager,
		r,
	)
	defer adapter.Close()
	adapter.Handle()
}
