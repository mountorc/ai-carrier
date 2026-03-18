package com.example.autodatasource.sdk;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.Map;
import java.util.Properties;

/**
 * AutoDataSource 统一客户端
 * 支持通过配置选择使用 HTTP API 或 Spark JDBC 方式获取数据
 * 
 * 使用示例：
 * <pre>
 * // 方式1：使用 HTTP API（默认）
 * AutoDataSourceUnifiedClient client = new AutoDataSourceUnifiedClient("http://localhost:8080");
 * 
 * // 方式2：使用 Spark JDBC
 * AutoDataSourceUnifiedClient client = new AutoDataSourceUnifiedClient("http://localhost:8080", DataSourceMode.SPARK_JDBC);
 * 
 * // 方式3：自动选择
 * AutoDataSourceUnifiedClient client = new AutoDataSourceUnifiedClient("http://localhost:8080", DataSourceMode.AUTO);
 * </pre>
 */
public class AutoDataSourceUnifiedClient {

    private static final Logger log = LoggerFactory.getLogger(AutoDataSourceUnifiedClient.class);
    
    private final String baseUrl;
    private final DataSourceMode mode;
    private final AutoDataSourceClient httpClient;
    private final AutoDataSourceSparkClient sparkClient;
    private final boolean sparkAvailable;
    
    /**
     * 构造函数（默认使用 HTTP API 模式）
     * @param baseUrl AutoDataSource API 基础 URL
     */
    public AutoDataSourceUnifiedClient(String baseUrl) {
        this(baseUrl, DataSourceMode.HTTP_API);
    }
    
    /**
     * 构造函数（指定模式）
     * @param baseUrl AutoDataSource API 基础 URL
     * @param mode 数据获取模式
     */
    public AutoDataSourceUnifiedClient(String baseUrl, DataSourceMode mode) {
        this.baseUrl = baseUrl;
        this.mode = mode;
        this.httpClient = new AutoDataSourceClient(baseUrl);
        this.sparkClient = new AutoDataSourceSparkClient(httpClient);
        this.sparkAvailable = checkSparkAvailable();
        
        log.info("AutoDataSourceUnifiedClient initialized, mode: {}, sparkAvailable: {}", mode, sparkAvailable);
    }
    
    /**
     * 检查 Spark 是否可用
     * @return Spark 是否可用
     */
    private boolean checkSparkAvailable() {
        try {
            Class.forName("org.apache.spark.sql.SparkSession");
            return true;
        } catch (ClassNotFoundException e) {
            log.debug("Spark not available in classpath");
            return false;
        }
    }
    
    /**
     * 获取当前使用的模式
     * @return 实际使用的数据获取模式
     */
    public DataSourceMode getEffectiveMode() {
        if (mode == DataSourceMode.AUTO) {
            return sparkAvailable ? DataSourceMode.SPARK_JDBC : DataSourceMode.HTTP_API;
        }
        return mode;
    }
    
    /**
     * 执行 SQL 查询
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @return 查询结果
     * @throws IOException IO 异常
     */
    public QueryResult executeQuery(String dataSourceId, String sql) throws IOException {
        return executeQuery(dataSourceId, sql, null);
    }
    
    /**
     * 执行 SQL 查询（带 SQL 类型）
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @param sqlType SQL类型（可选，用于 SQL 转换）
     * @return 查询结果
     * @throws IOException IO 异常
     */
    public QueryResult executeQuery(String dataSourceId, String sql, String sqlType) throws IOException {
        DataSourceMode effectiveMode = getEffectiveMode();
        log.info("Executing query, mode: {}, dataSourceId: {}, sql: {}", effectiveMode, dataSourceId, sql);
        
        if (effectiveMode == DataSourceMode.SPARK_JDBC) {
            return executeQueryWithSpark(dataSourceId, sql);
        } else {
            return executeQueryWithHttp(dataSourceId, sql, sqlType);
        }
    }
    
    /**
     * 使用 HTTP API 执行查询
     */
    private QueryResult executeQueryWithHttp(String dataSourceId, String sql, String sqlType) throws IOException {
        Map<String, Object> response = httpClient.executeSqlQuery(dataSourceId, sql, sqlType);
        return QueryResult.fromHttpResponse(response);
    }
    
