package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Server struct {
		IP    string `yaml:"ip" json:"ip"`
		Port  int    `yaml:"port" json:"port"`
		Token string `json:"token"`
		Auth  struct {
			Store struct {
				Type   string `yaml:"type" json:"type"`     // memory/file/redis
				Expiry int    `yaml:"expiry" json:"expiry"` // 过期时间(小时)
			} `yaml:"store" json:"store"`
		} `yaml:"auth" json:"auth"`
	} `yaml:"server" json:"server"`

	// 传输层配置
	Transport struct {
		WebSocket struct {
			Enabled bool   `yaml:"enabled" json:"enabled"`
			IP      string `yaml:"ip" json:"ip"`
			Port    int    `yaml:"port" json:"port"`
		} `yaml:"websocket" json:"websocket"`

		MQTTUDP struct {
			Enabled bool `yaml:"enabled" json:"enabled"`
			MQTT    struct {
				IP   string `yaml:"ip" json:"ip"`
				Port int    `yaml:"port" json:"port"`
				QoS  int    `yaml:"qos" json:"qos"`
			} `yaml:"mqtt" json:"mqtt"`
			UDP struct {
				IP                string `yaml:"ip" json:"ip"`
				ShowPort          int    `yaml:"show_port" json:"show_port"` // 显示端口
				Port              int    `yaml:"port" json:"port"`
				SessionTimeout    string `yaml:"session_timeout" json:"session_timeout"`
				MaxPacketSize     int    `yaml:"max_packet_size" json:"max_packet_size"`
				EnableReliability bool   `yaml:"enable_reliability" json:"enable_reliability"`
			} `yaml:"udp" json:"udp"`
		} `yaml:"mqtt_udp" json:"mqtt_udp"`
	} `yaml:"transport" json:"transport"`

	Log struct {
		LogLevel string `yaml:"log_level" json:"log_level"`
		LogDir   string `yaml:"log_dir" json:"log_dir"`
		LogFile  string `yaml:"log_file" json:"log_file"`
	} `yaml:"log" json:"log"`

	Web struct {
		Port         int    `yaml:"port" json:"port"`
		StaticDir    string `yaml:"static_dir" json:"static_dir"`
		Websocket    string `yaml:"websocket" json:"websocket"`
		VisionURL    string `yaml:"vision" json:"vision"`
		ActivateText string `yaml:"activate_text" json:"activate_text"` // 发送激活码时携带的文本
	} `yaml:"web" json:"web"`

	DefaultPrompt   string        `yaml:"prompt"             json:"prompt"`
	Roles           []Role        `yaml:"roles"              json:"roles"` // 角色列表
	DeleteAudio     bool          `yaml:"delete_audio"       json:"delete_audio"`
	QuickReply      bool          `yaml:"quick_reply"        json:"quick_reply"`
	QuickReplyWords []string      `yaml:"quick_reply_words"  json:"quick_reply_words"`
	LocalMCPFun     []LocalMCPFun `yaml:"local_mcp_fun"      json:"local_mcp_fun"` // 本地MCP函数映射
	SaveTTSAudio    bool          `yaml:"save_tts_audio"  json:"save_tts_audio"`   // 是否保存TTS音频文件
	SaveUserAudio   bool          `yaml:"save_user_audio" json:"save_user_audio"`  // 是否保存用户音频文件

	SelectedModule map[string]string `yaml:"selected_module" json:"selected_module"`

	ASR   map[string]ASRConfig  `yaml:"ASR"   json:"ASR"`
	TTS   map[string]TTSConfig  `yaml:"TTS"   json:"TTS"`
	LLM   map[string]LLMConfig  `yaml:"LLM"   json:"LLM"`
	VLLLM map[string]VLLMConfig `yaml:"VLLLM" json:"VLLLM"`
}

type LocalMCPFun struct {
	Name        string `yaml:"name"         json:"name"`        // 函数名称
	Description string `yaml:"description"  json:"description"` // 函数描述
	Enabled     bool   `yaml:"enabled"      json:"enabled"`     // 是否启用
}

