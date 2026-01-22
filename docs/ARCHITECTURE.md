# mcp-pprof 架构设计文档

## 项目概述

mcp-pprof 是一个基于 Go 语言实现的 MCP (Model Context Protocol) Server，用于提供 Go pprof 数据分析能力。该项目通过包装 `go tool pprof` 来复用现有功能，同时提供 MCP 工具和资源供 AI 使用。

## 设计目标

1. **易用性**：通过包装 `go tool pprof` 提供熟悉的命令行功能
2. **兼容性**：支持 MCP 协议标准，兼容所有支持 MCP 的 AI/Agent
3. **灵活性**：同时支持 stdio（本地）和 HTTP（mcp-remote）传输方式
4. **复用性**：最大化复用 `go tool pprof` 的现有能力，避免重复开发

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      mcp-pprof (Go)                             │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    MCP Server                            │  │
│  │  - 协议处理 (Request/Response)                           │  │
│  │  - Tool 管理 (ListTools/CallTool)                        │  │
│  │  - Resource 管理 (ListResources/ReadResource)            │  │
│  └───────────────────┬──────────────────────────────────────┘  │
│                      │                                         │
│  ┌───────────────────▼──────────────────────────────────────┐  │
│  │                   Tool Handlers                          │  │
│  │  - parse_profile: 解析 pprof 文件                         │  │
│  │  - top_functions: Top N 热点函数                           │  │
│  │  - list_callers: 查看调用关系                             │  │
│  │  - generate_svg: 生成 SVG 火焰图                          │  │
│  │  - analyze_performance: 性能分析                          │  │
│  │  - compare_profiles: 对比分析                             │  │
│  └───────────────────┬──────────────────────────────────────┘  │
│                      │                                         │
│  ┌───────────────────▼──────────────────────────────────────┐  │
│  │               go tool pprof Wrapper                       │  │
│  │  - 执行 go tool pprof 命令                                │  │
│  │  - 解析命令输出                                           │  │
│  │  - 格式化返回数据                                         │  │
│  └───────────────────┬──────────────────────────────────────┘  │
└──────────────────────┼─────────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┬────────────────┐
        │              │              │                │
   ┌────▼────┐    ┌─────▼─────┐  ┌────▼────┐    ┌──────▼──────┐
   │ Stdio   │    │   SSE/    │  │  HTTP   │    │  go tool    │
   │ Server  │    │   HTTP    │  │  Server │    │   pprof     │
   │ (stdio) │    │  Bridge   │  │ (Web)   │    │   (CLI)     │
   └─────────┘    └───────────┘  └─────────┘    └─────────────┘
