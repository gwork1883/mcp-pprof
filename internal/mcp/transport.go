package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gwork1883/mcp-pprof/pkg/protocol"
)

// Transport defines the interface for MCP transport
type Transport interface {
	Connect(context.Context) error
	Run(context.Context, *Server) error
	Close() error
}

// StdioTransport implements stdio-based MCP transport
type StdioTransport struct {
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(reader io.Reader, writer io.Writer) *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewReader(reader),
		writer: writer,
	}
}

// Connect initializes the stdio transport
func (t *StdioTransport) Connect(ctx context.Context) error {
	log.SetOutput(io.Discard) // Disable logging to stdout
	log.SetPrefix("[MCP] ")
	return nil
}

// Run starts processing requests
func (t *StdioTransport) Run(ctx context.Context, server *Server) error {
	decoder := json.NewDecoder(t.reader)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		var req protocol.JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			log.Printf("Error decoding request: %v", err)
			continue
		}
		
		resp, err := server.HandleRequest(ctx, &req)
		if err != nil {
			log.Printf("Error handling request: %v", err)
			continue
		}
		
		if resp != nil && req.ID != nil {
			t.mu.Lock()
			if err := json.NewEncoder(t.writer).Encode(resp); err != nil {
				t.mu.Unlock()
				return fmt.Errorf("error encoding response: %w", err)
			}
			t.mu.Unlock()
		}
	}
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	return nil
}

// HTTPTransport implements HTTP-based MCP transport for mcp-remote
type HTTPTransport struct {
	addr   string
	server *http.Server
	mu     sync.Mutex
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(addr string) *HTTPTransport {
	return &HTTPTransport{
		addr: addr,
	}
}

// Connect initializes the HTTP transport
func (t *HTTPTransport) Connect(ctx context.Context) error {
	return nil
}

// Run starts the HTTP server
func (t *HTTPTransport) Run(ctx context.Context, server *Server) error {
	mux := http.NewServeMux()
	
	// MCP endpoint for mcp-remote
	mux.HandleFunc("/mcp", t.handleMCPRequest(server))
	
	// Health check endpoint
	mux.HandleFunc("/health", t.handleHealth)
	
	t.server = &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}
	
	log.Printf("[MCP] HTTP server listening on %s", t.addr)
	
	errChan := make(chan error, 1)
	go func() {
		errChan <- t.server.ListenAndServe()
	}()
	
	select {
	case <-ctx.Done():
		log.Printf("[MCP] Shutting down HTTP server...")
		if err := t.server.Shutdown(ctx); err != nil {
			log.Printf("[MCP] Error shutting down server: %v", err)
		}
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// handleMCPRequest handles incoming MCP requests via HTTP
func (t *HTTPTransport) handleMCPRequest(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// For mcp-remote, we need to support both JSON RPC requests
		// and SSE for bidirectional communication
		contentType := r.Header.Get("Content-Type")
		
		if contentType == "application/json" {
			t.handleJSONRequest(w, r, server)
		} else {
			http.Error(w, "Unsupported content type", http.StatusBadRequest)
		}
	}
}

// handleJSONRequest handles JSON-RPC requests
func (t *HTTPTransport) handleJSONRequest(w http.ResponseWriter, r *http.Request, server *Server) {
	var req protocol.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	resp, err := server.HandleRequest(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error handling request: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}

// handleHealth handles health check requests
func (t *HTTPTransport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"server": "mcp-pprof",
	})
}

// Close closes the transport
func (t *HTTPTransport) Close() error {
	if t.server != nil {
		return t.server.Close()
	}
	return nil
}