    /**
     * 使用 Spark JDBC 执行查询
     * 注意：此方法需要用户自行使用获取的连接参数在 Spark 中执行
     * 这里提供连接配置信息
     */
    private QueryResult executeQueryWithSpark(String dataSourceId, String sql) throws IOException {
        Map<String, Object> dataSourceConfig = sparkClient.getDataSourceConfig(dataSourceId);
        String jdbcUrl = sparkClient.buildJdbcUrl(dataSourceConfig);
        Properties jdbcProps = sparkClient.buildJdbcProperties(dataSourceConfig);
        
        QueryResult result = new QueryResult();
        result.setSuccess(true);
        result.setMode(DataSourceMode.SPARK_JDBC);
        result.setJdbcUrl(jdbcUrl);
        result.setJdbcProperties(jdbcProps);
        result.setDataSourceConfig(dataSourceConfig);
        result.setMessage("Spark JDBC connection info provided. Use these parameters in your Spark code.");
        result.setHint("Use spark.read().jdbc(jdbcUrl, \"(" + sql + ") AS tmp\", jdbcProps) to execute query");
        
        return result;
    }
    
    /**
     * 执行 SQL 更新（INSERT, UPDATE, DELETE）
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @return 更新结果
     * @throws IOException IO 异常
     */
    public UpdateResult executeUpdate(String dataSourceId, String sql) throws IOException {
        return executeUpdate(dataSourceId, sql, null);
    }
    
    /**
     * 执行 SQL 更新（带 SQL 类型）
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @param sqlType SQL类型（可选）
     * @return 更新结果
     * @throws IOException IO 异常
     */
    public UpdateResult executeUpdate(String dataSourceId, String sql, String sqlType) throws IOException {
        DataSourceMode effectiveMode = getEffectiveMode();
        log.info("Executing update, mode: {}, dataSourceId: {}, sql: {}", effectiveMode, dataSourceId, sql);
        
        if (effectiveMode == DataSourceMode.SPARK_JDBC) {
            return executeUpdateWithSpark(dataSourceId, sql);
        } else {
            return executeUpdateWithHttp(dataSourceId, sql, sqlType);
        }
    }
    
    /**
     * 使用 HTTP API 执行更新
     */
    private UpdateResult executeUpdateWithHttp(String dataSourceId, String sql, String sqlType) throws IOException {
        Map<String, Object> response = httpClient.executeSqlUpdate(dataSourceId, sql, sqlType);
        return UpdateResult.fromHttpResponse(response);
    }
    
    /**
     * 使用 Spark JDBC 执行更新
     */
    private UpdateResult executeUpdateWithSpark(String dataSourceId, String sql) throws IOException {
        Map<String, Object> dataSourceConfig = sparkClient.getDataSourceConfig(dataSourceId);
        String jdbcUrl = sparkClient.buildJdbcUrl(dataSourceConfig);
        Properties jdbcProps = sparkClient.buildJdbcProperties(dataSourceConfig);
        
        UpdateResult result = new UpdateResult();
        result.setSuccess(true);
        result.setMode(DataSourceMode.SPARK_JDBC);
        result.setJdbcUrl(jdbcUrl);
        result.setJdbcProperties(jdbcProps);
        result.setDataSourceConfig(dataSourceConfig);
        result.setMessage("Spark JDBC connection info provided. Use these parameters in your Spark code.");
        result.setHint("Note: Spark SQL is primarily for queries. For updates, consider using JDBC directly.");
        
        return result;
    }
    
    /**
     * 获取表字段信息
     * @param dataSourceId 数据源ID
     * @param tableSchema 数据库名称
     * @param tableName 表名称
     * @return 表字段信息
     * @throws IOException IO 异常
     */
    public Map<String, Object> getTableFields(String dataSourceId, String tableSchema, String tableName) throws IOException {
        return httpClient.getTableFields(dataSourceId, tableSchema, tableName);
    }
    
    /**
     * 获取数据库表列表
     * @param dataSourceId 数据源ID
     * @param tableSchema 数据库名称（可选）
     * @return 表列表
     * @throws IOException IO 异常
     */
    public Map<String, Object> getTableList(String dataSourceId, String tableSchema) throws IOException {
        return httpClient.getTableList(dataSourceId, tableSchema);
    }
    
    /**
     * 获取底层 HTTP 客户端
     * @return AutoDataSourceClient 实例
     */
    public AutoDataSourceClient getHttpClient() {
        return httpClient;
    }
    
