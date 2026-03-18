# PostgreSQL SDK

PostgreSQL SDK 支持数据库连接和向量 embedding 功能，提供 Python、Go、Java 三种语言的实现。

## 功能特性

- ✅ **数据库连接管理**：支持 URL 和参数两种连接方式
- ✅ **SQL查询执行**：完整的查询和执行接口
- ✅ **向量操作**：向量插入、搜索（支持 pgvector）
- ✅ **集合管理**：创建、删除、列出集合（表）
- ✅ **事务支持**：完整的事务管理
- ✅ **表信息获取**：获取表列表、表字段信息

## 目录结构

```
postgres-sdk/
├── README.md
├── python/
│   ├── postgres_sdk/
│   │   ├── __init__.py
│   │   └── client.py
│   ├── example/
│   │   └── main.py
│   ├── requirements.txt
│   └── setup.py
├── go/
│   ├── postgres/
│   │   ├── client.go
│   │   ├── config.go
│   │   └── vector.go
│   ├── example/
│   │   └── main.go
│   ├── go.mod
│   └── README.md
└── java/
    ├── src/
    │   └── main/
    │       └── java/
    │           └── com/
    │               └── postgres/
    │                   ├── PostgresClient.java
    │                   └── VectorClient.java
    ├── example/
    │   └── Main.java
    └── pom.xml
```

## 快速开始

### Python SDK

```python
from postgres_sdk import PostgresClient

# 连接数据库
client = PostgresClient()
client.connect_from_url("postgresql://user:pass@host:5432/db")

# 创建向量集合
client.create_collection("test_vectors", 128, "IVFFLAT")

# 插入向量
client.insert_vectors("test_vectors", [[0.1, 0.2, ...]], [{"text": "hello"}])

# 搜索向量
results = client.search_vectors("test_vectors", [0.15, 0.25, ...], top_k=5)

client.disconnect()
```

### Go SDK

```go
package main

import "github.com/example/postgres-sdk-go/postgres"

func main() {
    config := postgres.NewConfig("host", 5432, "user", "pass", "db")
    client := postgres.NewClient(config)
    client.Connect()
    defer client.Disconnect()
}
```

### Java SDK

```java
import com.postgres.PostgresClient;

public class Main {
    public static void main(String[] args) {
        PostgresClient client = new PostgresClient();
        client.connectFromUrl("postgresql://user:pass@host:5432/db");
        client.disconnect();
    }
}
```

## 安装依赖

### Python

```bash
cd python
pip install -r requirements.txt
```

### Go

```bash
cd go
go mod download
```

### Java

```bash
cd java
mvn install
```

## 许可证

MIT License
