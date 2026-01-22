# mcp-pprof

[English](#english) | [中文](#中文)

---

<div id="english"></div>

## English

A Model Context Protocol (MCP) Server for Go pprof data analysis. This server wraps `go tool pprof` to provide AI-accessible tools for performance profiling and analysis.

### Overview

mcp-pprof enables AI agents to:
- Parse and analyze Go pprof files (CPU, heap, block, mutex, goroutine profiles)
- Generate SVG flamegraphs
- Identify performance bottlenecks and hotspots
- Compare different profile files
- Provide optimization suggestions

### Features

- **MCP Protocol Support**: Fully compatible with MCP protocol for seamless AI integration
- **Multiple Transport Modes**: Supports both stdio (local) and HTTP (mcp-remote) modes
- **Comprehensive Tools**:
  - `parse_profile` - Parse pprof files and return structured data
  - `top_functions` - Get top N hot functions
  - `generate_svg` - Generate SVG flamegraphs
  - `analyze_performance` - Deep performance analysis with suggestions
  - `compare_profiles` - Compare two profile files
  - `list_callers` - View function call relationships

### Installation

#### Build from source

```bash
# Clone the repository
git clone https://github.com/gwork1883/mcp-pprof.git
cd mcp-pprof

# Build binaries
go build -o build/mcp-pprof ./cmd/mcp-pprof
go build -o build/mcp-pprof-server ./cmd/mcp-pprof-server

# Or use Makefile
make build
```

#### Requirements

- Go 1.21 or higher
- `go tool pprof` (included with Go)

### Usage

#### Local Mode (stdio)

Add to your MCP configuration:

```json
{
  "mcpServers": {
    "pprof": {
      "type": "stdio",
      "command": "/path/to/mcp-pprof",
      "args": [],
      "disabled": false,
      "autoApprove": [
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

#### Remote Mode (mcp-remote)

##### 1. Start HTTP Server

```bash
./build/mcp-pprof-server -port 8080
```

##### 2. Configure mcp-remote

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
      ],
      "disabled": false,
      "autoApprove": [
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

### Tools

| Tool | Description |
|------|-------------|
| `parse_profile` | Parse a pprof file and return structured summary |
| `top_functions` | Get top N hot functions |
| `generate_svg` | Generate SVG flamegraph |
| `analyze_performance` | Deep performance analysis |
| `compare_profiles` | Compare two profile files |
| `list_callers` | View function call relationships |

### Example Usage with AI

```
Please analyze the CPU profile at path/to/profile.prof
```

```
Generate a flamegraph for the heap profile
```

```
Compare the before.prof and after.prof profiles
```

### Development

```bash
# Build
make build

# Run stdio mode
make run-stdio

# Run HTTP server
make run-server

# Test
make test
```

### Architecture

```
mcp-pprof/
├── cmd/
│   ├── mcp-pprof/           # stdio mode entry
│   └── mcp-pprof-server/    # HTTP server entry
├── internal/
│   ├── mcp/                 # MCP protocol implementation
│   ├── pprof/               # go tool pprof wrapper
│   └── tools/               # Tool handlers
├── pkg/
│   └── protocol/            # MCP protocol types
└── docs/                    # Documentation
```

### License

MIT License

---

<div id="中文"></div>

## 中文

用于 Go pprof 数据分析的 Model Context Protocol (MCP) Server。该服务器封装了 `go tool pprof`，为 AI 提供性能分析工具。

### 概述

mcp-pprof 使 AI 智能体能够：
- 解析和分析 Go pprof 文件（CPU、heap、block、mutex、goroutine profiles）
- 生成 SVG 火焰图
- 识别性能瓶颈和热点
- 对比不同的 profile 文件
- 提供优化建议

### 功能特性

- **MCP 协议支持**：完全兼容 MCP 协议，无缝集成 AI
- **多种传输模式**：支持 stdio（本地）和 HTTP（mcp-remote）两种模式
- **丰富的工具集**：
  - `parse_profile` - 解析 pprof 文件并返回结构化数据
  - `top_functions` - 获取 Top N 热点函数
  - `generate_svg` - 生成 SVG 火焰图
  - `analyze_performance` - 深度性能分析与优化建议
  - `compare_profiles` - 对比两个 profile 文件
  - `list_callers` - 查看函数调用关系

### 安装

#### 从源代码构建

```bash
# 克隆仓库
git clone https://github.com/gwork1883/mcp-pprof.git
cd mcp-pprof

# 构建二进制文件
go build -o build/mcp-pprof ./cmd/mcp-pprof
go build -o build/mcp-pprof-server ./cmd/mcp-pprof-server

# 或使用 Makefile
make build
```

#### 系统要求

- Go 1.21 或更高版本
- `go tool pprof`（Go 自带）

### 使用方式

#### 本地模式 (stdio)

将以下配置添加到您的 MCP 配置中：

```json
{
  "mcpServers": {
    "pprof": {
      "type": "stdio",
      "command": "/path/to/mcp-pprof",
      "args": [],
      "disabled": false,
      "autoApprove": [
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

#### 远程模式 (mcp-remote)

##### 1. 启动 HTTP 服务器

```bash
./build/mcp-pprof-server -port 8080
```

##### 2. 配置 mcp-remote

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
      ],
      "disabled": false,
      "autoApprove": [
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

### 工具

| 工具 | 描述 |
|------|------|
| `parse_profile` | 解析 pprof 文件并返回结构化摘要 |
| `top_functions` | 获取 Top N 热点函数 |
| `generate_svg` | 生成 SVG 火焰图 |
| `analyze_performance` | 深度性能分析 |
| `compare_profiles` | 对比两个 profile 文件 |
| `list_callers` | 查看函数调用关系 |

### AI 使用示例

```
请分析这个 CPU profile 文件，告诉我哪些函数占用最多时间
```

```
为这个 heap profile 生成火焰图
```

```
对比优化前后的两个 profile 文件，找出性能改进点
```

### 开发

```bash
# 构建
make build

# 运行 stdio 模式
make run-stdio

# 运行 HTTP 服务器
make run-server

# 运行测试
make test
```

### 项目架构

```
mcp-pprof/
├── cmd/
│   ├── mcp-pprof/           # stdio 模式入口
│   └── mcp-pprof-server/    # HTTP 服务器入口
├── internal/
│   ├── mcp/                 # MCP 协议实现
│   ├── pprof/               # go tool pprof 包装器
│   └── tools/               # 工具处理器
├── pkg/
│   └── protocol/            # MCP 协议类型定义
└── docs/                    # 文档
```

### 许可证

MIT License
