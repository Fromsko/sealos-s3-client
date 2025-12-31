# Sealos 对象存储（Object Storage）使用指南

> 📒 技术笔记 | 更新时间：2025-12-31

## 概述

Sealos Object Storage 是 Sealos 云平台内置的对象存储服务，主要用于存储和管理非结构化数据（如图片、视频、文档、备份文件等）。该服务完全兼容 **S3 API**，可以使用任何 S3 兼容的客户端或 SDK 进行访问。

## 核心概念

### 存储桶（Bucket）

- 存储桶是存储对象的容器，类似于文件系统中的顶级目录
- 每个存储桶有唯一的名称
- 在上传任何文件之前，必须先创建存储桶

### 对象（Object）

- 对象是存储在存储桶中的基本单元
- 每个对象由唯一的 Key（键名）标识
- 对象包含：数据内容、元数据、版本 ID

### 访问凭证

- **Access Key**：访问密钥 ID（相当于用户名）
- **Secret Key**：秘密访问密钥（相当于密码）

## 访问端点（Endpoint）

Sealos 对象存储提供两种访问端点：

| 类型                 | 说明         | 使用场景                            |
| -------------------- | ------------ | ----------------------------------- |
| **Internal（内部）** | 内部访问地址 | Sealos 平台内部应用访问，无流量费用 |
| **External（外部）** | 外部访问地址 | 外部网络访问，可能产生流量费用      |

## 使用方式

### 1. 通过 Sealos 控制台

