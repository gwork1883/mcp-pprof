package protocol

import "encoding/json"

// MCP Protocol Types

// ErrorCode represents JSON-RPC error codes
type ErrorCode int

const (
	// ParseInvalidRequest - Invalid JSON was received
	ParseInvalidRequest ErrorCode = -32700
	// InvalidRequest - The JSON sent is not a valid Request object
	InvalidRequest ErrorCode = -32600
	// MethodNotFound - The method does not exist
	MethodNotFound ErrorCode = -32601
	// InvalidParams - Invalid method parameter(s)
	InvalidParams ErrorCode = -32602
	// InternalError - Internal JSON-RPC error
	InternalError ErrorCode = -32603
)

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCNotification represents a JSON-RPC notification
type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// InitializeParams represents initialization parameters
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ImplementationInfo      `json:"clientInfo"`
	Metadata       map[string]any          `json:"metadata,omitempty"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Roots   RootsCapability   `json:"roots,omitempty"`
	Sampling SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability represents roots capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability represents sampling capability
type SamplingCapability map[string]any

// ImplementationInfo represents implementation info
type ImplementationInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents initialization result
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ServerCapabilities     `json:"capabilities"`
	ServerInfo      ImplementationInfo      `json:"serverInfo"`
	Metadata       map[string]any          `json:"metadata,omitempty"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Resources struct {
		Subscribe   bool `json:"subscribe,omitempty"`
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"resources,omitempty"`
	Tools struct {
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"tools,omitempty"`
	Logging struct{} `json:"logging,omitempty"`
	Prompts struct {
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"prompts,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]any         `json:"inputSchema"`
}

// ToolCallRequest represents a tool call request
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]any         `json:"arguments"`
}

// ToolCallResult represents a tool call result
type ToolCallResult struct {
	Content  []ContentBlock          `json:"content"`
	IsError  bool                    `json:"isError,omitempty"`
	Metadata map[string]any          `json:"metadata,omitempty"`
}

// ContentBlock represents a content block
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	MimeType    string            `json:"mimeType,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// ResourceTemplate represents a resource template
type ResourceTemplate struct {
	URITemplate string            `json:"uriTemplate"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	MimeType    string            `json:"mimeType,omitempty"`
}

// ResourceContent represents resource content
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

// ListToolsResult represents list tools result
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// ListResourcesResult represents list resources result
type ListResourcesResult struct {
	Resources         []Resource         `json:"resources"`
	ResourceTemplates []ResourceTemplate `json:"resourceTemplates,omitempty"`
}

// ReadResourceResult represents read resource result
type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}
