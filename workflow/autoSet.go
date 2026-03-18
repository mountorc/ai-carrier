package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trae/autoFlow/carriercore/common/db"
)

var (
	autosetDB *pgxpool.Pool
)

func InitAutosetDB() error {
	connString := db.GetDSN()

	var err error
	autosetDB, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("无法连接到数据库: %w", err)
	}

	if err := autosetDB.Ping(context.Background()); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Println("Autoset数据库连接成功")
	return nil
}

func CloseAutosetDB() {
	if autosetDB != nil {
		autosetDB.Close()
	}
}

func GetAutoSet(uuidWorkflow string) (map[string]interface{}, error) {
	if autosetDB == nil {
		if err := InitAutosetDB(); err != nil {
			return nil, fmt.Errorf("初始化数据库失败: %w", err)
		}
	}

	ctx := context.Background()
	query := "SELECT * FROM workflow WHERE uuid = $1"
	rows, err := autosetDB.Query(ctx, query, uuidWorkflow)
	if err != nil {
		log.Printf("查询workflow表失败: %v", err)
		return nil, fmt.Errorf("查询workflow表失败: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Printf("获取行数据失败: %v", err)
			return nil, fmt.Errorf("获取行数据失败: %w", err)
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			row[string(fd.Name)] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		log.Printf("遍历行数据失败: %v", err)
		return nil, fmt.Errorf("遍历行数据失败: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到指定的workflow记录")
	}

	return results[0], nil
}

func GetWorkflowList() ([]map[string]interface{}, error) {
	if autosetDB == nil {
		if err := InitAutosetDB(); err != nil {
			return nil, fmt.Errorf("初始化数据库失败: %w", err)
		}
	}

	ctx := context.Background()
	query := "SELECT * FROM workflow"
	rows, err := autosetDB.Query(ctx, query)
	if err != nil {
		log.Printf("查询workflow表失败: %v", err)
		return nil, fmt.Errorf("查询workflow表失败: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			log.Printf("获取行数据失败: %v", err)
			return nil, fmt.Errorf("获取行数据失败: %w", err)
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			row[string(fd.Name)] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		log.Printf("遍历行数据失败: %v", err)
		return nil, fmt.Errorf("遍历行数据失败: %w", err)
	}

	return results, nil
}
