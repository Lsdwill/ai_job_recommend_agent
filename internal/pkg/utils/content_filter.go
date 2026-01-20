package utils

import (
	"regexp"
	"strings"
)

// ThinkingTagPattern 匹配思维链标签的正则表达式
var ThinkingTagPattern = regexp.MustCompile(`<think>[\s\S]*?</think>`)

// FilterThinkingTags 过滤内容中的思维链标签
func FilterThinkingTags(content string) string {
	if content == "" {
		return content
	}

	// 移除 <think>...</think> 标签及其内容
	filtered := ThinkingTagPattern.ReplaceAllString(content, "")

	// 清理多余的空行
	filtered = strings.TrimSpace(filtered)

	// 将多个连续换行符替换为最多两个换行符
	multipleNewlines := regexp.MustCompile(`\n{3,}`)
	filtered = multipleNewlines.ReplaceAllString(filtered, "\n\n")

	return filtered
}

// ContainsThinkingTags 检查内容是否包含思维链标签
func ContainsThinkingTags(content string) bool {
	return ThinkingTagPattern.MatchString(content)
}
