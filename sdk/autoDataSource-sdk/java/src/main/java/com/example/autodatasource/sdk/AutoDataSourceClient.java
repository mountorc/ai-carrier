package com.example.autodatasource.sdk;

import com.fasterxml.jackson.databind.ObjectMapper;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.List;
import java.util.Map;

/**
 * AutoDataSource Java SDK 客户端
 */
public class AutoDataSourceClient {

    private static final Logger log = LoggerFactory.getLogger(AutoDataSourceClient.class);
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");

    private final String baseUrl;
    private final OkHttpClient client;
    private final ObjectMapper objectMapper;

    /**
     * 构造函数
     * @param baseUrl AutoDataSource API 基础 URL
     */
    public AutoDataSourceClient(String baseUrl) {
        this(baseUrl, 30); // 默认 30 秒超时
    }

    /**
     * 构造函数
     * @param baseUrl AutoDataSource API 基础 URL
     * @param timeout 超时时间（秒）
     */
    public AutoDataSourceClient(String baseUrl, int timeout) {
        this.baseUrl = baseUrl.endsWith("/") ? baseUrl.substring(0, baseUrl.length() - 1) : baseUrl;
        this.client = new OkHttpClient.Builder()
                .connectTimeout(timeout, java.util.concurrent.TimeUnit.SECONDS)
                .readTimeout(timeout, java.util.concurrent.TimeUnit.SECONDS)
                .writeTimeout(timeout, java.util.concurrent.TimeUnit.SECONDS)
                .build();
        this.objectMapper = new ObjectMapper();
    }

