package dataset

import (
	"fmt"
	"sync"

	"github.com/example/datasource"
)

var (
	manager     *datasource.DataSourceManager
	managerOnce sync.Once
	managerErr  error
)

func InitManager() error {
	managerOnce.Do(func() {
		manager = datasource.GetManager()

		defaultConfig := `{
			"uuid": "default-carrier",
			"type": "postgres",
			"config": {
				"host": "121.43.142.153",
				"port": 5432,
				"database": "carrier",
				"username": "carrier",
				"password": "GNerfiSP4dpZjwcJ"
			}
		}`

		if err := manager.AddDataSource(defaultConfig); err != nil {
			managerErr = fmt.Errorf("failed to add default datasource: %w", err)
			return
		}
	})
	return managerErr
}

func GetWorkflowConfig(uuidWorkflow string) (map[string]interface{}, error) {
	if err := InitManager(); err != nil {
		return nil, err
	}

	ds, err := manager.GetDataSource("default-carrier")
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource: %w", err)
	}

	return ds.GetWorkflow(uuidWorkflow)
}

func GetDatasetConfig(uuidDataset string) (map[string]interface{}, error) {
	if err := InitManager(); err != nil {
		return nil, err
	}

	ds, err := manager.GetDataSource("default-carrier")
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource: %w", err)
	}

	return ds.GetDataset(uuidDataset)
}

func GetAutoSetConfig(uuid string, tableName string) (map[string]interface{}, error) {
	if err := InitManager(); err != nil {
		return nil, err
	}

	ds, err := manager.GetDataSource("default-carrier")
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource: %w", err)
	}

	return ds.GetAutoSet(uuid, tableName)
}
