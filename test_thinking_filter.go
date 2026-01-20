package main

import (
	"fmt"
	"strings"
)

// 模拟思维链过滤逻辑
type thinkingTagState struct {
	insideThinking bool
	buffer         strings.Builder
}

var globalThinkingState = &thinkingTagState{}

func filterThinkingTagsRealtime(content string) string {
	if content == "" {
		return content
	}

	var result strings.Builder
	i := 0
	
	for i < len(content) {
		if globalThinkingState.insideThinking {
			// 当前在思维链内部，查找结束标签
			endIndex := strings.Index(content[i:], "</think>")
			if endIndex != -1 {
				// 找到结束标签，跳过思维链内容和结束标签
				i += endIndex + len("</think>")
				globalThinkingState.insideThinking = false
				globalThinkingState.buffer.Reset()
			} else {
				// 没找到结束标签，整个chunk都在思维链内部
				return ""
			}
		} else {
			// 当前在思维链外部，查找开始标签
			startIndex := strings.Index(content[i:], "<think>")
			if startIndex != -1 {
				// 找到开始标签，保留之前的内容
				result.WriteString(content[i : i+startIndex])
				i += startIndex + len("<think>")
				globalThinkingState.insideThinking = true
			} else {
				// 没找到开始标签，保留剩余内容
				result.WriteString(content[i:])
				break
			}
		}
	}

	return result.String()
}

func main() {
	// 重置状态
	globalThinkingState.insideThinking = false
	globalThinkingState.buffer.Reset()

	// 测试用例：模拟流式输出的chunks
	chunks := []string{
		"<think> 好的，用户打招呼说"你好啊"，我需要友好回应。先确认是否有需要调用工具的情况，但看起来只是普通问候，直接回复即可。保持自然，用中文回应。  "你好！有什么我可以帮助你的吗？"这样既礼貌又开放，鼓励用户进一步说明需求。暂时不需要调用任何工具，直接回复即可。 </think>  你好！有什么我可以帮助你的吗？",
	}

	fmt.Println("=== 测试思维链过滤 ===")
	
	for i, chunk := range chunks {
		fmt.Printf("Chunk %d 输入: %q\n", i+1, chunk)
		filtered := filterThinkingTagsRealtime(chunk)
		fmt.Printf("Chunk %d 输出: %q\n", i+1, filtered)
		fmt.Printf("状态: insideThinking=%v\n\n", globalThinkingState.insideThinking)
	}

	// 测试跨chunk的情况
	fmt.Println("=== 测试跨chunk思维链 ===")
	
	// 重置状态
	globalThinkingState.insideThinking = false
	globalThinkingState.buffer.Reset()
	
	crossChunks := []string{
		"<think> 这是思维链的开始",
		"，继续思维过程",
		"，思维链结束 </think> 这是正常内容",
	}
	
	for i, chunk := range crossChunks {
		fmt.Printf("Chunk %d 输入: %q\n", i+1, chunk)
		filtered := filterThinkingTagsRealtime(chunk)
		fmt.Printf("Chunk %d 输出: %q\n", i+1, filtered)
		fmt.Printf("状态: insideThinking=%v\n\n", globalThinkingState.insideThinking)
	}
}