package llm

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	apiClient *openai.Client
	model     string
}

// NewClient 初始化硅基流动 API 客户端
func NewClient(apiKey, baseURL, model string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("api key is required")
	}
	if baseURL == "" {
		baseURL = "https://api.siliconflow.cn/v1"
	}
	if model == "" {
		model = "deepseek-ai/DeepSeek-V4-Flash"
	}

	// 基于 openai.DefaultConfig(apiKey) 配置 BaseURL 并创建 Client
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	apiClient := openai.NewClientWithConfig(config)
	client := &Client{
		apiClient: apiClient,
		model:     model,
	}

	return client, nil
}

// Completion 发送对话历史，等待 LLM 的文本响应，然后返回 LLM 的文本响应
func (c *Client) Completion(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	// TODO: 构造 ChatCompletionRequest 并调用 CreateChatCompletion，提取第一个 choice 的 content 文本

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      false,
		Temperature: 0.2,
		MaxTokens:   4096,
	}
	resp, err := c.apiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("llm returned empty choices")
	}

	if len(resp.Choices[0].Message.Content) == 0 {
		return "", errors.New("llm returned empty content")
	}
	return resp.Choices[0].Message.Content, nil

}
