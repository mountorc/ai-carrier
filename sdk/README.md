# SDK 目录

本目录包含两个独立的 SDK：

1. **autoDataSource-sdk** - AutoDataSource 动态数据源管理系统的 API 封装
2. **seekdb-sdk** - SeekDB AI 原生数据库的封装

## 两个 SDK 的区别

### autoDataSource-sdk
- **用途**: 通过 HTTP API 访问 AutoDataSource 服务
- **适用场景**: 需要管理多种数据源的场景
- **特点**: 支持数据源动态管理、SQL 查询、SQL 转换、向量数据库管理等

### seekdb-sdk
- **用途**: 直接连接 SeekDB 数据库
- **适用场景**: 需要嵌入向量数据库功能的场景
- **特点**: 提供本地数据库能力，支持关系型和向量数据库功能

## 快速开始

### 使用 autoDataSource-sdk

请参考 [autoDataSource-sdk/README.md](./autoDataSource-sdk/README.md)

### 使用 seekdb-sdk

请参考 [seekdb-sdk/README.md](./seekdb-sdk/README.md)

## 目录结构

```
sdk/
├── autoDataSource-sdk/    # AutoDataSource SDK
│   ├── java/
│   ├── go/
│   ├── python/
│   └── README.md
├── seekdb-sdk/           # SeekDB SDK
│   ├── java/
│   ├── go/
│   └── README.md
└── README.md             # 本文件
```
