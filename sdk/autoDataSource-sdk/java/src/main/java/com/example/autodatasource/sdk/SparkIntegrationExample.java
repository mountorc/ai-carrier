package com.example.autodatasource.sdk;

import org.apache.spark.sql.Dataset;
import org.apache.spark.sql.Row;
import org.apache.spark.sql.SparkSession;

import java.io.IOException;
import java.util.Map;
import java.util.Properties;

/**
 * Spark 集成示例
 * 展示如何使用 AutoDataSourceSparkClient 配合 Spark 读取数据源数据
 * 
 * 使用前请确保项目中包含 Spark SQL 依赖：
 * &lt;dependency&gt;
 *   &lt;groupId&gt;org.apache.spark&lt;/groupId&gt;
 *   &lt;artifactId&gt;spark-sql_2.12&lt;/artifactId&gt;
 *   &lt;version&gt;3.3.0&lt;/version&gt;
 * &lt;/dependency&gt;
 */
public class SparkIntegrationExample {

    public static void main(String[] args) throws IOException {
        // 1. 创建 AutoDataSourceSparkClient
        AutoDataSourceSparkClient sparkClient = new AutoDataSourceSparkClient("http://localhost:8080");
        
        // 2. 获取数据源配置
        String dataSourceId = "your-data-source-id";
        Map<String, Object> dataSourceConfig = sparkClient.getDataSourceConfig(dataSourceId);
        
        // 3. 构建 JDBC 连接参数
        String jdbcUrl = sparkClient.buildJdbcUrl(dataSourceConfig);
        Properties jdbcProps = sparkClient.buildJdbcProperties(dataSourceConfig);
        
        // 4. 创建 SparkSession
        SparkSession spark = SparkSession.builder()
                .appName("AutoDataSource Spark Example")
                .master("local[*]")
                .getOrCreate();
        
        // 方式一：读取整个表
        System.out.println("=== 读取整个表 ===");
        Dataset<Row> tableData = spark.read()
                .jdbc(jdbcUrl, "your_table_name", jdbcProps);
        
        tableData.show();
        
        // 方式二：使用自定义 SQL 查询
        System.out.println("\n=== 使用自定义 SQL 查询 ===");
        Dataset<Row> queryData = spark.read()
                .format("jdbc")
                .option("url", jdbcUrl)
                .option("query", "SELECT * FROM your_table_name WHERE id > 100")
                .option("user", jdbcProps.getProperty("user"))
                .option("password", jdbcProps.getProperty("password"))
                .option("driver", jdbcProps.getProperty("driver"))
                .load();
        
        queryData.show();
        
        // 方式三：使用分区读取（大数据量优化）
        System.out.println("\n=== 分区读取（大数据量优化） ===");
        Dataset<Row> partitionedData = spark.read()
                .jdbc(jdbcUrl, "your_table_name", 
                       "id", 1, 1000, 10, jdbcProps);
        
        partitionedData.show();
        
        // 方式四：写入数据到数据库
        System.out.println("\n=== 写入数据到数据库 ===");
        Dataset<Row> dataToWrite = spark.read().json("path/to/data.json");
        dataToWrite.write()
                .mode("append")
                .jdbc(jdbcUrl, "target_table", jdbcProps);
        
        // 5. 关闭 SparkSession
        spark.stop();
    }
}
