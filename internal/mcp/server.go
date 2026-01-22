package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gwork1883/mcp-pprof/internal/pprof"
	"github.com/gwork1883/mcp-pprof/pkg/protocol"
)

// ToolHandler is a function that handles a tool call
type ToolHandler func(ctx context.Context, args map[string]any) (*protocol.ToolCallResult, error)

// Server represents the MCP server
type Server struct {
	serverInfo     protocol.ImplementationInfo
	tools          map[string]protocol.Tool
	toolHandlers   map[string]ToolHandler
	resources      map[string]protocol.Resource
	pprofWrapper   *pprof.Wrapper
	initialized    bool
	mu             sync.RWMutex
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	s := &Server{
		serverInfo: protocol.ImplementationInfo{
			Name:    name,
			Version: version,
		},
		tools:        make(map[string]protocol.Tool),
		toolHandlers: make(map[string]ToolHandler),
		resources:    make(map[string]protocol.Resource),
		pprofWrapper: pprof.NewWrapper(),
	}
	
	// Register default tools
	s.registerDefaultTools()
	s.registerDefaultResources()
	
	return s
}

// registerDefaultTools registers the default pprof tools
func (s *Server) registerDefaultTools() {
	// parse_profile tool
	s.RegisterTool(protocol.Tool{
		Name:        "parse_profile",
		Description: "Parse a pprof file and return structured summary",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "Path to the pprof file",
				},
				"profileType": map[string]any{
					"type":        "string",
					"default":     "auto",
					"enum":        []string{"auto", "cpu", "heap", "block", "mutex", "goroutine"},
					"description": "Type of profile",
				},
				"outputFormat": map[string]any{
					"type":        "string",
					"default":     "json",
					"enum":        []string{"json", "text", "proto"},
					"description": "Output format",
				},
			},
			"required": []string{"filePath"},
		},
	}, s.handleParseProfile)

	// top_functions tool
	s.RegisterTool(protocol.Tool{
		Name:        "top_functions",
		Description: "Get top N hot functions",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "Path to the pprof file",
				},
				"topN": map[string]any{
					"type":        "number",
					"default":     10,
					"minimum":     1,
					"maximum":     100,
					"description": "Number of top functions to return",
				},
			},
			"required": []string{"filePath"},
		},
	}, s.handleTopFunctions)

	// generate_svg tool
	s.RegisterTool(protocol.Tool{
		Name:        "generate_svg",
		Description: "Generate SVG flamegraph",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "Path to the pprof file",
				},
				"focus": map[string]any{
					"type":        "string",
					"default":     "",
					"description": "Focus on a specific function or pattern",
				},
				"ignore": map[string]any{
					"type":        "string",
					"default":     "",
					"description": "Ignore functions matching pattern",
				},
			},
			"required": []string{"filePath"},
		},
	}, s.handleGenerateSVG)

	// analyze_performance tool
	s.RegisterTool(protocol.Tool{
		Name:        "analyze_performance",
		Description: "Deep performance analysis",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "Path to the pprof file",
				},
				"focus": map[string]any{
					"type":        "string",
					"default":     "all",
					"enum":        []string{"bottlenecks", "hotspots", "all"},
					"description": "Analysis focus area",
				},
				"threshold": map[string]any{
					"type":        "number",
					"default":     5,
					"description": "Percentage threshold",
				},
			},
			"required": []string{"filePath"},
		},
	}, s.handleAnalyzePerformance)

	// compare_profiles tool
	s.RegisterTool(protocol.Tool{
		Name:        "compare_profiles",
		Description: "Compare two pprof files",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"baseFile": map[string]any{
					"type":        "string",
					"description": "Base profile file path",
				},
				"compareFile": map[string]any{
					"type":        "string",
					"description": "Comparison profile file path",
				},
			},
			"required": []string{"baseFile", "compareFile"},
		},
	}, s.handleCompareProfiles)

	// list_callers tool
	s.RegisterTool(protocol.Tool{
		Name:        "list_callers",
		Description: "List callers of a function",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"filePath": map[string]any{
					"type":        "string",
					"description": "Path to the pprof file",
				},
				"functionName": map[string]any{
					"type":        "string",
					"description": "Function name to list callers for",
				},
				"maxDepth": map[string]any{
					"type":        "number",
					"default":     10,
					"description": "Maximum depth",
				},
			},
			"required": []string{"filePath", "functionName"},
		},
	}, s.handleListCallers)
}

