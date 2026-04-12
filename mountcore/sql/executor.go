package sql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

type DatabaseConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

type SQLConfigEntry struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
}

type SQLConfig struct {
	Description string           `json:"description"`
	SQLs        []SQLConfigEntry `json:"sqls"`
}

type ModuleConfig struct {
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	APIConfig   string `json:"api_config"`
	SQLConfig   string `json:"sql_config"`
}

type MountConfig struct {
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Server      struct {
		Port  int    `json:"port"`
		Host  string `json:"host"`
		Debug bool   `json:"debug"`
	} `json:"server"`
	Modules    map[string]ModuleConfig `json:"modules"`
	Database   DatabaseConfig          `json:"database"`
	Embedding  struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	} `json:"embedding"`
}

type QueryResult struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Count   int                      `json:"count"`
	Error   string                   `json:"error,omitempty"`
}

type Executor struct {
	db          *sql.DB
	config      DatabaseConfig
	sqlRegistry map[string]SQLConfigEntry
}

var instance *Executor

func LoadMountConfig(configPath string) (*MountConfig, error) {
	var pathsToTry []string

	if configPath != "" {
		pathsToTry = append(pathsToTry, configPath)
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		pathsToTry = append(pathsToTry,
			filepath.Join(exeDir, "mount_config.json"),
			filepath.Join(exeDir, "../mount_config.json"),
			filepath.Join(exeDir, "../../mount_config.json"),
		)
	}

	if _, err := os.Getwd(); err == nil {
		pathsToTry = append(pathsToTry,
			"mount_config.json",
			filepath.Join("carriercore", "mount_config.json"),
			filepath.Join("..", "mount_config.json"),
			filepath.Join("..", "carriercore", "mount_config.json"),
		)
	}

	var lastErr error
	for _, path := range pathsToTry {
		data, err := os.ReadFile(path)
		if err == nil {
			var config MountConfig
			if err := json.Unmarshal(data, &config); err != nil {
				lastErr = fmt.Errorf("failed to parse mount config at %s: %v", path, err)
				continue
			}
			log.Printf("Loaded mount config from: %s", path)
			return &config, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to read mount config. Tried: %v. Last error: %v", pathsToTry, lastErr)
}

func NewExecutorFromConfig(configPath string) (*Executor, error) {
	mountConfig, err := LoadMountConfig(configPath)
	if err != nil {
		return nil, err
	}

	return NewExecutorWithSQLRegistry(mountConfig.Database, mountConfig.Modules)
}

func NewExecutorWithSQLRegistry(dbConfig DatabaseConfig, modules map[string]ModuleConfig) (*Executor, error) {
	executor, err := NewExecutor(dbConfig)
	if err != nil {
		return nil, err
	}

	err = executor.loadSQLConfigs(modules)
	if err != nil {
		executor.Close()
		return nil, fmt.Errorf("failed to load SQL configs: %v", err)
	}

	return executor, nil
}

func NewExecutor(config DatabaseConfig) (*Executor, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	db, err := sql.Open(config.Type, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Database connected successfully: %s@%s:%d/%s",
		config.Username, config.Host, config.Port, config.Database)

	executor := &Executor{
		db:          db,
		config:      config,
		sqlRegistry: make(map[string]SQLConfigEntry),
	}

	instance = executor
	return executor, nil
}

func (e *Executor) loadSQLConfigs(modules map[string]ModuleConfig) error {
	for name, module := range modules {
		if !module.Enabled {
			continue
		}

		if module.SQLConfig == "" {
			log.Printf("Module %s has no SQL config, skipping", name)
			continue
		}

		sqlConfig, err := LoadSQLConfig(module.SQLConfig)
		if err != nil {
			log.Printf("Warning: failed to load SQL config for module %s: %v", name, err)
			continue
		}

		for _, entry := range sqlConfig.SQLs {
			if existing, ok := e.sqlRegistry[entry.UUID]; ok {
				log.Printf("Warning: SQL UUID %s already registered (overriding %s with %s)",
					entry.UUID, existing.Name, entry.Name)
			}
			e.sqlRegistry[entry.UUID] = entry
			log.Printf("Registered SQL: %s (%s)", entry.Name, entry.UUID)
		}
	}

	log.Printf("Total SQL statements registered: %d", len(e.sqlRegistry))
	return nil
}

func LoadSQLConfig(path string) (*SQLConfig, error) {
	var pathsToTry []string

	pathsToTry = append(pathsToTry, path)

	absPath, err := filepath.Abs(path)
	if err == nil && absPath != path {
		pathsToTry = append(pathsToTry, absPath)
	}

	cwd, err := os.Getwd()
	if err == nil {
		pathsToTry = append(pathsToTry, filepath.Join(cwd, path))
	}

	pathsToTry = append(pathsToTry,
		filepath.Join("/Users/a1-6/Documents/code/trae/autoFlow/carriercore", path),
		filepath.Join("/Users/a1-6/Documents/code/trae/autoFlow", path),
	)

	var lastErr error
	for _, p := range pathsToTry {
		data, err := os.ReadFile(p)
		if err == nil {
			var config SQLConfig
			if err := json.Unmarshal(data, &config); err != nil {
				lastErr = fmt.Errorf("failed to parse SQL config file %s: %v", p, err)
				continue
			}
			log.Printf("Loaded SQL config from: %s", p)
			return &config, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to read SQL config file. Tried: %v. Last error: %v", pathsToTry, lastErr)
}

func GetInstance() *Executor {
	return instance
}

func (e *Executor) Close() error {
	if e.db != nil {
		return e.db.Close()
	}
	return nil
}

func decodeBase64JSON(s string) interface{} {
	if s == "" {
		return s
	}
	
	var decoded []byte
	var err error
	var data string
	
	decoded, err = base64.StdEncoding.DecodeString(s)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(s)
	}
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(s)
	}
	if err != nil {
		decoded, err = base64.RawURLEncoding.DecodeString(s)
	}
	
	if err == nil {
		data = string(decoded)
	} else {
		data = s
	}
	
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return s
	}
	return jsonData
}

func (e *Executor) Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				switch v := val.(type) {
				case string:
					row[col] = decodeBase64JSON(v)
				case []byte:
					row[col] = decodeBase64JSON(string(v))
				default:
					row[col] = val
				}
			} else {
				row[col] = nil
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return &QueryResult{
		Columns: columns,
		Rows:    results,
		Count:   len(results),
	}, nil
}