```


## MCP Tools 定义

### 1. parse_profile
解析 pprof 文件并返回结构化摘要数据。

```json
{
  "name": "parse_profile",
  "description": "解析 pprof 文件并返回结构化摘要",
  "inputSchema": {
    "type": "object",
    "properties": {
      "filePath": {
        "type": "string",
        "description": "pprof 文件路径"
      },
      "profileType": {
        "type": "string",
        "enum": ["auto", "cpu", "heap", "block", "mutex", "goroutine"],
        "default": "auto"
      },
      "outputFormat": {
        "type": "string",
        "enum": ["json", "text", "proto"],
        "default": "json"
      }
    },
    "required": ["filePath"]
  }
}
```

### 2. top_functions
获取 Top N 热点函数。

```json
{
  "name": "top_functions",
  "description": "获取 Top N 热点函数",
  "inputSchema": {
    "type": "object",
    "properties": {
      "filePath": {
        "type": "string"
      },
      "topN": {
        "type": "number",
        "default": 10,
        "minimum": 1,
        "maximum": 100
      }
    },
    "required": ["filePath"]
  }
}
```

### 3. generate_svg
生成 SVG 火焰图。

```json
{
  "name": "generate_svg",
  "description": "生成 SVG 火焰图",
  "inputSchema": {
    "type": "object",
    "properties": {
      "filePath": {
        "type": "string"
      },
      "focus": {
        "type": "string",
        "default": ""
      },
      "ignore": {
        "type": "string",
        "default": ""
      }
    },
    "required": ["filePath"]
  }
}
```

### 4. analyze_performance
深度性能分析。

```json
{
  "name": "analyze_performance",
  "description": "深度性能分析",
  "inputSchema": {
    "type": "object",
    "properties": {
      "filePath": {
        "type": "string"
      },
      "focus": {
        "type": "string",
        "enum": ["bottlenecks", "hotspots", "memory", "all"],
        "default": "all"
      },
      "threshold": {
        "type": "number",
        "default": 5,
        "description": "百分比阈值"
      }
    },
    "required": ["filePath"]
  }
}
```

### 5. compare_profiles
对比两个 pprof 文件。

```json
{
  "name": "compare_profiles",
  "description": "对比两个 pprof 文件",
  "inputSchema": {
    "type": "object",
    "properties": {
      "baseFile": {
        "type": "string"
      },
      "compareFile": {
        "type": "string"
      }
    },
    "required": ["baseFile", "compareFile"]
  }
}
```

### 6. list_callers
查看函数的调用关系。

```json
{
  "name": "list_callers",
  "description": "查看函数的调用关系",
  "inputSchema": {
    "type": "object",
    "properties": {
      "filePath": {
        "type": "string"
      },
      "functionName": {
        "type": "string"
      },
      "maxDepth": {
        "type": "number",
        "default": 10
      }
    },
    "required": ["filePath", "functionName"]
  }
}
```

## MCP Resources 定义

| URI Pattern | 描述 |
|-------------|------|
| `pprof://file/{filePath}` | 原始 pprof 数据 |
| `pprof://summary/{filePath}` | 摘要统计 |
| `pprof://text/{filePath}` | 文本格式输出 |
| `pprof://svg/{filePath}` | SVG 火焰图数据 |

## Transport 设计

### Transport 接口

```go
type Transport interface {
    Connect() error
    ReadMessage() (*mcp.JSONRPCMessage, error)
    WriteMessage(*mcp.JSONRPCMessage) error
    Close() error
}
```

### Stdio Transport
- 使用标准输入/输出进行通信
- 适用于本地使用场景
- 与 MCP Host 直接通信

### SSE Transport (mcp-remote)
- 通过 Server-Sent Events (SSE) 进行通信
- 支持 HTTP 连接
- 适用于远程访问场景

## go tool pprof 集成

### 执行模式

1. **命令执行模式**
   ```go
   cmd := exec.Command("go", "tool", "pprof", "-text", filePath)
   output, err := cmd.CombinedOutput()
   ```

2. **库调用模式**（优先）
   ```go
   // 使用 runtime/pprof 包
   profile, err := profile.Parse(filePath)
   ```

### 支持的命令映射

| Tool | go pprof 命令 |
|------|---------------|
| parse_profile | go tool pprof -text |
| top_functions | go tool pprof -top |
| generate_svg | go tool pprof -svg |
| list_callers | go tool pprof -list |

## 数据流程

```
AI Agent
   │
   │ 1. Request (JSON-RPC)
   ▼
MCP Server
   │
   │ 2. Route to Handler
   ▼
Tool Handler
   │
   │ 3. Parse arguments
   ▼
Pprof Wrapper
   │
   │ 4. Execute go tool pprof
   ▼
Parsing & Formatting
   │
   │ 5. Format output
   ▼
MCP Response
   │
   │ 6. Response (JSON-RPC)
   ▼
AI Agent
```

## 部署模式

### 1. 本地部署 (stdio)
```json
{
  "mcpServers": {
    "pprof": {
      "command": "mcp-pprof",
      "args": []
    }
  }
}
```

### 2. 远程部署 (mcp-remote)
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
