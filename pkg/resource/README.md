# Resource Internal Client# Resource Service 客户端



资源服务内部客户端封装，提供简洁易用的 gRPC 接口供内部微服务调用。资源服务的 Go 客户端封装，提供简洁易用的接口用于获取文件信息和URL。



## 功能特性## 功能特性



- ✅ **文件信息获取** - 单个/批量获取文件元数据- ✅ **批量获取文件URL** - 一次请求获取多个文件的URL，包括图片变体

- ✅ **文件URL获取** - 批量获取文件访问URL，支持图片变体- ✅ **服务发现支持** - 支持 Consul 等服务发现机制

- ✅ **下载URL生成** - 生成带预签名的下载URL- ✅ **自动重试和恢复** - 内置错误恢复机制

- ✅ **秒传检查** - 通过SHA256校验和检查文件是否存在- ✅ **详细日志** - 完整的调用日志，便于问题排查

- ✅ **配额管理** - 获取和检查租户配额- ✅ **超时控制** - 可配置的超时时间

- ✅ **服务发现支持** - 支持 Consul 等服务发现机制- ✅ **中文注释** - 所有代码都有详细的中文注释

- ✅ **自动恢复** - 内置 recovery 中间件

- ✅ **超时控制** - 可配置的请求超时时间## 快速开始



## 快速开始### 安装



### 安装```bash

go get github.com/heyinLab/common

```bash```

go get github.com/heyinLab/common

```### 基本使用



### 基本使用#### 1. 创建客户端（直连方式）



#### 1. 创建客户端（服务发现方式 - 推荐）```go

package main

```go

package mainimport (

    "context"

import (    "fmt"

    "context"    "log"

    "log"

    "github.com/heyinLab/common/pkg/resource"

    consulapi "github.com/hashicorp/consul/api")

    "github.com/go-kratos/kratos/contrib/registry/consul/v2"

    "github.com/heyinLab/common/pkg/resource"func main() {

)    // 使用默认配置

    client, err := resource.NewClient(resource.DefaultConfig())

func main() {    if err != nil {

    // 创建 Consul 客户端        log.Fatal(err)

    consulClient, err := consulapi.NewClient(consulapi.DefaultConfig())    }

    if err != nil {    defer client.Close()

        log.Fatal(err)

    }    // 或者自定义配置

    config := resource.DefaultConfig().

    // 创建 Consul 服务发现        WithAddress("resource-service:9000").

    discovery := consul.New(consulClient)        WithTimeout(5 * time.Second)

    client, err := resource.NewClient(config)

    // 创建资源服务内部客户端    if err != nil {

    client, err := resource.NewResourceClientWithDiscovery(        log.Fatal(err)

        resource.DefaultInternalConfig(),    }

        discovery,    defer client.Close()

    )}

    if err != nil {```

        log.Fatal(err)

    }#### 2. 创建客户端（服务发现方式 - 推荐）

    defer client.Close()

