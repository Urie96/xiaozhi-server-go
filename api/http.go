package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"github.com/urie96/xiaozhi-server-go/configs"
	"github.com/urie96/xiaozhi-server-go/httpsvr/ota"
	"github.com/urie96/xiaozhi-server-go/httpsvr/vision"
	"github.com/urie96/xiaozhi-server-go/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func StartHttpServer(config *configs.Config) {
	// 初始化Gin引擎
	if config.Log.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.SetTrustedProxies([]string{"0.0.0.0"})

	// 检查WebSocket传输层配置
	wsTransport := NewWebSocketTransport(config)
	logger.Debug("WebSocket传输层已注册")
	router.Any("/ws", func(ctx *gin.Context) {
		upgrader := &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源的连接
			},
		}
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			logger.Error("WebSocket升级失败: %v", err)
			return
		}

		wsTransport.HandleWebSocket(conn, ctx.Request)
	})

	// API路由全部挂载到/api前缀下
	apiGroup := router.Group("/api")

	// history 路由兜底，只处理 /web 下的 GET 请求
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"error": "api Not found"})
			return
		}
	})

	// 启动OTA服务
	otaService := ota.NewDefaultOTAService(config.Web.Websocket)
	apiGroup.Any("/ota/", otaService.HandleOTARequest())
	apiGroup.GET("/ota_bin/*filepath", otaService.HandleFirmwareDownload())

	// 启动Vision服务
	visionService, err := vision.NewDefaultVisionService(config)
	if err != nil {
		panic(fmt.Sprintf("Vision 服务初始化失败 %v", err))
	}

	apiGroup.GET("/vision", visionService.HandleGet)
	apiGroup.POST("/vision", visionService.HandlePost)
	apiGroup.OPTIONS("/vision", visionService.HandleOptions)

	// HTTP Server（支持优雅关机）
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Web.Port),
		Handler: router,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("HTTP 服务启动失败 %v", err))
	}

	return
}
