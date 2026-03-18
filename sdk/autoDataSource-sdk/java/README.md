# AutoDataSource Java SDK

Java SDK for AutoDataSource API，提供完整的数据源管理、SQL 查询、向量数据库操作等功能。

## 功能特性

- **数据源管理**：获取数据源列表、添加、删除、检查存在、测试连接
- **SQL 操作**：执行 SQL 查询、更新、获取表字段和表列表
- **SQL 转换**：不同数据库方言之间的 SQL 语句转换
- **数据集管理**：获取数据集列表，根据 UUID 获取数据
- **OSS 文件管理**：查看 OSS 文件结构
- **文档管理**：获取公共文档列表和内容
- **向量数据库**：连接管理、集合操作、向量插入和搜索
- **统一客户端**：支持 HTTP API 和 Spark JDBC 两种模式

## 快速开始

### 安装

#### 方式一：Maven 本地安装（推荐）

SDK 已经安装到本地 Maven 仓库，直接在项目中添加依赖即可：

```xml
<dependency>
    <groupId>com.example.autodatasource</groupId>
    <artifactId>autodatasource-sdk-java</artifactId>
    <version>1.0.0</version>
</dependency>
```

#### 方式二：手动安装

如果需要重新安装：

```bash
cd sdk/java
mvn clean install
```

#### 方式三：使用 JAR 包

将 `target/autodatasource-sdk-java-1.0.0.jar` 直接添加到项目的 classpath 中。

### 使用示例

```java
import com.example.autodatasource.sdk.AutoDataSourceClient;
import com.example.autodatasource.sdk.AutoDataSourceUnifiedClient;
import com.example.autodatasource.sdk.DataSourceMode;
import java.util.List;
import java.util.Map;

public class CompleteExample {
    public static void main(String[] args) {
        String baseUrl = "http://localhost:8080/autoDataSource";
        
        try {
            // ========== 方式一：使用基础客户端 ==========
            System.out.println("=== 基础客户端示例 ===");
            AutoDataSourceClient client = new AutoDataSourceClient(baseUrl);
            
            // 1. 获取数据源列表
            System.out.println("\n1. 获取数据源列表");
            Map<String, Object> dataSources = client.getDataSources();
            System.out.println("数据源: " + dataSources);
            
            // 2. 执行 SQL 查询
            System.out.println("\n2. 执行 SQL 查询");
            Map<String, Object> queryResult = client.executeSqlQuery(
                "your-data-source-id", 
                "SELECT * FROM users LIMIT 10"
            );
            System.out.println("查询结果: " + queryResult);
            
            // 3. 转换 SQL 语句
            System.out.println("\n3. 转换 SQL 语句");
            String originalSql = "SELECT * FROM users LIMIT 10";
            Map<String, Object> transformResult = client.transformSql(
                originalSql, 
                "mysql", 
                "oracle"
            );
            System.out.println("原始 SQL: " + originalSql);
            System.out.println("转换结果: " + transformResult);
            
            // ========== 方式二：使用统一客户端（推荐） ==========
            System.out.println("\n=== 统一客户端示例 ===");
            AutoDataSourceUnifiedClient unifiedClient = new AutoDataSourceUnifiedClient(
                baseUrl, 
                DataSourceMode.HTTP_API
            );
            
            // 4. 使用统一客户端执行查询
            System.out.println("\n4. 统一客户端查询");
            AutoDataSourceUnifiedClient.QueryResult result = unifiedClient.executeQuery(
                "your-data-source-id",
                "SELECT * FROM products"
            );
            
            if (result.isSuccess()) {
                System.out.println("查询成功！");
                System.out.println("记录数: " + result.getTotal());
                List<Map<String, Object>> data = result.getDataList();
                if (data != null) {
                    for (Map<String, Object> row : data) {
                        System.out.println("  行: " + row);
                    }
                }
            }
            
            // 5. 测试数据源连接
            System.out.println("\n5. 测试数据源连接");
            Map<String, Object> testConfig = Map.of(
                "url", "jdbc:mysql://localhost:3306/test",
                "username", "root",
                "password", "password",
                "driverClassName", "com.mysql.cj.jdbc.Driver"
            );
            Map<String, Object> testResult = client.testConnection(testConfig);
            System.out.println("连接测试: " + testResult);
            
            // 6. 获取表字段信息
            System.out.println("\n6. 获取表字段");
            Map<String, Object> tableFields = client.getTableFields(
                "your-data-source-id",
                "your_database",
                "your_table"
            );
            System.out.println("表字段: " + tableFields);
            
            // 7. 向量数据库操作示例
            System.out.println("\n7. 向量数据库示例");
            Map<String, Object> vectorConfig = Map.of(
                "databaseType", "milvus",
                "host", "localhost",
                "port", 19530
            );
            Map<String, Object> vectorTest = client.testVectorDatabaseConnection(vectorConfig);
            System.out.println("向量数据库连接测试: " + vectorTest);
            
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
```

