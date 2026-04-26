package skill

import (
	"context"
	"log"

	"github.com/trae/autoFlow/carriercore/mountcore/sql"
)

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

func ExecuteSkillSQLWithWhere(uuid string, where map[string]interface{}, args []interface{}) (*sql.QueryResult, error) {
	return sql.ExecuteSQLByUUIDWithWhere(context.Background(), uuid, where, args)
}

func SearchSkillsByVector(vectorText string) (*sql.QueryResult, error) {
	log.Printf("[DEBUG SKILL SQL_EXECUTOR] SearchSkillsByVector called with: %s", vectorText)
	where := map[string]interface{}{
		"vectorText": vectorText,
	}
	result, err := sql.ExecuteSQLByUUIDWithWhere(context.Background(), "330e8400-e22b-41d4-a716-446655440001", where, nil)
	if err != nil {
		log.Printf("[DEBUG SKILL SQL_EXECUTOR] SearchSkillsByVector error: %v", err)
	} else {
		log.Printf("[DEBUG SKILL SQL_EXECUTOR] SearchSkillsByVector success, found %d results", result.Count)
	}
	return result, err
}