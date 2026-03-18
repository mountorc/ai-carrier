# AutoDataSource Go SDK

Go SDK for AutoDataSource API，提供完整的数据源管理、SQL 查询、向量数据库操作等功能。

## 功能特性

- **数据源管理**：获取数据源列表、添加、删除、检查存在、测试连接
- **SQL 操作**：执行 SQL 查询、更新、获取表字段和表列表
- **SQL 转换**：不同数据库方言之间的 SQL 语句转换
- **数据集管理**：获取数据集列表，根据 UUID 获取数据
- **OSS 文件管理**：查看 OSS 文件结构
- **文档管理**：获取公共文档列表和内容
- **向量数据库**：连接管理、集合操作、向量插入和搜索

## 快速开始

### 本地安装（推荐）

由于这是本地 SDK，您可以通过以下方式使用：

#### 方式一：使用本地模块

1. 将 `sdk/go` 目录复制到您的项目中
2. 在您的 `go.mod` 中添加：

```go
replace github.com/example/autodatasource-sdk-go => ./path/to/sdk/go
```

3. 然后执行：

```bash
go get github.com/example/autodatasource-sdk-go
```

#### 方式二：直接复制源代码

将 `sdk/go` 目录下的 `autodatasource.go` 文件复制到您的项目中使用。

### 使用示例

```go
package main

import (
	"fmt"
	"log"

	"github.com/example/autodatasource-sdk-go"
)

func main() {
	// 初始化客户端
	baseURL := "http://localhost:8080/autoDataSource"
	client := autodatasource.NewClient(baseURL)

	// 1. 获取数据源列表
	fmt.Println("=== 获取数据源列表 ===")
	dataSourcesResp, err := client.GetDataSources()
	if err != nil {
		log.Printf("获取数据源列表失败: %v\n", err)
	} else {
		fmt.Printf("数据源列表: %+v\n", dataSourcesResp)
	}

	// 2. 执行 SQL 查询
	fmt.Println("\n=== 执行 SQL 查询 ===")
	queryResult, err := client.ExecuteSqlQuery("your-data-source-id", "SELECT * FROM users LIMIT 10")
	if err != nil {
		log.Printf("SQL 查询失败: %v\n", err)
	} else if queryResult.Success {
		fmt.Printf("查询成功，返回 %d 条记录\n", queryResult.Total)
		fmt.Printf("数据: %+v\n", queryResult.DataList)
	} else {
		fmt.Printf("查询失败: %s\n", queryResult.Message)
	}

	// 3. 转换 SQL 语句
	fmt.Println("\n=== 转换 SQL 语句 ===")
	sql := "SELECT * FROM users LIMIT 10"
	transformResp, err := client.TransformSql(sql, "mysql", "oracle")
	if err != nil {
		log.Printf("转换 SQL 失败: %v\n", err)
	} else {
		fmt.Printf("原始 SQL: %s\n", sql)
		if data, ok := transformResp.Data.(map[string]interface{}); ok {
			fmt.Printf("转换后 SQL: %s\n", data["transformedSql"])
		}
	}

	// 4. 获取表字段信息
	fmt.Println("\n=== 获取表字段信息 ===")
	tableFields, err := client.GetTableFields("your-data-source-id", "your_database", "your_table")
	if err != nil {
		log.Printf("获取表字段失败: %v\n", err)
	} else {
		fmt.Printf("表字段: %+v\n", tableFields)
	}

	// 5. 测试向量数据库连接
	fmt.Println("\n=== 测试向量数据库连接 ===")
	vectorConfig := map[string]interface{}{
		"databaseType": "milvus",
		"host":         "localhost",
		"port":         19530,
	}
	vectorTest, err := client.TestVectorDatabaseConnection(vectorConfig)
	if err != nil {
		log.Printf("测试向量数据库连接失败: %v\n", err)
	} else {
		fmt.Printf("连接测试结果: %+v\n", vectorTest)
	}
}
```

