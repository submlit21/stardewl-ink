.PHONY: all build test clean run-signaling run-example deps

all: build test

build:
	@echo "Building Stardewl-Ink..."
	@./scripts/build.sh

test:
	@echo "Running tests..."
	@cd core && go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -rf dist/
	@rm -f core/*.test
	@rm -f signaling/stardewl-signaling

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

help:
	@echo "Available targets:"
	@echo "  build        - Build the project"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  run-signaling - Start signaling server"
	@echo "  run-cli      - Run CLI application (interactive)"
	@echo "  run-example  - Run example demo"
	@echo "  deps         - Download dependencies"
	@echo "  dev-setup    - Setup development environment"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  update-deps  - Update dependencies"
	@echo "  doc          - Generate documentation"
	@echo "  bench        - Run benchmarks"
	@echo "  help         - Show this help"