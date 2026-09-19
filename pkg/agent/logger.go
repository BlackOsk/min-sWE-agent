package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TraceLogger struct {
	file *os.File
}

func NewTraceLogger(workDir string, task string) (*TraceLogger, error) {
	filename := fmt.Sprintf("trace_%s.md", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(workDir, filename)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("TraceLogger error : %v", err)
	}

	logger := &TraceLogger{file: f}
	logger.WriteHeader(task)
	return logger, nil
}

func (l *TraceLogger) WriteHeader(task string) {
	header := fmt.Sprintf("# Agent Task Execution Trace\n- **Task**: %s\n- **Start Time**: %s\n\n---\n\n",
		task, time.Now().Format("2006-01-02 15:04:05"))
	l.file.WriteString(header)
}

func (l *TraceLogger) LogStep(step int, thought string, toolName string, args map[string]string, observation string) {
	entry := fmt.Sprintf("## Step %d\n", step)
	if thought != "" {
		entry += fmt.Sprintf("### Thought\n%s\n\n", thought)
	}
	if toolName != "" {
		entry += "### Action\n"
		entry += fmt.Sprintf("- **Tool**: `%s`\n", toolName)

		// 针对不同工具类型做语义化排版
		switch toolName {
		case "write_file":
			path := args["path"]
			content := args["content"]
			lang := getLanguageTag(path)

			entry += fmt.Sprintf("- **Path**: `%s`\n", path)
			entry += fmt.Sprintf("- **Content**:\n```%s\n%s\n```\n\n", lang, content)

		case "read_file":
			entry += fmt.Sprintf("- **Path**: `%s`\n\n", args["path"])

		case "exec":
			entry += fmt.Sprintf("- **Command**:\n```bash\n%s\n```\n\n", args["command"])

		default:
			// 通用工具处理：禁用 HTML 转义以确保 < > 正常显示
			entry += fmt.Sprintf("- **Args**: `%s`\n\n", formatUnescapedJSON(args))
		}
	}
	if observation != "" {
		entry += fmt.Sprintf("### Observation\n```text\n%s\n```\n\n", observation)
	}
	entry += "---\n\n"
	l.file.WriteString(entry)
}

func (l *TraceLogger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// getLanguageTag 根据文件后缀推导 Markdown 代码块的高亮语言标记
func getLanguageTag(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".json":
		return "json"
	case ".sh", ".bash":
		return "bash"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".md":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	default:
		return ""
	}
}

// formatUnescapedJSON 将 map 转为 JSON 文本，禁用 HTML 转义 (< > 不会被替换为 \u003c \u003e)
func formatUnescapedJSON(v interface{}) string {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false) // 禁用 HTML 字符转义
	if err := enc.Encode(v); err != nil {
		return fmt.Sprintf("%v", v)
	}
	return strings.TrimSpace(buf.String())
}
