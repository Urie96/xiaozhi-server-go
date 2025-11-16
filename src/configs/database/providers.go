package database

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

func RemoveSensitiveFields(data datatypes.JSON) (string, error) {
	configData := make(map[string]interface{})
	if err := json.Unmarshal(data, &configData); err != nil {
		return "", fmt.Errorf("反序列化ASR提供者数据失败: %v", err)
	}
	// 移除敏感字段
	delete(configData, "token")
	delete(configData, "access_token")
	delete(configData, "api_key")
	delete(configData, "appid")

	// 将处理后的数据重新序列化为JSON字符串
	configJson, err := json.Marshal(configData)
	if err != nil {
		return "", fmt.Errorf("序列化处理后的ASR提供者数据失败: %v", err)
	}
	return string(configJson), nil
}