```go

    // 使用客户端...package main

}

```import (

    "context"

#### 2. 创建客户端（直连方式）    "log"



```go    "github.com/heyinLab/common/pkg/resource"

package main    "github.com/go-kratos/kratos/v2/registry"

    // 导入你的服务发现实现，如 Consul

import ()

    "time"

    "github.com/heyinLab/common/pkg/resource"func main() {

)    // 创建服务发现实例（以 Consul 为例）

    consulRegistry := NewConsulRegistry(consulConfig)

func main() {

    // 直连到资源服务    // 创建带服务发现的客户端

    config := resource.DefaultInternalConfig().    config := resource.DefaultConfig().

        WithEndpoint("localhost:9000").        WithAddress("discovery:///resource-service")  // 使用服务发现地址

        WithTimeout(5 * time.Second)

    client, err := resource.NewClientWithDiscovery(config, consulRegistry)

    client, err := resource.NewResourceClient(config)    if err != nil {

    if err != nil {        log.Fatal(err)

        log.Fatal(err)    }

    }    defer client.Close()

    defer client.Close()}

}```

```

## 核心功能

## 核心功能

### 1. 批量获取文件URL（最常用）

### 1. 获取文件信息

适用场景：商品列表页、用户头像、相册等需要展示多个图片的场景

```go

// 获取单个文件信息```go

file, err := client.GetFile(ctx, tenantID, fileID)// 场景：商品列表页获取图片

if err != nil {func GetProductImages(client *resource.Client, productIDs []string) error {

    return err    ctx := context.Background()

}

    // 假设每个商品有一个主图文件ID

fmt.Printf("文件名: %s\n", file.Filename)    fileIDs := []string{

fmt.Printf("大小: %d 字节\n", file.Size)        "file_id_001",

fmt.Printf("类型: %s\n", file.ContentType)        "file_id_002",

fmt.Printf("状态: %s\n", file.Status)        "file_id_003",

fmt.Printf("分类: %s\n", file.FileCategory) // image/video/document/audio/archive/other    }



// 批量获取文件信息（最多100个）    // 批量获取文件URL（包含缩略图）

files, failedIDs, err := client.GetFiles(ctx, tenantID, []string{    urls, err := client.BatchGetFileUrls(ctx, fileIDs)

    "file_id_001",    if err != nil {

    "file_id_002",        return fmt.Errorf("获取图片URL失败: %w", err)

    "file_id_003",    }

})

if err != nil {    // 处理结果

    return err    for fileID, info := range urls {

}        if !info.Success {

            log.Printf("文件 %s 获取失败: %s", fileID, info.Error)

for fileID, file := range files {            continue

    fmt.Printf("文件: %s -> %s\n", fileID, file.Filename)        }

}

        fmt.Printf("文件ID: %s\n", fileID)

if len(failedIDs) > 0 {        fmt.Printf("原图URL: %s\n", info.URL)

    fmt.Printf("获取失败的文件: %v\n", failedIDs)        fmt.Printf("文件名: %s\n", info.Filename)

}        fmt.Printf("大小: %d 字节\n", info.Size)

```        fmt.Printf("类型: %s\n", info.ContentType)



### 2. 获取文件URL        // 获取不同尺寸的缩略图

        if thumbnailURL, ok := info.VariantUrls["thumbnail_200x200"]; ok {

```go            fmt.Printf("小缩略图: %s\n", thumbnailURL)

// 获取单个文件URL（便捷方法）        }

url, err := client.GetFileUrl(ctx, tenantID, fileID)        if mediumURL, ok := info.VariantUrls["thumbnail_800x800"]; ok {

if err != nil {            fmt.Printf("中等尺寸: %s\n", mediumURL)

    return err        }

}

fmt.Printf("文件URL: %s\n", url)        // 判断URL类型

        if info.IsPublic {

// 批量获取文件URL（最多100个）            fmt.Println("这是永久有效的公开URL（可长期缓存）")

results, err := client.GetFileUrls(ctx, tenantID, fileIDs, &resource.GetFileUrlsOptions{        } else {

    IncludeVariants: true,  // 包含缩略图等变体            fmt.Printf("这是临时URL，%d秒后过期（需定期刷新）\n", info.ExpiresIn)

    ExpiresIn:       7200,  // URL有效期2小时        }

})    }

if err != nil {

    return err    return nil

}}

```

for fileID, info := range results {

    if !info.Success {### 2. 批量获取文件URL（自定义选项）

        fmt.Printf("文件 %s 获取失败: %s\n", fileID, info.Error)

        continue```go

    }// 只获取原图URL，不要缩略图（节省响应大小）

func GetOriginalImagesOnly(client *resource.Client, fileIDs []string) error {

    fmt.Printf("原图URL: %s\n", info.Url)    ctx := context.Background()

    

    // 获取缩略图    urls, err := client.BatchGetFileUrlsWithOptions(ctx, &resource.BatchGetFileUrlsRequest{

    if thumb, ok := info.VariantUrls["thumbnail_200x200"]; ok {        FileIDs:         fileIDs,

        fmt.Printf("缩略图: %s\n", thumb)        IncludeVariants: false,  // 不包含变体

    }        ExpiresIn:       7200,    // 2小时有效期

    })

    // 判断URL类型    if err != nil {

    if info.IsPublic {        return err

        fmt.Println("永久有效的公开URL")    }

    } else {

        fmt.Printf("临时URL，%d秒后过期\n", info.ExpiresIn)    // 处理结果...

    }    return nil

}}

