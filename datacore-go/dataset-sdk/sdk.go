// Package datasetsdk 提供数据集和工作流处理功能
package datasetsdk

import (
	"github.com/example/dataset-sdk/database"
	"github.com/example/dataset-sdk/datasource"
	"github.com/example/dataset-sdk/scheduler"
	"github.com/example/dataset-sdk/server"
)

// 重新导出常用类型和函数，方便使用
type (
	DataSource        = datasource.DataSource
	DataSourceConfig  = datasource.DataSourceConfig
	DataSourceManager = datasource.DataSourceManager
	Database          = database.Database
	Server            = server.Server
	Scheduler         = scheduler.Scheduler
)

var (
	GetManager      = datasource.GetManager
	NewDataSource   = datasource.NewDataSource
	NewDatabase     = database.NewDatabase
	NewServer       = server.NewServer
	NewScheduler    = scheduler.NewScheduler
	DefaultCarrierDSN = datasource.DefaultCarrierDSN
)
