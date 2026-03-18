# Dataset Go

一个轻量级的Go语言SQL调度器内核，支持从JSON或JSON文件加载SQL语句到内存，并提供REST API通过UUID获取SQL。

## 项目结构

```
dataset-go/
├── go.mod
├── main.go
├── README.md
├── scheduler/
│   └── scheduler.go      # 核心调度器
├── server/
│   └── server.go        # HTTP服务器
└── example/
    └── test.go          # 测试示例
```

## 功能特性

1. **从JSON或JSON文件加载SQL配置**
2. **线程安全的内存存储**（使用读写锁）
3. **REST API接口**
4. **支持通过UUID获取SQL语句**

## 使用方法

### 1. 作为库使用

```go
package main

import (
    "fmt"
    "github.com/example/dataset-go/scheduler"
)

func main() {
    // 创建调度器
    sched := scheduler.NewScheduler()

    // 从文件加载
    err := sched.LoadFromFile("/path/to/sql_scheduler.json")
    if err != nil {
        panic(err)
    }

    // 通过UUID获取SQL
    sql, exists := sched.GetSQL("your-uuid-here")
    if exists {
        fmt.Println("SQL:", sql)
    }

    // 获取完整的SQL项信息
    item, exists := sched.GetSQLItem("your-uuid-here")
    if exists {
        fmt.Printf("Name: %s, Description: %s\n", item.Name, item.Description)
    }
}
```

### 2. 作为HTTP服务器启动

```bash
go run main.go -file /path/to/sql_scheduler.json -addr :8084
```

参数说明：
- `-file`: SQL配置JSON文件路径（必需）
- `-addr`: 服务器监听地址（可选，默认`:8084`）

## REST API接口

### 1. 健康检查

```
GET /health
```

响应示例：
```json
{
  "success": true,
  "status": "healthy",
  "count": 24
}
```

### 2. 通过UUID获取SQL

```
GET /sql?uuid=550e8400-e29b-41d4-a716-446655440001
```

响应示例：
```json
{
  "success": true,
  "uuid": "550e8400-e29b-41d4-a716-446655440001",
  "sql": "SELECT * FROM table_name"
}
```

### 3. 通过UUID获取完整SQL项信息

```
GET /sql/item?uuid=550e8400-e29b-41d4-a716-446655440001
```

响应示例：
```json
{
  "success": true,
  "item": {
    "uuid": "550e8400-e29b-41d4-a716-446655440001",
    "name": "listInstances",
    "description": "获取实例列表",
    "sql": "SELECT * FROM scheduler_instances"
  }
}
```

### 4. 列出所有UUID

```
GET /sql/list
```

响应示例：
```json
{
  "success": true,
  "count": 24,
  "uuids": ["uuid1", "uuid2", "uuid3"]
}
```

## 配置文件格式

配置文件必须是JSON格式，包含以下结构：

```json
{
  "description": "调度器 SQL 配置文件",
  "sqls": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440001",
      "name": "listInstances",
      "description": "获取实例列表",
      "sql": "SELECT * FROM scheduler_instances"
    }
  ]
}
```

## 运行测试

```bash
go run example/test.go
```

## 核心API

### Scheduler 方法

- `NewScheduler() *Scheduler` - 创建新的调度器
- `LoadFromFile(filePath string) error` - 从文件加载配置
- `LoadFromJSON(data []byte) error` - 从JSON数据加载配置
- `GetSQL(uuid string) (string, bool)` - 获取SQL语句
- `GetSQLItem(uuid string) (SQLItem, bool)` - 获取完整的SQL项
- `ListAllUUIDs() []string` - 列出所有UUID
- `Count() int` - 获取SQL语句总数
