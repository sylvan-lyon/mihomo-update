# mihomo-update Go项目的Makefile
.PHONY: all build test clean fmt lint mod-tidy help

# Go参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOLINT=$(GOCMD) vet

# 二进制文件名
BINARY_NAME=mihomo-update

# 构建目录
BUILD_DIR=bin

# 默认目标
all: test build

# 构建二进制文件
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mihomo-update

# 为多平台构建
build-all: build-linux build-windows build-darwin

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mihomo-update

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/mihomo-update

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/mihomo-update

# 运行测试
test:
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 清理构建产物
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 格式化代码
fmt:
	$(GOFMT) ./...

# 运行go vet（代码检查器）
lint:
	$(GOLINT) ./...

# 下载并整理依赖
mod-tidy:
	$(GOMOD) tidy

# 安装依赖
deps: mod-tidy

# 运行所有检查（格式化、代码检查、测试）
check: fmt lint test

# 显示帮助
help:
	@echo "可用目标:"
	@echo "  all        - 运行测试并构建（默认）"
	@echo "  build      - 为当前平台构建二进制文件"
	@echo "  build-all  - 为Linux、Windows和macOS构建"
	@echo "  test       - 运行测试"
	@echo "  test-coverage - 运行测试并生成覆盖率报告"
	@echo "  clean      - 清理构建产物"
	@echo "  fmt        - 格式化代码"
	@echo "  lint       - 运行go vet"
	@echo "  mod-tidy   - 下载并整理依赖"
	@echo "  deps       - 安装依赖"
	@echo "  check      - 运行fmt、lint和test"
	@echo "  help       - 显示此帮助"

# 展示的最佳实践:
# 1. 使用伪目标防止与文件冲突
# 2. 变量定义便于配置
# 3. 跨平台构建目标
# 4. 测试覆盖率报告
# 5. 依赖管理
# 6. 清洁构建过程