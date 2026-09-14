package executor

import (
	"context"
	"testing"
	"time"
)

func TestTruncateOutput(t *testing.T) {

	notOverLen := "This is a test string that is not over the maximum length."
	overLen := "This is a test string that is over the maximum length. It should be truncated to fit within the specified maximum length. The quick brown fox jumps over the lazy dog. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."

	maxLen := 200

	truncatedNotOver := TruncateOutput(notOverLen, maxLen)
	truncatedOver := TruncateOutput(overLen, maxLen)

	t.Logf("Truncated not over length string: %s", truncatedNotOver)

	t.Logf("Truncated over length string: %s", truncatedOver)
}

func TestExecCommand_SuccessAndTruncate(t *testing.T) {
	tempDir := t.TempDir()

	testExecutor := NewExecutor(tempDir)
	testExecutor.MaxOutputLen = 50 // 设置最大输出长度为 50

	// 测试成功执行命令
	successCommand := "echo 'Hello, World! This is a test string that is over the maximum length!'"

	testCtx := context.Background()

	output, err := testExecutor.ExecCommand(testCtx, successCommand)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	t.Log(output)

}

func TestExecCommand_Timeout(t *testing.T) {
	tempDir := t.TempDir()

	testExecutor := NewExecutor(tempDir)
	testExecutor.MaxOutputLen = 500               // 设置最大输出长度为 500
	testExecutor.Timeout = 100 * time.Millisecond // 设置超时时间为 100 毫秒

	// 测试成功执行命令
	successCommand := "ping 127.0.0.1 -n 3"

	testCtx := context.Background()

	output, err := testExecutor.ExecCommand(testCtx, successCommand)
	if err == nil {
		t.Errorf("Expected timeout error, got no error")
	}
	t.Logf("Output: %s", output)
	t.Logf("Error: %v", err)
}

func TestFileOperations_And_Security(t *testing.T) {
	tempDir := t.TempDir()
	//正常读写流
	testExecutor := NewExecutor(tempDir)
	testExecutor.MaxOutputLen = 500                // 设置最大输出长度为 500
	testExecutor.Timeout = 2000 * time.Millisecond // 设置超时时间为 2000 毫秒

	testExecutor.WriteFile("sub/test/demo.txt", "content test")
	testRead, err := testExecutor.ReadFile("sub/test/demo.txt")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	t.Logf("Read content: %s", testRead)

	//测试越界读写
	err = testExecutor.WriteFile("../demo.txt", "content test")
	if err == nil {
		t.Errorf("Expected error for writing outside workdir, got no error")
	}
	t.Logf("Write outside workdir error: %v", err)

	testRead, err = testExecutor.ReadFile("../demo.txt")
	if err == nil {
		t.Errorf("Expected error for reading outside workdir, got no error")
	}
	t.Logf("Read outside workdir error: %v", err)

	// 异常文件场景
	// 例如：尝试读取不存在的文件
	testRead, err = testExecutor.ReadFile("nonexistent.txt")
	if err == nil {
		t.Errorf("Expected error for reading nonexistent file, got no error")
	}
	t.Logf("Read nonexistent file error: %v", err)

}
