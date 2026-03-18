# Dataset SDK

一个完整的 Go SDK，用于处理数据集和工作流，支持自动解析 JSONB 字段和自动展开 autoset 字段。

## 功能特性

- 支持通过 uuid_datasource 管理多个数据源
- 自动识别和解析 JSONB 字段为 JSON 对象
- 自动展开 autoset 或 autoSet 字段内容
- 支持 dataset 和 workflow 表查询
- 提供简单易用的 API

## 安装

将 `dataset-sdk` 文件夹复制到您的项目中，然后在您的 `go.mod` 文件中添加：

```go
replace github.com/example/dataset-sdk => ./dataset-sdk
```

或者直接引用本地路径：

```go
import (
    "github.com/example/dataset-sdk/database"
    "github.com/example/dataset-sdk/datasource"
    "github.com/example/dataset-sdk/server"
)
```

## 快速开始

### 1. 添加数据源

```go
package main

import (
    "github.com/example/dataset-sdk/datasource"
)

func main() {
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
}
```

### 2. 获取 Dataset 数据

```go
ds, err := manager.GetDataSource("post4bc7-9a41-4332-93a1-a60c4d8a7e19")
if err != nil {
    panic(err)
}

result, err := ds.GetDataset("dataset1")
if err != nil {
    panic(err)
}

// result 会自动处理 autoset 字段
// 例如: {"autodatakey": "1", "sql": "select * from auto_ability_template", ...}
```

### 3. 获取 Workflow 数据

```go
result, err := ds.GetWorkflow("workflow1")
if err != nil {
    panic(err)
}
```

### 4. 使用 HTTP 服务器

```go
package main

import (
    "github.com/example/dataset-sdk/database"
    "github.com/example/dataset-sdk/scheduler"
    "github.com/example/dataset-sdk/server"
)

func main() {
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

## HTTP API

### 添加数据源
```
POST /datasource/add
Content-Type: application/json

{
  "config_json": "{\"uuid\":\"...\", \"type\":\"postgres\", \"config\":{...}}"
}
```

### 获取 Dataset
```
GET /getAutoSet?uuid_datasource=xxx&uuid_dataset=xxx
```

### 获取 Workflow
```
GET /getAutoSet?uuid_datasource=xxx&uuid_workflow=xxx
```

## 项目结构

```
dataset-sdk/
├── go.mod
├── README.md
├── postgres/          # PostgreSQL 客户端
├── datasource/        # 数据源管理
├── database/          # 数据库操作
├── scheduler/         # SQL 调度器
└── server/            # HTTP 服务器
```

## 依赖

- github.com/lib/pq v1.10.9
