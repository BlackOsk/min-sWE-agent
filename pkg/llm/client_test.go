package llm

import (
	"context"

	"os"

	"testing"

	"time"

	"github.com/sashabaranov/go-openai"
)

func TestClient(t *testing.T) {

	apiKey := os.Getenv("SILICONFLOW_API_KEY")

	if apiKey == "" {

		t.Skip("SILICONFLOW_API_KEY is not set, skipping integration test")

	}

	testClient, err := NewClient(apiKey, "", "")

	if err != nil {

		t.Fatalf("Failed to create client: %v", err)

	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	messages := []openai.ChatCompletionMessage{

		{Role: openai.ChatMessageRoleUser, Content: "你好，请介绍一下你自己"},
	}

	response, err := testClient.Completion(ctx, messages)

	if err != nil {

		t.Errorf("Error during completion: %v", err)

	}

	t.Logf("LLM respons: %v\n", response)

}
