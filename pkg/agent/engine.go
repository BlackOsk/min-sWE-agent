package agent

import (
	"context"
	"fmt"
	"minsweagent/pkg/executor"
	"minsweagent/pkg/llm"
	"minsweagent/pkg/parser"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const SystemPrompt = `你是一个专业的 SWE Agent (软件工程智能助手)。你拥有在一个隔离终端中操作项目代码库的能力。

你可以使用以下工具来完成用户的任务：
1. exec: 执行 Shell 命令
   参数: {"command": "要执行的命令"}
2. read_file: 读取文件内容
   参数: {"path": "相对工作区的操作路径"}
3. write_file: 写入/覆盖文件内容，若文件不存在，则创建新文件
   参数: {"path": "相对工作区的操作路径", "content": "要写入的文件内容"}

【输出协议规范】
你的每次响应必须严格包含 <thought> 标签说明推理逻辑。
如果需要调用工具，必须包含 <call name="工具名">JSON参数</call> 标签。
如果任务已经完成，或者不需要调用工具，则不要输出 <call> 标签。

格式示例1：
<thought>
我需要查看当前目录下的文件列表
</thought>
<call name="exec">
{
  "command": "ls -la"
}
</call>

格式示例2：
<thought>
我需要查看"/testRead"目录下的文件"testRead.txt"的内容
</thought>
<call name="read_file">
{
  "path": "testRead/testRead.txt"
}
</call>

格式示例3：
<thought>
我需要将文本写入"/testWrite"目录下的文件"testWrite.txt"的内容
</thought>
<call name="write_file">
{
  "path": "testWrite/testWrite.txt",
  "content": "这是要写入的文本内容"
}
</call>
【任务规划与状态追踪规范 (Task Planner)】
当面对包含 2 个以上步骤的复杂任务时，你必须按以下标准流程执行：
1. 【规划阶段】：首先使用 write_file 工具在工作区根目录创建 "todo.md"，将任务拆解为具体的子任务清单（使用 markdown 复选框格式 "- [ ] 子任务"）。
2. 【执行与追踪】：每完成一个子任务，在后续步骤中更新 "todo.md"（将完成项改为 "- [x]"），以保持上下文的清晰连贯。
`

type Engine struct {
	llmClient   *llm.Client
	executor    *executor.Executor
	maxSteps    int
	traceLogger *TraceLogger
}

func NewEngine(llmClient *llm.Client, exec *executor.Executor, maxSteps int, traceLogger *TraceLogger) *Engine {
	if maxSteps <= 0 {
		maxSteps = 10
	}
	return &Engine{
		llmClient:   llmClient,
		executor:    exec,
		maxSteps:    maxSteps,
		traceLogger: traceLogger,
	}
}

// executeToolCall 解析 ToolCall 并路由到 Executor 执行，返回执行结果字符串
func (e *Engine) executeToolCall(ctx context.Context, call *parser.ToolCall) string {
	if call == nil {
		return ""
	}

	switch call.Name {
	case "exec":
		cmd, ok := call.Args["command"]
		if !ok {
			return "Error: missing 'command' argument for exec tool"
		}
		out, err := e.executor.ExecCommand(ctx, cmd)
		if err != nil {
			return fmt.Sprintf("Command failed with error (%v):\n%s", err, out)
		}
		return fmt.Sprintf("Command output:\t%s", out)

	case "read_file":
		// 获取 path 参数并调用 e.executor.ReadFile
		path, ok := call.Args["path"]
		if !ok {
			return "Error: missing 'path' argument for read_file tool"
		}
		content, err := e.executor.ReadFile(path)
		if err != nil {
			return fmt.Sprintf("Read file failed with error (%v)", err)
		}
		return fmt.Sprintf("File content:\t%s", content)

	case "write_file":
		// 获取 path 和 content 参数并调用 e.executor.WriteFile
		path, ok := call.Args["path"]
		if !ok {
			return "Error: missing 'path' argument for write_file tool"
		}
		content, ok := call.Args["content"]
		if !ok {
			return "Error: missing 'content' argument for write_file tool"
		}
		err := e.executor.WriteFile(path, content)
		if err != nil {
			return fmt.Sprintf("Write file failed with error (%v)", err)
		}
		return "File written successfully"

	default:
		return fmt.Sprintf("Error: unknown tool '%s'", call.Name)
	}
}

// Run 启动 ReAct 主循环
func (e *Engine) Run(ctx context.Context, task string) (string, error) {

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: SystemPrompt},
		{Role: openai.ChatMessageRoleUser, Content: task},
	}

	fmt.Printf("\n启动 Agent 任务: %s\n", task)
	fmt.Println(strings.Repeat("=", 60))

	for step := 0; step < e.maxSteps; step++ {

		fmt.Printf("\n[Step %d/%d]\n", step+1, e.maxSteps)
		// 发起 LLM 请求前，先对上下文进行裁剪清洗
		prunedMessages := e.pruneMessages(messages)

		// 调用 e.llmClient.Completion(ctx, messages) 获取 LLM 输出
		llmOutput, err := e.llmClient.Completion(ctx, prunedMessages)
		if err != nil {
			return "", fmt.Errorf("LLM completion failed: %w", err)
		}

		// 将 LLM 输出追加到 messages 历史中 (LLM Role: Assistant)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: llmOutput,
		})

		// 调用 parser.Parse(llmOutput) 解析结果
		result, err := parser.Parse(llmOutput)
		if err != nil {
			return "", fmt.Errorf("failed to parse LLM output: %w", err)
		}

		// 检查 result.ToolCall
		//         - 若 result.ToolCall == nil，说明模型完成了推理/回答，直接 return result.Thought, nil
		//         - 若 result.ToolCall != nil，调用 e.executeToolCall(ctx, result.ToolCall) 获取结果
		var toolResult string
		if result.ToolCall == nil {
			fmt.Println("任务完成，无后续动作。")
			fmt.Println(strings.Repeat("=", 60))
			if result.Thought != "" {
				return result.Thought, nil
			}
			return llmOutput, nil
		} else {

			fmt.Printf("Thought: %s\n", result.Thought)
			fmt.Printf("Tool Call: [%s] | Args: %v\n", result.ToolCall.Name, result.ToolCall.Args)
			toolResult = e.executeToolCall(ctx, result.ToolCall)
			fmt.Printf("Tool Output:\n\t%s\n", toolResult)

			// 将工具执行结果作为 User 消息追加到 messages 中：
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Tool Response:\n\t %s", toolResult),
			})
		}

		e.traceLogger.LogStep(step+1, result.Thought, result.ToolCall.Name, result.ToolCall.Args, toolResult)

	}

	return "", fmt.Errorf("reached maximum steps (%d) without completing the task", e.maxSteps)
}

