# AutoDataSource SDK

AutoDataSource SDK 是 AutoDataSource 动态数据源管理系统的 API 封装，提供 Java 和 Go 两种语言的 SDK。

## 概述

AutoDataSource 是一个动态数据源管理系统，支持多种数据库类型，提供数据源的动态添加、删除、查询，以及 SQL 查询、转换和表字段获取等功能。

## 功能特性

- 动态数据源管理（添加、删除、查询）
- SQL 查询执行
- SQL 转换功能
- 向量数据库管理
- 表字段获取

## 项目结构

```
autoDataSource-sdk/
├── java/           # Java SDK
│   ├── pom.xml
│   └── src/
├── go/            # Go SDK
│   ├── go.mod
│   ├── autodatasource.go
│   └── example/
├── python/        # Python SDK
│   └── autodatasource/
└── README.md
```

## Java SDK 使用说明

### 安装

将 SDK 安装到本地 Maven 仓库：

```bash
cd autoDataSource-sdk/java
mvn clean install
```

### 依赖配置

在项目的 `pom.xml` 中添加依赖：

```xml
<dependency>
    <groupId>com.example</groupId>
    <artifactId>autoDataSource-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

### 快速开始

```java
import com.example.autodatasource.sdk.AutoDataSourceClient;

public class Example {
    public static void main(String[] args) {
        AutoDataSourceClient client = new AutoDataSourceClient("http://localhost:8080/autoDataSource");
        
        try {
            // 获取数据源列表
            List<String> dataSources = client.getLocalDataSources();
            
            // 执行 SQL 查询
            Map<String, Object> result = client.executeQuery("test-ds", "SELECT * FROM test_table");
            
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
```

## Go SDK 使用说明

### 安装

```bash
cd autoDataSource-sdk/go
go mod tidy
```

### 快速开始

```go
package main

import (
    "fmt"
    autodatasource "github.com/example/autodatasource-sdk-go"
)

func main() {
    client := autodatasource.NewAutoDataSourceClient("http://localhost:8080/autoDataSource")
    
    // 获取数据源列表
    dataSources, err := client.GetLocalDataSources()
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Data sources:", dataSources)
}
```

## API 文档

详细的 API 文档请参考各语言 SDK 目录下的 README。

## 与 SeekDB SDK 的区别

- **AutoDataSource SDK**: 通过 HTTP API 访问 AutoDataSource 服务，适合需要管理多种数据源的场景
- **SeekDB SDK**: 直接连接 SeekDB 数据库，提供本地数据库能力，适合需要嵌入向量数据库功能的场景
