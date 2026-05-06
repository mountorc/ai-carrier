package ability

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/xmzail/ai-carrier-dev/carriercore/common/db"
	sqlutil "github.com/xmzail/ai-carrier-dev/carriercore/common/sql"
	"github.com/xmzail/ai-carrier-dev/carriercore/project"
	"github.com/xmzail/ai-carrier-dev/common/capability"
	"github.com/xmzail/ai-carrier-dev/mounts/workflow/workers"
	api "github.com/xmzail/ai-carrier-dev/sdk/mountcore-sdk/go"
)

//go:embed sql_ability.json
var sqlAbilityData []byte

func LoadSQLConfig() error {
	return sqlutil.LoadSQLConfigData(sqlAbilityData, "ability/sql_ability.json")
}

type ExecuteWorkerRequest struct {
	WorkerID   string                 `json:"worker_id"`
	WorkerName string                 `json:"worker_name"`
	Params     map[string]interface{} `json:"params"`
}

type ExecuteWorkerResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

//go:embed api_ability.json
var apiAbilityData []byte

type Config struct {
	Port string
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ExecutionLog struct {
	Timestamp int64       `json:"timestamp"`
	Level     string      `json:"level"`
	LogType   string      `json:"log_type"`
	NodeName  string      `json:"node_name"`
	Content   interface{} `json:"content"`
}

type NodeExecutionDetail struct {
	NodeID     string                 `json:"node_id"`
	NodeType   string                 `json:"node_type"`
	NodeName   string                 `json:"node_name"`
	Properties map[string]interface{} `json:"properties"`
	State      int                    `json:"state"`
	Logs       []ExecutionLog         `json:"logs"`
}

type SupervisorExecutionResult struct {
	Success       bool                  `json:"success"`
	Logs          []ExecutionLog        `json:"logs"`
	ExecutedNodes int                   `json:"executed_nodes"`
	Error         string                `json:"error"`
	NodeDetails   []NodeExecutionDetail `json:"node_details"`
}

type Node struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Sources    []string               `json:"sources"`
}

var (
	config    Config
	pgDB      *sql.DB
	capClient interface {
		Register(ctx context.Context, capability *capability.Capability) error
		Unregister(ctx context.Context, id string) error
		Discover(ctx context.Context, id string, filter *capability.CapabilityFilter) ([]*capability.Capability, error)
		UpdateHeartbeat(ctx context.Context, id string) error
		ListAll(ctx context.Context) ([]*capability.Capability, error)
		GetMetadata(ctx context.Context, id string) (*capability.Capability, error)
		Close() error
	}
)

