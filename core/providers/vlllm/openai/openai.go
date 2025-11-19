package openai

import (
	"github.com/urie96/xiaozhi-server-go/core/providers/vlllm"
	"github.com/urie96/xiaozhi-server-go/logger"
)

// OpenAIVLLMProvider OpenAI类型的VLLLM提供者
type OpenAIVLLMProvider struct {
	*vlllm.Provider
}

// NewProvider 创建OpenAI VLLLM提供者实例
func NewProvider(config *vlllm.Config) (*vlllm.Provider, error) {
	// 直接使用基础VLLLM Provider，因为它已经复用了LLM架构
	// OpenAI类型的VLLLM只需要确保使用正确的模型名称（如glm-4v-flash）
	provider, err := vlllm.NewProvider(config)
	if err != nil {
		return nil, err
	}

	logger.Debug("OpenAI VLLLM Provider创建成功 %v", map[string]any{
		"model_name": config.ModelName,
		"base_url":   config.BaseURL,
	})

	return provider, nil
}

// init 注册OpenAI VLLLM提供者
func init() {
	vlllm.Register("openai", NewProvider)
}
