// @title 小智服务端 API 文档
// @version 1.0
// @description 小智服务端，包含OTA与Vision等接口
// @host localhost:8080
// @BasePath /api
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"xiaozhi-server-go/src/api"
	"xiaozhi-server-go/src/configs"
	_ "xiaozhi-server-go/src/core/providers/asr/deepgram"
	_ "xiaozhi-server-go/src/core/providers/asr/doubao"
	_ "xiaozhi-server-go/src/core/providers/asr/gosherpa"
	_ "xiaozhi-server-go/src/core/providers/asr/stepfun"
	_ "xiaozhi-server-go/src/core/providers/llm/coze"
	_ "xiaozhi-server-go/src/core/providers/llm/doubao"
	_ "xiaozhi-server-go/src/core/providers/llm/ollama"
	_ "xiaozhi-server-go/src/core/providers/llm/openai"
	_ "xiaozhi-server-go/src/core/providers/tts/doubao"
	_ "xiaozhi-server-go/src/core/providers/tts/edge"
	_ "xiaozhi-server-go/src/core/providers/tts/gosherpa"
	_ "xiaozhi-server-go/src/core/providers/vlllm/ollama"
	_ "xiaozhi-server-go/src/core/providers/vlllm/openai"
	"xiaozhi-server-go/src/httpsvr/ota"
	"xiaozhi-server-go/src/httpsvr/vision"
	"xiaozhi-server-go/src/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

func LoadConfigAndLogger() (*configs.Config, error) {
	// 加载配置,默认使用.config.yaml
	config, err := configs.LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		return nil, err
	}

	return config, nil
}

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
	wsTransport := api.NewWebSocketTransport(config)
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

func GracefulShutdown(cancel context.CancelFunc, g *errgroup.Group) {
	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// 等待信号
	sig := <-sigChan
	logger.Info("接收到系统信号: %v，开始优雅关闭服务", sig)

	// 取消上下文，通知所有服务开始关闭
	cancel()

	// 等待所有服务关闭，设置超时保护
	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.Error("服务关闭过程中出现错误 %v", err)
			os.Exit(1)
		}
		logger.Info("所有服务已优雅关闭")
	case <-time.After(15 * time.Second):
		logger.Error("服务关闭超时，强制退出")
		os.Exit(1)
	}
}

func main() {
	// 加载配置和初始化日志系统
	config, err := LoadConfigAndLogger()
	if err != nil {
		fmt.Println("加载配置或初始化日志系统失败:", err)
		os.Exit(1)
	}

	StartHttpServer(config)

	logger.Info("程序已成功退出")
}
