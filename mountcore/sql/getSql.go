package sql

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WhereSetConfig struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type SQLConfigWithWhere struct {
	Description string                 `json:"description"`
	SQLs        []SQLConfigEntryWithWhere `json:"sqls"`
}

type SQLConfigEntryWithWhere struct {
	UUID      string              `json:"uuid"`
	Name      string              `json:"name"`
	Description string           `json:"description"`
	SQL       string              `json:"sql"`
	WhereSet  map[string]WhereSetConfig `json:"whereSet,omitempty"`
}

func GetSQL(uuid string, where map[string]interface{}) (string, error) {
	sqlConfig, err := loadSQLAbilityConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load SQL ability config: %v", err)
	}

	for _, entry := range sqlConfig.SQLs {
		if entry.UUID == uuid {
			return buildSQL(entry, where)
		}
	}

	return "", fmt.Errorf("未找到 UUID 为 %s 的 SQL 配置", uuid)
}

func loadSQLAbilityConfig() (*SQLConfigWithWhere, error) {
	pathsToTry := []string{
		"./ability/sql_ability.json",
		"../ability/sql_ability.json",
		"carriercore/ability/sql_ability.json",
		"../carriercore/ability/sql_ability.json",
	}

	cwd, err := os.Getwd()
	if err == nil {
		pathsToTry = append(pathsToTry,
			filepath.Join(cwd, "ability/sql_ability.json"),
			filepath.Join(cwd, "../ability/sql_ability.json"),
			filepath.Join(cwd, "carriercore/ability/sql_ability.json"),
		)
	}

	pathsToTry = append(pathsToTry,
		"/Users/a1-6/Documents/code/trae/autoFlow/carriercore/ability/sql_ability.json",
		"/Users/a1-6/Documents/code/trae/autoFlow/ability/sql_ability.json",
	)

	for _, path := range pathsToTry {
		data, err := os.ReadFile(path)
		if err == nil {
			var config SQLConfigWithWhere
			if err := json.Unmarshal(data, &config); err != nil {
				continue
			}
			return &config, nil
		}
	}

	return nil, fmt.Errorf("failed to read SQL ability config file")
}

func buildSQL(entry SQLConfigEntryWithWhere, where map[string]interface{}) (string, error) {
	sql := entry.SQL
	whereClauses := []string{}

	if entry.WhereSet != nil && where != nil {
		for key, value := range where {
			if whereSetConfig, ok := entry.WhereSet[key]; ok {
				clause, err := buildWhereClause(key, value, whereSetConfig)
				if err != nil {
					return "", err
				}
				whereClauses = append(whereClauses, clause)
			}
		}
	}

	if len(whereClauses) > 0 {
		whereSQL := strings.Join(whereClauses, " AND ")
		if strings.Contains(strings.ToUpper(sql), " WHERE ") {
			sql = fmt.Sprintf("%s AND %s", sql, whereSQL)
		} else {
			sql = fmt.Sprintf("%s WHERE %s", sql, whereSQL)
		}
	}

	return sql, nil
}

func buildWhereClause(key string, value interface{}, config WhereSetConfig) (string, error) {
	switch config.Type {
	case "vector":
		return buildVectorClause(key, value, config)
	case "string":
		return buildStringClause(key, value)
	case "number":
		return buildNumberClause(key, value)
	case "boolean":
		return buildBooleanClause(key, value)
	default:
		return buildDefaultClause(key, value)
	}
}

func buildVectorClause(key string, value interface{}, config WhereSetConfig) (string, error) {
	vectorField := config.Name
	if vectorField == "" {
		vectorField = key
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("vector search value must be a string")
	}

	return fmt.Sprintf("%s <-> '%s' < 0.8", vectorField, escapeValue(valueStr)), nil
}

func buildStringClause(key string, value interface{}) (string, error) {
	valueStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("string value must be a string")
	}

	if strings.Contains(valueStr, "%") {
		return fmt.Sprintf("%s LIKE '%s'", key, escapeValue(valueStr)), nil
	}

	return fmt.Sprintf("%s = '%s'", key, escapeValue(valueStr)), nil
}

func buildNumberClause(key string, value interface{}) (string, error) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%s = %d", key, v), nil
	case float32, float64:
		return fmt.Sprintf("%s = %f", key, v), nil
	default:
		return "", fmt.Errorf("invalid number type")
	}
}

func buildBooleanClause(key string, value interface{}) (string, error) {
	switch v := value.(type) {
	case bool:
		return fmt.Sprintf("%s = %t", key, v), nil
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "1" {
			return fmt.Sprintf("%s = true", key), nil
		} else if lower == "false" || lower == "0" {
			return fmt.Sprintf("%s = false", key), nil
		}
		return "", fmt.Errorf("invalid boolean string value")
	default:
		return "", fmt.Errorf("invalid boolean type")
	}
}

func buildDefaultClause(key string, value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%s = '%s'", key, escapeValue(v)), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%s = %d", key, v), nil
	case float32, float64:
		return fmt.Sprintf("%s = %f", key, v), nil
	case bool:
		return fmt.Sprintf("%s = %t", key, v), nil
	case nil:
		return fmt.Sprintf("%s IS NULL", key), nil
	default:
		return "", fmt.Errorf("unsupported value type")
	}
}

func escapeValue(value string) string {
	value = strings.ReplaceAll(value, "'", "''")
	value = strings.ReplaceAll(value, "\\", "\\\\")
	return value
}