``````



### 3. 获取下载URL### 3. 获取单个文件元数据



```go适用场景：需要获取文件的详细信息（上传时间、状态等）

// 获取单个文件下载URL（便捷方法）

downloadUrl, err := client.GetDownloadUrl(ctx, tenantID, fileID)```go

if err != nil {func GetFileInfo(client *resource.Client, fileID string) error {

    return err    ctx := context.Background()

}

    file, err := client.GetFile(ctx, fileID)

// 批量获取下载URL（最多50个）    if err != nil {

results, err := client.GetDownloadUrls(ctx, tenantID, []resource.DownloadFileRequest{        return fmt.Errorf("获取文件信息失败: %w", err)

    {FileID: "file_001"},    }

    {FileID: "file_002", DownloadFilename: "自定义文件名.pdf"},

    {FileID: "file_003", VariantID: "thumbnail_800x800"}, // 下载缩略图    fmt.Printf("文件ID: %s\n", file.Id)

}, 3600) // URL有效期1小时    fmt.Printf("文件名: %s\n", file.Filename)

    fmt.Printf("大小: %d 字节\n", file.Size)

for fileID, info := range results {    fmt.Printf("类型: %s\n", file.ContentType)

    if info.Success {    fmt.Printf("状态: %s\n", file.Status)

        fmt.Printf("下载URL: %s\n", info.DownloadUrl)    fmt.Printf("上传时间: %s\n", file.CreatedAt.AsTime())

        fmt.Printf("文件名: %s, 大小: %d\n", info.Filename, info.Size)    fmt.Printf("上传者ID: %s\n", file.UploaderId)

    }    fmt.Printf("分类: %s\n", file.FileCategory)  // image/video/document等

}

```    // 检查自定义元数据

    if file.Metadata != nil {

### 4. 秒传检查        fmt.Printf("自定义元数据: %v\n", file.Metadata)

    }

```go

// 通过文件SHA256校验和检查文件是否已存在    return nil

exists, existingFile, err := client.CheckFileExists(ctx, tenantID, checksumSHA256, fileSize)}

