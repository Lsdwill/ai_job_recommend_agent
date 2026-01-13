package utils

import (
	"encoding/json"
	"fmt"
)

// ToJSONStringPretty 将对象转换为格式化的JSON字符串
func ToJSONStringPretty(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %w", err)
	}
	return string(data), nil
}
