package main

import (
	"fmt"

	"./"
)

func main() {
	// 初始化客户端
	baseURL := "http://localhost:8080/autoDataSource"
	client := NewClient(baseURL)

	// 获取数据源列表
	dataSourcesResp, err := client.GetDataSources()
	if err != nil {
		fmt.Printf("获取数据源列表失败: %v\n", err)
	} else {
		fmt.Printf("数据源列表响应: %+v\n", dataSourcesResp)
		fmt.Println("成功获取数据源列表！")
	}
}