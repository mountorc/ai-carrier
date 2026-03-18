package sql

import (
	"context"
)

type QueryResult struct {
	UUID        string                   `json:"uuid"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	SQL         string                   `json:"sql"`
	Results     []map[string]interface{} `json:"results"`
	Count       int                      `json:"count"`
}

type DatabaseQuerier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
}

type Rows interface {
	Close()
	Next() bool
	FieldDescriptions() []FieldDescription
	Values() ([]interface{}, error)
}

type FieldDescription interface {
	Name() string
}

func ExecuteSQLByUUID(
	ctx context.Context,
	db DatabaseQuerier,
	uuid string,
	args []interface{},
) (*QueryResult, error) {
	sqlEntry, err := GetSQLByUUID(uuid)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, sqlEntry.SQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()

	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			row[fd.Name()] = values[i]
		}
		results = append(results, row)
	}

	return &QueryResult{
		UUID:        sqlEntry.UUID,
		Name:        sqlEntry.Name,
		Description: sqlEntry.Description,
		SQL:         sqlEntry.SQL,
		Results:     results,
		Count:       len(results),
	}, nil
}
