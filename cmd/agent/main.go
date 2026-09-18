package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"minsweagent/pkg/agent"
	"minsweagent/pkg/executor"
	"minsweagent/pkg/llm"
	"os"
)

func main() {
	baseURL := "https://api.siliconflow.cn/v1"

	// 1. 解析命令行参数
	taskFlag := flag.String("task", "", "要执行的任务描述 (例如: '帮我编写一个 Golang HTTP 服务并进行测试')")
	workDirFlag := flag.String("dir", ".", "Agent 执行工作的相对/绝对路径")
	modelFlag := flag.String("model", "deepseek-ai/DeepSeek-V4-Flash", "使用的 LLM 模型名称")
	maxStepsFlag := flag.Int("max-steps", 20, "ReAct 引擎的最大轮转步数")
	flag.Parse()

	if *taskFlag == "" {
		fmt.Println("❌ 错误: 必须通过 -task 参数指定任务描述")
		flag.Usage()
		os.Exit(1)
	}

	// 2. 获取 API Key
	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ 错误: 未找到 SILICONFLOW_API_KEY 环境变量，请先设置后再试！")
	}

	// 初始化 llm.NewClient
	AgentClient, err := llm.NewClient(apiKey, baseURL, *modelFlag)
	if err != nil {
		log.Printf("Failed to create LLM client: %v", err)
	}

	// 初始化 executor.NewExecutor
	AgentExecutor := executor.NewExecutor(*workDirFlag)
	// 初始化 agent.NewEngine
	AgentEngine := agent.NewEngine(AgentClient, AgentExecutor, *maxStepsFlag)

	// 启动 engine.Run(context.Background(), *taskFlag) 并打印最终结果
	ctx := context.Background()
	result, err := AgentEngine.Run(ctx, *taskFlag)
	log.Printf("Engine run result: %s", result)

}
