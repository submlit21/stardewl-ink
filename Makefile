.PHONY: all build test clean run-signaling run-example deps cross-build-all cross-build-windows cross-build-macos cross-build-linux

all: build test

build:
	@echo "Building Stardewl-Ink for current platform..."
	@./scripts/build.sh

test:
	@echo "Running tests..."
	@cd core && go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -rf dist/

run-signaling: build
	@echo "Starting signaling server..."
	@./dist/stardewl-signaling

run-cli: build
	@echo "Starting CLI application..."
	@./dist/stardewl --interactive

run-example: build
	@echo "Starting example demo..."
	@./dist/stardewl-demo

deps:
	@echo "Downloading dependencies..."
	@go mod download

# 开发相关
dev-setup: deps
	@echo "Development setup complete"

# 代码质量检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 更新依赖
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# 生成文档
doc:
	@echo "Generating documentation..."
	@go doc -all ./core

# 性能测试
bench:
	@echo "Running benchmarks..."
	@cd core && go test -bench=. -benchmem

# 交叉编译
cross-build-all: cross-build-windows cross-build-macos cross-build-linux
	@echo "✅ All cross-platform builds completed"

cross-build-windows:
	@echo "🪟 Building for Windows (amd64)..."
	@mkdir -p dist/windows
	GOOS=windows GOARCH=amd64 go build -o dist/windows/stardewl.exe ./cmd/stardewl
	GOOS=windows GOARCH=amd64 go build -o dist/windows/stardewl-signaling.exe ./signaling
	GOOS=windows GOARCH=amd64 go build -o dist/windows/stardewl-demo.exe ./examples/simple_demo.go
	@echo "✅ Windows builds saved to dist/windows/"

cross-build-macos:
	@echo "🍎 Building for macOS (arm64)..."
	@mkdir -p dist/macos
	GOOS=darwin GOARCH=arm64 go build -o dist/macos/stardewl ./cmd/stardewl
	GOOS=darwin GOARCH=arm64 go build -o dist/macos/stardewl-signaling ./signaling
	GOOS=darwin GOARCH=arm64 go build -o dist/macos/stardewl-demo ./examples/simple_demo.go
	@echo "✅ macOS builds saved to dist/macos/"

cross-build-linux:
	@echo "🐧 Building for Linux (amd64)..."
	@mkdir -p dist/linux
	GOOS=linux GOARCH=amd64 go build -o dist/linux/stardewl ./cmd/stardewl
	GOOS=linux GOARCH=amd64 go build -o dist/linux/stardewl-signaling ./signaling
	GOOS=linux GOARCH=amd64 go build -o dist/linux/stardewl-demo ./examples/simple_demo.go
	@echo "✅ Linux builds saved to dist/linux/"

# 平台特定构建
build-windows: cross-build-windows
build-macos: cross-build-macos
build-linux: cross-build-linux

help:
	@echo "Available targets:"
	@echo "  build                    - Build for current platform"
	@echo "  test                     - Run tests"
	@echo "  clean                    - Clean build artifacts"
	@echo "  run-signaling            - Start signaling server"
	@echo "  run-cli                  - Run CLI application (interactive)"
	@echo "  run-example              - Run example demo"
	@echo "  deps                     - Download dependencies"
	@echo "  dev-setup                - Setup development environment"
	@echo "  lint                     - Run linter"
	@echo "  fmt                      - Format code"
	@echo "  update-deps              - Update dependencies"
	@echo "  doc                      - Generate documentation"
	@echo "  bench                    - Run benchmarks"
	@echo ""
	@echo "  Cross-compilation targets:"
	@echo "  cross-build-all          - Build for all platforms (Windows, macOS, Linux)"
	@echo "  cross-build-windows      - Build Windows executables (.exe)"
	@echo "  cross-build-macos        - Build macOS binaries"
	@echo "  cross-build-linux        - Build Linux binaries"
	@echo "  build-windows            - Alias for cross-build-windows"
	@echo "  build-macos              - Alias for cross-build-macos"
	@echo "  build-linux              - Alias for cross-build-linux"
	@echo ""
	@echo "  help                     - Show this help"