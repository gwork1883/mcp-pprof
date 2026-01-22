package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gwork1883/mcp-pprof/internal/mcp"
)

var (
	port    = flag.String("port", "8080", "Port to listen on")
	debug   = flag.Bool("debug", false, "Enable debug logging")
	address = flag.String("address", "0.0.0.0", "Address to bind to")
)

func main() {
	flag.Parse()
	
	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("[MCP] Debug mode enabled")
	}
	
	// Create MCP server
	server := mcp.NewServer("mcp-pprof", "0.1.0")
	
	// Create HTTP transport
	addr := *address + ":" + *port
	transport := mcp.NewHTTPTransport(addr)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("[MCP] Received shutdown signal")
		cancel()
	}()
	
	// Run the server
	log.Printf("[MCP] Starting mcp-pprof HTTP server on %s", addr)
	if err := transport.Run(ctx, server); err != nil && err != context.Canceled {
		log.Printf("[MCP] Error running server: %v", err)
		os.Exit(1)
	}
	
	log.Printf("[MCP] Server stopped")
}
