# Sealos S3 Integration Client

中文 | [English](README.md)

一个生产级的 Go S3 集成客户端，用于与 Sealos 对象存储服务交互。

## 特性

- ✅ 完整的 S3 操作支持（上传、下载、列表、删除）
- ✅ 自动重试机制（指数退避）
- ✅ 并发操作支持（带限流）
- ✅ 预签名 URL 生成
- ✅ 生产级错误处理
- ✅ 详细的日志记录
- ✅ 连接池优化

## 快速开始

### 安装依赖

```bash
go mod download
```

### 基本使用

```bash
# 运行演示
go run . -cmd demo

# 上传文件
go run . -cmd upload -object myfile.txt -file ./myfile.txt

# 下载文件
go run . -cmd download -object myfile.txt -file ./downloaded.txt

# 列出对象
go run . -cmd list -prefix "documents/"

# 删除对象
go run . -cmd delete -object myfile.txt

# 生成预签名 URL
go run . -cmd presigned -object myfile.txt
```

## 配置

### 命令行参数

```bash
-endpoint string      S3 端点
-key string          访问密钥
-secret string       秘密密钥
-bucket string       Bucket 名称
-object string       对象名称
-file string         文件路径
-prefix string       列表前缀
-external            使用外部端点
```

## 端点

- **内部端点**（推荐）：`object-storage.objectstorage-system.svc.cluster.local`

  - 用于 Sealos 平台内应用
  - 无流量费用
  - 低延迟

- **外部端点**：`objectstorageapi.hzh.sealos.run`
  - 用于外部网络访问
  - 可能产生流量费用

## 项目结构

```
sealos-s3-client/
├── main.go           # 主程序和 CLI 接口
├── config.go         # 配置管理
├── client.go         # S3 客户端核心
├── operations.go     # 基础操作（上传、下载等）
├── concurrent.go     # 并发操作
├── go.mod            # Go 模块定义
└── README.md         # 本文件
```

## 代码示例

### 初始化客户端

```go
package main

import (
    "context"
    "log"
    "os"
)

func main() {
    logger := log.New(os.Stdout, "[S3] ", log.LstdFlags)

    config := LoadConfig(
        "object-storage.objectstorage-system.svc.cluster.local",
        "your-access-key",
        "your-secret-key",
        "your-bucket-name",
    )

    client, err := NewS3Client(config, logger)
    if err != nil {
        logger.Fatalf("Failed to create client: %v", err)
    }

    ctx := context.Background()

    // 使用 client 进行操作
}
```

### 上传文件

```go
err := client.UploadFile(ctx, "myfile.txt", "./local-file.txt")
if err != nil {
    logger.Fatalf("Upload failed: %v", err)
}
```

### 下载文件

```go
err := client.DownloadFile(ctx, "myfile.txt", "./downloaded.txt")
if err != nil {
    logger.Fatalf("Download failed: %v", err)
}
```

### 列出对象

```go
objects, err := client.ListObjects(ctx, "prefix/")
if err != nil {
    logger.Fatalf("List failed: %v", err)
}

for _, obj := range objects {
    println(obj)
}
```

### 并发上传

```go
files := map[string]string{
    "file1.txt": "./local1.txt",
    "file2.txt": "./local2.txt",
    "file3.txt": "./local3.txt",
}

results := client.UploadFilesParallel(ctx, files, 10)
client.PrintResults(results)
```

### 生成预签名 URL

```go
import "time"

url, err := client.GeneratePresignedURL(ctx, "myfile.txt", 24*time.Hour)
if err != nil {
    logger.Fatalf("Failed to generate URL: %v", err)
}

println("Download URL:", url)
```

## 错误处理

客户端自动处理以下错误：

- **网络错误**：连接超时、连接重置等（自动重试）
- **临时错误**：服务暂时不可用（自动重试）
- **永久错误**：权限不足、对象不存在等（立即返回）

重试策略：

- 最多重试 3 次
- 使用指数退避：1s, 2s, 4s

## 性能优化

### 连接池

客户端自动配置连接池：

- 最大空闲连接：100
- 每个主机最大空闲连接：10
- 空闲连接超时：90 秒

### 并发限制

并发操作默认限制为 10 个并发任务，可通过参数调整：

```go
results := client.UploadFilesParallel(ctx, files, 20) // 20 个并发
```

### 大文件处理

minio-go 自动处理大文件分段上传：

- 文件 > 64MB 时自动分段
- 可配置分段大小和并发数

## 最佳实践

1. **凭证管理**

   - 不要硬编码凭证
   - 使用环境变量或配置文件
   - 定期轮换访问密钥

2. **错误处理**

   - 总是检查返回的 error
   - 区分临时错误和永久错误
   - 记录详细的错误信息

3. **资源管理**

   - 及时关闭文件和连接
   - 使用 defer 确保资源释放
   - 监控内存使用

4. **性能**

   - 使用内部端点避免流量费用
   - 并发操作时合理设置并发数
   - 监控 S3 操作的延迟

5. **安全**
   - 使用预签名 URL 而不是暴露凭证
   - 验证上传文件的完整性
   - 定期审查访问日志

## 故障排除

### 连接失败

```
failed to verify S3 connection: failed to check bucket existence
```

**解决方案**：

- 检查端点是否正确
- 检查网络连接
- 验证凭证是否正确

### 权限错误

```
upload failed: AccessDenied
```

**解决方案**：

- 验证 Access Key 和 Secret Key
- 检查 Bucket 权限
- 确保凭证有足够的权限

### 超时错误

```
upload failed: i/o timeout
```

**解决方案**：

- 检查网络连接
- 增加超时时间
- 检查文件大小

## 许可证

MIT

## 参考资源

- [Sealos 官方文档](https://docs.sealos.io)
- [MinIO Go SDK](https://github.com/minio/minio-go)
- [AWS S3 API 参考](https://docs.aws.amazon.com/s3/)
