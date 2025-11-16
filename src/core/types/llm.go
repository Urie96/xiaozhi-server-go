package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// Action represents the type of action.
type Action int

const (
	ActionTypeError       Action = -1
	ActionTypeNotFound    Action = 0
	ActionTypeNone        Action = 1
	ActionTypeResponse    Action = 2
	ActionTypeReqLLM      Action = 3
	ActionTypeCallHandler Action = 4
)

// ActionResponse holds the result of an action.
type ActionResponse struct {
	Action   Action      // 动作类型
	Result   any // 动作产生的结果
	Response any // 直接回复的内容
}

type ActionResponseCall struct {
	FuncName string      // 函数名
	Args     any // 函数参数
}

// Message 对话消息结构
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

func (m *Message) Print() {
	// 转为json字符串
	jsonStr, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Println("json marshal error:", err)
		return
	}
	// fmt.Println("Message:")
	fmt.Println(string(jsonStr))
}

// ToolCall 工具调用结构
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
	Index    int          `json:"index"`
}

// FunctionCall 函数调用结果
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Response LLM响应结构
type Response struct {
	Content              string     `json:"content,omitempty"`
	ReasonContent        string     `json:"reasoning_content,omitempty"`
	UpdateConversationID string     `json:"update_conversation_id,omitempty"`
	ToolCalls            []ToolCall `json:"tool_calls,omitempty"`
	StopReason           string     `json:"stop_reason,omitempty"`
	Error                string     `json:"error,omitempty"`
}

// Provider 基础提供者接口
type Provider interface {
	Initialize() error
	Cleanup() error
}

type FunctionRegistryInterface interface {
	RegisterFunction(name string, function openai.Tool) error
	GetFunction(name string) (openai.Tool, error)
	GetAllFunctions() []openai.Tool
	UnregisterFunction(name string) error
	UnregisterAllFunctions() error
	FunctionExists(name string) bool
}

// LLMProvider 大语言模型提供者接口
type LLMProvider interface {
	Provider
	Response(ctx context.Context, sessionID string, messages []Message) (<-chan string, error)
	ResponseWithFunctions(
		ctx context.Context,
		sessionID string,
		messages []Message,
		tools []openai.Tool,
	) (<-chan Response, error)
	GetSessionID() string                       // 获取当前会话ID
	SetIdentityFlag(idType string, flag string) // 设置身份标识
}
