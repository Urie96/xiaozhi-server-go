// @title 小智服务端 API 文档
// @version 1.0
// @description 小智服务端，包含OTA与Vision等接口
// @host localhost:8080
// @BasePath /api
package main

import (
	"fmt"
	"os"
	"github.com/urie96/xiaozhi-server-go/api"
	"github.com/urie96/xiaozhi-server-go/configs"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/asr/deepgram"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/asr/doubao"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/asr/gosherpa"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/asr/stepfun"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/llm/coze"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/llm/doubao"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/llm/ollama"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/llm/openai"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/tts/doubao"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/tts/edge"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/tts/gosherpa"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/vlllm/ollama"
	_ "github.com/urie96/xiaozhi-server-go/core/providers/vlllm/openai"
	"github.com/urie96/xiaozhi-server-go/logger"
)

func LoadConfigAndLogger() (*configs.Config, error) {
	// 加载配置,默认使用.config.yaml
	config, err := configs.LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		return nil, err
	}

	return config, nil
}

func main() {
	// 加载配置和初始化日志系统
	config, err := LoadConfigAndLogger()
	if err != nil {
		fmt.Println("加载配置或初始化日志系统失败:", err)
		os.Exit(1)
	}

	api.StartHttpServer(config)

	logger.Info("程序已成功退出")
}
