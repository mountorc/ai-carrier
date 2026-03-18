#!/usr/bin/env python3
content = '''package datasource

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	postgres "github.com/example/postgres-sdk-go/postgres"
)

const DefaultCarrierDSN = "host=121.43.142.153 port=5432 user=carrier password=GNerfiSP4dpZjwcJ dbname=carrier sslmode=disable"

type DataSourceConfig struct {
	UUID   string                 `json:"uuid"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type DataSource struct {
	client      *postgres.Client
	config      *DataSourceConfig
	mu          sync.RWMutex
}

type DataSourceManager struct {
	sources map[string]*DataSource
	mu      sync.RWMutex
}

var manager *DataSourceManager
var managerOnce sync.Once

func GetManager() *DataSourceManager {
	managerOnce.Do(func() {
		manager = &DataSourceManager{
			sources: make(map[string]*DataSource),
		}
	})
	return manager
}

func (m *DataSourceManager) AddDataSource(configJSON string) error {
	var config DataSourceConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	dsn, err := configToDSN(&config)
	if err != nil {
		return fmt.Errorf("failed to convert config to DSN: %w", err)
	}

	client, err := postgres.NewClientFromURL(dsn)
	if err != nil {
		return fmt.Errorf("failed to create datasource: %w", err)
	}

	ds := &DataSource{
		client: client,
		config: &config,
	}

	m.mu.Lock()
	m.sources[config.UUID] = ds
	m.mu.Unlock()

	return nil
}

func (m *DataSourceManager) GetDataSource(uuid string) (*DataSource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ds, exists := m.sources[uuid]
	if !exists {
		return nil, fmt.Errorf("datasource not found for uuid: %s", uuid)
	}
	return ds, nil
}

func (m *DataSourceManager) RemoveDataSource(uuid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ds, exists := m.sources[uuid]; exists {
		ds.Close()
		delete(m.sources, uuid)
	}
}

func (m *DataSourceManager) QueryWithUUID(uuidDatasource string, sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	ds, err := m.GetDataSource(uuidDatasource)
	if err != nil {
		return nil, err
	}
	return ds.Query(sqlStr, args...)
}

func configToDSN(config *DataSourceConfig) (string, error) {
	if config.Type != "postgres" {
		return "", fmt.Errorf("unsupported datasource type: %s", config.Type)
	}

	cfg := config.Config
	host := getString(cfg, "host", "localhost")
	port := getInt(cfg, "port", 5432)
	database := getString(cfg, "database", "")
	username := getString(cfg, "username", "")
	password := getString(cfg, "password", "")

	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, database), nil
}

func getString(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getInt(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return defaultValue
}

func NewDataSource(dsn string) (*DataSource, error) {
	client, err := postgres.NewClientFromURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create datasource: %w", err)
	}
	return &DataSource{client: client}, nil
}

func (ds *DataSource) Close() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if ds.client != nil {
		ds.client.Disconnect()
	}
}

func (ds *DataSource) GetAutoSet(uuid string, tableName string) (map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	sqlStr := fmt.Sprintf("SELECT * FROM %s WHERE uuid = $1", tableName)
	results, err := ds.client.Query(sqlStr, uuid)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no record found for uuid: %s in table: %s", uuid, tableName)
	}

	originalRow := results[0]
	newRow := make(map[string]interface{})

	for key, value := range originalRow {
		if strVal, ok := value.(string); ok {
			if isJSONString(strVal) {
				var jsonVal interface{}
				if err := json.Unmarshal([]byte(strVal), &jsonVal); err == nil {
					newRow[key] = jsonVal
					continue
				}
			}

			decoded, decodeErr := base64.StdEncoding.DecodeString(strVal)
			if decodeErr == nil {
				if isJSONString(string(decoded)) {
					var jsonVal interface{}
					if err := json.Unmarshal(decoded, &jsonVal); err == nil {
						newRow[key] = jsonVal
						continue
					}
				}
			}
		}
		newRow[key] = value
	}

	return newRow, nil
}

func (ds *DataSource) GetDataset(uuid string) (map[string]interface{}, error) {
	return ds.GetAutoSet(uuid, "dataset")
}

func (ds *DataSource) GetWorkflow(uuid string) (map[string]interface{}, error) {
	return ds.GetAutoSet(uuid, "workflow")
}

func (ds *DataSource) Query(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.client.Query(sqlStr, args...)
}

func (ds *DataSource) QueryJSON(sqlStr string, args ...interface{}) (string, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.client.QueryJSON(sqlStr, args...)
}

func isJSONString(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) || (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}
'''

with open('/Users/a1-6/Documents/code/JavaProject/autoDataSource/datacore-go/datasource/datasource.go', 'w') as f:
    f.write(content)

print("datasource.go written successfully")