// registerDefaultResources registers default resources
func (s *Server) registerDefaultResources() {
	// Template for pprof资源
	s.resources["pprof://summary/{filePath}"] = protocol.Resource{
		URI:         "pprof://summary/{filePath}",
		Name:        "Profile Summary",
		Description: "Get summary of a pprof file",
		MimeType:    "application/json",
	}

	s.resources["pprof://text/{filePath}"] = protocol.Resource{
		URI:         "pprof://text/{filePath}",
		Name:        "Profile Text Output",
		Description: "Get text format output from pprof",
		MimeType:    "text/plain",
	}

	s.resources["pprof://svg/{filePath}"] = protocol.Resource{
		URI:         "pprof://svg/{filePath}",
		Name:        "Profile SVG",
		Description: "Get SVG flamegraph from pprof",
		MimeType:    "image/svg+xml",
	}
}

// RegisterTool registers a new tool
func (s *Server) RegisterTool(tool protocol.Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.toolHandlers[tool.Name] = handler
	log.Printf("[MCP] Registered tool: %s", tool.Name)
}

// HandleRequest handles an incoming MCP request
func (s *Server) HandleRequest(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, req)
	case "initialized":
		return s.handleInitialized(ctx, req)
	case "tools/list":
		return s.handleListTools(ctx, req)
	case "tools/call":
		return s.handleCallTool(ctx, req)
	case "resources/list":
		return s.handleListResources(ctx, req)
	case "resources/read":
		return s.handleReadResource(ctx, req)
	case "shutdown":
		return s.handleShutdown(ctx, req)
	default:
		return s.errorResponse(req.ID, protocol.MethodNotFound, fmt.Sprintf("unknown method: %s", req.Method)), nil
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	var params protocol.InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.errorResponse(req.ID, protocol.InvalidParams, "invalid params"), nil
	}

	log.Printf("[MCP] Initialize request from %s %s", params.ClientInfo.Name, params.ClientInfo.Version)

	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()

	result := protocol.InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: protocol.ServerCapabilities{
			Tools: struct {
				ListChanged bool `json:"listChanged,omitempty"`
			}{
				ListChanged: false,
			},
			Resources: struct {
				Subscribe   bool `json:"subscribe,omitempty"`
				ListChanged bool `json:"listChanged,omitempty"`
			}{
				ListChanged: false,
			},
		},
		ServerInfo: s.serverInfo,
	}

	return s.successResponse(req.ID, result), nil
}

// handleInitialized handles the initialized notification
func (s *Server) handleInitialized(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	log.Printf("[MCP] Initialized")
	return &protocol.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}, nil
}

// handleListTools handles the tools/list request
func (s *Server) handleListTools(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]protocol.Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	return s.successResponse(req.ID, protocol.ListToolsResult{
		Tools: tools,
	}), nil
}

// handleCallTool handles the tools/call request
func (s *Server) handleCallTool(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.errorResponse(req.ID, protocol.InvalidParams, "invalid params"), nil
	}

	handler, exists := s.toolHandlers[params.Name]
	if !exists {
		return s.errorResponse(req.ID, protocol.MethodNotFound, fmt.Sprintf("tool not found: %s", params.Name)), nil
	}

	result, err := handler(ctx, params.Arguments)
	if err != nil {
		return &protocol.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: &protocol.ToolCallResult{
				Content: []protocol.ContentBlock{
					{
						Type: "text",
						Text: fmt.Sprintf("Error: %v", err),
					},
				},
				IsError: true,
			},
		}, nil
	}

	return s.successResponse(req.ID, result), nil
}

// handleListResources handles the resources/list request
func (s *Server) handleListResources(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]protocol.Resource, 0, len(s.resources))
	for _, res := range s.resources {
		resources = append(resources, res)
	}

	return s.successResponse(req.ID, protocol.ListResourcesResult{
		Resources: resources,
	}), nil
}

// handleReadResource handles the resources/read request
func (s *Server) handleReadResource(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	// TODO: Implement resource reading
	return s.errorResponse(req.ID, protocol.InternalError, "not implemented"), nil
}

// handleShutdown handles the shutdown request
func (s *Server) handleShutdown(ctx context.Context, req *protocol.JSONRPCRequest) (*protocol.JSONRPCResponse, error) {
	log.Printf("[MCP] Shutdown")
	return s.successResponse(req.ID, nil), nil
}

// successResponse creates a success response
func (s *Server) successResponse(id any, result any) *protocol.JSONRPCResponse {
	return &protocol.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// errorResponse creates an error response
func (s *Server) errorResponse(id any, code protocol.ErrorCode, message string) *protocol.JSONRPCResponse {
	return &protocol.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &protocol.JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}
