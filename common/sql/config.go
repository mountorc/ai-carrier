package sql

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type SQLEntry struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
}

type SQLConfig struct {
	Description string     `json:"description"`
	SQLs        []SQLEntry `json:"sqls"`
}

var (
	sqlConfig     SQLConfig
	sqlConfigOnce sync.Once
	sqlConfigErr  error
)

func LoadSQLConfig(configPath string) error {
	return LoadSQLConfigData([]byte{}, configPath)
}

func LoadSQLConfigData(embeddedData []byte, configPath string) error {
	sqlConfigOnce.Do(func() {
		var data []byte
		var err error

		if len(embeddedData) > 0 {
			data = embeddedData
		} else {
			exePath, exeErr := os.Executable()
			if exeErr == nil {
				exeDir := filepath.Dir(exePath)
				fullPath := filepath.Join(exeDir, configPath)
				data, err = os.ReadFile(fullPath)
				if err != nil {
					data, err = os.ReadFile(configPath)
				}
			} else {
				data, err = os.ReadFile(configPath)
			}
			if err != nil {
				sqlConfigErr = fmt.Errorf("无法读取 SQL 配置文件: %w", err)
				return
			}
		}

		if err := json.Unmarshal(data, &sqlConfig); err != nil {
			sqlConfigErr = fmt.Errorf("无法解析 SQL 配置文件: %w", err)
			return
		}
	})
	return sqlConfigErr
}

func GetSQLByUUID(uuid string) (*SQLEntry, error) {
	for _, sqlEntry := range sqlConfig.SQLs {
		if sqlEntry.UUID == uuid {
			return &sqlEntry, nil
		}
	}
	return nil, fmt.Errorf("未找到 UUID 为 %s 的 SQL", uuid)
}

func GetAllSQLs() []SQLEntry {
	return sqlConfig.SQLs
}
