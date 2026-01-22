# mcp-pprof Usage Guide

[English](#english) | [中文](#中文)

---

<div id="english"></div>

## English

### Quick Start

#### 1. Build the Project

```bash
git clone https://github.com/gwork1883/mcp-pprof.git
cd mcp-pprof
make build
```

Or build manually:
```bash
go build -o mcp-pprof ./cmd/mcp-pprof
go build -o mcp-pprof-server ./cmd/mcp-pprof-server
```

#### 2. Configure AI Assistant

Add the following to your AI assistant's MCP configuration:

**For Cline (VS Code Extension):**
Open settings and add to `cline_mcp_settings.json`:

```json
{
  "mcpServers": {
    "pprof": {
      "command": "/absolute/path/to/mcp-pprof",
      "env": {},
      "disabled": false,
      "alwaysAllow": [
        "parse_profile",
        "top_functions", 
        "generate_svg",
        "analyze_performance",
        "compare_profiles",
        "list_callers"
      ]
    }
  }
}
```

**For Claude Desktop:**
Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or the config file in your settings directory:

```json
{
  "mcpServers": {
    "pprof": {
      "command": "/absolute/path/to/mcp-pprof",
      "args": [],
      "env": {}
    }
  }
}
```

#### 3. Collect pprof Data

Generate a CPU profile from your Go application:

```bash
# Run your Go program with CPU profiling enabled
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Or use pprof in your code
import _ "net/http/pprof"
# Then: curl http://localhost:6060/debug/pprof/profile > cpu.prof
```

#### 4. Ask AI to Analyze

Start a conversation with your AI assistant:

```
Analyze the CPU profile file at /path/to/cpu.prof and tell me about the top functions
```

```
Generate a flamegraph for /path/to/cpu.prof
```

### Available Tools

#### 1. parse_profile

Parse a pprof file and return structured data.

**Parameters:**
- `filePath` (required): Path to the pprof file
- `profileType` (optional, default: "auto"): Type of profile ("cpu", "heap", "block", "mutex", "goroutine", "auto")
- `outputFormat` (optional, default: "json"): Output format ("json", "text", "proto")

**Example:**
```
Parse the profile at /path/to/cpu.prof with auto-detect
```

#### 2. top_functions

Get the top N hot functions from a profile.

**Parameters:**
- `filePath` (required): Path to the pprof file
- `topN` (optional, default: 10): Number of top functions to return (1-100)

**Example:**
```
Show me the top 20 functions from /path/to/cpu.prof
```

#### 3. generate_svg

Generate an SVG flamegraph.

**Parameters:**
- `filePath` (required): Path to the pprof file
- `focus` (optional): Focus on a specific function or pattern (regex)
- `ignore` (optional): Ignore functions matching pattern (regex)

**Example:**
```
Generate a flamegraph for /path/to/cpu.prof, focusing on handlers
```

#### 4. analyze_performance

Perform deep performance analysis with suggestions.

**Parameters:**
- `filePath` (required): Path to the pprof file
- `focus` (optional, default: "all"): Analysis focus ("bottlenecks", "hotspots", "all")
- `threshold` (optional, default: 5): Percentage threshold for hotspot detection

**Example:**
```
Perform deep performance analysis on /path/to/cpu.prof with 10% threshold
```

#### 5. compare_profiles

Compare two profile files to find differences.

**Parameters:**
- `baseFile` (required): Base profile file path
- `compareFile` (required): Comparison profile file path

**Example:**
```
Compare before.prof and after.prof to see performance improvements
```

#### 6. list_callers

View function call relationships.

**Parameters:**
- `filePath` (required): Path to the pprof file
- `functionName` (required): Function name to list callers for
- `maxDepth` (optional, default: 10): Maximum depth of call stack

**Example:**
```
Show all callers of the function named "MainHandler" from /path/to/cpu.prof
```

### Remote Mode (mcp-remote)

For remote access, use mcp-remote with HTTP transport:

#### 1. Start HTTP Server

```bash
./mcp-pprof-server -port 8080
```

Options:
- `-port`: Port to listen on (default: 8080)
- `-address`: Address to bind to (default: 0.0.0.0)
- `-debug`: Enable debug logging

#### 2. Configure Client

```json
{
  "mcpServers": {
    "pprof": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-remote@latest",
        "http://localhost:8080/mcp",
        "--allow-http",
        "--reconnect",
        "--reconnect-delay",
        "5000"
      ]
    }
  }
}
```

### Common Use Cases

#### 1. Finding Performance Bottlenecks

```
Analyze the heap profile at memory.prof and identify memory allocation hotspots
```

#### 2. Before/After Comparison

```
Compare cpu-before.prof and cpu-after.prof to measure performance improvements
```

#### 3. Focus on Specific Functions

```
Generate a flamegraph for cpu.prof, focusing only on database related functions
```

### Tips

1. **Use meaningful profile names**: Include timestamp and type in filename (e.g., `cpu-2025-01-22.prof`)
2. **Focus on hot paths**: Use the `focus` parameter to zoom into specific areas
3. **Compare before/after**: Always compare profiles to measure optimization impact
4. **Regular profiling**: Profile during peak load to catch production issues

---

<div id="中文"></div>

## 中文

### 快速开始

#### 1. 构建项目

```bash
git clone https://github.com/gwork1883/mcp-pprof.git
cd mcp-pprof
make build
```

或手动构建：
```bash
go build -o mcp-pprof ./cmd/mcp-pprof
go build -o mcp-pprof-server ./cmd/mcp-pprof-server
```

#### 2. 配置 AI 助手

将以下配置添加到您的 AI 助手的 MCP 配置中：

**对于 Cline (VS Code 扩展):**
打开设置并编辑 `cline_mcp_settings.json`：

```json
{
  "mcpServers": {
    "pprof": {
      "command": "/绝对路径/to/mcp-pprof",
      "env": {},
      "disabled": false,
      "alwaysAllow": [
        "parse_profile",
        "top_functions", 
        "generate_svg",
        "analyze_performance",
        "compare_profiles",
        "list_callers"
      ]
    }
  }
}
```

**对于 Claude Desktop:**
编辑 `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) 或您设置目录中的配置文件：

```json
{
  "mcpServers": {
    "pprof": {
      "command": "/绝对路径/to/mcp-pprof",
      "args": [],
      "env": {}
    }
  }
}
```

#### 3. 收集 pprof 数据

从您的 Go 应用程序生成 CPU profile：

```bash
# 使用 pprof 工具采集30秒的 CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 或在代码中启用 pprof
import _ "net/http/pprof"
# 然后执行: curl http://localhost:6060/debug/pprof/profile > cpu.prof
```

#### 4. 让 AI 进行分析

与您的 AI 助手开始对话：

```
分析 /path/to/cpu.prof 文件，告诉我哪些函数占用最多时间
```

```
为 /path/to/cpu.prof 生成火焰图
```

### 可用工具

#### 1. parse_profile

解析 pprof 文件并返回结构化数据。

**参数：**
- `filePath` (必需): pprof 文件路径
- `profileType` (可选，默认: "auto"): profile 类型 ("cpu", "heap", "block", "mutex", "goroutine", "auto")
- `outputFormat` (可选，默认: "json"): 输出格式 ("json", "text", "proto")

**示例：**
```
解析 /path/to/cpu.prof 文件，自动检测类型
```

#### 2. top_functions

获取 profile 中的 Top N 热点函数。

**参数：**
- `filePath` (必需): pprof 文件路径
- `topN` (可选，默认: 10): 返回的函数数量 (1-100)

**示例：**
```
显示 /path/to/cpu.prof 中前20个函数
```

#### 3. generate_svg

生成 SVG 火焰图。

**参数：**
- `filePath` (必需): pprof 文件路径
- `focus` (可选): 聚焦于特定函数或模式 (正则表达式)
- `ignore` (可选): 忽略匹配模式的函数 (正则表达式)

**示例：**
```
为 /path/to/cpu.prof 生成火焰图，聚焦于 handlers 相关函数
```

#### 4. analyze_performance

执行深度性能分析并提供优化建议。

**参数：**
- `filePath` (必需): pprof 文件路径
- `focus` (可选，默认: "all"): 分析重点 ("bottlenecks", "hotspots", "all")
- `threshold` (可选，默认: 5): 热点检测的百分比阈值

**示例：**
```
对 /path/to/cpu.prof 进行深度性能分析，阈值为10%
```

#### 5. compare_profiles

对比两个 profile 文件找出差异。

**参数：**
- `baseFile` (必需): 基准 profile 文件路径
- `compareFile` (必需): 对比 profile 文件路径

**示例：**
```
对比 before.prof 和 after.prof，查看性能改进
```

#### 6. list_callers

查看函数调用关系。

**参数：**
- `filePath` (必需): pprof 文件路径
- `functionName` (必需): 要查看调用链的函数名
- `maxDepth` (可选，默认: 10): 最大调用栈深度

**示例：**
```
显示调用 "MainHandler" 函数的所有调用者，来源文件为 /path/to/cpu.prof
```

### 远程模式 (mcp-remote)

使用 mcp-remote 进行远程访问，基于 HTTP 传输：

#### 1. 启动 HTTP 服务器

```bash
./mcp-pprof-server -port 8080
```

选项：
- `-port`: 监听端口 (默认: 8080)
- `-address`: 绑定地址 (默认: 0.0.0.0)
- `-debug`: 启用调试日志

#### 2. 配置客户端

```json
{
  "mcpServers": {
    "pprof": {
      "command": "npx",
      "args": [
        "-y",
        "mcp-remote@latest",
        "http://localhost:8080/mcp",
        "--allow-http",
        "--reconnect",
        "--reconnect-delay",
        "5000"
      ]
    }
  }
}
```

### 常见使用场景

#### 1. 发现性能瓶颈

```
分析 memory.prof heap profile，找出内存分配热点
```

#### 2. 前后对比

```
对比 cpu-before.prof 和 cpu-after.prof，测量性能改进效果
```

#### 3. 聚焦特定函数

```
为 cpu.prof 生成火焰图，只显示数据库相关函数
```

### 使用建议

1. **使用有意义的文件名**：在文件名中包含时间戳和类型（例如：`cpu-2025-01-22.prof`）
2. **聚焦热点路径**：使用 `focus` 参数放大查看特定区域
3. **对比前后变化**：始终对比 profile 以测量优化效果
4. **定期 profiling**：在高峰负载期间进行 profiling，发现生产环境问题