1. 登录 [Sealos Cloud](https://cloud.sealos.io)
2. 在应用面板中找到 **Object Storage** 应用
3. 创建存储桶，设置名称和访问权限
4. 获取 Access Key 和 Secret Key
5. 上传/下载/管理文件

### 2. 通过 S3 兼容 SDK

由于 Sealos 对象存储兼容 S3 API，可以使用各种语言的 S3 SDK：

#### Go 语言示例（使用 minio-go）

```go
package main

import (
    "context"
    "log"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
    // Sealos 对象存储配置
    endpoint := "your-endpoint.sealos.run"  // 替换为实际端点
    accessKey := "your-access-key"           // 替换为实际 Access Key
    secretKey := "your-secret-key"           // 替换为实际 Secret Key
    useSSL := true

    // 初始化客户端
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        log.Fatalln(err)
    }

    // 创建存储桶
    bucketName := "my-bucket"
    ctx := context.Background()

    err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
    if err != nil {
        // 检查存储桶是否已存在
        exists, errBucketExists := client.BucketExists(ctx, bucketName)
        if errBucketExists == nil && exists {
            log.Printf("Bucket %s already exists\n", bucketName)
        } else {
            log.Fatalln(err)
        }
    } else {
        log.Printf("Successfully created bucket %s\n", bucketName)
    }

    // 上传文件
    objectName := "example.txt"
    filePath := "/path/to/file.txt"
    contentType := "text/plain"

    info, err := client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
        ContentType: contentType,
    })
    if err != nil {
        log.Fatalln(err)
    }
    log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

    // 下载文件
    err = client.FGetObject(ctx, bucketName, objectName, "/path/to/download.txt", minio.GetObjectOptions{})
    if err != nil {
        log.Fatalln(err)
    }
    log.Println("Successfully downloaded file")
}
```

#### JavaScript/Node.js 示例

```javascript
const Minio = require("minio");

// 初始化客户端
const minioClient = new Minio.Client({
  endPoint: "your-endpoint.sealos.run", // 替换为实际端点
  port: 443,
  useSSL: true,
  accessKey: "your-access-key", // 替换为实际 Access Key
  secretKey: "your-secret-key", // 替换为实际 Secret Key
});

// 创建存储桶
const bucketName = "my-bucket";

minioClient.makeBucket(bucketName, "", (err) => {
  if (err) {
    if (err.code === "BucketAlreadyOwnedByYou") {
      console.log("Bucket already exists");
    } else {
      return console.log("Error creating bucket:", err);
    }
  }
  console.log("Bucket created successfully");
});

// 上传文件
const objectName = "example.txt";
const filePath = "/path/to/file.txt";

minioClient.fPutObject(bucketName, objectName, filePath, {}, (err, etag) => {
  if (err) return console.log(err);
  console.log("File uploaded successfully. ETag:", etag);
});

// 下载文件
minioClient.fGetObject(
  bucketName,
  objectName,
  "/path/to/download.txt",
  (err) => {
    if (err) return console.log(err);
    console.log("File downloaded successfully");
  }
);

// 列出存储桶中的对象
const stream = minioClient.listObjects(bucketName, "", true);
stream.on("data", (obj) => console.log(obj));
stream.on("error", (err) => console.log(err));
```

#### Python 示例（使用 boto3）

```python
import boto3
from botocore.config import Config

# 配置 S3 客户端
s3_client = boto3.client(
    's3',
    endpoint_url='https://your-endpoint.sealos.run',  # 替换为实际端点
    aws_access_key_id='your-access-key',               # 替换为实际 Access Key
    aws_secret_access_key='your-secret-key',           # 替换为实际 Secret Key
    config=Config(signature_version='s3v4')
)

# 创建存储桶
bucket_name = 'my-bucket'
try:
    s3_client.create_bucket(Bucket=bucket_name)
    print(f'Bucket {bucket_name} created successfully')
except Exception as e:
    print(f'Bucket may already exist: {e}')

# 上传文件
s3_client.upload_file('/path/to/file.txt', bucket_name, 'example.txt')
print('File uploaded successfully')

# 下载文件
s3_client.download_file(bucket_name, 'example.txt', '/path/to/download.txt')
print('File downloaded successfully')

# 列出对象
response = s3_client.list_objects_v2(Bucket=bucket_name)
for obj in response.get('Contents', []):
    print(f"Object: {obj['Key']}, Size: {obj['Size']}")
```

### 3. 通过 AWS CLI

```bash
# 配置 AWS CLI
aws configure set aws_access_key_id your-access-key
aws configure set aws_secret_access_key your-secret-key

# 创建存储桶
aws --endpoint-url https://your-endpoint.sealos.run s3 mb s3://my-bucket

# 上传文件
aws --endpoint-url https://your-endpoint.sealos.run s3 cp ./file.txt s3://my-bucket/

# 下载文件
aws --endpoint-url https://your-endpoint.sealos.run s3 cp s3://my-bucket/file.txt ./

# 列出存储桶内容
aws --endpoint-url https://your-endpoint.sealos.run s3 ls s3://my-bucket/

# 删除文件
aws --endpoint-url https://your-endpoint.sealos.run s3 rm s3://my-bucket/file.txt
```

## 常用操作

| 操作                 | 说明               |
| -------------------- | ------------------ |
| `MakeBucket`         | 创建存储桶         |
| `ListBuckets`        | 列出所有存储桶     |
| `BucketExists`       | 检查存储桶是否存在 |
| `RemoveBucket`       | 删除存储桶         |
| `PutObject`          | 上传对象           |
| `GetObject`          | 下载对象           |
| `ListObjects`        | 列出存储桶中的对象 |
| `RemoveObject`       | 删除对象           |
| `CopyObject`         | 复制对象           |
| `PresignedGetObject` | 生成预签名下载 URL |
| `PresignedPutObject` | 生成预签名上传 URL |

## 最佳实践

1. **安全性**

   - 不要在代码中硬编码 Access Key 和 Secret Key
   - 使用环境变量或配置文件管理凭证
   - 定期轮换访问密钥

2. **性能优化**

   - 大文件使用分段上传（Multipart Upload）
   - 内部应用优先使用 Internal 端点
   - 合理设置对象的 Content-Type

3. **成本控制**
   - 设置生命周期策略自动清理过期数据
   - 监控存储使用量
   - 内部访问使用内部端点避免流量费用

## 参考资源

- [Sealos 官方文档](https://docs.sealos.io)
- [MinIO Go SDK](https://github.com/minio/minio-go)
- [MinIO JavaScript SDK](https://www.npmjs.com/package/minio)
- [AWS S3 API 参考](https://docs.aws.amazon.com/s3/index.html)
- [boto3 文档](https://boto3.amazonaws.com/v1/documentation/api/latest/index.html)

---

_Content was rephrased for compliance with licensing restrictions_
_来源：[Sealos 官方文档](https://docs.sealos.io)_
