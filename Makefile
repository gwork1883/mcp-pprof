.PHONY: build install clean test run-stdio run-server help

# Binary names
STDIO_BIN = mcp-pprof
SERVER_BIN = mcp-pprof-server

# Go build flags
GO = go
GOFLAGS = -v
LDFLAGS = -ldflags "-s -w"

# Build directory
BUILD_DIR = build

help:
	@echo "Available targets:"
	@echo "  build          - Build the binaries"
	@echo "  install        - Install the binaries to GOPATH/bin"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  run-stdio      - Run stdio mode MCP server"
	@echo "  run-server     - Run HTTP mode MCP server (mcp-remote)"
	@echo "  tidy           - Run go mod tidy"

build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(STDIO_BIN) ./cmd/mcp-pprof
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(SERVER_BIN) ./cmd/mcp-pprof-server
	@echo "Build completed: $(BUILD_DIR)/$(STDIO_BIN), $(BUILD_DIR)/$(SERVER_BIN)"

install: build
	$(GO) install ./cmd/mcp-pprof
	$(GO) install ./cmd/mcp-pprof-server
	@echo "Installation completed"

clean:
	rm -rf $(BUILD_DIR)
	$(GO) clean
	@echo "Clean completed"

tidy:
	$(GO) mod tidy
	$(GO) mod verify
	@echo "Tidy completed"

test:
	$(GO) test -v ./...

run-stdio: build
	$(BUILD_DIR)/$(STDIO_BIN) -debug

run-server: build
	$(BUILD_DIR)/$(SERVER_BIN) -port 8080 -debug