func initConfig() {
	config = Config{
		Port: getEnv("PORT", "2428"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func initDB() error {
	dsn := db.GetDSN()

	log.Println("Initializing Postgres connection...")

	var err error
	pgDB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open Postgres connection: %w", err)
	}

	if err := pgDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping Postgres: %w", err)
	}

	pgDB.SetMaxOpenConns(25)
	pgDB.SetMaxIdleConns(5)
	pgDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Postgres connected successfully")

	log.Println("Creating ability_proxy table if not exists...")
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS ability_proxy (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		worker_type TEXT NOT NULL,
		enabled BOOLEAN DEFAULT true,
		description TEXT,
		api_address TEXT,
		param_configs JSONB,
		project TEXT,
		uuid_project TEXT,
		last_registered TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err = pgDB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create ability_proxy table: %w", err)
	}
	log.Println("ability_proxy table created successfully")

	log.Println("Adding project and uuid_project columns if not exists...")
	_, err = pgDB.Exec(`
		ALTER TABLE ability_proxy ADD COLUMN IF NOT EXISTS project TEXT,
		ADD COLUMN IF NOT EXISTS uuid_project TEXT
	`)
	if err != nil {
		log.Printf("Warning: failed to add project and uuid_project columns: %v", err)
	} else {
		log.Println("project and uuid_project columns added successfully")
	}

	log.Println("Adding anx_config column if not exists...")
	_, err = pgDB.Exec(`
		ALTER TABLE ability_proxy ADD COLUMN IF NOT EXISTS anx_config JSONB
	`)
	if err != nil {
		log.Printf("Warning: failed to add anx_config column: %v", err)
	} else {
		log.Println("anx_config column added successfully")
	}

	return nil
}

func getHandlerMap() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"startSchedulerInstanceHandler":    startSchedulerInstanceHandler,
		"getNextNodesHandler":              getNextNodesHandler,
		"executeNextNodeHandler":           executeNextNodeHandler,
		"workerRegisterHandler":            workerRegisterHandler,
		"addProxyAbilityHandler":           addProxyAbilityHandler,
		"heartbeatHandler":                 heartbeatHandler,
		"discoverHandler":                  discoverHandler,
		"listHandler":                      listHandler,
		"unregisterHandler":                unregisterHandler,
		"getWorkerHandler":                 getWorkerHandler,
		"getWorkListHandler":               getWorkListHandler,
		"getProxyAbilityHandler":           getProxyAbilityHandler,
		"updateProxyAbilityHandler":        updateProxyAbilityHandler,
		"toggleProxyAbilityEnabledHandler": toggleProxyAbilityEnabledHandler,
		"getProxyAbilityListHandler":       getProxyAbilityListHandler,
		"executeWorkerHandler":             executeWorkerHandler,
		"getAbilityHandler":                getAbilityHandler,
		"vectorSearchHandler":              vectorSearchHandler,
		"textSearchHandler":                textSearchHandler,
		"listProjectsHandler":              listProjectsHandler,
		"getProjectHandler":                getProjectHandler,
		"getAgentListHandler":              getAgentListHandler,
		"GetAgentListDBHandler":            GetAgentListDBHandler,
		"AgentRegisterHandler":             AgentRegisterHandler,
		"executeAbilitySQLByUUID":          executeAbilitySQLByUUID,
		"getANXConfigHandler":              getANXConfigHandler,
	}
}

func RegisterRoutes(router gin.IRouter) {
	if err := api.RegisterRoutesFromJSONData(router, apiAbilityData, "ability/api_ability.json", getHandlerMap()); err != nil {
		log.Fatalf("Failed to register ability routes: %v", err)
	}
}