## API 文档

### AutoDataSourceClient（基础客户端）

#### 构造函数

```java
public AutoDataSourceClient(String baseUrl)
public AutoDataSourceClient(String baseUrl, int timeout)
```

- `baseUrl`: AutoDataSource API 基础 URL
- `timeout`: 超时时间（秒）

#### 数据源管理方法

1. **获取外部数据源列表**
   ```java
   public Map<String, Object> getDataSources() throws IOException
   ```

2. **获取本地数据源列表**
   ```java
   public Map<String, Object> getLocalDataSources() throws IOException
   ```

3. **添加数据源**
   ```java
   public Map<String, Object> addDataSource(String dataSourceId, Map<String, Object> properties) throws IOException
   ```

4. **删除数据源**
   ```java
   public Map<String, Object> removeDataSource(String dataSourceId) throws IOException
   ```

5. **检查数据源是否存在**
   ```java
   public Map<String, Object> checkDataSourceExists(String dataSourceId) throws IOException
   ```

6. **测试数据源连接**
   ```java
   public Map<String, Object> testConnection(Map<String, Object> properties) throws IOException
   ```

#### SQL 操作方法

1. **执行 SQL 查询**
   ```java
   public Map<String, Object> executeSqlQuery(String dataSourceId, String sql) throws IOException
   public Map<String, Object> executeSqlQuery(String dataSourceId, String sql, String sqlType) throws IOException
   ```

2. **执行 SQL 更新**
   ```java
   public Map<String, Object> executeSqlUpdate(String dataSourceId, String sql) throws IOException
   public Map<String, Object> executeSqlUpdate(String dataSourceId, String sql, String sqlType) throws IOException
   ```

3. **获取表字段信息**
   ```java
   public Map<String, Object> getTableFields(String dataSourceId, String tableSchema, String tableName) throws IOException
   ```

4. **获取表列表**
   ```java
   public Map<String, Object> getTableList(String dataSourceId, String tableSchema) throws IOException
   ```

#### SQL 转换方法

1. **转换 SQL 语句**
   ```java
   public Map<String, Object> transformSql(String sql, String sourceType, String targetType) throws IOException
   ```

#### 数据集管理方法

1. **获取数据集列表**
   ```java
   public Map<String, Object> getDataSets() throws IOException
   ```

2. **根据 UUID 获取数据**
   ```java
   public Map<String, Object> getDataByUuid(String uuidAutoData) throws IOException
   ```

3. **根据数据源 ID 预览数据集**
   ```java
   public Map<String, Object> previewDataSetsByDataSourceId(String dataSourceId) throws IOException
   ```

#### 向量数据库方法

1. **测试向量数据库连接**
   ```java
   public Map<String, Object> testVectorDatabaseConnection(Map<String, Object> properties) throws IOException
   ```

2. **连接到向量数据库**
   ```java
   public Map<String, Object> connectToVectorDatabase(Map<String, Object> properties) throws IOException
   ```

