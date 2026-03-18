package main

import (
	"fmt"

	"github.com/example/autodatasource-sdk-go"
)

func main() {
	// 创建客户端
	client := autodatasource.NewClient("http://localhost:8080/autoDataSource")

	// 测试根据数据源ID预览数据集
	fmt.Println("=== 测试: 根据数据源ID预览数据集 ===")
	dataSourceId := "90a49r2l313l4243robp5lbqf678vfn10719"
	response, err := client.PreviewDataSetsByDataSourceId(dataSourceId)
	if err != nil {
		fmt.Printf("调用 API 失败: %v\n", err)
		return
	}

	fmt.Printf("成功: %v\n", response.Success)
	fmt.Printf("数据: %v\n", response.Data)
	fmt.Printf("消息: %v\n", response.Message)
	fmt.Println()
}
