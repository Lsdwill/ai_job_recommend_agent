package utils

import (
	"testing"
)

func TestFilterThinkingTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "无思维链标签",
			input:    "这是一个正常的回复，没有思维链。",
			expected: "这是一个正常的回复，没有思维链。",
		},
		{
			name: "包含思维链标签",
			input: `<think>好吧，用户问石河子有什么薪资大于10000的岗位。首先，我需要明确用户的需求，他们可能是在寻找高薪的工作机会。</think>

好的，我将为您查询石河子市薪资大于10000的岗位信息。`,
			expected: "好的，我将为您查询石河子市薪资大于10000的岗位信息。",
		},
		{
			name: "多个思维链标签",
			input: `<think>第一段思维</think>

这是正常内容。

<think>第二段思维</think>

这是另一段正常内容。`,
			expected: "这是正常内容。\n\n这是另一段正常内容。",
		},
		{
			name: "思维链标签跨行",
			input: `<think>
这是一个
跨行的思维链
包含多行内容
</think>

正常的回复内容在这里。`,
			expected: "正常的回复内容在这里。",
		},
		{
			name: "思维链标签在中间",
			input: `开始的内容。

<think>中间的思维过程</think>

结束的内容。`,
			expected: "开始的内容。\n\n结束的内容。",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterThinkingTags(tt.input)
			if result != tt.expected {
				t.Errorf("FilterThinkingTags() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestContainsThinkingTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: false,
		},
		{
			name:     "无思维链标签",
			input:    "这是一个正常的回复。",
			expected: false,
		},
		{
			name:     "包含思维链标签",
			input:    "<think>这是思维过程</think>正常内容",
			expected: true,
		},
		{
			name:     "只有开始标签",
			input:    "<think>这是不完整的思维",
			expected: false,
		},
		{
			name:     "只有结束标签",
			input:    "这是不完整的思维</think>",
			expected: false,
		},
		{
			name:     "多个思维链标签",
			input:    "<think>第一个</think>内容<think>第二个</think>",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsThinkingTags(tt.input)
			if result != tt.expected {
				t.Errorf("ContainsThinkingTags() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
