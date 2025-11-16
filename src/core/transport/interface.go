package transport

import (
	"context"
	"net/http"
	"xiaozhi-server-go/src/core"
)

// Transport 传输层接口
type Transport interface {
	// 启动传输服务
	Start(ctx context.Context) error
	// 停止传输服务
	Stop() error
	// 设置连接处理器工厂
	SetConnectionHandler(handler ConnectionHandlerFactory)
}

type Connection = core.Connection

// ConnectionHandler 连接处理器接口
type ConnectionHandler interface {
	// 处理连接
	Handle()
	// 关闭处理器
	Close()
}

// ConnectionHandlerFactory 连接处理器工厂接口
type ConnectionHandlerFactory interface {
	// 创建连接处理器
	CreateHandler(conn Connection, req *http.Request) ConnectionHandler
}
