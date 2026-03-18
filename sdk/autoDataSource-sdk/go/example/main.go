package main

import (
	"fmt"

	"github.com/example/autodatasource-sdk-go"
)

func main() {
	// 初始化客户端
	baseURL := "http://localhost:8080/autoDataSource"
	client := autodatasource.NewClient(baseURL)

	// 示例 1: 获取数据源列表
	fmt.Println("=== 示例 1: 获取数据源列表 ===")
	dataSourcesResp, err := client.GetDataSources()
	if err != nil {
		fmt.Printf("获取数据源列表失败: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", dataSourcesResp)
	}

	// 示例 2: 转换 SQL 语句
	fmt.Println("\n=== 示例 2: 转换 SQL 语句 ===")
	sql := "SELECT * FROM users LIMIT 10"
	transformResp, err := client.TransformSql(sql, "mysql", "oracle")
	if err != nil {
		fmt.Printf("转换 SQL 失败: %v\n", err)
	} else {
		fmt.Printf("原始 SQL: %s\n", sql)
		if data, ok := transformResp.Data.(map[string]interface{}); ok {
			fmt.Printf("转换后 SQL: %s\n", data["transformedSql"])
		}
	}

	// 示例 3: 获取数据集列表
	fmt.Println("\n=== 示例 3: 获取数据集列表 ===")
	dataItemsResp, err := client.GetDataSets()
	if err != nil {
		fmt.Printf("获取数据集列表失败: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", dataItemsResp)
	}

	// 示例 4: 获取 OSS 文件列表
	fmt.Println("\n=== 示例 4: 获取 OSS 文件列表 ===")
	ossFilesResp, err := client.GetOssFiles("")
	if err != nil {
		fmt.Printf("获取 OSS 文件列表失败: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", ossFilesResp)
	}

	// 示例 5: 获取公共文档列表
	fmt.Println("\n=== 示例 5: 获取公共文档列表 ===")
	docsListResp, err := client.GetPublicDocsList()
	if err != nil {
		fmt.Printf("获取公共文档列表失败: %v\n", err)
	} else {
		fmt.Printf("响应: %+v\n", docsListResp)
	}
}
