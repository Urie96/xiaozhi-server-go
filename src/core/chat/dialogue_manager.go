package chat

import (
	"xiaozhi-server-go/src/core/types"
	"xiaozhi-server-go/src/core/utils"
)

type Message = types.Message

// DialogueManager 管理对话上下文和历史
type DialogueManager struct {
	logger   *utils.Logger
	dialogue []Message
}

// NewDialogueManager 创建对话管理器实例
func NewDialogueManager() *DialogueManager {
	return &DialogueManager{
		dialogue: make([]Message, 0),
	}
}

func (dm *DialogueManager) SetSystemMessage(systemMessage string) {
	if systemMessage == "" {
		return
	}

	// 如果对话中已经有系统消息，则不再添加
	if len(dm.dialogue) > 0 && dm.dialogue[0].Role == "system" {
		dm.dialogue[0].Content = systemMessage
		return
	}

	// 添加新的系统消息到对话开头
	dm.dialogue = append([]Message{
		{Role: "system", Content: systemMessage},
	}, dm.dialogue...)
}

// 保留最近的几条对话消息
func (dm *DialogueManager) KeepRecentMessages(maxMessages int) {
	if maxMessages <= 0 || len(dm.dialogue) <= maxMessages {
		return
	}
	// 保留system消息和最近的 maxMessages 条消息
	if len(dm.dialogue) > 0 && dm.dialogue[0].Role == "system" {
		// 保留system消息
		dm.dialogue = append(dm.dialogue[:1], dm.dialogue[len(dm.dialogue)-maxMessages:]...)
		return
	}
	// 如果没有system消息，直接保留最近的 maxMessages 条消息
	if len(dm.dialogue) > maxMessages {
		dm.dialogue = dm.dialogue[len(dm.dialogue)-maxMessages:]
	}
}

// Put 添加新消息到对话
func (dm *DialogueManager) Put(message Message) {
	// 如果最近一条是user消息且当前也是user消息，则插入一个空的assistant消息
	if len(dm.dialogue) > 0 && dm.dialogue[len(dm.dialogue)-1].Role == "user" && message.Role == "user" {
		dm.dialogue = append(dm.dialogue, Message{Role: "assistant", Content: "..."})
	}
	dm.dialogue = append(dm.dialogue, message)
}

// GetLLMDialogue 获取完整对话历史
func (dm *DialogueManager) GetLLMDialogue() []Message {
	return dm.dialogue
}
