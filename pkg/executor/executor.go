package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Executor struct {
	WorkDir      string        // Agent 工作目录
	Timeout      time.Duration // 命令执行超时时间，默认 30s
	MaxOutputLen int           // 输出最大长度，超过则截断，默认 2000
}

func NewExecutor(workDir string) *Executor {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	return &Executor{
		WorkDir:      workDir,
		Timeout:      60 * time.Second,
		MaxOutputLen: 2000,
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

	// 区分运行系统
	// 创建 bash/sh 或 cmd 执行对象，设置 e.WorkDir 为 Cmd.Dir
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctxWithTimeout, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctxWithTimeout, "sh", "-c", command)
	}
	cmd.Dir = e.WorkDir

	// 捕获 CombinedOutput (包含 stdout 和 stderr)
	output, err := cmd.CombinedOutput()
	truncate := TruncateOutput(string(output), e.MaxOutputLen)
	if err != nil {
		// 先判断是否是超时错误，再返回截断后的输出和错误信息
		if errors.Is(ctxWithTimeout.Err(), context.DeadlineExceeded) {
			return truncate, fmt.Errorf("command execution timed out after %v: %w", e.Timeout, ctxWithTimeout.Err())
		}
		return truncate, fmt.Errorf("command execution failed (%v)", err)
	}
	//  对输出进行 TruncateOutput 截断并返回
	return TruncateOutput(string(output), e.MaxOutputLen), nil
}

// ReadFile 读取指定文件内容
func (e *Executor) ReadFile(filePath string) (string, error) {
	// TODO: 校验并读取 e.WorkDir 路径下的文件内容

	absPath := filepath.Join(e.WorkDir, filePath)

	// 检测路径保持在工作区内
	rel, err := filepath.Rel(e.WorkDir, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("安全拦截：试图读取工作区之外的文件: %s", filePath)
	}

	// 校验文件存在
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", filePath)
		} else {
			return "", fmt.Errorf("failed to stat file: %w", err)
		}

	}
	// 校验不是目录
	if info.IsDir() {
		return "", fmt.Errorf("path %s is a directory, not a file", filePath)
	}

	// 读取文件内容
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 对过长文本进行截取（设置100000字节为上限）
	content := string(data)
	truncatedContent := TruncateOutput(content, 100000)

	return truncatedContent, nil
}

// WriteFile 覆盖或新建文件
func (e *Executor) WriteFile(filePath string, content string) error {
	// TODO: 在 e.WorkDir 路径下写入文件，若目录不存在需自动创建 (os.MkdirAll)
	absPath := filepath.Join(e.WorkDir, filePath)

	// 检测路径保持在工作区内
	rel, err := filepath.Rel(e.WorkDir, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("安全拦截：试图写入工作区之外的文件: %s", filePath)
	}

	// 创建目录
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件内容
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