if err != nil {```

    return err

}### 4. 获取单个文件下载URL



if exists {```go

    fmt.Printf("文件已存在，ID: %s\n", existingFile.Id)func DownloadFile(client *resource.Client, fileID string) error {

    // 可以直接使用已存在的文件，实现秒传    ctx := context.Background()

} else {

    fmt.Println("文件不存在，需要上传")    url, variantUrls, err := client.GetDownloadUrl(ctx, fileID)

}    if err != nil {

```        return fmt.Errorf("获取下载URL失败: %w", err)

    }

### 5. 配额管理

    fmt.Printf("下载URL: %s\n", url)

```go

// 获取租户配额信息    // 如果是图片，还会返回变体URL

quota, err := client.GetQuota(ctx, tenantID)    for variantID, variantURL := range variantUrls {

if err != nil {        fmt.Printf("变体 %s URL: %s\n", variantID, variantURL)

    return err    }

}

    return nil

fmt.Printf("存储配额: %d 字节\n", quota.StorageQuota)}

fmt.Printf("已用存储: %d 字节 (%.2f%%)\n", quota.StorageUsed, quota.StorageUsagePercent)```

fmt.Printf("文件数配额: %d\n", quota.FileCountQuota)

fmt.Printf("已用文件数: %d\n", quota.FileCountUsed)### 5. 列出文件

fmt.Printf("配额状态: %s\n", quota.Status) // active, suspended, exceeded

适用场景：文件管理后台、用户文件列表

// 在上传前检查配额

result, err := client.CheckQuota(ctx, tenantID, resource.CheckQuotaTypeUpload, fileSize)```go

if err != nil {func ListUserFiles(client *resource.Client) error {

    return err    ctx := context.Background()

}

    // 第1页，每页20条

if !result.Allowed {    files, total, err := client.ListFiles(ctx, 1, 20)

    return fmt.Errorf("配额不足: %s", result.Reason)    if err != nil {

}        return fmt.Errorf("列出文件失败: %w", err)

    }

// 继续上传操作...

```    fmt.Printf("总共 %d 个文件\n", total)



## 配置选项    for i, file := range files {

        fmt.Printf("%d. %s (%d 字节) - %s\n",

### InternalConfig 结构            i+1, file.Filename, file.Size, file.Status)

    }

```go

type InternalConfig struct {    return nil

    Endpoint    string        // 服务端点}

    ServiceName string        // 服务名称（用于服务发现）```

    Timeout     time.Duration // 请求超时时间

}## 在 Kratos 服务中集成

```

### 1. 在 Wire 中配置

### 配置示例

```go

```go// internal/server/server.go

// 默认配置（使用服务发现）package server

config := resource.DefaultInternalConfig()

// Endpoint: "discovery:///resource-service"import (

// ServiceName: "resource-service"    "github.com/heyinLab/common/pkg/resource"

// Timeout: 10s    "github.com/go-kratos/kratos/v2/registry"

    "github.com/google/wire"

// 链式配置)

config := resource.DefaultInternalConfig().

    WithServiceName("my-resource-service").  // 自定义服务名// ProviderSet 服务提供者集合

    WithTimeout(5 * time.Second)var ProviderSet = wire.NewSet(

    NewGRPCServer,

// 直连配置    NewHTTPServer,

config := resource.DefaultInternalConfig().    NewConsulRegistry,

    WithEndpoint("192.168.1.100:9000").    NewConsulDiscovery,

    WithTimeout(5 * time.Second)    NewResourceClient,  // 添加资源服务客户端

```)



## 最佳实践// NewResourceClient 创建资源服务客户端

func NewResourceClient(discovery registry.Discovery) (*resource.Client, error) {

### 1. 使用服务发现    config := resource.DefaultConfig().

        WithAddress("discovery:///resource-service").  // 使用服务发现

推荐使用 Consul 服务发现而不是硬编码IP地址：        WithTimeout(10 * time.Second)



```go    return resource.NewClientWithDiscovery(config, discovery)

// ✅ 推荐：使用服务发现}

client, err := resource.NewResourceClientWithDiscovery(```

    resource.DefaultInternalConfig(),

    consulDiscovery,### 2. 在 Biz 层注入使用

)

```go

// ❌ 不推荐：硬编码IP// internal/biz/product_usecase.go

config := resource.DefaultInternalConfig().package biz

    WithEndpoint("192.168.1.100:9000")

client, err := resource.NewResourceClient(config)import (

```    "context"



### 2. 合理设置超时时间    "github.com/heyinLab/common/pkg/resource"

)

根据业务场景设置合适的超时时间：

type ProductUsecase struct {

```go    resourceClient *resource.Client

// 快速响应场景（如用户头像）    // ... 其他依赖

config.WithTimeout(3 * time.Second)}



// 普通场景（如商品列表）func NewProductUsecase(

config.WithTimeout(5 * time.Second)    resourceClient *resource.Client,

    // ... 其他依赖

// 批量处理场景（如后台导出）) *ProductUsecase {

config.WithTimeout(30 * time.Second)    return &ProductUsecase{

```        resourceClient: resourceClient,

    }

### 3. 批量操作优化}



尽量使用批量接口减少网络往返：// GetProductDetail 获取商品详情（包含图片）