type Role struct {
	Name        string `yaml:"name"         json:"name"`        // 角色名称
	Description string `yaml:"description"  json:"description"` // 角色描述
	Enabled     bool   `yaml:"enabled"      json:"enabled"`     // 是否启用
}

// ASRConfig ASR配置结构
type ASRConfig map[string]interface{}

// TTSConfig TTS配置结构
type TTSConfig struct {
	Type      string `yaml:"type"             json:"type"`       // TTS类型
	Voice     string `yaml:"voice"            json:"voice"`      // 语音名称
	Format    string `yaml:"format"           json:"format"`     // 输出格式
	OutputDir string `yaml:"output_dir"       json:"output_dir"` // 输出目录
	AppID     string `yaml:"appid"            json:"appid"`      // 应用ID
	Token     string `yaml:"token"            json:"token"`      // API密钥
	Cluster   string `yaml:"cluster"          json:"cluster"`    // 集群信息
}

// LLMConfig LLM配置结构
type LLMConfig struct {
	Type        string                 `yaml:"type"        json:"type"`        // LLM类型
	ModelName   string                 `yaml:"model_name"  json:"model_name"`  // 模型名称
	BaseURL     string                 `yaml:"url"         json:"url"`         // API地址
	APIKey      string                 `yaml:"api_key"     json:"api_key"`     // API密钥
	Temperature float64                `yaml:"temperature" json:"temperature"` // 温度参数
	MaxTokens   int                    `yaml:"max_tokens"  json:"max_tokens"`  // 最大令牌数
	TopP        float64                `yaml:"top_p"       json:"top_p"`       // TopP参数
	Extra       map[string]interface{} `yaml:",inline"     json:"extra"`       // 额外配置
}

// SecurityConfig 图片安全配置结构
type SecurityConfig struct {
	MaxFileSize       int64    `yaml:"max_file_size"      json:"max_file_size"`      // 最大文件大小（字节）
	MaxPixels         int64    `yaml:"max_pixels"         json:"max_pixels"`         // 最大像素数量
	MaxWidth          int      `yaml:"max_width"          json:"max_width"`          // 最大宽度
	MaxHeight         int      `yaml:"max_height"         json:"max_height"`         // 最大高度
	AllowedFormats    []string `yaml:"allowed_formats"    json:"allowed_formats"`    // 允许的图片格式
	EnableDeepScan    bool     `yaml:"enable_deep_scan"   json:"enable_deep_scan"`   // 启用深度安全扫描
	ValidationTimeout string   `yaml:"validation_timeout" json:"validation_timeout"` // 验证超时时间
}

// VLLMConfig VLLLM配置结构（视觉语言大模型）
type VLLMConfig struct {
	Type        string                 `yaml:"type"        json:"type"`        // API类型，复用LLM的类型
	ModelName   string                 `yaml:"model_name"  json:"model_name"`  // 模型名称，使用支持视觉的模型
	BaseURL     string                 `yaml:"url"         json:"url"`         // API地址
	APIKey      string                 `yaml:"api_key"     json:"api_key"`     // API密钥
	Temperature float64                `yaml:"temperature" json:"temperature"` // 温度参数
	MaxTokens   int                    `yaml:"max_tokens"  json:"max_tokens"`  // 最大令牌数
	TopP        float64                `yaml:"top_p"       json:"top_p"`       // TopP参数
	Security    SecurityConfig         `yaml:"security"    json:"security"`    // 图片安全配置
	Extra       map[string]interface{} `yaml:",inline"     json:"extra"`       // 额外配置
}

var Cfg *Config

func (cfg *Config) ToString() string {
	data, _ := yaml.Marshal(cfg)
	return string(data)
}

// LoadConfig 加载配置
// 完全从数据库加载配置，如果数据库为空则使用默认配置并初始化数据库
func LoadConfig(path string) (*Config, error) {
	config := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	Cfg = config
	return config, nil
}
