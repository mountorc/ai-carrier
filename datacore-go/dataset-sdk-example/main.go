package main

import (
	"fmt"
	"github.com/example/datasource"
	"github.com/example/dataset-go/database"
	"github.com/example/dataset-go/scheduler"
	"github.com/example/dataset-go/server"
)

func main() {
	fmt.Println("=== Dataset SDK 使用示例 ===")
	fmt.Println()

	// 1. 使用数据源管理器
	fmt.Println("步骤 1: 初始化数据源管理器")
	manager := datasource.GetManager()

	configJSON := `{
		"uuid": "post4bc7-9a41-4332-93a1-a60c4d8a7e19",
		"type": "postgres",
		"config": {
			"host": "121.43.142.153",
			"port": 5432,
			"database": "carrier",
			"username": "carrier",
			"password": "GNerfiSP4dpZjwcJ"
		}
	}`

	err := manager.AddDataSource(configJSON)
	if err != nil {
		panic(fmt.Sprintf("添加数据源失败: %v", err))
	}
	fmt.Println("✓ 数据源添加成功")
	fmt.Println()

	// 2. 获取数据源
	fmt.Println("步骤 2: 获取数据源")
	ds, err := manager.GetDataSource("post4bc7-9a41-4332-93a1-a60c4d8a7e19")
	if err != nil {
		panic(fmt.Sprintf("获取数据源失败: %v", err))
	}
	fmt.Println("✓ 数据源获取成功")
	fmt.Println()

	// 3. 查询 Dataset
	fmt.Println("步骤 3: 查询 Dataset (uuid_dataset=dataset1)")
	datasetResult, err := ds.GetDataset("dataset1")
	if err != nil {
		panic(fmt.Sprintf("查询 Dataset 失败: %v", err))
	}
	fmt.Printf("✓ 查询成功，结果: %#v\n", datasetResult)
	fmt.Println()

	// 4. 查询 Workflow
	fmt.Println("步骤 4: 查询 Workflow (uuid_workflow=workflow1)")
	workflowResult, err := ds.GetWorkflow("workflow1")
	if err != nil {
		panic(fmt.Sprintf("查询 Workflow 失败: %v", err))
	}
	fmt.Printf("✓ 查询成功，结果: %#v\n", workflowResult)
	fmt.Println()

	// 5. 启动 HTTP 服务器（可选）
	fmt.Println("步骤 5: 启动 HTTP 服务器（可选）")
	fmt.Println("按 Ctrl+C 停止服务器")
	fmt.Println()

	sched := scheduler.NewScheduler()
	db, err := database.NewDatabase("")
	if err != nil {
		panic(fmt.Sprintf("初始化数据库失败: %v", err))
	}
	defer db.Close()

	srv := server.NewServer(sched, db)
	if err := srv.Start(":8084"); err != nil {
		panic(fmt.Sprintf("启动服务器失败: %v", err))
	}
}
