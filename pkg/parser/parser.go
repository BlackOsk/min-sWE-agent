package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ToolCall 代表一次被解析出的工具调用动作
type ToolCall struct {
	Name string            `json:"name"`
	Args map[string]string `json:"args"`
}

// ParseResult 包含从 LLM 原始文本中提取的所有结构化信息
type ParseResult struct {
	Thought  string    `json:"thought"`
	ToolCall *ToolCall `json:"tool_call,omitempty"`
}

var (
	thoughtRegexp = regexp.MustCompile(`(?s)<thought>(.*?)</thought>`)
	callRegexp    = regexp.MustCompile(`(?s)<call\s+name\s*=\s*"([^"]+)"\s*>(.*?)</call>`)
)

func NewParseResult() *ParseResult {
	return &ParseResult{
		Thought:  "",
		ToolCall: nil,
	}
}

// Parse 解析 LLM 输出的原始文本，提取 Thought 与 ToolCall
func Parse(llmOutput string) (*ParseResult, error) {
	result := NewParseResult()

	// 使用 thoughtRegexp 提取 <thought> 中的内容，并使用 strings.TrimSpace 去除多余空白
	matchThought := thoughtRegexp.FindStringSubmatch(llmOutput)
	if len(matchThought) > 1 {
		result.Thought = strings.TrimSpace(matchThought[1])
	}

	// 使用 callRegexp 匹配 <call name="...">内容</call>
	matchCall := callRegexp.FindStringSubmatch(llmOutput)
	if len(matchCall) > 2 {
		toolName := matchCall[1]
		jsonArgStr := matchCall[2]

		args := make(map[string]string)

		err := json.Unmarshal([]byte(jsonArgStr), &args)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON arguments: %v", err)
		}

		result.ToolCall = &ToolCall{Name: toolName, Args: args}
	}

	return result, nil
}
