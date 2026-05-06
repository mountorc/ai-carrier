package scheduler

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	api "github.com/xmzail/ai-carrier-dev/sdk/mountcore-sdk/go"
	"github.com/xmzail/ai-carrier-dev/carriercore/common/db"
	sqlutil "github.com/xmzail/ai-carrier-dev/carriercore/common/sql"
)

//go:embed api_scheduler.json
var apiSchedulerData []byte

//go:embed sql_scheduler.json
var sqlSchedulerData []byte

func LoadSQLConfig() error {
	return sqlutil.LoadSQLConfigData(sqlSchedulerData, "scheduler/sql_scheduler.json")
}

type Config struct {
	Port string
}

type Instance struct {
	ID            string          `json:"id"`
	UUID          string          `json:"uuid"`
	Name          string          `json:"name"`
	ProjectID     string          `json:"projectId"`
	TemplateID    string          `json:"templateId"`
	Status        string          `json:"status"`
	Owner         string          `json:"owner"`
	Description   string          `json:"description"`
	StartupParams json.RawMessage `json:"startupParams"`
	DataUUID      string          `json:"dataUUID"`
	Result        json.RawMessage `json:"result"`
	Nodes         []Node          `json:"nodes,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	StartedAt     *time.Time      `json:"startedAt"`
	CompletedAt   *time.Time      `json:"completedAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type Node struct {
	ID            string          `json:"id"`
	UUIDInstances string          `json:"uuidInstances"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	CapabilityID  string          `json:"capabilityId"`
	Status        string          `json:"status"`
	Input         json.RawMessage `json:"input"`
	Output        json.RawMessage `json:"output"`
	Result        json.RawMessage `json:"result"`
	Properties    json.RawMessage `json:"properties"`
	Description   string          `json:"description"`
	Sources       json.RawMessage `json:"sources"`
	Position      json.RawMessage `json:"position"`
	CreatedAt     time.Time       `json:"createdAt"`
	StartedAt     *time.Time      `json:"startedAt"`
	CompletedAt   *time.Time      `json:"completedAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type Connection struct {
	ID           string          `json:"id"`
	InstanceID   string          `json:"instanceId"`
	SourceNodeID string          `json:"sourceNodeId"`
	TargetNodeID string          `json:"targetNodeId"`
	Points       json.RawMessage `json:"points"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type Execution struct {
	ID           string          `json:"id"`
	InstanceID   string          `json:"instanceId"`
	Status       string          `json:"status"`
	StartTime    time.Time       `json:"startTime"`
	EndTime      *time.Time      `json:"endTime"`
	Duration     int             `json:"duration"`
	Parameters   json.RawMessage `json:"parameters"`
	Result       json.RawMessage `json:"result"`
	ErrorMessage string          `json:"errorMessage"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	config Config
	dbPool *pgxpool.Pool
)

func initConfig() {
	config = Config{
		Port: getEnv("PORT", "2430"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func initDB() error {
	connString := db.GetDSN()

	var err error
	dbPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("无法连接到数据库: %w", err)
	}

	if err := dbPool.Ping(context.Background()); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Println("数据库连接成功")
	return nil
}

func getHandlerMap() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"listInstances":    listInstances,
		"createInstance":   createInstance,
		"getInstance":      getInstance,
		"updateInstance":   updateInstance,
		"deleteInstance":   deleteInstance,
		"startInstance":    startInstance,
		"stopInstance":     stopInstance,
		"restartInstance":  restartInstance,
		"listNodes":        listNodes,
		"createNode":       createNode,
		"updateNode":       updateNode,
		"updateNodeStatus": updateNodeStatus,
		"deleteNode":       deleteNode,
		"listConnections":  listConnections,
		"createConnection": createConnection,
		"updateConnection": updateConnection,
		"deleteConnection": deleteConnection,
		"listExecutions":   listExecutions,
		"getExecution":     getExecution,
		"executeSQLByUUID": executeSQLByUUID,
	}
}

func RegisterRoutes(router gin.IRouter) {
	if err := api.RegisterRoutesFromJSONData(router, apiSchedulerData, "scheduler/api_scheduler.json", getHandlerMap()); err != nil {
		log.Fatalf("Failed to register scheduler routes: %v", err)
	}
}

func Init() error {
	initConfig()
	if err := initDB(); err != nil {
		return err
	}
	return nil
}

func Close() {
	if dbPool != nil {
		dbPool.Close()
	}
}
