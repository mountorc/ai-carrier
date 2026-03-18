# AutoDataSource Python SDK

AutoDataSource Python SDK 是一个用于与 AutoDataSource API 交互的 Python 客户端库。

## 功能特性

- 获取数据源列表（使用外部数据源列表接口）
- 转换 SQL 语句
- 获取数据集列表
- 根据 UUID 获取数据
- 获取 OSS 文件列表
- 获取公共文档列表和内容
- 根据数据源ID预览数据集

## 安装

### 使用 pip 安装

```bash
# 从本地安装
cd sdk/python
pip install .

# 或从源码包安装
pip install autodatasource-sdk-python-1.0.0.tar.gz
```

## 依赖

- Python 3.6+
- requests 2.28.0+

## 快速开始

### 初始化客户端

```python
from autodatasource import AutoDataSourceClient

# 初始化客户端
client = AutoDataSourceClient("http://localhost:8080/autoDataSource")
```

### 获取数据源列表

```python
try:
    response = client.get_data_sources()
    print("数据源列表:", response)
except Exception as e:
    print("错误:", str(e))
```

### 转换 SQL 语句

```python
try:
    sql = "SELECT * FROM users LIMIT 10"
    response = client.transform_sql(sql, "mysql", "oracle")
    print("转换后的 SQL:", response)
except Exception as e:
    print("错误:", str(e))
```

### 获取数据集列表

```python
try:
    response = client.get_data_items()
    print("数据集列表:", response)
except Exception as e:
    print("错误:", str(e))
```

### 根据 UUID 获取数据

```python
try:
    uuid = "e38ff77ff5b84985aff4cb97ebb87409"
    response = client.get_data_by_uuid(uuid)
    print("数据:", response)
except Exception as e:
    print("错误:", str(e))
```

### 获取 OSS 文件列表

```python
try:
    response = client.get_oss_files()
    print("OSS 文件列表:", response)
except Exception as e:
    print("错误:", str(e))
```

### 获取公共文档列表

```python
try:
    response = client.get_public_docs_list()
    print("公共文档列表:", response)
except Exception as e:
    print("错误:", str(e))
```

### 获取公共文档内容

```python
try:
    response = client.get_public_doc_content("README.md")
    print("文档内容:", response)
except Exception as e:
    print("错误:", str(e))
```

### 根据数据源ID预览数据集

```python
try:
    data_source_id = "90a49r2l313l4243robp5lbqf678vfn10719"
    response = client.preview_data_items_by_data_source_id(data_source_id)
    print("数据集预览:", response)
except Exception as e:
    print("错误:", str(e))
```

## API 参考

### 客户端方法

#### `__init__(base_url)`
- `base_url`: AutoDataSource API 基础 URL

#### `get_data_sources()`
- 返回: 数据源列表响应

#### `transform_sql(sql, source_type, target_type)`
- `sql`: SQL 语句
- `source_type`: 源数据库类型
- `target_type`: 目标数据库类型
- 返回: SQL 转换响应

#### `get_data_items()`
- 返回: 数据集列表响应

#### `get_data_by_uuid(uuid_auto_data)`
- `uuid_auto_data`: 数据集 UUID
- 返回: 数据响应

#### `get_oss_files(prefix=None)`
- `prefix`: 目录前缀（可选）
- 返回: OSS 文件列表响应

#### `get_public_docs_list()`
- 返回: 公共文档列表响应

#### `get_public_doc_content(file_name)`
- `file_name`: 文件名
- 返回: 文档内容响应

#### `preview_data_items_by_data_source_id(data_source_id)`
- `data_source_id`: 数据源ID
- 返回: 数据集预览响应

## 错误处理

所有方法都会抛出 `requests.exceptions.HTTPError` 异常，当 API 请求失败时。

## 版本历史

### v1.0.0
- 初始版本
- 支持所有核心 API 功能
- 添加根据数据源ID预览数据集功能

## 许可证

MIT