    /**
     * 获取数据源列表
     * @return 数据源列表响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getDataSources() throws IOException {
        String url = baseUrl + "/api/data-sources/external/list";
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 转换 SQL 语句
     * @param sql SQL 语句
     * @param sourceType 源数据库类型
     * @param targetType 目标数据库类型
     * @return SQL 转换响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> transformSql(String sql, String sourceType, String targetType) throws IOException {
        String url = baseUrl + "/api/sql/transform";

        Map<String, Object> requestBody = Map.of(
                "sql", sql,
                "sourceType", sourceType,
                "targetType", targetType
        );

        String json = objectMapper.writeValueAsString(requestBody);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取数据集列表
     * @return 数据集列表响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getDataSets() throws IOException {
        String url = baseUrl + "/api/data-sets/list";
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 根据 UUID 获取数据
     * @param uuidAutoData 数据集 UUID
     * @return 数据响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getDataByUuid(String uuidAutoData) throws IOException {
        String url = baseUrl + "/api/data-sets/data/" + uuidAutoData;
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取 OSS 文件列表
     * @param prefix 目录前缀
     * @return OSS 文件列表响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getOssFiles(String prefix) throws IOException {
        String url = baseUrl + "/api/oss/files";
        if (prefix != null && !prefix.isEmpty()) {
            url += "?prefix=" + prefix;
        }

        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取公共文档列表
     * @return 文档列表响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getPublicDocsList() throws IOException {
        String url = baseUrl + "/docs-public/list";
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取公共文档内容
     * @param fileName 文件名
     * @return 文档内容响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getPublicDocContent(String fileName) throws IOException {
        String url = baseUrl + "/docs-public/docs?fileName=" + fileName;
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 根据数据源ID预览数据集
     * @param dataSourceId 数据源ID
     * @return 数据集预览响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> previewDataSetsByDataSourceId(String dataSourceId) throws IOException {
        String url = baseUrl + "/api/data-sets/preview/" + dataSourceId;
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取数据抽取记录
     * @return 数据抽取记录响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getExtractRecords() throws IOException {
        String url = baseUrl + "/api/extract-records/list";
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 执行 SQL 查询
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @return 查询结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> executeSqlQuery(String dataSourceId, String sql) throws IOException {
        return executeSqlQuery(dataSourceId, sql, null);
    }

    /**
     * 执行 SQL 查询（带 SQL 类型）
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @param sqlType SQL类型（可选）
     * @return 查询结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> executeSqlQuery(String dataSourceId, String sql, String sqlType) throws IOException {
        String url = baseUrl + "/api/data-sources/" + dataSourceId + "/query/sql";

        Map<String, String> requestBody = new java.util.HashMap<>();
        requestBody.put("sql", sql);
        if (sqlType != null && !sqlType.isEmpty()) {
            requestBody.put("sqlType", sqlType);
        }

        String json = objectMapper.writeValueAsString(requestBody);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 执行 SQL 更新（INSERT, UPDATE, DELETE）
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @return 更新结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> executeSqlUpdate(String dataSourceId, String sql) throws IOException {
        return executeSqlUpdate(dataSourceId, sql, null);
    }

    /**
     * 执行 SQL 更新（带 SQL 类型）
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @param sqlType SQL类型（可选）
     * @return 更新结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> executeSqlUpdate(String dataSourceId, String sql, String sqlType) throws IOException {
        String url = baseUrl + "/api/data-sources/" + dataSourceId + "/query/update";

        Map<String, String> requestBody = new java.util.HashMap<>();
        requestBody.put("sql", sql);
        if (sqlType != null && !sqlType.isEmpty()) {
            requestBody.put("sqlType", sqlType);
        }

        String json = objectMapper.writeValueAsString(requestBody);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取表字段信息
     * @param dataSourceId 数据源ID
     * @param tableSchema 数据库名称
     * @param tableName 表名称
     * @return 表字段信息响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getTableFields(String dataSourceId, String tableSchema, String tableName) throws IOException {
        String url = baseUrl + "/api/data-sources/" + dataSourceId + "/query/table-fields";

        Map<String, String> requestBody = new java.util.HashMap<>();
        requestBody.put("tableSchema", tableSchema);
        requestBody.put("tableName", tableName);

        String json = objectMapper.writeValueAsString(requestBody);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取数据库表列表
     * @param dataSourceId 数据源ID
     * @param tableSchema 数据库名称（可选）
     * @return 表列表响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getTableList(String dataSourceId, String tableSchema) throws IOException {
        String url = baseUrl + "/api/data-sources/" + dataSourceId + "/query/tables";

        Map<String, String> requestBody = new java.util.HashMap<>();
        if (tableSchema != null && !tableSchema.isEmpty()) {
            requestBody.put("tableSchema", tableSchema);
        }

        String json = objectMapper.writeValueAsString(requestBody);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取本地数据源列表
     * @return 本地数据源列表响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getLocalDataSources() throws IOException {
        String url = baseUrl + "/api/data-sources";
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 添加数据源
     * @param dataSourceId 数据源ID
     * @param properties 数据源配置
     * @return 添加结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> addDataSource(String dataSourceId, Map<String, Object> properties) throws IOException {
        String url = baseUrl + "/api/data-sources/add/" + dataSourceId;

        String json = objectMapper.writeValueAsString(properties);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 删除数据源
     * @param dataSourceId 数据源ID
     * @return 删除结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> removeDataSource(String dataSourceId) throws IOException {
        String url = baseUrl + "/api/data-sources/remove/" + dataSourceId;

        Request request = new Request.Builder()
                .url(url)
                .delete()
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 检查数据源是否存在
     * @param dataSourceId 数据源ID
     * @return 检查结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> checkDataSourceExists(String dataSourceId) throws IOException {
        String url = baseUrl + "/api/data-sources/" + dataSourceId + "/exists";
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 测试数据源连接
     * @param properties 数据源配置
     * @return 测试结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> testConnection(Map<String, Object> properties) throws IOException {
        String url = baseUrl + "/api/data-sources/test-connection";

        String json = objectMapper.writeValueAsString(properties);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 测试向量数据库连接
     * @param properties 数据源配置
     * @return 测试结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> testVectorDatabaseConnection(Map<String, Object> properties) throws IOException {
        String url = baseUrl + "/api/vector-databases/test-connection";

        String json = objectMapper.writeValueAsString(properties);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 连接到向量数据库
     * @param properties 数据源配置
     * @return 连接结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> connectToVectorDatabase(Map<String, Object> properties) throws IOException {
        String url = baseUrl + "/api/vector-databases/connect";

        String json = objectMapper.writeValueAsString(properties);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 断开向量数据库连接
     * @param connectionId 连接ID
     * @param databaseType 数据库类型
     * @return 断开结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> disconnectFromVectorDatabase(String connectionId, String databaseType) throws IOException {
        String url = baseUrl + "/api/vector-databases/disconnect/" + connectionId + "?databaseType=" + databaseType;

        Request request = new Request.Builder()
                .url(url)
                .delete()
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 创建向量数据库集合
     * @param connectionId 连接ID
     * @param databaseType 数据库类型
     * @param collectionName 集合名称
     * @param dimension 向量维度
     * @param indexType 索引类型（可选）
     * @return 创建结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> createVectorCollection(String connectionId, String databaseType, String collectionName, int dimension, String indexType) throws IOException {
        String url = baseUrl + "/api/vector-databases/collections?connectionId=" + connectionId + "&databaseType=" + databaseType + "&collectionName=" + collectionName + "&dimension=" + dimension;
        if (indexType != null && !indexType.isEmpty()) {
            url += "&indexType=" + indexType;
        }

        Request request = new Request.Builder()
                .url(url)
                .post(RequestBody.create("", JSON))
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 删除向量数据库集合
     * @param connectionId 连接ID
     * @param databaseType 数据库类型
     * @param collectionName 集合名称
     * @return 删除结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> dropVectorCollection(String connectionId, String databaseType, String collectionName) throws IOException {
        String url = baseUrl + "/api/vector-databases/collections/" + collectionName + "?connectionId=" + connectionId + "&databaseType=" + databaseType;

        Request request = new Request.Builder()
                .url(url)
                .delete()
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取向量数据库集合列表
     * @param connectionId 连接ID
     * @param databaseType 数据库类型
     * @return 集合列表响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> listVectorCollections(String connectionId, String databaseType) throws IOException {
        String url = baseUrl + "/api/vector-databases/collections?connectionId=" + connectionId + "&databaseType=" + databaseType;
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 获取向量数据库集合信息
     * @param connectionId 连接ID
     * @param databaseType 数据库类型
     * @param collectionName 集合名称
     * @return 集合信息响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> getVectorCollectionInfo(String connectionId, String databaseType, String collectionName) throws IOException {
        String url = baseUrl + "/api/vector-databases/collections/" + collectionName + "?connectionId=" + connectionId + "&databaseType=" + databaseType;
        Request request = new Request.Builder()
                .url(url)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 插入向量数据
     * @param connectionId 连接ID
     * @param databaseType 数据库类型
     * @param collectionName 集合名称
     * @param vectors 向量列表
     * @param metadata 元数据列表
     * @return 插入结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> insertVectors(String connectionId, String databaseType, String collectionName, List<float[]> vectors, List<Map<String, Object>> metadata) throws IOException {
        String url = baseUrl + "/api/vector-databases/collections/" + collectionName + "/vectors?connectionId=" + connectionId + "&databaseType=" + databaseType;

        Map<String, Object> requestBody = new java.util.HashMap<>();
        requestBody.put("vectors", vectors);
        requestBody.put("metadata", metadata);

        String json = objectMapper.writeValueAsString(requestBody);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

    /**
     * 搜索向量数据
     * @param connectionId 连接ID
     * @param databaseType 数据库类型
     * @param collectionName 集合名称
     * @param queryVector 查询向量
     * @param topK 返回结果数
     * @param filter 过滤条件（可选）
     * @return 搜索结果响应
     * @throws IOException IO 异常
     */
    public Map<String, Object> searchVectors(String connectionId, String databaseType, String collectionName, float[] queryVector, int topK, String filter) throws IOException {
        String url = baseUrl + "/api/vector-databases/collections/" + collectionName + "/search?connectionId=" + connectionId + "&databaseType=" + databaseType;

        Map<String, Object> requestBody = new java.util.HashMap<>();
        requestBody.put("queryVector", queryVector);
        requestBody.put("topK", topK);
        if (filter != null && !filter.isEmpty()) {
            requestBody.put("filter", filter);
        }

        String json = objectMapper.writeValueAsString(requestBody);
        RequestBody body = RequestBody.create(json, JSON);

        Request request = new Request.Builder()
                .url(url)
                .post(body)
                .build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new IOException("Unexpected code " + response);
            }
            String responseBody = response.body().string();
            return objectMapper.readValue(responseBody, Map.class);
        }
    }

}