func (uc *ProductUsecase) GetProductDetail(ctx context.Context, productID string) (*Product, error) {

```go    // 1. 查询商品信息（包含图片文件ID列表）

// ✅ 推荐：批量获取    product, err := uc.repo.GetProduct(ctx, productID)

files, _, err := client.GetFiles(ctx, tenantID, fileIDs)    if err != nil {

        return nil, err

// ❌ 不推荐：循环单个获取    }

for _, id := range fileIDs {

    file, err := client.GetFile(ctx, tenantID, id)    // 2. 批量获取图片URL

    // ...    if len(product.ImageFileIDs) > 0 {

}        imageURLs, err := uc.resourceClient.BatchGetFileUrls(ctx, product.ImageFileIDs)

```        if err != nil {

            uc.log.Errorf("获取商品图片URL失败: %v", err)

### 4. 错误处理            // 降级处理：图片获取失败不影响商品信息返回

        } else {

```go            // 填充图片URL到商品对象

file, err := client.GetFile(ctx, tenantID, fileID)            for _, fileID := range product.ImageFileIDs {

if err != nil {                if urlInfo, ok := imageURLs[fileID]; ok && urlInfo.Success {

    // 检查是否是 gRPC 错误                    product.Images = append(product.Images, &ProductImage{

    if st, ok := status.FromError(err); ok {                        FileID:       fileID,

        switch st.Code() {                        OriginalURL:  urlInfo.URL,

        case codes.NotFound:                        ThumbnailURL: urlInfo.VariantUrls["thumbnail_200x200"],

            return fmt.Errorf("文件不存在: %s", fileID)                        MediumURL:    urlInfo.VariantUrls["thumbnail_800x800"],

        case codes.PermissionDenied:                    })

            return fmt.Errorf("无权访问文件: %s", fileID)                }

        case codes.DeadlineExceeded:            }

            return fmt.Errorf("请求超时")        }

        default:    }

            return fmt.Errorf("获取文件失败: %v", err)

        }    return product, nil

    }}

    return err```

}

```## 配置选项



## 完整示例### Config 结构



### 商品服务获取商品图片```go

type Config struct {

```go    Address     string        // 服务地址

package product    Timeout     time.Duration // 超时时间

    EnableTrace bool          // 是否启用追踪

import (    EnableLog   bool          // 是否启用日志

    "context"}

    "github.com/heyinLab/common/pkg/resource"```

)

### 配置示例

type ProductService struct {

    resourceClient *resource.ResourceClient```go

}// 默认配置

config := resource.DefaultConfig()

func NewProductService(resourceClient *resource.ResourceClient) *ProductService {

    return &ProductService{// 链式配置

        resourceClient: resourceClient,config := resource.DefaultConfig().

    }    WithAddress("resource-service:9000").

}    WithTimeout(5 * time.Second).

    WithTrace(true).

// GetProductWithImages 获取商品信息和图片URL    WithLog(true)

func (s *ProductService) GetProductWithImages(ctx context.Context, tenantID uint32, product *Product) error {

    // 收集所有图片ID// 或者直接创建

    fileIDs := []string{product.MainImageID}config := &resource.Config{

    fileIDs = append(fileIDs, product.DetailImageIDs...)    Address:     "discovery:///resource-service",

    Timeout:     10 * time.Second,

    // 批量获取图片URL    EnableTrace: true,

    urls, err := s.resourceClient.GetFileUrls(ctx, tenantID, fileIDs, &resource.GetFileUrlsOptions{    EnableLog:   true,

        IncludeVariants: true,}

    })```

    if err != nil {

        return err## 最佳实践

    }

### 1. 使用服务发现

    // 填充主图URL

    if info, ok := urls[product.MainImageID]; ok && info.Success {推荐使用服务发现而不是硬编码IP地址：

        product.MainImageURL = info.Url

        if thumb, ok := info.VariantUrls["thumbnail_400x400"]; ok {```go

            product.MainImageThumb = thumb// ✅ 推荐：使用服务发现

        }config := resource.DefaultConfig().

    }    WithAddress("discovery:///resource-service")



    // 填充详情图URLclient, err := resource.NewClientWithDiscovery(config, consulDiscovery)

    for _, imageID := range product.DetailImageIDs {

        if info, ok := urls[imageID]; ok && info.Success {// ❌ 不推荐：硬编码IP

            product.DetailImageURLs = append(product.DetailImageURLs, info.Url)config := resource.DefaultConfig().

        }    WithAddress("192.168.1.100:9000")

    }```



    return nil### 2. 合理设置超时时间

}

```根据业务场景设置合适的超时时间：



### 导出服务生成下载链接```go

// 快速响应场景（如用户头像）

```goconfig.WithTimeout(3 * time.Second)

package export

// 普通场景（如商品列表）

import (config.WithTimeout(5 * time.Second)

    "context"

    "github.com/heyinLab/common/pkg/resource"// 批量处理场景（如后台导出）

)config.WithTimeout(30 * time.Second)

```

type ExportService struct {

    resourceClient *resource.ResourceClient### 3. 错误处理和降级

}

图片获取失败不应该影响核心业务：

// GenerateExportDownloadLinks 为导出文件生成下载链接

func (s *ExportService) GenerateExportDownloadLinks(ctx context.Context, tenantID uint32, exportFiles []ExportFile) ([]DownloadLink, error) {```go

    // 构建下载请求func GetProductList(ctx context.Context) ([]*Product, error) {

    requests := make([]resource.DownloadFileRequest, len(exportFiles))    // 1. 查询商品信息

    for i, f := range exportFiles {    products, err := repo.ListProducts(ctx)

        requests[i] = resource.DownloadFileRequest{    if err != nil {

            FileID:           f.FileID,        return nil, err

            DownloadFilename: f.DisplayName, // 使用友好的文件名    }

        }

    }    // 2. 收集所有图片ID

    var fileIDs []string

    // 批量获取下载URL，有效期24小时    for _, p := range products {

    results, err := s.resourceClient.GetDownloadUrls(ctx, tenantID, requests, 86400)        if p.ImageFileID != "" {

    if err != nil {            fileIDs = append(fileIDs, p.ImageFileID)

        return nil, err        }

    }    }



    // 构建下载链接    // 3. 批量获取图片URL（失败时降级）

    links := make([]DownloadLink, 0, len(exportFiles))    imageURLs := make(map[string]*resource.FileUrlInfo)

    for _, f := range exportFiles {    if len(fileIDs) > 0 {

        if info, ok := results[f.FileID]; ok && info.Success {        urls, err := resourceClient.BatchGetFileUrls(ctx, fileIDs)

            links = append(links, DownloadLink{        if err != nil {

                Name:      info.Filename,            log.Errorf("获取图片URL失败（降级处理）: %v", err)

                URL:       info.DownloadUrl,            // 降级：使用默认占位图

                Size:      info.Size,        } else {

                ExpiresIn: info.ExpiresIn,            imageURLs = urls

            })        }

        }    }

    }

    // 4. 填充图片URL

    return links, nil    for _, p := range products {

}        if urlInfo, ok := imageURLs[p.ImageFileID]; ok && urlInfo.Success {

```            p.ImageURL = urlInfo.URL

            p.ThumbnailURL = urlInfo.VariantUrls["thumbnail_200x200"]

## API 参考        } else {

            p.ImageURL = "https://cdn.example.com/placeholder.png"  // 占位图

### ResourceClient 方法        }

    }

| 方法 | 描述 | 参数限制 |

|------|------|---------|    return products, nil

| `GetFile` | 获取单个文件信息 | - |}

| `GetFiles` | 批量获取文件信息 | 最多100个 |```

| `GetFileUrl` | 获取单个文件URL | - |

| `GetFileUrls` | 批量获取文件URL | 最多100个 |### 4. 缓存优化

| `GetDownloadUrl` | 获取单个下载URL | - |

| `GetDownloadUrls` | 批量获取下载URL | 最多50个 |对于公开URL，可以长期缓存：

| `CheckFileExists` | 检查文件是否存在 | - |

| `GetQuota` | 获取租户配额 | - |```go

| `CheckQuota` | 检查配额 | - |func GetImageURL(ctx context.Context, fileID string) (string, error) {

    // 1. 先从缓存获取

### 检查配额类型    if cachedURL, ok := cache.Get(fileID); ok {

        return cachedURL, nil

| 类型 | 常量 | 描述 |    }

|------|------|------|

| upload | `CheckQuotaTypeUpload` | 上传检查 |    // 2. 调用资源服务

| download | `CheckQuotaTypeDownload` | 下载检查 |    urls, err := resourceClient.BatchGetFileUrls(ctx, []string{fileID})

| storage | `CheckQuotaTypeStorage` | 存储检查 |    if err != nil {

        return "", err
    }

    urlInfo := urls[fileID]
    if !urlInfo.Success {
        return "", fmt.Errorf(urlInfo.Error)
    }

    // 3. 根据URL类型设置缓存时间
    if urlInfo.IsPublic {
        // 公开URL永久有效，缓存1天
        cache.Set(fileID, urlInfo.URL, 24*time.Hour)
    } else {
        // 私有URL有过期时间，提前5分钟刷新
        cacheDuration := time.Duration(urlInfo.ExpiresIn-300) * time.Second
        cache.Set(fileID, urlInfo.URL, cacheDuration)
    }

    return urlInfo.URL, nil
}
```

## 完整示例

### PWA 服务集成示例

```go
// cmd/pwa/wire.go
//go:build wireinject

package main

import (
    "github.com/heyinLab/common/pkg/resource"
    "github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
    panic(wire.Build(
        server.ProviderSet,
        data.ProviderSet,
        biz.ProviderSet,
        service.ProviderSet,
        newApp,
    ))
}

// internal/server/server.go
package server

import (
    "github.com/heyinLab/common/pkg/resource"
)

var ProviderSet = wire.NewSet(
    NewResourceClient,
    // ... 其他
)

func NewResourceClient(discovery registry.Discovery, logger log.Logger) *resource.Client {
    config := resource.DefaultConfig().
        WithAddress("discovery:///resource-service")

    client, err := resource.NewClientWithDiscovery(config, discovery)
    if err != nil {
        logger.Fatal("创建资源服务客户端失败", err)
    }

    return client
}

// internal/biz/app_usecase.go
package biz

type AppUsecase struct {
    resourceClient *resource.Client
    log            *log.Helper
}

func NewAppUsecase(resourceClient *resource.Client, logger log.Logger) *AppUsecase {
    return &AppUsecase{
        resourceClient: resourceClient,
        log:            log.NewHelper(logger),
    }
}

func (uc *AppUsecase) GetAppWithIcon(ctx context.Context, appID string) (*App, error) {
    // 查询应用信息
    app, err := uc.repo.GetApp(ctx, appID)
    if err != nil {
        return nil, err
    }

    // 获取图标URL
    if app.IconFileID != "" {
        urls, err := uc.resourceClient.BatchGetFileUrls(ctx, []string{app.IconFileID})
        if err != nil {
            uc.log.Warnf("获取应用图标失败: %v", err)
        } else if urlInfo, ok := urls[app.IconFileID]; ok && urlInfo.Success {
            app.IconURL = urlInfo.URL
            app.IconThumbnail = urlInfo.VariantUrls["thumbnail_200x200"]
        }
    }

    return app, nil
}
```

## 常见问题

### Q1: 如何处理批量获取时部分文件失败？

A: `BatchGetFileUrls` 返回的是 map，每个文件都有 `Success` 字段，失败的文件不会影响其他文件：

```go
urls, err := client.BatchGetFileUrls(ctx, fileIDs)
if err != nil {
    return err  // 整个请求失败
}

for fileID, info := range urls {
    if !info.Success {
        log.Errorf("文件 %s 失败: %s", fileID, info.Error)
        continue  // 跳过失败的文件
    }
    // 处理成功的文件
}
```

### Q2: URL什么时候会过期？

A: 取决于文件的公开设置：
- **公开文件**（`IsPublic=true`）：返回CDN URL，永久有效
- **私有文件**（`IsPublic=false`）：返回预签名URL，有过期时间（默认1小时）

### Q3: 如何获取不同尺寸的缩略图？

A: 使用 `VariantUrls` 字段：

```go
urlInfo := urls[fileID]
thumbnailURL := urlInfo.VariantUrls["thumbnail_200x200"]  // 小图
mediumURL := urlInfo.VariantUrls["thumbnail_800x800"]      // 中图
largeURL := urlInfo.VariantUrls["thumbnail_1600x1600"]     // 大图
```

具体有哪些变体取决于上传时的策略配置。

### Q4: 性能怎么样？

A: 内网 gRPC 调用通常在 10ms 以内：
- 单个文件查询：~5-10ms
- 批量查询（50个文件）：~10-20ms
- 建议一次不超过100个文件

## API 文档

详细的 API 文档请参考：
- [资源服务 API 文档](../../docs/integration/service-integration.md)
- [Proto 定义](../../api/protos/resource/v1/)

## 版本历史

- v1.0.0 (2025-11-21): 初始版本
  - 支持批量获取文件URL
  - 支持服务发现
  - 完整的中文注释

## 许可证

内部项目，仅供公司内部使用。
