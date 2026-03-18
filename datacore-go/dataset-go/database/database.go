package database

import (
	"fmt"

	"github.com/example/datasource"
)

const DefaultDSN = datasource.DefaultCarrierDSN

type Database struct {
	ds *datasource.DataSource
}

func GetAutoSetFromDataSource(uuidDatasource string, uuid string, tableName string) (map[string]interface{}, error) {
	manager := datasource.GetManager()
	ds, err := manager.GetDataSource(uuidDatasource)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource: %w", err)
	}
	return ds.GetAutoSet(uuid, tableName)
}

func GetDatasetFromDataSource(uuidDatasource string, uuid string) (map[string]interface{}, error) {
	return GetAutoSetFromDataSource(uuidDatasource, uuid, "dataset")
}

func GetWorkflowFromDataSource(uuidDatasource string, uuid string) (map[string]interface{}, error) {
	return GetAutoSetFromDataSource(uuidDatasource, uuid, "workflow")
}

func NewDatabase(dsn string) (*Database, error) {
	if dsn == "" {
		dsn = DefaultDSN
	}

	ds, err := datasource.NewDataSource(dsn)
	if err != nil {
		return nil, err
	}

	return &Database{ds: ds}, nil
}

func (d *Database) Close() {
	if d.ds != nil {
		d.ds.Close()
	}
}

func (d *Database) GetAutoSet(uuid string, tableName string) (map[string]interface{}, error) {
	return d.ds.GetAutoSet(uuid, tableName)
}

func (d *Database) GetDataset(uuid string) (map[string]interface{}, error) {
	return d.ds.GetDataset(uuid)
}

func (d *Database) GetWorkflow(uuid string) (map[string]interface{}, error) {
	return d.ds.GetWorkflow(uuid)
}

func (d *Database) Query(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	return d.ds.Query(sqlStr, args...)
}

func (d *Database) QueryJSON(sqlStr string, args ...interface{}) (string, error) {
	return d.ds.QueryJSON(sqlStr, args...)
}