    /**
     * 获取底层 Spark 客户端
     * @return AutoDataSourceSparkClient 实例
     */
    public AutoDataSourceSparkClient getSparkClient() {
        return sparkClient;
    }
    
    /**
     * 查询结果类
     */
    public static class QueryResult {
        private boolean success;
        private String message;
        private DataSourceMode mode;
        private java.util.List<Map<String, Object>> dataList;
        private Integer total;
        private String jdbcUrl;
        private Properties jdbcProperties;
        private Map<String, Object> dataSourceConfig;
        private String hint;
        
        public static QueryResult fromHttpResponse(Map<String, Object> response) {
            QueryResult result = new QueryResult();
            result.setSuccess(Boolean.TRUE.equals(response.get("success")));
            result.setMessage((String) response.get("message"));
            result.setMode(DataSourceMode.HTTP_API);
            result.setDataList((java.util.List<Map<String, Object>>) response.get("dataList"));
            Object total = response.get("total");
            result.setTotal(total != null ? ((Number) total).intValue() : null);
            return result;
        }
        
        // Getters and Setters
        public boolean isSuccess() { return success; }
        public void setSuccess(boolean success) { this.success = success; }
        public String getMessage() { return message; }
        public void setMessage(String message) { this.message = message; }
        public DataSourceMode getMode() { return mode; }
        public void setMode(DataSourceMode mode) { this.mode = mode; }
        public java.util.List<Map<String, Object>> getDataList() { return dataList; }
        public void setDataList(java.util.List<Map<String, Object>> dataList) { this.dataList = dataList; }
        public Integer getTotal() { return total; }
        public void setTotal(Integer total) { this.total = total; }
        public String getJdbcUrl() { return jdbcUrl; }
        public void setJdbcUrl(String jdbcUrl) { this.jdbcUrl = jdbcUrl; }
        public Properties getJdbcProperties() { return jdbcProperties; }
        public void setJdbcProperties(Properties jdbcProperties) { this.jdbcProperties = jdbcProperties; }
        public Map<String, Object> getDataSourceConfig() { return dataSourceConfig; }
        public void setDataSourceConfig(Map<String, Object> dataSourceConfig) { this.dataSourceConfig = dataSourceConfig; }
        public String getHint() { return hint; }
        public void setHint(String hint) { this.hint = hint; }
    }
    
    /**
     * 更新结果类
     */
    public static class UpdateResult {
        private boolean success;
        private String message;
        private DataSourceMode mode;
        private Integer affectedRows;
        private String jdbcUrl;
        private Properties jdbcProperties;
        private Map<String, Object> dataSourceConfig;
        private String hint;
        
        public static UpdateResult fromHttpResponse(Map<String, Object> response) {
            UpdateResult result = new UpdateResult();
            result.setSuccess(Boolean.TRUE.equals(response.get("success")));
            result.setMessage((String) response.get("message"));
            result.setMode(DataSourceMode.HTTP_API);
            Object affectedRows = response.get("affectedRows");
            result.setAffectedRows(affectedRows != null ? ((Number) affectedRows).intValue() : null);
            return result;
        }
        
        // Getters and Setters
        public boolean isSuccess() { return success; }
        public void setSuccess(boolean success) { this.success = success; }
        public String getMessage() { return message; }
        public void setMessage(String message) { this.message = message; }
        public DataSourceMode getMode() { return mode; }
        public void setMode(DataSourceMode mode) { this.mode = mode; }
        public Integer getAffectedRows() { return affectedRows; }
        public void setAffectedRows(Integer affectedRows) { this.affectedRows = affectedRows; }
        public String getJdbcUrl() { return jdbcUrl; }
        public void setJdbcUrl(String jdbcUrl) { this.jdbcUrl = jdbcUrl; }
        public Properties getJdbcProperties() { return jdbcProperties; }
        public void setJdbcProperties(Properties jdbcProperties) { this.jdbcProperties = jdbcProperties; }
        public Map<String, Object> getDataSourceConfig() { return dataSourceConfig; }
        public void setDataSourceConfig(Map<String, Object> dataSourceConfig) { this.dataSourceConfig = dataSourceConfig; }
        public String getHint() { return hint; }
        public void setHint(String hint) { this.hint = hint; }
    }
}
