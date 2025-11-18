package transport

import (
	"net/http"
	"xiaozhi-server-go/src/core"
)

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
