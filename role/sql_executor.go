package role

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type QueryResult struct {
	Count int
	Rows  []map[string]interface{}
}

func GetRoleList() (*QueryResult, error) {
	rows, err := pgDB.Query(`
		SELECT id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, created_at, updated_at 
		FROM role_store 
		WHERE status != 'deleted' 
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows)
}

func GetRoleByUUID(uuid string) (*QueryResult, error) {
	rows, err := pgDB.Query(`
		SELECT id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, created_at, updated_at 
		FROM role_store 
		WHERE uuid = $1 AND status != 'deleted'`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows)
}

func GetRoleByName(name string) (*QueryResult, error) {
	rows, err := pgDB.Query(`
		SELECT id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, created_at, updated_at 
		FROM role_store 
		WHERE name = $1 AND status != 'deleted'`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRows(rows)
}

func scanRows(rows *sql.Rows) (*QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result QueryResult
	result.Rows = make([]map[string]interface{}, 0)

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		if row["skills"] != nil {
			skillsStr, ok := row["skills"].(string)
			if ok && skillsStr != "" {
				var skills []interface{}
				if err := json.Unmarshal([]byte(skillsStr), &skills); err == nil {
					row["skills"] = skills
				}
			}
		}

		if row["tags"] != nil {
			tagsStr, ok := row["tags"].(string)
			if ok && tagsStr != "" {
				var tags []interface{}
				if err := json.Unmarshal([]byte(tagsStr), &tags); err == nil {
					row["tags"] = tags
				}
			}
		}

		result.Rows = append(result.Rows, row)
	}

	result.Count = len(result.Rows)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &result, nil
}

func checkError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("role module db error: %v", err)
}
