package com.example.autodatasource.sdk;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.Map;
import java.util.Properties;

/**
 * AutoDataSource Spark 集成客户端
 * 提供与 Spark 集成的便捷方法，利用 Spark 的 JDBC 能力直接获取数据
 * 注意：使用此类需要项目中包含 Spark SQL 依赖
 */
public class AutoDataSourceSparkClient {

    private static final Logger log = LoggerFactory.getLogger(AutoDataSourceSparkClient.class);
    
    private final AutoDataSourceClient client;
    private final ObjectMapper objectMapper;
    
    /**
     * 构造函数
     * @param baseUrl AutoDataSource API 基础 URL
     */
    public AutoDataSourceSparkClient(String baseUrl) {
        this.client = new AutoDataSourceClient(baseUrl);
        this.objectMapper = new ObjectMapper();
    }
    
    /**
     * 构造函数
     * @param client AutoDataSourceClient 实例
     */
    public AutoDataSourceSparkClient(AutoDataSourceClient client) {
        this.client = client;
        this.objectMapper = new ObjectMapper();
    }
    
    /**
     * 获取数据源配置（用于 Spark JDBC 连接）
     * @param dataSourceId 数据源ID
     * @return 数据源配置信息
     * @throws IOException IO 异常
     */
    public Map<String, Object> getDataSourceConfig(String dataSourceId) throws IOException {
        Map<String, Object> dataSources = client.getDataSources();
        if (dataSources != null && dataSources.containsKey("list")) {
            java.util.List<?> list = (java.util.List<?>) dataSources.get("list");
            for (Object item : list) {
                if (item instanceof Map) {
                    Map<String, Object> ds = (Map<String, Object>) item;
                    if (dataSourceId.equals(ds.get("id")) || dataSourceId.equals(ds.get("uuid_dataSource"))) {
                        return ds;
                    }
                }
            }
        }
        throw new IOException("DataSource not found: " + dataSourceId);
    }
    
    /**
     * 构建 Spark JDBC 连接 URL
     * @param dataSourceConfig 数据源配置
     * @return JDBC URL
     */
    public String buildJdbcUrl(Map<String, Object> dataSourceConfig) {
        String databaseType = (String) dataSourceConfig.get("databaseType");
        String host = (String) dataSourceConfig.get("host");
        Integer port = (Integer) dataSourceConfig.get("port");
        String database = (String) dataSourceConfig.get("database");
        
        if (databaseType == null) {
            databaseType = "mysql";
        }
        
        switch (databaseType.toLowerCase()) {
            case "mysql":
                return "jdbc:mysql://" + host + ":" + (port != null ? port : 3306) + "/" + database;
            case "postgresql":
                return "jdbc:postgresql://" + host + ":" + (port != null ? port : 5432) + "/" + database;
            case "oracle":
                return "jdbc:oracle:thin:@" + host + ":" + (port != null ? port : 1521) + ":" + database;
            case "sqlserver":
                return "jdbc:sqlserver://" + host + ":" + (port != null ? port : 1433) + ";databaseName=" + database;
            default:
                throw new IllegalArgumentException("Unsupported database type: " + databaseType);
        }
    }
    
    /**
     * 构建 Spark JDBC 连接属性
     * @param dataSourceConfig 数据源配置
     * @return Properties 对象
     */
    public Properties buildJdbcProperties(Map<String, Object> dataSourceConfig) {
        Properties props = new Properties();
        String user = (String) dataSourceConfig.get("username");
        String password = (String) dataSourceConfig.get("password");
        String databaseType = (String) dataSourceConfig.get("databaseType");
        
        if (user != null) {
            props.put("user", user);
        }
        if (password != null) {
            props.put("password", password);
        }
        
        // 设置合适的驱动类名
        if (databaseType != null) {
            switch (databaseType.toLowerCase()) {
                case "mysql":
                    props.put("driver", "com.mysql.cj.jdbc.Driver");
                    break;
                case "postgresql":
                    props.put("driver", "org.postgresql.Driver");
                    break;
                case "oracle":
                    props.put("driver", "oracle.jdbc.OracleDriver");
                    break;
                case "sqlserver":
                    props.put("driver", "com.microsoft.sqlserver.jdbc.SQLServerDriver");
                    break;
            }
        }
        
        return props;
    }
    
    /**
     * 获取底层的 AutoDataSourceClient
     * @return AutoDataSourceClient 实例
     */
    public AutoDataSourceClient getClient() {
        return client;
    }
    
    /**
     * 获取 Java SDK 中的查询结果（备用方案，当 Spark 不可用时使用）
     * @param dataSourceId 数据源ID
     * @param sql SQL语句
     * @return 查询结果
     * @throws IOException IO 异常
     */
    public Map<String, Object> executeQueryViaSdk(String dataSourceId, String sql) throws IOException {
        return client.executeSqlQuery(dataSourceId, sql);
    }
}
