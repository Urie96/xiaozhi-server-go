package pool

import (
	"fmt"
	"github.com/urie96/xiaozhi-server-go/configs"
	"github.com/urie96/xiaozhi-server-go/core/mcp"
	"github.com/urie96/xiaozhi-server-go/core/providers"
	"github.com/urie96/xiaozhi-server-go/core/providers/asr"
	"github.com/urie96/xiaozhi-server-go/core/providers/llm"
	"github.com/urie96/xiaozhi-server-go/core/providers/tts"
	"github.com/urie96/xiaozhi-server-go/core/providers/vlllm"
	"github.com/urie96/xiaozhi-server-go/logger"
)

/*
* 工厂类，用于创建不同类型的资源池工厂。
* 通过配置文件和提供者类型，动态创建资源池工厂。
* 支持ASR、LLM、TTS和VLLLM等多种提供者类型。
* 每个工厂实现了ResourceFactory接口，提供Create和Destroy方法。
 */

// ProviderFactory 简化的提供者工厂
type ProviderFactory struct {
	Name         string // 提供者名称
	providerType string
	config       any
	params       map[string]any // 可选参数
}

func (f *ProviderFactory) Create() (any, error) {
	return f.createProvider()
}

func (f *ProviderFactory) Destroy(resource any) error {
	logger.Info("[Destroy] %s 资源池关闭，销毁资源", f.Name)

	if provider, ok := resource.(providers.Provider); ok {
		return provider.Cleanup()
	}

	if resource != nil {
		// 使用反射或类型断言来调用Cleanup方法
		if cleaner, ok := resource.(interface{ Cleanup() error }); ok {
			return cleaner.Cleanup()
		}
	}
	return nil
}

func (f *ProviderFactory) createProvider() (any, error) {
	switch f.providerType {
	case "asr":
		cfg := f.config.(*asr.Config)
		params := f.params
		delete_audio, _ := params["delete_audio"].(bool)
		asrType, _ := params["type"].(string)
		return asr.Create(asrType, cfg, delete_audio)
	case "llm":
		cfg := f.config.(*llm.Config)
		return llm.Create(cfg.Type, cfg)
	case "tts":
		cfg := f.config.(*tts.Config)
		params := f.params
		delete_audio, _ := params["delete_audio"].(bool)
		return tts.Create(cfg.Type, cfg, delete_audio)
	case "vlllm":
		cfg := f.config.(*configs.VLLMConfig)
		return vlllm.Create(cfg.Type, cfg)
	case "mcp":
		cfg := f.config.(*configs.Config)
		return mcp.NewManagerForPool(cfg), nil
	default:
		return nil, fmt.Errorf("未知的提供者类型: %s", f.providerType)
	}
}

// 创建各类型工厂的便利函数
func NewASRFactory(asrType string, config *configs.Config) ResourceFactory {
	if asrCfg, ok := config.ASR[asrType]; ok {
		return &ProviderFactory{
			providerType: "asr",
			config: &asr.Config{
				Name: asrType,
				Type: asrType,
				Data: asrCfg,
			},
			params: map[string]any{
				"type":         asrCfg["type"],
				"delete_audio": config.DeleteAudio,
			},
		}
	}
	return nil
}

func NewLLMFactory(llmType string, config *configs.Config) ResourceFactory {
	if llmCfg, ok := config.LLM[llmType]; ok {
		return &ProviderFactory{
			providerType: "llm",
			config: &llm.Config{
				Name:        llmType,
				Type:        llmCfg.Type,
				ModelName:   llmCfg.ModelName,
				BaseURL:     llmCfg.BaseURL,
				APIKey:      llmCfg.APIKey,
				Temperature: llmCfg.Temperature,
				MaxTokens:   llmCfg.MaxTokens,
				TopP:        llmCfg.TopP,
				Extra:       llmCfg.Extra,
			},
		}
	}
	return nil
}

func NewTTSFactory(ttsType string, config *configs.Config) ResourceFactory {
	if ttsCfg, ok := config.TTS[ttsType]; ok {
		return &ProviderFactory{
			providerType: "tts",
			config: &tts.Config{
				Name:      ttsType,
				Type:      ttsCfg.Type,
				Voice:     ttsCfg.Voice,
				Format:    ttsCfg.Format,
				OutputDir: ttsCfg.OutputDir,
				AppID:     ttsCfg.AppID,
				Token:     ttsCfg.Token,
				Cluster:   ttsCfg.Cluster,
			},
			params: map[string]any{
				"type":         ttsCfg.Type,
				"delete_audio": config.DeleteAudio,
			},
		}
	}
	return nil
}

func NewVLLLMFactory(
	vlllmType string,
	config *configs.Config,
) ResourceFactory {
	if vlllmCfg, ok := config.VLLLM[vlllmType]; ok {
		return &ProviderFactory{
			Name:         vlllmType,
			providerType: "vlllm",
			config:       &vlllmCfg,
		}
	}
	return nil
}

func NewMCPFactory(config *configs.Config) ResourceFactory {
	return &ProviderFactory{
		providerType: "mcp",
		config:       config,
		params:       map[string]any{},
	}
}
