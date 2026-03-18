package com.example.autodatasource.sdk;

import java.util.Map;

/**
 * 测试预览数据集方法
 */
public class TestPreviewDataSets {
    public static void main(String[] args) {
        // 创建客户端
        AutoDataSourceClient client = new AutoDataSourceClient("http://localhost:8080/autoDataSource");

        try {
            // 测试根据数据源ID预览数据集
            System.out.println("=== 测试: 根据数据源ID预览数据集 ===");
            String dataSourceId = "90a49r2l313l4243robp5lbqf678vfn10719";
            Map<String, Object> previewResponse = client.previewDataSetsByDataSourceId(dataSourceId);
            System.out.println("响应: " + previewResponse);
            System.out.println();

        } catch (Exception e) {
            System.err.println("调用 API 失败: " + e.getMessage());
            e.printStackTrace();
        }
    }
}