func Init() error {
	initConfig()
	if err := initDB(); err != nil {
		return err
	}

	// Initialize API logs table
	initApiLogTable()

	projectRoot := getProjectRoot()

	configFile := filepath.Join(projectRoot, "data", "proxy_worker_registrations.json")
	log.Printf("Initializing proxy registry with config file: %s", configFile)
	if err := workers.InitProxyRegistry(configFile); err != nil {
		log.Printf("Warning: failed to initialize proxy registry: %v", err)
	} else {
		log.Printf("Proxy registry initialized successfully")
	}

	// Initialize capability client
	capClient = workers.GetRegistryClient()
	if capClient == nil {
		log.Printf("Registry client is nil, falling back to local client")
		capClient = capability.NewLocalClient()
		log.Printf("Local client created as fallback")
	} else {
		log.Printf("Registry client obtained successfully")
	}

	// Initialize project manager
	// Use the absolute path directly
	projectConfigFile := "/Users/a1-6/Documents/code/trae/autoFlow/services/ability/data/project.json"
	log.Printf("Initializing ProjectManager with config file: %s", projectConfigFile)
	pm, err := project.NewProjectManager(projectConfigFile)
	if err != nil {
		log.Printf("Warning: failed to initialize ProjectManager: %v, continuing without project validation", err)
	} else {
		log.Printf("ProjectManager initialized successfully, loaded %d projects", len(pm.GetAllProjects()))
	}
	// 调用 SetProjectManager 函数，将初始化的 ProjectManager 实例传递给 handlers.go 文件
	SetProjectManager(pm)

	return nil
}

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: failed to get working directory: %v, trying executable path", err)
	} else {
		dir := wd
		for {
			goModPath := filepath.Join(dir, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: failed to get executable path: %v, using current directory", err)
		return "."
	}

	exeDir := filepath.Dir(exePath)

	for {
		goModPath := filepath.Join(exeDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return exeDir
		}

		parent := filepath.Dir(exeDir)
		if parent == exeDir {
			break
		}
		exeDir = parent
	}

	log.Printf("Warning: go.mod not found, using current directory")
	return "."
}

func ExecuteWorker(ctx context.Context, req ExecuteWorkerRequest) (*ExecuteWorkerResponse, error) {
	if req.WorkerID == "" {
		return &ExecuteWorkerResponse{
			Success: false,
			Message: "Invalid request",
			Error:   "Worker ID is required",
		}, nil
	}

	apiAddress := ""

	filter := &capability.CapabilityFilter{
		OnlineOnly: true,
	}

	capabilities, err := capClient.Discover(ctx, req.WorkerID, filter)
	if err == nil && len(capabilities) > 0 {
		worker := capabilities[0]
		if worker.Labels != nil && worker.Labels["api_address"] != "" {
			apiAddress = worker.Labels["api_address"]
		}
	}

	if apiAddress == "" {
		return &ExecuteWorkerResponse{
			Success: false,
			Message: "Worker not found",
			Error:   "Worker with ID " + req.WorkerID + " not found or offline",
		}, nil
	}

	var requestBody []byte

	if inputData, ok := req.Params["input"].(string); ok && inputData != "" {
		log.Printf("Using input field as request body: %s", inputData)
		requestBody = []byte(inputData)
	} else {
		var err error
		requestBody, err = json.Marshal(req.Params)
		if err != nil {
			log.Printf("Error marshaling params: %v", err)
			return &ExecuteWorkerResponse{
				Success: false,
				Message: "Failed to marshal params",
				Error:   fmt.Sprintf("Failed to marshal parameters: %v", err),
			}, nil
		}
		log.Printf("Request Body: %s", string(requestBody))
	}

	externalReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiAddress, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return &ExecuteWorkerResponse{
			Success: false,
			Message: "Failed to create request",
			Error:   fmt.Sprintf("Failed to create HTTP request: %v", err),
		}, nil
	}

	externalReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	log.Printf("Sending request to worker API...")
	externalResp, err := client.Do(externalReq)
	if err != nil {
		log.Printf("Error executing worker: %v", err)
		return &ExecuteWorkerResponse{
			Success: false,
			Message: "Failed to execute worker",
			Error:   fmt.Sprintf("Failed to call worker API: %v", err),
			Data: map[string]interface{}{
				"api_address":    apiAddress,
				"request_params": req.Params,
				"error_details":  err.Error(),
			},
		}, nil
	}
	defer externalResp.Body.Close()

	externalRespBody, err := io.ReadAll(externalResp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return &ExecuteWorkerResponse{
			Success: false,
			Message: "Failed to read response",
			Error:   fmt.Sprintf("Failed to read worker response: %v", err),
		}, nil
	}

	log.Printf("Worker Response Status: %d", externalResp.StatusCode)
	log.Printf("Worker Response Body: %s", string(externalRespBody))

	if externalResp.StatusCode != http.StatusOK {
		log.Printf("Worker execution failed with status: %d", externalResp.StatusCode)
		return &ExecuteWorkerResponse{
			Success: false,
			Message: "Worker execution failed",
			Error:   fmt.Sprintf("Worker API returned status %d: %s", externalResp.StatusCode, string(externalRespBody)),
			Data: map[string]interface{}{
				"api_address":     apiAddress,
				"request_params":  req.Params,
				"response_status": externalResp.StatusCode,
				"response_body":   string(externalRespBody),
			},
		}, nil
	}

	var workerResult map[string]interface{}
	if err := json.Unmarshal(externalRespBody, &workerResult); err != nil {
		log.Printf("Worker executed successfully, but response parsing failed: %v", err)
		return &ExecuteWorkerResponse{
			Success: true,
			Message: "Worker executed successfully",
			Data: map[string]interface{}{
				"api_address":    apiAddress,
				"request_params": req.Params,
				"response_body":  string(externalRespBody),
				"parsing_error":  err.Error(),
			},
		}, nil
	}

	return &ExecuteWorkerResponse{
		Success: true,
		Message: "Worker executed successfully",
		Data:    workerResult,
	}, nil
}

func Close() {
	if capClient != nil {
		capClient.Close()
	}
	if pgDB != nil {
		pgDB.Close()
	}
}