3. **断开向量数据库连接**
   ```java
   public Map<String, Object> disconnectFromVectorDatabase(String connectionId, String databaseType) throws IOException
   ```

4. **创建向量集合**
   ```java
   public Map<String, Object> createVectorCollection(String connectionId, String databaseType, String collectionName, int dimension, String indexType) throws IOException
   ```

5. **删除向量集合**
   ```java
   public Map<String, Object> dropVectorCollection(String connectionId, String databaseType, String collectionName) throws IOException
   ```

6. **获取向量集合列表**
   ```java
   public Map<String, Object> listVectorCollections(String connectionId, String databaseType) throws IOException
   ```

7. **获取向量集合信息**
   ```java
   public Map<String, Object> getVectorCollectionInfo(String connectionId, String databaseType, String collectionName) throws IOException
   ```

8. **插入向量数据**
   ```java
   public Map<String, Object> insertVectors(String connectionId, String databaseType, String collectionName, List<float[]> vectors, List<Map<String, Object>> metadata) throws IOException
   ```

9. **搜索向量数据**
   ```java
   public Map<String, Object> searchVectors(String connectionId, String databaseType, String collectionName, float[] queryVector, int topK, String filter) throws IOException
   ```

#### 其他方法

1. **获取 OSS 文件列表**
   ```java
   public Map<String, Object> getOssFiles(String prefix) throws IOException
   ```

2. **获取公共文档列表**
   ```java
   public Map<String, Object> getPublicDocsList() throws IOException
   ```

3. **获取公共文档内容**
   ```java
   public Map<String, Object> getPublicDocContent(String fileName) throws IOException
   ```

4. **获取数据抽取记录**
   ```java
   public Map<String, Object> getExtractRecords() throws IOException
   ```

### AutoDataSourceUnifiedClient（统一客户端）

#### 构造函数

```java
public AutoDataSourceUnifiedClient(String baseUrl)
public AutoDataSourceUnifiedClient(String baseUrl, DataSourceMode mode)
```

- `baseUrl`: AutoDataSource API 基础 URL
- `mode`: 数据获取模式（HTTP_API、SPARK_JDBC、AUTO）

#### 主要方法

1. **执行查询**
   ```java
   public QueryResult executeQuery(String dataSourceId, String sql) throws IOException
   public QueryResult executeQuery(String dataSourceId, String sql, String sqlType) throws IOException
   ```

2. **执行更新**
   ```java
   public UpdateResult executeUpdate(String dataSourceId, String sql) throws IOException
   public UpdateResult executeUpdate(String dataSourceId, String sql, String sqlType) throws IOException
   ```

3. **获取实际使用的模式**
   ```java
   public DataSourceMode getEffectiveMode()
   ```

4. **获取底层客户端**
   ```java
   public AutoDataSourceClient getHttpClient()
   public AutoDataSourceSparkClient getSparkClient()
   ```

## 依赖项

SDK 依赖以下库（Maven 会自动处理）：

- OkHttp 4.9.3 - HTTP 客户端
- Jackson Databind 2.13.5 - JSON 处理
- SLF4J API 1.7.36 - 日志接口
- Spark SQL 3.3.0（可选，provided）

## 错误处理

SDK 会抛出 `IOException` 异常，建议在使用时进行适当的异常处理。

```java
try {
    Map<String, Object> result = client.executeSqlQuery(dataSourceId, sql);
    if (Boolean.TRUE.equals(result.get("success"))) {
        // 处理成功响应
    } else {
        // 处理业务错误
        System.out.println("错误: " + result.get("message"));
    }
} catch (IOException e) {
    // 处理网络或 IO 错误
    e.printStackTrace();
}
```

## 版本历史

- v1.0.0: 完整功能版本
  - 支持所有数据源管理 API
  - 支持 SQL 查询和更新
  - 支持向量数据库操作
  - 支持 SQL 转换功能
  - 提供统一客户端接口

## 许可证

MIT License
