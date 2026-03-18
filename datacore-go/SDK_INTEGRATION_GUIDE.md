# Dataset SDK 集成指南

本指南将帮助您在其他 Go 项目中使用 dataset-go SDK。

## 方法一：直接复制整个 datacore-go 目录（推荐）

最简单的方法是直接将整个 `datacore-go` 目录复制到您的项目中，然后配置 go.mod 使用本地路径引用。

### 步骤：

1. 将 `datacore-go` 文件夹复制到您的项目根目录

2. 在您的项目 `go.mod` 文件中添加以下 replace 指令：

```go
module your-project

go 1.21

require (
    github.com/example/dataset-go v0.0.0
    github.com/example/datasource v0.0.0
    github.com/example/postgres-sdk-go v0.0.0
)

replace github.com/example/dataset-go => ./datacore-go/dataset-go
replace github.com/example/datasource => ./datacore-go/datasource
replace github.com/example/postgres-sdk-go => ./datacore-go/../sdk/postgres-sdk/go
```

3. 在您的代码中导入并使用：

```go
package main

import (
    "fmt"
    "github.com/example/datasource"
    "github.com/example/dataset-go/database"
    "github.com/example/dataset-go/server"
    "github.com/example/dataset-go/scheduler"
)

func main() {
    // 1. 使用数据源管理
    manager := datasource.GetManager()
    
    configJSON := `{
        "uuid": "post4bc7-9a41-4332-93a1-a60c4d8a7e19",
        "type": "postgres",
        "config": {
            "host": "121.43.142.153",
            "port": 5432,
            "database": "carrier",
            "username": "carrier",
            "password": "GNerfiSP4dpZjwcJ"
        }
    }`
    
    err := manager.AddDataSource(configJSON)
    if err != nil {
        panic(err)
    }
    
    // 2. 获取数据
    ds, err := manager.GetDataSource("post4bc7-9a41-4332-93a1-a60c4d8a7e19")
    if err != nil {
        panic(err)
    }
    
    // 获取 dataset
    datasetResult, err := ds.GetDataset("dataset1")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Dataset: %#v\n", datasetResult)
    
    // 获取 workflow
    workflowResult, err := ds.GetWorkflow("workflow1")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Workflow: %#v\n", workflowResult)
    
    // 3. 启动 HTTP 服务器（可选）
    sched := scheduler.NewScheduler()
    db, err := database.NewDatabase("")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    srv := server.NewServer(sched, db)
    srv.Start(":8084")
}
```

## 方法二：创建独立的 SDK 包

如果您想创建一个更整洁的 SDK 结构，可以按照以下步骤：

### 1. 目录结构

```
your-project/
├── go.mod
├── main.go
└── datacore-go/          # 复制过来的完整目录
    ├── dataset-go/
    ├── datasource/
    └── ...
```

### 2. 配置 go.mod

```go
module your-project

go 1.21

require (
    github.com/example/dataset-go v0.0.0
    github.com/lib/pq v1.10.9
)

replace github.com/example/dataset-go => ./datacore-go/dataset-go
replace github.com/example/datasource => ./datacore-go/datasource
replace github.com/example/postgres-sdk-go => ./datacore-go/../sdk/postgres-sdk/go
```

## 功能特性

SDK 提供以下功能：

1. **数据源管理**
   - 添加、获取、删除数据源
   - 通过 uuid_datasource 管理多个数据源

2. **自动 JSONB 解析**
   - 自动识别 Base64 编码的 JSON 字符串
   - 自动解析为 JSON 对象

3. **AutoSet 自动展开**
   - 自动识别 `autoset` 或 `autoSet` 字段
   - 将字段内容展开到主对象中
   - 删除原始的 autoset 字段

4. **Dataset 和 Workflow 查询**
   - `GetDataset(uuid)` - 查询 dataset 表
   - `GetWorkflow(uuid)` - 查询 workflow 表
   - `GetAutoSet(uuid, tableName)` - 通用查询方法

5. **HTTP 服务器**
   - 提供 RESTful API 接口
   - 支持 GET 和 POST 请求

## API 参考

### 数据源管理

```go
import "github.com/example/datasource"

// 获取管理器单例
manager := datasource.GetManager()

// 添加数据源
err := manager.AddDataSource(configJSON)

// 获取数据源
ds, err := manager.GetDataSource(uuidDatasource)

// 删除数据源
manager.RemoveDataSource(uuidDatasource)
```

### 数据查询

```go
// 查询 dataset
result, err := ds.GetDataset(uuidDataset)

// 查询 workflow
result, err := ds.GetWorkflow(uuidWorkflow)

// 通用查询
result, err := ds.GetAutoSet(uuid, tableName)

// 执行 SQL
results, err := ds.Query(sql, args...)
```

### HTTP 服务器

```go
import (
    "github.com/example/dataset-go/database"
    "github.com/example/dataset-go/scheduler"
    "github.com/example/dataset-go/server"
)

sched := scheduler.NewScheduler()
db, err := database.NewDatabase(dsn)
srv := server.NewServer(sched, db)
srv.Start(":8084")
```

## 依赖

SDK 依赖以下包：
- `github.com/lib/pq v1.10.9` - PostgreSQL 驱动

运行 `go mod tidy` 自动安装依赖。

## 示例项目

您可以参考 `datacore-go/dataset-go/main.go` 作为完整的使用示例。
