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
	debug = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	flag.Parse()
	
	if *debug {
		// Enable stderr logging for debug mode
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("[MCP] Debug mode enabled")
	}
	
	// Create MCP server
	server := mcp.NewServer("mcp-pprof", "0.1.0")
	
	// Create stdio transport
	transport := mcp.NewStdioTransport(os.Stdin, os.Stdout)
	
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
	if err := transport.Run(ctx, server); err != nil {
		log.Printf("[MCP] Error running server: %v", err)
		os.Exit(1)
	}
	
	log.Printf("[MCP] Server stopped")
}
