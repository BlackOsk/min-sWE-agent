package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Executor struct {
	WorkDir string        // Agent 工作目录
	Timeout time.Duration // 命令执行超时时间，默认 30s
}

func NewExecutor(workDir string) *Executor {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	return &Executor{
		WorkDir: workDir,
		Timeout: 30 * time.Second,
	}
}

// TruncateOutput 智能截断过长输出，保留头部与尾部
func TruncateOutput(output string, maxLen int) string {
	if len(output) <= maxLen {
		return output
	}
	// 截断长输出逻辑
	return output[:maxLen/3] + "[... Output Truncated ...]" + output[len(output)-maxLen*2/3:]
}

// ExecCommand 执行 Shell 命令并返回结果（带超时控制与智能截断）
func (e *Executor) ExecCommand(ctx context.Context, command string) (string, error) {
	// 基于 e.Timeout 创建 WithTimeout context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	// 创建 bash/sh 或 cmd 执行对象，设置 e.WorkDir 为 Cmd.Dir
	cmd := exec.CommandContext(ctxWithTimeout, "bash", "-c", command)
	cmd.Dir = e.WorkDir

	// 捕获 CombinedOutput (包含 stdout 和 stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}
	//  对输出进行 TruncateOutput 截断并返回
	return TruncateOutput(string(output), int(e.Timeout)), nil
}

// ReadFile 读取指定文件内容
func (e *Executor) ReadFile(filePath string) (string, error) {
	// TODO: 校验并读取 e.WorkDir 路径下的文件内容
	return "", nil
}

// WriteFile 覆盖或新建文件
func (e *Executor) WriteFile(filePath string, content string) error {
	// TODO: 在 e.WorkDir 路径下写入文件，若目录不存在需自动创建 (os.MkdirAll)
	return nil
}