// pruneMessages 对历史对话中的冗余 Tool Output 进行清理，防止上下文无限膨胀
func (e *Engine) pruneMessages(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	// 消息数量较少时不做处理（如小于 6 条）
	if len(messages) <= 6 {
		return messages
	}

	pruned := make([]openai.ChatCompletionMessage, len(messages))
	copy(pruned, messages)

	// 保留最新 4 条消息（即最近约 2 轮交互）不压缩
	keepRecentIndex := len(pruned) - 4

	// 遍历中间的历史消息 (跳过 index 0 的 System Prompt 和 index 1 的初始 Task)
	for i := 2; i < keepRecentIndex; i++ {
		msg := &pruned[i]
		// 仅对 User 角色且包含 Tool Response 的历史长文本进行清理
		if msg.Role == openai.ChatMessageRoleUser && strings.HasPrefix(msg.Content, "Tool Response:\n") {
			// 如果历史工具输出超过 300 字符，进行压缩处理,保留前后 150 字符，并在末尾添加提示
			if len(msg.Content) > 300 {
				header := msg.Content[:150]
				tail := msg.Content[len(msg.Content)-150:]
				msg.Content = fmt.Sprintf("%s\n\n... [Historical Tool Log Pruned to save Context] ...\n\n%s", header, tail)
			}
		}
	}

	return pruned
}
