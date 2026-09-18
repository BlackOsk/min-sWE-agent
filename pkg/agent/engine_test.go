package agent

import (
	"context"
	"minsweagent/pkg/executor"
	"minsweagent/pkg/llm"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
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
	testPrompt := `请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 "print('Hello, World!')"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。`
	result, err := testEngine.Run(testCtx, testPrompt)
	if err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}
	t.Logf("Engine run result: %s", result)
}

func TestToolCallExecution(t *testing.T) {
	// 1. 使用 t.TempDir() 自动生成与清理临时目录
	tempDir := t.TempDir()
	testExecutor := executor.NewExecutor(tempDir)

	// 2. 写入测试 py 文件
	err := testExecutor.WriteFile("hello.py", "print('Hello, World!')")
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	ctx := context.Background()

	// 3. 直接写相对文件名即可，不需要重复拼接 tempDir
	pythonPath := "D:\\anaconda\\envs\\migrate1\\python.exe"
	commandString := pythonPath + " hello.py"

	str, err := testExecutor.ExecCommand(ctx, commandString)
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	t.Logf("Command output: %s", str)
}

func TestPruneMessages(t *testing.T) {
	SILICONFLOW_API_KEY := os.Getenv("SILICONFLOW_API_KEY")
	tempDir := t.TempDir()
	//testCtx := context.Background()

	testClient, err := llm.NewClient(SILICONFLOW_API_KEY, "", "")
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}

	testExecutor := executor.NewExecutor(tempDir)

	testEngine := NewEngine(testClient, testExecutor, 10)

	testMessage := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: SystemPrompt},
		{Role: openai.ChatMessageRoleUser, Content: "请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目录下创建Tool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAW;EOIGJAQW;POEIUFHALKGJawilouhgv;owaieghnap;eoirsghanploei;htjiauw4y6t0-2w47u6yhw49j8roiae5usertswryjwq458ir6k3w456ike6thq357u5rhts请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrknhygtweTool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrkofijpklsdcdr6tfyughbjinkkyhbgvfctrd5yvhujio98klyhgtrfd567ol;8io775tr fv7bm,nsze w jnmdxebhnuytjuaz 5r68 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输Tool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk目录下创Tool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目录下创建一个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目录下创建一Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令Tool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "Tool Response:\n请在当前目录下创建一Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令Tool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "请在当前目录下创建一Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令Tool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
		{Role: openai.ChatMessageRoleUser, Content: "请在当前目录下创建一Tool Response:\nsdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk个名为 “hello.py” 的文件，写入内容 \"print('Hello, World!')\"，然后在命令Tool Response:sdlfksjdfsldkfjsldkvnal;weoig8jqaeo;rigkhAWaworegiu20-87utip9johglrk行中调用 python 执行这个文件，并确认输出结果是否和预期一致。请严格按照输出协议规范返回结果。"},
	}

	prunedMessages := testEngine.pruneMessages(testMessage)

	t.Logf("length of prunedMessages: %v", len(prunedMessages))
	for _, msg := range prunedMessages {
		t.Logf("Role: %s, Content: %s", msg.Role, msg.Content)
	}

}
