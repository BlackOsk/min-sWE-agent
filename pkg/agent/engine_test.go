package agent

import (
	"context"
	"minsweagent/pkg/executor"
	"minsweagent/pkg/llm"
	"os"
	"testing"
)

func TestEngine(t *testing.T) {

	SILICONFLOW_API_KEY := os.Getenv("SILICONFLOW_API_KEY")
	tempDir := t.TempDir()
	testCtx := context.Background()

	testClient, err := llm.NewClient(SILICONFLOW_API_KEY, "", "")
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}

	testExecutor := executor.NewExecutor(tempDir)

	testEngine := NewEngine(testClient, testExecutor, 10)

	// 测试用例：执行一个简单的命令
	testPrompt := `请在当前目录下创建一个叫 hello.py 的文件，写入内容 "print('Hello, World!')"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。`
	result, err := testEngine.Run(testCtx, testPrompt)
	if err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}
	t.Logf("Engine run result: %s", result)
}
