# LicenseCenter Makefile

# 变量定义
APP_NAME=license-server
TEST_CLIENT=test-client
VERSION=1.0.0
BUILD_DIR=bin
CONFIG_DIR=configs
DATA_DIR=data
LOGS_DIR=logs

# Go相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date +%Y-%m-%d_%H:%M:%S)"

.PHONY: all build clean test deps run dev help

# 默认目标
all: deps build

# 显示帮助信息
help:
	@echo "LicenseCenter 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build      - 构建所有程序"
	@echo "  server     - 构建服务端程序"
	@echo "  client     - 构建测试客户端"
	@echo "  init-tool  - 构建初始化工具"
	@echo "  clean      - 清理构建文件"
	@echo "  test       - 运行测试"
	@echo "  deps       - 安装依赖"
	@echo "  run        - 运行服务端"
	@echo "  dev        - 开发模式运行"
	@echo "  setup      - 初始化项目环境"
	@echo "  init-system - 初始化系统数据"
	@echo "  reset-db   - 重置数据库"
	@echo "  machine-id - 机器ID调试工具"
	@echo "  machine-id-debug - 机器ID详细调试"
	@echo "  network-debug - 网络接口调试"
	@echo ""

# 安装依赖
deps:
	@echo "📦 安装Go依赖..."
	$(GOMOD) tidy
	$(GOMOD) download

# 构建所有程序
build: server client init-tool
	@echo "✅ 构建完成"

# 构建服务端
server:
	@echo "🔨 构建服务端程序..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) cmd/server/main.go

# 构建测试客户端
client:
	@echo "🔨 构建测试客户端..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(TEST_CLIENT) test_client/main.go

# 构建初始化工具
init-tool:
	@echo "🔨 构建初始化工具..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/init cmd/init/main.go

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f *.bind *.license *.unbind

# 运行测试
test:
	@echo "🧪 运行测试..."
	$(GOTEST) -v ./...

# 运行测试（详细输出）
test-verbose:
	@echo "🧪 运行测试（详细模式）..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# 运行服务端（生产模式）
run: server
	@echo "🚀 启动服务端..."
	./$(BUILD_DIR)/$(APP_NAME) $(CONFIG_DIR)/app.yaml

# 开发模式运行
dev:
	@echo "🔧 开发模式启动..."
	$(GOCMD) run cmd/server/main.go $(CONFIG_DIR)/app.yaml

# 运行测试客户端
run-client: client
	@echo "🖥️  运行测试客户端..."
	./$(BUILD_DIR)/$(TEST_CLIENT) show-machine

# 生成绑定文件
generate-bind: client
	@echo "📄 生成绑定文件..."
	./$(BUILD_DIR)/$(TEST_CLIENT) generate-bind

# 机器ID调试工具
machine-id:
	@echo "🔍 机器ID调试..."
	$(GOCMD) run cmd/machine-id-debug/main.go

# 机器ID详细调试
machine-id-debug:
	@echo "🔍 机器ID详细调试..."
	$(GOTEST) ./pkg/utils/ -v -run TestGetMachineIDDebug

# 网络接口调试
network-debug:
	@echo "🌐 网络接口调试..."
	$(GOTEST) ./pkg/utils/ -v -run TestNetworkInterfaces

# 初始化项目环境
setup:
	@echo "🔧 初始化项目环境..."
	@mkdir -p $(DATA_DIR) $(LOGS_DIR) uploads
	@if [ ! -f $(CONFIG_DIR)/app.local.yaml ]; then \
		cp $(CONFIG_DIR)/app.yaml $(CONFIG_DIR)/app.local.yaml; \
		echo "✅ 创建本地配置文件: $(CONFIG_DIR)/app.local.yaml"; \
	fi
	@echo "✅ 环境初始化完成"

# 初始化系统数据
init-system: init-tool
	@echo "🔧 初始化系统数据..."
	./$(BUILD_DIR)/init

# 重置数据库
reset-db:
	@echo "🗃️  重置数据库..."
	@read -p "确定要重置数据库吗？这将删除所有数据！(y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -f $(DATA_DIR)/license.db; \
		echo "✅ 数据库已重置"; \
	else \
		echo "❌ 操作已取消"; \
	fi

# 查看日志
logs:
	@echo "📋 查看日志..."
	@if [ -f $(LOGS_DIR)/app.log ]; then \
		tail -f $(LOGS_DIR)/app.log; \
	else \
		echo "❌ 日志文件不存在"; \
	fi

# 格式化代码
fmt:
	@echo "💅 格式化代码..."
	$(GOCMD) fmt ./...

# 代码检查
lint:
	@echo "🔍 代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过检查"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 安装开发工具
install-tools:
	@echo "🔧 安装开发工具..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/air-verse/air@latest

# 热重载开发
watch:
	@echo "🔥 启动热重载开发模式..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "❌ air 未安装，请先运行: make install-tools"; \
	fi

# 显示项目状态
status:
	@echo "📊 项目状态:"
	@echo "版本: $(VERSION)"
	@echo "构建目录: $(BUILD_DIR)"
	@echo "配置目录: $(CONFIG_DIR)"
	@echo "数据目录: $(DATA_DIR)"
	@echo ""
	@echo "构建文件:"
	@ls -la $(BUILD_DIR)/ 2>/dev/null || echo "  无构建文件"
	@echo ""
	@echo "数据文件:"
	@ls -la $(DATA_DIR)/ 2>/dev/null || echo "  无数据文件"

# 打包发布
package: build
	@echo "📦 打包发布版本..."
	@mkdir -p release/$(VERSION)
	@cp -r $(BUILD_DIR) release/$(VERSION)/
	@cp -r $(CONFIG_DIR) release/$(VERSION)/
	@cp README.md release/$(VERSION)/
	@cd release && tar -czf $(APP_NAME)-$(VERSION).tar.gz $(VERSION)
	@echo "✅ 发布包已创建: release/$(APP_NAME)-$(VERSION).tar.gz" 