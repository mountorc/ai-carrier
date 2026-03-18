package scheduler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/trae/autoFlow/carriercore/common/sql"
)

type pgxRowsAdapter struct {
	rows pgx.Rows
}

func (a *pgxRowsAdapter) Close() {
	a.rows.Close()
}

func (a *pgxRowsAdapter) Next() bool {
	return a.rows.Next()
}

func (a *pgxRowsAdapter) FieldDescriptions() []sql.FieldDescription {
	fds := a.rows.FieldDescriptions()
	result := make([]sql.FieldDescription, len(fds))
	for i := range fds {
		result[i] = &pgxFieldDescriptionAdapter{name: string(fds[i].Name)}
	}
	return result
}

func (a *pgxRowsAdapter) Values() ([]interface{}, error) {
	return a.rows.Values()
}

type pgxFieldDescriptionAdapter struct {
	name string
}

func (a *pgxFieldDescriptionAdapter) Name() string {
	return a.name
}

type pgxQuerierAdapter struct {
	db *pgxpool.Pool
}

func (a *pgxQuerierAdapter) Query(ctx context.Context, sqlStr string, args ...interface{}) (sql.Rows, error) {
	rows, err := a.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRowsAdapter{rows: rows}, nil
}

// 通过 UUID 执行 SQL
func executeSQLByUUID(c *gin.Context) {
	ctx := context.Background()
	uuid := c.Param("uuid")

	var reqBody struct {
		Args []interface{} `json:"args"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		reqBody.Args = []interface{}{}
	}

	querier := &pgxQuerierAdapter{db: dbPool}
	result, err := sql.ExecuteSQLByUUID(ctx, querier, uuid, reqBody.Args)
	if err != nil {
		status := http.StatusInternalServerError
		if len(err.Error()) >= 16 && err.Error()[:16] == "未找到 UUID 为" {
			status = http.StatusNotFound
		}
		c.JSON(status, Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "执行 SQL 成功",
		Data:    result,
	})
}