## API 文档

### AutoDataSourceClient

#### NewClient 创建新的 AutoDataSource 客户端

```go
func NewClient(baseURL string) *AutoDataSourceClient
```

- `baseURL`: AutoDataSource API 基础 URL，例如 `http://localhost:8080/autoDataSource`

#### NewClientWithTimeout 创建带超时设置的客户端

```go
func NewClientWithTimeout(baseURL string, timeout int) *AutoDataSourceClient
```

- `baseURL`: AutoDataSource API 基础 URL
- `timeout`: 超时时间（秒）

#### 数据源管理方法

1. **获取外部数据源列表**

   ```go
   func (c *AutoDataSourceClient) GetDataSources() (map[string]interface{}, error)
   ```

2. **获取本地数据源列表**

   ```go
   func (c *AutoDataSourceClient) GetLocalDataSources() (*Response, error)
   ```

3. **添加数据源**

   ```go
   func (c *AutoDataSourceClient) AddDataSource(dataSourceId string, properties map[string]interface{}) (*Response, error)
   ```

4. **删除数据源**

   ```go
   func (c *AutoDataSourceClient) RemoveDataSource(dataSourceId string) (*Response, error)
   ```

5. **检查数据源是否存在**

   ```go
   func (c *AutoDataSourceClient) CheckDataSourceExists(dataSourceId string) (map[string]interface{}, error)
   ```

6. **测试数据源连接**

   ```go
   func (c *AutoDataSourceClient) TestConnection(properties map[string]interface{}) (*Response, error)
   ```

#### SQL 操作方法

1. **执行 SQL 查询**

   ```go
   func (c *AutoDataSourceClient) ExecuteSqlQuery(dataSourceId, sql string) (*Response, error)
   func (c *AutoDataSourceClient) ExecuteSqlQueryWithType(dataSourceId, sql, sqlType string) (*Response, error)
   ```

2. **执行 SQL 更新**

   ```go
   func (c *AutoDataSourceClient) ExecuteSqlUpdate(dataSourceId, sql string) (*Response, error)
   func (c *AutoDataSourceClient) ExecuteSqlUpdateWithType(dataSourceId, sql, sqlType string) (*Response, error)
   ```

3. **获取表字段信息**

   ```go
   func (c *AutoDataSourceClient) GetTableFields(dataSourceId, tableSchema, tableName string) (*Response, error)
   ```

4. **获取表列表**

   ```go
   func (c *AutoDataSourceClient) GetTableList(dataSourceId, tableSchema string) (*Response, error)
   ```

#### SQL 转换方法

1. **转换 SQL 语句**

   ```go
   func (c *AutoDataSourceClient) TransformSql(sql, sourceType, targetType string) (*Response, error)
   ```

#### 数据集管理方法

1. **获取数据集列表**

   ```go
   func (c *AutoDataSourceClient) GetDataSets() (*Response, error)
   ```

2. **根据 UUID 获取数据**

   ```go
   func (c *AutoDataSourceClient) GetDataByUuid(uuidAutoData string) (*Response, error)
   ```

3. **根据数据源 ID 预览数据集**

   ```go
   func (c *AutoDataSourceClient) PreviewDataSetsByDataSourceId(dataSourceId string) (*Response, error)
   ```

#### 向量数据库方法

1. **测试向量数据库连接**

   ```go
   func (c *AutoDataSourceClient) TestVectorDatabaseConnection(properties map[string]interface{}) (*Response, error)
   ```

2. **连接到向量数据库**

   ```go
   func (c *AutoDataSourceClient) ConnectToVectorDatabase(properties map[string]interface{}) (*Response, error)
   ```

3. **断开向量数据库连接**

   ```go
   func (c *AutoDataSourceClient) DisconnectFromVectorDatabase(connectionId, databaseType string) (*Response, error)
   ```

