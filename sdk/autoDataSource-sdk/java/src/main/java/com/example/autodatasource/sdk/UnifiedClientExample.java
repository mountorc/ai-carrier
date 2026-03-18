package com.example.autodatasource.sdk;

import java.io.IOException;
import java.util.Properties;

/**
 * AutoDataSource 统一客户端使用示例
 * 
 * 这个示例展示了如何使用统一客户端在不同模式下获取数据
 */
public class UnifiedClientExample {

    public static void main(String[] args) throws IOException {
        String baseUrl = "http://localhost:8080";
        String dataSourceId = "your-data-source-id";
        
        System.out.println("=== AutoDataSource 统一客户端示例 ===\n");
        
        // 示例1：使用 HTTP API 模式（默认）
        System.out.println("--- 示例1：HTTP API 模式 ---");
        httpApiModeExample(baseUrl, dataSourceId);
        
        System.out.println("\n" + "=".repeat(60) + "\n");
        
        // 示例2：使用 Spark JDBC 模式
        System.out.println("--- 示例2：Spark JDBC 模式 ---");
        sparkJdbcModeExample(baseUrl, dataSourceId);
        
        System.out.println("\n" + "=".repeat(60) + "\n");
        
        // 示例3：使用自动模式
        System.out.println("--- 示例3：自动模式 ---");
        autoModeExample(baseUrl, dataSourceId);
    }
    
    /**
     * 示例1：HTTP API 模式
     * 这是默认模式，通过 AutoDataSource API 获取数据
     */
    private static void httpApiModeExample(String baseUrl, String dataSourceId) throws IOException {
        // 创建客户端（默认使用 HTTP_API 模式）
        AutoDataSourceUnifiedClient client = new AutoDataSourceUnifiedClient(baseUrl);
        
        System.out.println("当前模式: " + client.getEffectiveMode());
        
        // 执行查询
        String sql = "SELECT * FROM users LIMIT 10";
        AutoDataSourceUnifiedClient.QueryResult result = client.executeQuery(dataSourceId, sql);
        
        System.out.println("查询成功: " + result.isSuccess());
        System.out.println("消息: " + result.getMessage());
        System.out.println("数据行数: " + result.getTotal());
        
        if (result.getDataList() != null) {
            System.out.println("\n前3条数据:");
            result.getDataList().stream().limit(3).forEach(row -> {
                System.out.println("  " + row);
            });
        }
        
        // 执行更新
        String updateSql = "UPDATE users SET status = 'active' WHERE id = 1";
        AutoDataSourceUnifiedClient.UpdateResult updateResult = client.executeUpdate(dataSourceId, updateSql);
        
        System.out.println("\n更新成功: " + updateResult.isSuccess());
        System.out.println("影响行数: " + updateResult.getAffectedRows());
    }
    
    /**
     * 示例2：Spark JDBC 模式
     * 使用本地 Spark 的 JDBC 能力直接连接数据库
     */
    private static void sparkJdbcModeExample(String baseUrl, String dataSourceId) throws IOException {
        // 创建客户端，指定使用 SPARK_JDBC 模式
        AutoDataSourceUnifiedClient client = new AutoDataSourceUnifiedClient(baseUrl, DataSourceMode.SPARK_JDBC);
        
        System.out.println("当前模式: " + client.getEffectiveMode());
        
        // 获取连接配置
        String sql = "SELECT * FROM users WHERE age > 18";
        AutoDataSourceUnifiedClient.QueryResult result = client.executeQuery(dataSourceId, sql);
        
        System.out.println("成功: " + result.isSuccess());
        System.out.println("消息: " + result.getMessage());
        System.out.println("提示: " + result.getHint());
        
        // 获取 JDBC 连接信息
        String jdbcUrl = result.getJdbcUrl();
        Properties jdbcProps = result.getJdbcProperties();
        
        System.out.println("\nJDBC URL: " + jdbcUrl);
        System.out.println("JDBC Properties:");
        jdbcProps.forEach((key, value) -> {
            if (!"password".equals(key)) {
                System.out.println("  " + key + "=" + value);
            }
        });
        
        // 展示如何在 Spark 中使用
        System.out.println("\n--- 在 Spark 中使用的代码示例 ---");
        System.out.println("SparkSession spark = SparkSession.builder()");
        System.out.println("    .appName(\"MyApp\")");
        System.out.println("    .getOrCreate();");
        System.out.println("");
        System.out.println("Dataset<Row> df = spark.read()");
        System.out.println("    .format(\"jdbc\")");
        System.out.println("    .option(\"url\", \"" + jdbcUrl + "\")");
        System.out.println("    .option(\"query\", \"(" + sql + ") AS tmp\")");
        System.out.println("    .option(\"user\", \"" + jdbcProps.getProperty("user") + "\")");
        System.out.println("    .option(\"password\", \"***\")");
        System.out.println("    .option(\"driver\", \"" + jdbcProps.getProperty("driver") + "\")");
        System.out.println("    .load();");
        System.out.println("");
        System.out.println("df.show();");
    }
    
    /**
     * 示例3：自动模式
     * 优先使用 Spark（如果可用），否则回退到 HTTP API
     */
    private static void autoModeExample(String baseUrl, String dataSourceId) throws IOException {
        // 创建客户端，使用 AUTO 模式
        AutoDataSourceUnifiedClient client = new AutoDataSourceUnifiedClient(baseUrl, DataSourceMode.AUTO);
        
        System.out.println("配置模式: AUTO");
        System.out.println("实际使用模式: " + client.getEffectiveMode());
        
        // 执行查询
        String sql = "SELECT * FROM products LIMIT 5";
        AutoDataSourceUnifiedClient.QueryResult result = client.executeQuery(dataSourceId, sql);
        
        System.out.println("\n查询成功: " + result.isSuccess());
        System.out.println("使用模式: " + result.getMode());
        System.out.println("消息: " + result.getMessage());
        
        // 根据不同模式处理结果
        if (result.getMode() == DataSourceMode.HTTP_API) {
            System.out.println("\n通过 HTTP API 获取的数据:");
            if (result.getDataList() != null) {
                result.getDataList().forEach(row -> System.out.println("  " + row));
            }
        } else {
            System.out.println("\nSpark JDBC 连接信息已获取");
            System.out.println("JDBC URL: " + result.getJdbcUrl());
        }
    }
}
