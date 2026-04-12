package ability

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trae/autoFlow/carriercore/mountcore/sql"
)

func executeAbilitySQLByUUID(c *gin.Context) {
	ctx := context.Background()
	uuid := c.Param("uuid")

	var reqBody struct {
		Args []interface{} `json:"args"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		reqBody.Args = []interface{}{}
	}

	result, err := sql.ExecuteSQLByUUID(ctx, uuid, reqBody.Args)
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

func ExecuteSkillSQL(uuid string, args ...interface{}) (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUID(context.Background(), uuid, args)
}

func GetSkillList() (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUID(context.Background(), "330e8400-e22b-41d4-a716-446655440001", []interface{}{})
}

func GetSkillByUUID(uuid string) (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUID(context.Background(), "330e8400-e29b-41d4-a716-446655440002", []interface{}{uuid})
}

func SearchSkillsByName(name string) (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUID(context.Background(), "330e8400-e29b-41d4-a716-446655440003", []interface{}{"%" + name + "%"})
}

func SearchSkillsByNick(nick string) (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUID(context.Background(), "330e8400-e29b-41d4-a716-446655440004", []interface{}{"%" + nick + "%"})
}

func SearchSkillsByKeyword(keyword string) (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUID(context.Background(), "330e8400-e29b-41d4-a716-446655440006", []interface{}{"%" + keyword + "%"})
}

func CountSkills() (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUID(context.Background(), "330e8400-e29b-41d4-a716-446655440005", []interface{}{})
}