4. **创建向量集合**

   ```go
   func (c *AutoDataSourceClient) CreateVectorCollection(connectionId, databaseType, collectionName string, dimension int, indexType string) (*Response, error)
   ```

5. **删除向量集合**

   ```go
   func (c *AutoDataSourceClient) DropVectorCollection(connectionId, databaseType, collectionName string) (*Response, error)
   ```

6. **获取向量集合列表**

   ```go
   func (c *AutoDataSourceClient) ListVectorCollections(connectionId, databaseType string) (*Response, error)
   ```

7. **获取向量集合信息**

   ```go
   func (c *AutoDataSourceClient) GetVectorCollectionInfo(connectionId, databaseType, collectionName string) (*Response, error)
   ```

8. **插入向量数据**

   ```go
   func (c *AutoDataSourceClient) InsertVectors(connectionId, databaseType, collectionName string, vectors [][]float32, metadata []map[string]interface{}) (*Response, error)
   ```

9. **搜索向量数据**

   ```go
   func (c *AutoDataSourceClient) SearchVectors(connectionId, databaseType, collectionName string, queryVector []float32, topK int, filter string) (*Response, error)
   ```

#### 其他方法

1. **获取 OSS 文件列表**

   ```go
   func (c *AutoDataSourceClient) GetOssFiles(prefix string) (*Response, error)
   ```

2. **获取公共文档列表**

   ```go
   func (c *AutoDataSourceClient) GetPublicDocsList() (map[string]interface{}, error)
   ```

3. **获取公共文档内容**

   ```go
   func (c *AutoDataSourceClient) GetPublicDocContent(fileName string) (map[string]interface{}, error)
   ```

4. **获取数据抽取记录**

   ```go
   func (c *AutoDataSourceClient) GetExtractRecords() (*Response, error)
   ```

## 响应结构

### Response 统一响应结构

```go
type Response struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	DataList interface{} `json:"dataList,omitempty"`
	Message  string      `json:"message,omitempty"`
	Total    int         `json:"total,omitempty"`
}
```

## 向量数据库使用示例

```go
package main

import (
	"fmt"
	"log"

	"github.com/example/autodatasource-sdk-go"
)

func vectorExample() {
	client := autodatasource.NewClient("http://localhost:8080/autoDataSource")

	// 1. 连接向量数据库
	connResult, err := client.ConnectToVectorDatabase(map[string]interface{}{
		"databaseType": "milvus",
		"host":         "localhost",
		"port":         19530,
	})
	if err != nil {
		log.Fatal(err)
	}
	connectionId := "your-connection-id" // 从响应中获取

	// 2. 创建集合
	createResult, err := client.CreateVectorCollection(
		connectionId,
		"milvus",
		"test_collection",
		128, // 维度
		"FLAT",
	)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 插入向量
	vectors := [][]float32{
		make([]float32, 128),
		make([]float32, 128),
	}
	metadata := []map[string]interface{}{
		{"id": 1, "name": "item1"},
		{"id": 2, "name": "item2"},
	}
	insertResult, err := client.InsertVectors(connectionId, "milvus", "test_collection", vectors, metadata)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 搜索向量
	queryVector := make([]float32, 128)
	searchResult, err := client.SearchVectors(connectionId, "milvus", "test_collection", queryVector, 10, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("搜索结果: %+v\n", searchResult)
}
```

## 错误处理

SDK 会返回 `error` 类型的错误，建议在使用时进行适当的错误处理。

```go
result, err := client.ExecuteSqlQuery(dataSourceId, sql)
if err != nil {
    // 处理网络错误等
    log.Fatalf("请求失败: %v", err)
}
if !result.Success {
    // 处理业务错误
    log.Fatalf("操作失败: %s", result.Message)
}
```

## 版本历史

- v1.0.0: 完整功能版本
  - 支持所有数据源管理 API
  - 支持 SQL 查询和更新
  - 支持向量数据库操作
  - 支持 SQL 转换功能

## 许可证

MIT License
