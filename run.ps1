# run.ps1 - 自动识别脚本所在目录并启动沙箱
param (
    # 默认路径自动设为脚本所在目录 ($PSScriptRoot)
    [string]$WorkDir =$PSScriptRoot,
    [string]$Model   = "deepseek-ai/DeepSeek-V4-Flash"
)

# 路径兜底：如果 $PSScriptRoot 为空，自动退回当前终端所在路径
if (-not $WorkDir) { $WorkDir = $PSScriptRoot }
if (-not $WorkDir) { $WorkDir = (Get-Location).Path }

# 1. 检查 API Key 环境变量
if (-not $env:SILICONFLOW_API_KEY) {
    Write-Host "错误: 当前终端未检测到 SILICONFLOW_API_KEY 环境变量！" -ForegroundColor Red
    Write-Host "提示: 请先运行 `$env:SILICONFLOW_API_KEY='your_key_here'" -ForegroundColor Yellow
    exit 1
}

# 2. 获取工作区绝对路径
$AbsWorkDir = (Resolve-Path $WorkDir -ErrorAction SilentlyContinue).Path
if (-not $AbsWorkDir) {
    Write-Host "错误: 指定的工作区目录 [$WorkDir] 不存在！" -ForegroundColor Red
    exit 1
}

# 3. 检查与脚本同目录下的 prompt.txt
$PromptPath = Join-Path $AbsWorkDir "prompt.txt"
if (-not (Test-Path $PromptPath)) {
    Write-Host "错误: 在 [$AbsWorkDir] 未找到 prompt.txt 文件！" -ForegroundColor Red
    exit 1
}

Write-Host "当前工作区: $AbsWorkDir" -ForegroundColor Green
Write-Host "调用的模型: $Model" -ForegroundColor Cyan
Write-Host "------------------------------------------------------------" -ForegroundColor Gray

# 4. 启动 Docker 沙箱
docker run --rm `
  -e SILICONFLOW_API_KEY="$env:SILICONFLOW_API_KEY" `
  -v "${AbsWorkDir}:/workspace" `
  minimal-swe-agent:v1 `
  -task-file "/workspace/prompt.txt" `
  -model "$Model"