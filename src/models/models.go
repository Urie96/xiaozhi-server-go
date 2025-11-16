package models

import (
	//"gorm.io/gorm"
	"time"
)

// 智能体结构：智能体属于某个用户，拥有多个设备
type Agent struct {
	ID                 uint      `gorm:"primaryKey"           json:"id"`
	Name               string    `gorm:"not null"             json:"name"` // 智能体名称
	LLM                string    `gorm:"default:'ChatGLMLLM'" json:"LLM"`
	Language           string    `gorm:"default:'普通话'"        json:"language"`                            // 语言，默认为中文
	Voice              string    `gorm:"default:'zh_female_wanwanxiaohe_moon_bigtts'"       json:"voice"` // 语音，默认为zh_female_wanwanxiaohe_moon_bigtts
	VoiceName          string    `gorm:"default:'湾湾小何'"       json:"voiceName"`                           // 语音，默认为湾湾小何
	Prompt             string    `gorm:"type:text"            json:"prompt"`
	ASRSpeed           int       `gorm:"default:2"            json:"asrSpeed"`   // ASR 语音识别速度，1=耐心，2=正常，3=快速
	SpeakSpeed         int       `gorm:"default:2"            json:"speakSpeed"` // TTS 角色语速，1=慢速，2=正常，3=快速
	Tone               int       `gorm:"default:50"           json:"tone"`       // TTS 角色音调，1-100，低音-高音
	UserID             uint      `gorm:"not null"             json:"-"`
	CreatedAt          time.Time `                            json:"createdAt"`          // 创建时间
	UpdatedAt          time.Time `                            json:"updatedAt"`          // 更新时间
	LastConversationAt time.Time `                            json:"lastConversationAt"` // 最后对话时间
	EnabledTools       string    `gorm:"type:text"            json:"enabledTools"`       // 启用的工具列表，字符串格式，如 "tool1,tool2"
	Conversationid     string    `                            json:"conversationId"`     // 关联的对话AgentDialog的ID
	HeadImg            string    `gorm:"type:varchar(255)"    json:"head_img"`           // 头像URL
	Description        string    `gorm:"type:text"            json:"description"`        // 智能体描述
	CatalogyID         uint      `                            json:"catalogy_id"`        // 分类ID
	Extra              string    `gorm:"type:text"            json:"extra"`              // 额外信息，JSON格式
}
