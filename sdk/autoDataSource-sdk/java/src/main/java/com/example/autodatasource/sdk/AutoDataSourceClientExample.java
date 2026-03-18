package com.example.autodatasource.sdk;

import java.io.IOException;
import java.util.Map;

/**
 * AutoDataSource Java SDK 使用示例
 */
public class AutoDataSourceClientExample {

    public static void main(String[] args) {
        // 初始化客户端
        String baseUrl = "http://localhost:8080/autoDataSource";
        AutoDataSourceClient client = new AutoDataSourceClient(baseUrl);

        try {
            // 示例 1: 获取数据源列表
            System.out.println("=== 示例 1: 获取数据源列表 ===");
            Map<String, Object> dataSourcesResponse = client.getDataSources();
            System.out.println("响应: " + dataSourcesResponse);

            // 示例 2: 转换 SQL 语句
            System.out.println("\n=== 示例 2: 转换 SQL 语句 ===");
            String sql = "SELECT * FROM users LIMIT 10";
            Map<String, Object> transformResponse = client.transformSql(sql, "mysql", "oracle");
            System.out.println("原始 SQL: " + sql);
            System.out.println("转换后 SQL: " + ((Map<?, ?>) transformResponse.get("data")).get("transformedSql"));

            // 示例 3: 获取数据集列表
            System.out.println("\n=== 示例 3: 获取数据集列表 ===");
            Map<String, Object> dataItemsResponse = client.getDataSets();
            System.out.println("响应: " + dataItemsResponse);

            // 示例 4: 获取 OSS 文件列表
            System.out.println("\n=== 示例 4: 获取 OSS 文件列表 ===");
            Map<String, Object> ossFilesResponse = client.getOssFiles(null);
            System.out.println("响应: " + ossFilesResponse);

            // 示例 5: 获取公共文档列表
            System.out.println("\n=== 示例 5: 获取公共文档列表 ===");
            Map<String, Object> docsListResponse = client.getPublicDocsList();
            System.out.println("响应: " + docsListResponse);

        } catch (IOException e) {
            System.err.println("调用 API 失败: " + e.getMessage());
            e.printStackTrace();
        }
    }
}