func (e *Executor) Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return e.db.ExecContext(ctx, query, args...)
}

func (e *Executor) QueryOne(ctx context.Context, query string, args ...interface{}) (map[string]interface{}, error) {
	result, err := e.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	if result.Count == 0 {
		return nil, sql.ErrNoRows
	}

	return result.Rows[0], nil
}

func (e *Executor) GetConfig() DatabaseConfig {
	return e.config
}

func (e *Executor) IsConnected() bool {
	if e.db == nil {
		return false
	}
	err := e.db.Ping()
	return err == nil
}

func (e *Executor) GetSQLByUUID(uuid string) (*SQLConfigEntry, error) {
	entry, ok := e.sqlRegistry[uuid]
	if !ok {
		return nil, fmt.Errorf("未找到 UUID 为 %s 的 SQL 配置", uuid)
	}
	return &entry, nil
}

func (e *Executor) ExecuteSQLByUUID(ctx context.Context, uuid string, args []interface{}) (*QueryResult, error) {
	entry, err := e.GetSQLByUUID(uuid)
	if err != nil {
		return nil, err
	}

	log.Printf("Executing SQL: %s (%s)", entry.Name, uuid)
	return e.Query(ctx, entry.SQL, args...)
}

func (e *Executor) ListRegisteredSQL() []SQLConfigEntry {
	entries := make([]SQLConfigEntry, 0, len(e.sqlRegistry))
	for _, entry := range e.sqlRegistry {
		entries = append(entries, entry)
	}
	return entries
}

func Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	if instance == nil {
		return nil, fmt.Errorf("executor not initialized")
	}
	return instance.Query(ctx, query, args...)
}

func Execute(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if instance == nil {
		return nil, fmt.Errorf("executor not initialized")
	}
	return instance.Execute(ctx, query, args...)
}

func ExecuteSQLByUUID(ctx context.Context, uuid string, args []interface{}) (*QueryResult, error) {
	if instance == nil {
		return nil, fmt.Errorf("executor not initialized")
	}
	return instance.ExecuteSQLByUUID(ctx, uuid, args)
}
