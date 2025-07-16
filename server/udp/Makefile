# UDP Server Makefile

.PHONY: test bench build clean example client dtls-example dtls-client certs

# 运行测试
test:
	go test -v -cover ./...

# 运行基准测试
bench:
	go test -bench=. -benchmem ./...

# 构建示例程序
build:
	go build -o bin/udp_server ./example
	go build -o bin/udp_client ./client_example

# 构建DTLS示例程序
build-dtls:
	go build -o bin/dtls_server ./dtls_example
	go build -o bin/dtls_client ./dtls_client_example

# 运行示例服务器
example:
	go run ./example/main.go

# 运行示例客户端
client:
	go run ./client_example/main.go

# 运行DTLS示例服务器
dtls-example: certs
	go run ./dtls_example/main.go

# 运行DTLS示例客户端
dtls-client:
	go run ./dtls_client_example/main.go

# 生成测试证书
certs:
	@if [ "$(OS)" = "Windows_NT" ]; then \
		scripts/generate_dtls_certs.bat; \
	else \
		chmod +x scripts/generate_dtls_certs.sh && ./scripts/generate_dtls_certs.sh; \
	fi

# 测试DTLS功能
test-dtls: certs
	@echo "Testing DTLS functionality..."
	@go run ./dtls_example/main.go &
	@sleep 2
	@go run ./dtls_client_example/main.go
	@pkill -f "dtls_example" || true

# 清理构建文件
clean:
	rm -rf bin/
	rm -f server.crt server.key

# 创建bin目录
bin:
	mkdir -p bin

# 构建所有
all: bin build build-dtls

# 安装依赖
deps:
	go mod tidy

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
vet:
	go vet ./...

# 完整检查
check: fmt vet test

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  test         - 运行测试"
	@echo "  bench        - 运行基准测试"
	@echo "  build        - 构建UDP示例程序"
	@echo "  build-dtls   - 构建DTLS示例程序"
	@echo "  example      - 运行UDP服务器示例"
	@echo "  client       - 运行UDP客户端示例"
	@echo "  dtls-example - 运行DTLS服务器示例"
	@echo "  dtls-client  - 运行DTLS客户端示例"
	@echo "  certs        - 生成测试证书"
	@echo "  test-dtls    - 测试DTLS功能"
	@echo "  clean        - 清理构建文件"
	@echo "  all          - 构建所有程序"
	@echo "  deps         - 安装依赖"
	@echo "  fmt          - 代码格式化"
	@echo "  vet          - 代码检查"
	@echo "  check        - 完整检查"
	@echo "  help         - 显示此帮助信息"