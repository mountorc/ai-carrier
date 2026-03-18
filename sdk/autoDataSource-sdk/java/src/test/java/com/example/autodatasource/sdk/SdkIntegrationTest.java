package com.example.autodatasource.sdk;

import org.junit.Test;
import static org.junit.Assert.*;
import java.util.Map;

public class SdkIntegrationTest {

    private static final String BASE_URL = "http://localhost:8080/autoDataSource";

    @Test
    public void testSdkInitialization() {
        System.out.println("=== 测试 SDK 初始化 ===");
        AutoDataSourceClient client = new AutoDataSourceClient(BASE_URL);
        assertNotNull("SDK 客户端初始化失败", client);
        System.out.println("✓ SDK 初始化成功");
    }

    @Test
    public void testGetPublicDocsList() throws Exception {
        System.out.println("\n=== 测试获取公共文档列表 ===");
        AutoDataSourceClient client = new AutoDataSourceClient(BASE_URL);
        Map<String, Object> result = client.getPublicDocsList();
        assertNotNull("响应不应为空", result);
        System.out.println("✓ 获取公共文档列表成功");
        System.out.println("  响应: " + result);
    }

    @Test
    public void testGetDataSources() throws Exception {
        System.out.println("\n=== 测试获取数据源列表 ===");
        AutoDataSourceClient client = new AutoDataSourceClient(BASE_URL);
        Map<String, Object> result = client.getDataSources();
        assertNotNull("响应不应为空", result);
        System.out.println("✓ 获取数据源列表成功");
        System.out.println("  响应: " + result);
    }
}
