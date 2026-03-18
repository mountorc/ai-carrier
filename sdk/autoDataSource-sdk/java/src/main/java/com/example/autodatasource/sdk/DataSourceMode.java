package com.example.autodatasource.sdk;

/**
 * 数据源获取模式枚举
 */
public enum DataSourceMode {
    /**
     * HTTP API 模式（默认）：通过 AutoDataSource API 获取数据
     */
    HTTP_API,
    
    /**
     * Spark JDBC 模式：使用本地 Spark 的 JDBC 能力直接连接数据库
     */
    SPARK_JDBC,
    
    /**
     * 自动模式：优先使用 Spark（如果可用），否则回退到 HTTP API
     */
    AUTO
}
