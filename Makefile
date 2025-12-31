.PHONY: help build run demo upload download list delete clean

# 默认目标
help:
	@echo "Sealos S3 Client - Available Commands"
	@echo ""
	@echo "Build:"
	@echo "  make build          Build the binary"
	@echo ""
	@echo "Run:"
	@echo "  make demo           Run demo"
	@echo "  make upload         Upload test file"
	@echo "  make download       Download test file"
	@echo "  make list           List objects"
	@echo "  make delete         Delete test object"
	@echo ""
	@echo "Development:"
	@echo "  make clean          Clean build artifacts"
	@echo "  make deps           Download dependencies"
	@echo ""

# 构建
build:
	go build -o s3-client .

# 运行演示
demo:
	go run . -cmd demo

# 上传测试
upload:
	@echo "Creating test file..."
	@echo "Hello, Sealos S3!" > test-upload.txt
	go run . -cmd upload -object test-upload.txt -file test-upload.txt

# 下载测试
download:
	go run . -cmd download -object test-upload.txt -file test-download.txt
	@echo "Downloaded content:"
	@cat test-download.txt

# 列表测试
list:
	go run . -cmd list

# 删除测试
delete:
	go run . -cmd delete -object test-upload.txt

# 清理
clean:
	rm -f s3-client test-upload.txt test-download.txt

# 下载依赖
deps:
	go mod download
	go mod tidy

# 完整测试流程
test-flow: upload list download delete
	@echo "Test flow completed!"
