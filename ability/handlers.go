package ability

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trae/autoFlow/carriercore/project"
	"github.com/trae/autoFlow/common/capability"
	"github.com/trae/autoFlow/common/embedding"
	"github.com/trae/autoFlow/mounts/workflow/workers"
)

var (
	// embeddingService 全局嵌入服务实例
	embeddingService *embedding.EmbeddingService
	// ProjectManager 全局项目管理器实例
	ProjectManager *project.ProjectManager
)

// SetEmbeddingService 设置嵌入服务实例
func SetEmbeddingService(service *embedding.EmbeddingService) {
	embeddingService = service
	log.Println("Embedding service set successfully in ability module")
}

// SetProjectManager 设置项目管理器实例
func SetProjectManager(pm *project.ProjectManager) {
	ProjectManager = pm
	log.Println("Project manager set successfully in ability module")
}

func listProjectsHandler(c *gin.Context) {
	if ProjectManager == nil {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Project manager not initialized",
			Data:    []interface{}{},
		})
		return
	}

	projects := ProjectManager.GetAllProjects()
	// 转换为 []interface{} 类型
	var data []interface{}
	for _, project := range projects {
		data = append(data, project)
	}
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Projects retrieved successfully",
		Data:    data,
	})
}

func getProjectHandler(c *gin.Context) {
	if ProjectManager == nil {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Project manager not initialized",
			Data:    nil,
		})
		return
	}

	uuidProject := c.Query("uuid_project")
	if uuidProject == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing required parameter: uuid_project",
		})
		return
	}

	// 使用 ValidateUUIDProject 方法，传递空字符串作为 project 参数
	// 由于我们只需要通过 UUID 查找项目，不需要验证 project 名称
	project, err := ProjectManager.ValidateUUIDProject(uuidProject, "")
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Project not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Project retrieved successfully",
		Data:    project,
	})
}

type RegisterRequest struct {
	UUIDProject  string                   `json:"uuid_project"`
	Project      string                   `json:"project"`
	WorkerNick   string                   `json:"workerNick"`
	WorkerNick2  string                   `json:"worker_nick"`
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	Address      string                   `json:"address"`
	Type         string                   `json:"type"`
	Language     string                   `json:"language"`
	Version      string                   `json:"version"`
	Description  string                   `json:"description"`
	Labels       map[string]string        `json:"labels"`
	Tags         []string                 `json:"tags"`
	Owner        string                   `json:"owner"`
	Permission   string                   `json:"permission"`
	AiPermission string                   `json:"ai_permission"`
	Quota        int                      `json:"quota"`
	MaxInstances int                      `json:"max_instances"`
	ParamConfigs []capability.ParamConfig `json:"param_configs"`
	InvokeMethod string                   `json:"invoke_method"`
	Timeout      int                      `json:"timeout"`
	ProxyType    string                   `json:"proxy_type"`
	ApiAddress   string                   `json:"api_address"`
	Enabled      bool                     `json:"enabled"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type HeartbeatRequest struct {
	ID string `json:"id"`
}

type HeartbeatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DiscoverRequest struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Language   string            `json:"language"`
	Labels     map[string]string `json:"labels"`
	OnlineOnly bool              `json:"online_only"`
}

type DiscoverResponse struct {
	Success      bool                     `json:"success"`
	Message      string                   `json:"message"`
	Capabilities []*capability.Capability `json:"capabilities"`
}

type ProxyRegisterRequest struct {
	ID           string                   `json:"id"`
	Name         string                   `json:"name"`
	Type         string                   `json:"type"`
	Language     string                   `json:"language"`
	Version      string                   `json:"version"`
	Address      string                   `json:"address"`
	Description  string                   `json:"description"`
	InvokeMethod string                   `json:"invoke_method"`
	Timeout      int                      `json:"timeout"`
	ProxyType    string                   `json:"proxy_type"`
	VirtualIDs   []string                 `json:"virtual_ids"`
	Owner        string                   `json:"owner"`
	Permission   string                   `json:"permission"`
	AiPermission string                   `json:"ai_permission"`
	Quota        int                      `json:"quota"`
	MaxInstances int                      `json:"max_instances"`
	ParamConfigs []capability.ParamConfig `json:"param_configs"`
	ApiAddress   string                   `json:"api_address"`
	Enabled      bool                     `json:"enabled"`
}

func startSchedulerInstanceHandler(c *gin.Context) {
	instanceID := c.Query("uuid_instance")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing required parameter: uuid_instance",
		})
		return
	}

	var name, projectID, templateID, status, owner, description string
	var dataUUID sql.NullString
	var startupParams []byte
	var createdAt, updatedAt time.Time
	var startedAt, completedAt sql.NullTime

	err := pgDB.QueryRow(`
		SELECT id, name, project_id, template_id, status, owner, description, startup_params, data_uuid, created_at, started_at, completed_at, updated_at
		FROM scheduler_instances
		WHERE id = $1
	`, instanceID).Scan(
		&instanceID, &name, &projectID, &templateID, &status, &owner, &description, &startupParams, &dataUUID,
		&createdAt, &startedAt, &completedAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Message: "Instance not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to check instance: " + err.Error(),
		})
		return
	}

	nodeSQL := `
		SELECT id, name, type, capability_id, status, description, input, output, result, properties, sources, position, created_at, started_at, completed_at, updated_at
		FROM scheduler_nodes
		WHERE instance_id = $1
		ORDER BY created_at ASC
	`

	nodeRows, err := pgDB.Query(nodeSQL, instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to query nodes: " + err.Error(),
		})
		return
	}
	defer nodeRows.Close()

	var nodes []*Node
	for nodeRows.Next() {
		var id, name, nodeType, status, description string
		var capabilityID sql.NullString
		var input, output, result, properties, sources, position sql.NullString
		var createdAt, updatedAt time.Time
		var startedAt, completedAt sql.NullTime

		err := nodeRows.Scan(
			&id, &name, &nodeType, &capabilityID, &status, &description, &input, &output, &result, &properties, &sources, &position,
			&createdAt, &startedAt, &completedAt, &updatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to scan node row: " + err.Error(),
			})
			return
		}

		var propertiesMap map[string]interface{}
		if properties.Valid && properties.String != "" {
			if err := json.Unmarshal([]byte(properties.String), &propertiesMap); err != nil {
				propertiesMap = make(map[string]interface{})
			}
		} else {
			propertiesMap = make(map[string]interface{})
		}

		var sourcesArray []string
		if sources.Valid && sources.String != "" {
			if err := json.Unmarshal([]byte(sources.String), &sourcesArray); err != nil {
				sourcesArray = []string{}
			}
		} else {
			sourcesArray = []string{}
		}

		node := &Node{
			ID:         id,
			Type:       nodeType,
			Properties: propertiesMap,
			Sources:    sourcesArray,
		}

		nodes = append(nodes, node)
	}

	_, err = pgDB.Exec(`
		UPDATE scheduler_instances 
		SET status = 'running', started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`, instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to start instance: " + err.Error(),
		})
		return
	}

	executionResult := executeFlow(nodes, instanceID)

	instance := map[string]interface{}{
		"id":          instanceID,
		"name":        name,
		"projectId":   projectID,
		"templateId":  templateID,
		"status":      "running",
		"owner":       owner,
		"description": description,
		"dataUUID":    dataUUID.String,
		"createdAt":   createdAt,
		"updatedAt":   updatedAt,
	}

	if startedAt.Valid {
		instance["startedAt"] = startedAt.Time
	} else {
		instance["startedAt"] = time.Now()
	}

	if completedAt.Valid {
		instance["completedAt"] = completedAt.Time
	} else {
		instance["completedAt"] = nil
	}

	if len(startupParams) > 0 {
		var startupParamsMap map[string]interface{}
		if err := json.Unmarshal(startupParams, &startupParamsMap); err == nil {
			instance["startupParams"] = startupParamsMap
		}
	}

	instance["nodes"] = nodes

	response := map[string]interface{}{
		"instance_id":      instanceID,
		"status":           "running",
		"message":          "Instance started successfully",
		"started_at":       time.Now(),
		"instance":         instance,
		"execution_result": executionResult,
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Scheduler instance started successfully",
		Data:    response,
	})
}

func getNextNodesHandler(c *gin.Context) {
	instanceID := c.Query("uuid_instance")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing required parameter: uuid_instance",
		})
		return
	}

	var instanceStatus string
	err := pgDB.QueryRow("SELECT status FROM scheduler_instances WHERE id = $1", instanceID).Scan(&instanceStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Message: "Instance not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to check instance: " + err.Error(),
		})
		return
	}

	nodeSQL := `
		SELECT id, name, type, capability_id, status, description, input, output, result, properties, sources, position, created_at, started_at, completed_at, updated_at
		FROM scheduler_nodes
		WHERE instance_id = $1
		ORDER BY created_at ASC
	`

	nodeRows, err := pgDB.Query(nodeSQL, instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to query nodes: " + err.Error(),
		})
		return
	}
	defer nodeRows.Close()

	var nodes []*Node
	var nodeStates map[string]int = make(map[string]int)

	for nodeRows.Next() {
		var id, name, nodeType, status, description string
		var capabilityID sql.NullString
		var input, output, result, properties, sources, position sql.NullString
		var createdAt, updatedAt time.Time
		var startedAt, completedAt sql.NullTime

		err := nodeRows.Scan(
			&id, &name, &nodeType, &capabilityID, &status, &description, &input, &output, &result, &properties, &sources, &position,
			&createdAt, &startedAt, &completedAt, &updatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to scan node row: " + err.Error(),
			})
			return
		}

		var propertiesMap map[string]interface{}
		if properties.Valid && properties.String != "" {
			if err := json.Unmarshal([]byte(properties.String), &propertiesMap); err != nil {
				propertiesMap = make(map[string]interface{})
			}
		} else {
			propertiesMap = make(map[string]interface{})
		}

		var sourcesArray []string
		if sources.Valid && sources.String != "" {
			if err := json.Unmarshal([]byte(sources.String), &sourcesArray); err != nil {
				sourcesArray = []string{}
			}
		} else {
			sourcesArray = []string{}
		}

		node := &Node{
			ID:         id,
			Type:       nodeType,
			Properties: propertiesMap,
			Sources:    sourcesArray,
		}

		nodes = append(nodes, node)

		nodeState := 0
		if status == "running" {
			nodeState = 1
		} else if status == "completed" {
			nodeState = 2
		} else if status == "failed" {
			nodeState = -1
		}
		nodeStates[id] = nodeState
	}

	nextNodes := findNextNodes(nodes, nodeStates)

	response := map[string]interface{}{
		"instance_id": instanceID,
		"next_nodes":  nextNodes,
		"total_nodes": len(nodes),
		"message":     fmt.Sprintf("Found %d next executable nodes", len(nextNodes)),
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Successfully retrieved next nodes",
		Data:    response,
	})
}

func executeNextNodeHandler(c *gin.Context) {
	instanceID := c.Query("uuid_instance")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing required parameter: uuid_instance",
		})
		return
	}

	var instanceStatus string
	err := pgDB.QueryRow("SELECT status FROM scheduler_instances WHERE id = $1", instanceID).Scan(&instanceStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Message: "Instance not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to check instance: " + err.Error(),
		})
		return
	}

	nodeSQL := `
		SELECT id, name, type, capability_id, status, description, input, output, result, properties, sources, position, created_at, started_at, completed_at, updated_at
		FROM scheduler_nodes
		WHERE instance_id = $1
		ORDER BY created_at ASC
	`

	nodeRows, err := pgDB.Query(nodeSQL, instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to query nodes: " + err.Error(),
		})
		return
	}
	defer nodeRows.Close()

	var nodes []*Node
	var nodeStates map[string]int = make(map[string]int)

	for nodeRows.Next() {
		var id, name, nodeType, status, description string
		var capabilityID sql.NullString
		var input, output, result, properties, sources, position sql.NullString
		var createdAt, updatedAt time.Time
		var startedAt, completedAt sql.NullTime

		err := nodeRows.Scan(
			&id, &name, &nodeType, &capabilityID, &status, &description, &input, &output, &result, &properties, &sources, &position,
			&createdAt, &startedAt, &completedAt, &updatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "Failed to scan node row: " + err.Error(),
			})
			return
		}

		var propertiesMap map[string]interface{}
		if properties.Valid && properties.String != "" {
			if err := json.Unmarshal([]byte(properties.String), &propertiesMap); err != nil {
				propertiesMap = make(map[string]interface{})
			}
		} else {
			propertiesMap = make(map[string]interface{})
		}

		var sourcesArray []string
		if sources.Valid && sources.String != "" {
			if err := json.Unmarshal([]byte(sources.String), &sourcesArray); err != nil {
				sourcesArray = []string{}
			}
		} else {
			sourcesArray = []string{}
		}

		node := &Node{
			ID:         id,
			Type:       nodeType,
			Properties: propertiesMap,
			Sources:    sourcesArray,
		}

		nodes = append(nodes, node)

		nodeState := 0
		if status == "running" {
			nodeState = 1
		} else if status == "completed" {
			nodeState = 2
		} else if status == "failed" {
			nodeState = -1
		}
		nodeStates[id] = nodeState
	}

	nextNodes := findNextNodes(nodes, nodeStates)
	if len(nextNodes) == 0 {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "No executable nodes found",
			Data: map[string]interface{}{
				"instance_id": instanceID,
				"message":     "No executable nodes found",
				"next_nodes":  []interface{}{},
			},
		})
		return
	}

	nodeToExecute := nextNodes[0]

	tx, err := pgDB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to begin transaction: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE scheduler_nodes 
		SET status = 'running', started_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`, nodeToExecute.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to update node status: " + err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to commit transaction: " + err.Error(),
		})
		return
	}

	executionResult, executionError := executeNodeWithResultSimple(nodeToExecute)

	tx, err = pgDB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to begin transaction: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()

	resultJSON, err := json.Marshal(executionResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to marshal result: " + err.Error(),
		})
		return
	}

	finalStatus := "completed"
	if executionError != nil {
		finalStatus = "failed"
	}

	_, err = tx.Exec(`
		UPDATE scheduler_nodes 
		SET status = $1, result = $2, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3
	`, finalStatus, string(resultJSON), nodeToExecute.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to update node result: " + err.Error(),
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to commit transaction: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"instance_id":      instanceID,
		"executed_node":    nodeToExecute,
		"execution_result": executionResult,
		"status":           finalStatus,
		"message":          fmt.Sprintf("Node %s executed successfully", nodeToExecute.ID),
	}

	if executionError != nil {
		response["error"] = executionError.Error()
		response["message"] = fmt.Sprintf("Node %s execution failed: %v", nodeToExecute.ID, executionError)
	}

	c.JSON(http.StatusOK, Response{
		Success: executionError == nil,
		Message: "Node execution completed",
		Data:    response,
	})
}

func executeFlow(nodes []*Node, instanceID string) *SupervisorExecutionResult {
	result := &SupervisorExecutionResult{
		Success:       true,
		Logs:          make([]ExecutionLog, 0),
		ExecutedNodes: 0,
		Error:         "",
		NodeDetails:   make([]NodeExecutionDetail, 0),
	}

	for _, node := range nodes {
		nodeName := node.ID
		if name, ok := node.Properties["name"].(string); ok && name != "" {
			nodeName = name
		} else if nick, ok := node.Properties["nick"].(string); ok && nick != "" {
			nodeName = nick
		}
		nodeDetail := NodeExecutionDetail{
			NodeID:     node.ID,
			NodeType:   node.Type,
			NodeName:   nodeName,
			Properties: node.Properties,
			State:      0,
			Logs:       make([]ExecutionLog, 0),
		}
		result.NodeDetails = append(result.NodeDetails, nodeDetail)
	}

	result.Logs = append(result.Logs, ExecutionLog{
		Timestamp: getCurrentTimestamp(),
		Level:     "info",
		LogType:   "execution",
		NodeName:  "supervisor",
		Content:   fmt.Sprintf("收到节点列表，开始执行流程，共 %d 个节点", len(nodes)),
	})

	var startNode *Node
	for _, node := range nodes {
		if node.Type == "start" {
			startNode = node
			break
		}
	}

	if startNode != nil {
		startNodeName := startNode.ID
		if name, ok := startNode.Properties["name"].(string); ok && name != "" {
			startNodeName = name
		} else if nick, ok := startNode.Properties["nick"].(string); ok && nick != "" {
			startNodeName = nick
		}
		result.Logs = append(result.Logs, ExecutionLog{
			Timestamp: getCurrentTimestamp(),
			Level:     "info",
			LogType:   "execution",
			NodeName:  "supervisor",
			Content:   fmt.Sprintf("找到start节点: %s (%s)", startNodeName, startNode.ID),
		})
	} else {
		result.Logs = append(result.Logs, ExecutionLog{
			Timestamp: getCurrentTimestamp(),
			Level:     "warning",
			LogType:   "execution",
			NodeName:  "supervisor",
			Content:   "未找到start节点，将执行所有没有依赖的节点",
		})
	}

	if startNode != nil {
		executeNode(startNode, result)
	} else {
		for _, node := range nodes {
			if len(node.Sources) == 0 {
				executeNode(node, result)
			}
		}
	}

	for _, node := range nodes {
		if node != startNode && len(node.Sources) > 0 {
			executeNode(node, result)
		}
	}

	result.ExecutedNodes = len(nodes)
	return result
}

func executeNode(node *Node, result *SupervisorExecutionResult) {
	nodeName := node.ID
	if name, ok := node.Properties["name"].(string); ok && name != "" {
		nodeName = name
	} else if nick, ok := node.Properties["nick"].(string); ok && nick != "" {
		nodeName = nick
	}

	result.Logs = append(result.Logs, ExecutionLog{
		Timestamp: getCurrentTimestamp(),
		Level:     "info",
		LogType:   "status_change",
		NodeName:  nodeName,
		Content:   "节点状态变更: pending -> running",
	})

	result.Logs = append(result.Logs, ExecutionLog{
		Timestamp: getCurrentTimestamp(),
		Level:     "info",
		LogType:   "execution",
		NodeName:  nodeName,
		Content:   fmt.Sprintf("开始执行节点: %s", nodeName),
	})

	time.Sleep(100 * time.Millisecond)

	result.Logs = append(result.Logs, ExecutionLog{
		Timestamp: getCurrentTimestamp(),
		Level:     "info",
		LogType:   "status_change",
		NodeName:  nodeName,
		Content:   "节点状态变更: running -> completed",
	})

	result.Logs = append(result.Logs, ExecutionLog{
		Timestamp: getCurrentTimestamp(),
		Level:     "info",
		LogType:   "execution",
		NodeName:  nodeName,
		Content:   fmt.Sprintf("节点执行完成: %s", nodeName),
	})

	for i, detail := range result.NodeDetails {
		if detail.NodeID == node.ID {
			result.NodeDetails[i].State = 2
			break
		}
	}
}

func getCurrentTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func findNextNodes(nodes []*Node, nodeStates map[string]int) []*Node {
	var nextNodes []*Node

	for _, node := range nodes {
		state, exists := nodeStates[node.ID]
		if !exists || state != 0 {
			continue
		}

		allSourcesCompleted := true
		for _, sourceID := range node.Sources {
			depState, exists := nodeStates[sourceID]
			if !exists || depState != 2 {
				allSourcesCompleted = false
				break
			}
		}

		if allSourcesCompleted {
			nextNodes = append(nextNodes, node)
		}
	}

	return nextNodes
}

func executeNodeWithResultSimple(node *Node) (map[string]interface{}, error) {
	time.Sleep(500 * time.Millisecond)

	result := map[string]interface{}{
		"node_id":    node.ID,
		"node_type":  node.Type,
		"status":     "executed",
		"timestamp":  time.Now().Unix(),
		"properties": node.Properties,
	}

	switch node.Type {
	case "start":
		result["output"] = map[string]interface{}{
			"message": "Flow started",
			"status":  "success",
		}
	case "end":
		result["output"] = map[string]interface{}{
			"message": "Flow completed",
			"status":  "success",
		}
	case "data-fetch-worker":
		result["output"] = map[string]interface{}{
			"message": "Data fetched successfully",
			"data":    []string{"item1", "item2", "item3"},
		}
	case "data-insert-worker":
		result["output"] = map[string]interface{}{
			"message": "Data inserted successfully",
			"count":   3,
		}
	case "json-worker":
		result["output"] = map[string]interface{}{
			"message":   "JSON processed successfully",
			"processed": true,
		}
	default:
		result["output"] = map[string]interface{}{
			"message": "Node executed",
			"status":  "success",
		}
	}

	return result, nil
}

func workerRegisterHandler(c *gin.Context) {
	var req RegisterRequest
	var rawReq map[string]interface{}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	if err := json.Unmarshal(bodyBytes, &rawReq); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// 如果ProjectManager已初始化，则验证项目；否则跳过验证（向后兼容）
	if ProjectManager != nil {
		if _, err := ProjectManager.ValidateUUIDProject(req.UUIDProject, req.Project); err != nil {
			log.Printf("Warning: project validation failed, but continuing with registration: %v", err)
		}
	}

	if req.WorkerNick == "" {
		if workerNick, ok := rawReq["worker_nick"]; ok {
			req.WorkerNick = workerNick.(string)
		}
	}

	workerID := fmt.Sprintf("%s.%s@%s", req.Project, req.WorkerNick, req.Version)

	if req.Labels == nil {
		req.Labels = make(map[string]string)
	}
	req.Labels["project"] = req.Project
	req.Labels["workerNick"] = req.WorkerNick
	req.Labels["worker_nick"] = req.WorkerNick

	cap := &capability.Capability{
		ID:           workerID,
		Name:         req.Name,
		Address:      req.Address,
		Type:         capability.CapabilityType(req.Type),
		Language:     capability.Language(req.Language),
		Version:      req.Version,
		Description:  req.Description,
		Labels:       req.Labels,
		Tags:         req.Tags,
		Owner:        req.Owner,
		Permission:   req.Permission,
		AiPermission: req.AiPermission,
		Quota:        req.Quota,
		MaxInstances: req.MaxInstances,
		ParamConfigs: req.ParamConfigs,
		InvokeMethod: req.InvokeMethod,
		Timeout:      req.Timeout,
		ProxyType:    req.ProxyType,
		Online:       true,
		Heartbeat:    time.Now(),
	}

	if req.ApiAddress != "" {
		cap.Labels["api_address"] = req.ApiAddress
	}
	if req.Enabled {
		cap.Labels["enabled"] = "true"
	}

	err = capClient.Register(context.Background(), cap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to register: %v", err),
		})
		return
	}

	// 生成向量嵌入
	var embedding []float32
	if embeddingService != nil {
		// 为能力生成嵌入向量，将需要的字段合到一起
		embeddingText := req.Name + " " + req.Description + " " + req.Type + " " + req.Language + " " + strings.Join([]string{"workflow", "worker", req.Type}, " ")
		// 添加tags
		for _, tag := range req.Tags {
			embeddingText += " " + tag
		}
		// 添加param_configs中的字段
		for _, param := range req.ParamConfigs {
			embeddingText += " " + param.Title + " " + param.Nick + " " + param.Mode
		}
		// 添加api_address
		if req.ApiAddress != "" {
			embeddingText += " " + req.ApiAddress
		}
		log.Printf("Generating embedding for ability text: %s", embeddingText)
		var err error
		embedding, err = embeddingService.GetEmbedding(embeddingText)
		if err != nil {
			log.Printf("Warning: failed to generate embedding: %v", err)
			// 如果生成嵌入失败，使用空的1024维向量
			embedding = make([]float32, 1024)
		} else {
			log.Printf("Successfully generated embedding with dimension: %d", len(embedding))
			// 将向量扩展为1024维
			if len(embedding) < 1024 {
				extendedEmbedding := make([]float32, 1024)
				copy(extendedEmbedding, embedding)
				embedding = extendedEmbedding
				log.Printf("Extended embedding to 1024 dimensions")
			}
		}
	} else {
		// 如果嵌入服务不可用，使用空的1024维向量
		embedding = make([]float32, 1024)
		log.Printf("EmbeddingService not available, using empty vector")
	}

	// 将嵌入向量转换为JSON字符串，避免空字节问题
	embeddingJSON, _ := json.Marshal(embedding)

	// 将能力保存到数据库
	tagsJSON, _ := json.Marshal(req.Tags)
	paramConfigsJSON, _ := json.Marshal(req.ParamConfigs)
	inputSchemaJSON, _ := json.Marshal(map[string]interface{}{})
	outputSchemaJSON, _ := json.Marshal(map[string]interface{}{})

	sql := `
		INSERT INTO ability_template 
		(id, project, worker_type, name, description, type, version, 
		 input_schema, output_schema, tags, param_configs, embedding, embedding_type, is_enabled, is_auto_created, last_register_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
		name = $4,
		description = $5,
		type = $6,
		input_schema = $8,
		output_schema = $9,
		tags = $10,
		param_configs = $11,
		embedding = $12,
		embedding_type = $13,
		is_enabled = $14,
		last_register_at = NOW(),
		updated_at = NOW()
	`

	_, dbErr := pgDB.Exec(sql,
		workerID,
		req.Project,
		req.WorkerNick,
		req.Name,
		req.Description,
		req.Type,
		req.Version,
		inputSchemaJSON,
		outputSchemaJSON,
		tagsJSON,
		paramConfigsJSON,
		embeddingJSON, // embedding as JSON
		"qwen",        // embedding_type
		req.Enabled,
		true, // is_auto_created
	)

	if dbErr != nil {
		log.Printf("Warning: Failed to save capability to database: %v", dbErr)
		// 不影响注册成功，只是记录警告
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: true,
		Message: "Worker registered successfully",
	})
}

func addProxyWorkerHandler(c *gin.Context) {
	var req ProxyRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	if req.Enabled {
		options := workers.DefaultWorkerOptions()
		options.Description = req.Description
		options.ApiAddress = req.ApiAddress
		options.ParamConfigs = req.ParamConfigs
		options.Project = "proxyWorkers"
		options.WorkerNick = req.ID

		if err := workers.RegisterWorkerWithOptions(req.ID, req.Name, req.Type, options); err != nil {
			c.JSON(http.StatusInternalServerError, RegisterResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to register proxy worker: %v", err),
			})
			return
		}

		address := req.Address
		if address == "" {
			address = "localhost"
		}

		version := req.Version
		if version == "" {
			version = "1.0.0"
		}
		workerID := fmt.Sprintf("proxyWorkers.%s@%s", req.ID, version)

		cap := &capability.Capability{
			ID:               workerID,
			Name:             req.Name,
			Address:          address,
			Type:             capability.CapabilityType(req.Type),
			Language:         capability.Language(req.Language),
			Version:          version,
			Description:      req.Description,
			InvokeMethod:     req.InvokeMethod,
			Timeout:          req.Timeout,
			ProxyType:        req.ProxyType,
			VirtualIDs:       req.VirtualIDs,
			VirtualIDDescMap: make(map[string]string),
			Labels: map[string]string{
				"env":         "development",
				"component":   "worker",
				"api_address": req.ApiAddress,
				"project":     "proxyWorkers",
				"workerNick":  req.ID,
			},
			Tags:         []string{"workflow", "worker", req.Type},
			Owner:        req.Owner,
			Permission:   req.Permission,
			AiPermission: req.AiPermission,
			Quota:        req.Quota,
			MaxInstances: req.MaxInstances,
			ParamConfigs: req.ParamConfigs,
			Online:       true,
			Heartbeat:    time.Now(),
		}

		if err := capClient.Register(context.Background(), cap); err != nil {
			log.Printf("Warning: failed to register proxy worker to capability center: %v", err)
		} else {
			log.Printf("Proxy worker %s registered to capability center successfully", req.ID)
		}
	} else {
		log.Printf("Proxy worker %s is disabled, skipping registration", req.ID)
	}

	projectRoot := getProjectRoot()
	proxyConfigFile := filepath.Join(projectRoot, "data", "proxy_worker_registrations.json")
	log.Printf("Saving proxy worker to: %s", proxyConfigFile)

	var proxyWorkers []map[string]interface{}
	data, err := os.ReadFile(proxyConfigFile)
	if err == nil {
		if err := json.Unmarshal(data, &proxyWorkers); err != nil {
			log.Printf("Warning: failed to unmarshal existing proxy worker config: %v", err)
			proxyWorkers = []map[string]interface{}{}
		}
	} else {
		proxyWorkers = []map[string]interface{}{}
	}

	newWorker := map[string]interface{}{
		"ID":         req.ID,
		"Name":       req.Name,
		"WorkerType": req.Type,
		"Enabled":    req.Enabled,
		"Options": map[string]interface{}{
			"Description":  req.Description,
			"ApiAddress":   req.ApiAddress,
			"ParamConfigs": req.ParamConfigs,
		},
		"LastRegistered": time.Now().Format(time.RFC3339),
	}

	updated := false
	for i, worker := range proxyWorkers {
		if workerID, ok := worker["ID"].(string); ok && workerID == req.ID {
			proxyWorkers[i] = newWorker
			updated = true
			break
		}
	}

	if !updated {
		proxyWorkers = append(proxyWorkers, newWorker)
	}

	updatedData, err := json.MarshalIndent(proxyWorkers, "", "  ")
	if err != nil {
		log.Printf("Warning: failed to marshal proxy worker config: %v", err)
	} else if err := os.WriteFile(proxyConfigFile, updatedData, 0644); err != nil {
		log.Printf("Warning: failed to write proxy worker config: %v", err)
	} else {
		log.Printf("Proxy worker saved to %s successfully", proxyConfigFile)
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: true,
		Message: "Proxy worker registered successfully and saved to proxy_worker_registrations.json",
	})
}

func heartbeatHandler(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	err := capClient.UpdateHeartbeat(context.Background(), req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, HeartbeatResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update heartbeat: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, HeartbeatResponse{
		Success: true,
		Message: "Heartbeat updated successfully",
	})
}

func discoverHandler(c *gin.Context) {
	var req DiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	filter := &capability.CapabilityFilter{
		OnlineOnly: req.OnlineOnly,
		Labels:     req.Labels,
	}

	if req.Type != "" {
		capType := capability.CapabilityType(req.Type)
		filter.Type = &capType
	}

	if req.Language != "" {
		lang := capability.Language(req.Language)
		filter.Language = &lang
	}

	capabilities, err := capClient.Discover(context.Background(), req.ID, filter)
	if err != nil {
		response := map[string]interface{}{
			"success":        false,
			"message":        fmt.Sprintf("Failed to discover: %v", err),
			"request_params": req,
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	if len(capabilities) == 0 && req.ID != "" {
		projectRoot := getProjectRoot()
		proxyConfigFile := filepath.Join(projectRoot, "data", "proxy_worker_registrations.json")
		log.Printf("Reading proxy worker config from: %s", proxyConfigFile)

		data, err := os.ReadFile(proxyConfigFile)
		if err == nil {
			var proxyWorkers []map[string]interface{}
			if err := json.Unmarshal(data, &proxyWorkers); err == nil {
				for _, worker := range proxyWorkers {
					if workerID, ok := worker["ID"].(string); ok && workerID == req.ID {
						options, _ := worker["Options"].(map[string]interface{})
						apiAddress := ""
						description := ""
						var paramConfigs []capability.ParamConfig

						if options != nil {
							if addr, ok := options["ApiAddress"].(string); ok {
								apiAddress = addr
							}
							if desc, ok := options["Description"].(string); ok {
								description = desc
							}
							if params, ok := options["ParamConfigs"].([]interface{}); ok {
								for _, p := range params {
									if paramMap, ok := p.(map[string]interface{}); ok {
										param := capability.ParamConfig{}
										if title, ok := paramMap["Title"].(string); ok {
											param.Title = title
										}
										if nick, ok := paramMap["Nick"].(string); ok {
											param.Nick = nick
										}
										if mode, ok := paramMap["Mode"].(string); ok {
											param.Mode = mode
										}
										if must, ok := paramMap["Must"].(bool); ok {
											param.Must = must
										}
										if placeholder, ok := paramMap["Placeholder"].(string); ok {
											param.Placeholder = placeholder
										}
										if value, ok := paramMap["Value"].(string); ok {
											param.Value = value
										}
										paramConfigs = append(paramConfigs, param)
									}
								}
							}
						}

						cap := &capability.Capability{
							ID:           req.ID,
							Name:         worker["Name"].(string),
							Address:      "localhost",
							Type:         capability.CapabilityType(worker["WorkerType"].(string)),
							Language:     capability.Language("go"),
							Version:      "1.0.0",
							Description:  description,
							InvokeMethod: "local",
							Timeout:      30,
							ProxyType:    "proxy",
							VirtualIDs:   []string{},
							Labels: map[string]string{
								"api_address": apiAddress,
								"component":   "worker",
								"env":         "development",
							},
							Tags:         []string{"workflow", "worker", worker["WorkerType"].(string)},
							Online:       true,
							Heartbeat:    time.Now(),
							ParamConfigs: paramConfigs,
						}

						capabilities = append(capabilities, cap)
						break
					}
				}
			}
		}
	}

	response := map[string]interface{}{
		"success":        true,
		"message":        "Discovered successfully",
		"capabilities":   capabilities,
		"request_params": req,
	}

	if len(capabilities) > 0 {
		for _, cap := range capabilities {
			if cap.Labels != nil && cap.Labels["api_address"] != "" {
				response["api_address"] = cap.Labels["api_address"]
				break
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

func listHandler(c *gin.Context) {
	capabilities, err := capClient.ListAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to list: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":      true,
		"message":      "Listed successfully",
		"capabilities": capabilities,
	})
}

func unregisterHandler(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	err := capClient.Unregister(context.Background(), req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RegisterResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to unregister: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: true,
		Message: "Unregistered successfully",
	})
}

func getWorkerHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing worker id",
		})
		return
	}

	filter := &capability.CapabilityFilter{
		OnlineOnly: true,
	}

	capabilities, err := capClient.Discover(context.Background(), id, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DiscoverResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get worker: %v", err),
		})
		return
	}

	if len(capabilities) == 0 {
		c.JSON(http.StatusNotFound, DiscoverResponse{
			Success: false,
			Message: "Worker not found",
		})
		return
	}

	c.JSON(http.StatusOK, DiscoverResponse{
		Success:      true,
		Message:      "Worker found successfully",
		Capabilities: capabilities,
	})
}

func getWorkListHandler(c *gin.Context) {
	filter := &capability.CapabilityFilter{
		OnlineOnly: true,
	}

	capType := capability.CapabilityTypeWorker
	filter.Type = &capType

	capabilities, err := capClient.Discover(context.Background(), "", filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DiscoverResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get work list: %v", err),
		})
		return
	}

	deduplicatedCapabilities := make([]*capability.Capability, 0)
	seenIDs := make(map[string]bool)
	for _, cap := range capabilities {
		if !seenIDs[cap.ID] {
			deduplicatedCapabilities = append(deduplicatedCapabilities, cap)
			seenIDs[cap.ID] = true
		}
	}

	c.JSON(http.StatusOK, DiscoverResponse{
		Success:      true,
		Message:      "Work list retrieved successfully",
		Capabilities: deduplicatedCapabilities,
	})
}

func getProxyWorkerHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing worker id",
		})
		return
	}

	projectRoot := getProjectRoot()
	proxyConfigFile := filepath.Join(projectRoot, "data", "proxy_worker_registrations.json")
	log.Printf("Reading proxy worker config from: %s", proxyConfigFile)

	data, err := os.ReadFile(proxyConfigFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to read proxy worker config: %v", err),
		})
		return
	}

	var proxyWorkers []map[string]interface{}
	if err := json.Unmarshal(data, &proxyWorkers); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to parse proxy worker config: %v", err),
		})
		return
	}

	var targetWorker map[string]interface{}
	for _, worker := range proxyWorkers {
		if workerID, ok := worker["ID"].(string); ok && workerID == id {
			targetWorker = worker
			break
		}
	}

	if targetWorker == nil {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Proxy worker with ID %s not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":      true,
		"message":      fmt.Sprintf("Proxy worker %s retrieved successfully", id),
		"proxy_worker": targetWorker,
	})
}

func updateProxyWorkerHandler(c *gin.Context) {
	var updateRequest map[string]interface{}
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	log.Printf("=== updateProxyWorkerHandler called ===")
	log.Printf("Received update request: %+v", updateRequest)

	id, ok := updateRequest["id"].(string)
	if !ok || id == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing worker id",
		})
		return
	}
	log.Printf("Updating worker with ID: %s", id)

	projectRoot := getProjectRoot()
	proxyConfigFile := filepath.Join(projectRoot, "data", "proxy_worker_registrations.json")
	log.Printf("Updating proxy worker config in: %s", proxyConfigFile)

	data, err := os.ReadFile(proxyConfigFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to read proxy worker config: %v", err),
		})
		return
	}

	var proxyWorkers []map[string]interface{}
	if err := json.Unmarshal(data, &proxyWorkers); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to parse proxy worker config: %v", err),
		})
		return
	}

	updated := false
	var updatedWorker map[string]interface{}
	var workerIndex int = -1
	for i, worker := range proxyWorkers {
		if workerID, ok := worker["ID"].(string); ok && workerID == id {
			if name, ok := updateRequest["name"].(string); ok {
				proxyWorkers[i]["Name"] = name
			}
			if options, ok := updateRequest["options"].(map[string]interface{}); ok {
				if workerOptions, ok := proxyWorkers[i]["Options"].(map[string]interface{}); ok {
					if apiAddress, ok := options["api_address"].(string); ok {
						workerOptions["ApiAddress"] = apiAddress
					}
					if description, ok := options["description"].(string); ok {
						workerOptions["Description"] = description
					}
					if paramConfigs, ok := options["param_configs"].([]interface{}); ok {
						workerOptions["ParamConfigs"] = paramConfigs
					}
				}
			}
			proxyWorkers[i]["LastRegistered"] = time.Now().Format(time.RFC3339)
			updatedWorker = proxyWorkers[i]
			workerIndex = i
			updated = true
			break
		}
	}

	if !updated || workerIndex < 0 {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Proxy worker with ID %s not found", id),
		})
		return
	}

	isEnabled := true
	if enabledVal, ok := updatedWorker["Enabled"].(bool); ok {
		isEnabled = enabledVal
	}

	if isEnabled && updatedWorker != nil {
		log.Printf("Proxy worker %s is enabled, re-registering to worker registry and capability center", id)

		name, _ := updatedWorker["Name"].(string)
		log.Printf("Re-registering with new name: %s", name)
		workerType, _ := updatedWorker["WorkerType"].(string)
		options, _ := updatedWorker["Options"].(map[string]interface{})

		if options != nil {
			apiAddress, _ := options["ApiAddress"].(string)
			description, _ := options["Description"].(string)
			paramConfigs, _ := options["ParamConfigs"].([]interface{})

			workerOptions := workers.DefaultWorkerOptions()
			workerOptions.Description = description
			workerOptions.ApiAddress = apiAddress
			workerOptions.Project = "proxyWorkers"
			workerOptions.WorkerNick = id

			var configs []capability.ParamConfig
			if paramConfigs != nil {
				for _, pc := range paramConfigs {
					if pcMap, ok := pc.(map[string]interface{}); ok {
						config := capability.ParamConfig{}
						if title, ok := pcMap["Title"].(string); ok {
							config.Title = title
						} else if title, ok := pcMap["title"].(string); ok {
							config.Title = title
						}
						if nick, ok := pcMap["Nick"].(string); ok {
							config.Nick = nick
						} else if nick, ok := pcMap["nick"].(string); ok {
							config.Nick = nick
						}
						if mode, ok := pcMap["Mode"].(string); ok {
							config.Mode = mode
						} else if mode, ok := pcMap["mode"].(string); ok {
							config.Mode = mode
						}
						if must, ok := pcMap["Must"].(bool); ok {
							config.Must = must
						} else if must, ok := pcMap["must"].(bool); ok {
							config.Must = must
						}
						if placeholder, ok := pcMap["Placeholder"].(string); ok {
							config.Placeholder = placeholder
						} else if placeholder, ok := pcMap["placeholder"].(string); ok {
							config.Placeholder = placeholder
						}
						if value, ok := pcMap["Value"].(string); ok {
							config.Value = value
						} else if value, ok := pcMap["value"].(string); ok {
							config.Value = value
						}
						configs = append(configs, config)
					}
				}
				workerOptions.ParamConfigs = configs
			}

			if err := capClient.Unregister(context.Background(), id); err != nil {
				log.Printf("Warning: failed to unregister old proxy worker from capability center: %v", err)
			}

			if err := workers.RegisterWorkerWithOptions(id, name, workerType, workerOptions); err != nil {
				log.Printf("Warning: failed to re-register proxy worker to worker registry: %v", err)
			} else {
				log.Printf("Proxy worker %s re-registered to worker registry successfully", id)
			}

			cap := &capability.Capability{
				ID:               id,
				Name:             name,
				Address:          "localhost",
				Type:             capability.CapabilityType(workerType),
				Language:         capability.Language("go"),
				Version:          "1.0.0",
				Description:      description,
				InvokeMethod:     "local",
				Timeout:          30,
				ProxyType:        "proxy",
				VirtualIDs:       []string{},
				VirtualIDDescMap: make(map[string]string),
				Labels: map[string]string{
					"env":         "development",
					"component":   "worker",
					"api_address": apiAddress,
				},
				Tags:         []string{"workflow", "worker", workerType},
				Owner:        "",
				Permission:   "",
				AiPermission: "",
				Quota:        0,
				MaxInstances: 0,
				ParamConfigs: configs,
				Online:       true,
				Heartbeat:    time.Now(),
			}

			if err := capClient.Register(context.Background(), cap); err != nil {
				log.Printf("Warning: failed to re-register proxy worker to capability center: %v", err)
			} else {
				log.Printf("Proxy worker %s re-registered to capability center successfully", id)
			}
		}
	} else if !isEnabled {
		log.Printf("Proxy worker %s is disabled, unregistering from capability center", id)
		workerID := fmt.Sprintf("proxyWorkers.%s@1.0.0", id)
		if err := capClient.Unregister(context.Background(), workerID); err != nil {
			log.Printf("Warning: failed to unregister proxy worker from capability center: %v", err)
		} else {
			log.Printf("Proxy worker %s unregistered from capability center successfully", id)
		}
	}

	updatedData, err := json.MarshalIndent(proxyWorkers, "", "  ")
	if err != nil {
		log.Printf("Warning: failed to marshal proxy worker config: %v", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to marshal proxy worker config: %v", err),
		})
		return
	}

	if err := os.WriteFile(proxyConfigFile, updatedData, 0644); err != nil {
		log.Printf("Warning: failed to write proxy worker config: %v", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to write proxy worker config: %v", err),
		})
		return
	}

	log.Printf("Proxy worker %s updated successfully", id)

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Proxy worker %s updated successfully", id),
	})
}

func getAgentListHandler(c *gin.Context) {
	// 直接使用绝对路径读取agent_list.json文件
	agentListPath := "/Users/a1-6/Documents/code/trae/autoFlow/carriercore/data/agent_list.json"
	log.Printf("Reading agent list from: %s", agentListPath)

	data, err := os.ReadFile(agentListPath)
	if err != nil {
		log.Printf("Error reading agent_list.json: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to read agent list",
		})
		return
	}

	// 解析JSON数据
	var agentList []interface{}
	if err := json.Unmarshal(data, &agentList); err != nil {
		log.Printf("Error parsing agent_list.json: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to parse agent list",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Agent list retrieved successfully",
		Data:    agentList,
	})
}

func getSkillListHandler(c *gin.Context) {
	log.Println("Getting skill list from database...")

	result, err := GetSkillList()
	if err != nil {
		log.Printf("Error getting skill list: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get skills: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d skills", result.Count),
		Data:    result.Rows,
	})
}

func getSkillHandler(c *gin.Context) {
	skillUUID := c.Query("uuid")
	if skillUUID == "" {
		skillUUID = c.Query("id")
		if skillUUID == "" {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "Missing required parameter: uuid or id",
			})
			return
		}
	}

	log.Printf("Getting skill with UUID: %s", skillUUID)

	var (
		uuid        string
		name        string
		nick        string
		description string
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := pgDB.QueryRow(`SELECT uuid, name, nick, description, created_at, updated_at FROM skill_store_skills WHERE uuid = $1`, skillUUID).
		Scan(&uuid, &name, &nick, &description, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Skill with UUID %s not found", skillUUID),
		})
		return
	}

	if err != nil {
		log.Printf("Error querying skill: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get skill: %v", err),
		})
		return
	}

	skill := map[string]interface{}{
		"uuid":        uuid,
		"name":        name,
		"nick":        nick,
		"description": description,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill retrieved successfully",
		Data:    skill,
	})
}

func toggleProxyWorkerEnabledHandler(c *gin.Context) {
	var req struct {
		ID      string `json:"id"`
		Enabled bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	if req.ID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing worker id",
		})
		return
	}

	projectRoot := getProjectRoot()
	proxyConfigFile := filepath.Join(projectRoot, "data", "proxy_worker_registrations.json")
	log.Printf("Toggling proxy worker enabled in: %s", proxyConfigFile)

	data, err := os.ReadFile(proxyConfigFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to read proxy worker config: %v", err),
		})
		return
	}

	var proxyWorkers []map[string]interface{}
	if err := json.Unmarshal(data, &proxyWorkers); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to parse proxy worker config: %v", err),
		})
		return
	}

	updated := false
	var updatedWorker map[string]interface{}
	for i, worker := range proxyWorkers {
		if workerID, ok := worker["ID"].(string); ok && workerID == req.ID {
			proxyWorkers[i]["Enabled"] = req.Enabled
			updatedWorker = proxyWorkers[i]
			updated = true
			break
		}
	}

	if !updated {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Proxy worker with ID %s not found", req.ID),
		})
		return
	}

	if req.Enabled {
		log.Printf("Proxy worker %s is being enabled, registering to worker registry and capability center", req.ID)

		name, _ := updatedWorker["Name"].(string)
		workerType, _ := updatedWorker["WorkerType"].(string)
		options, _ := updatedWorker["Options"].(map[string]interface{})

		if options != nil {
			apiAddress, _ := options["ApiAddress"].(string)
			description, _ := options["Description"].(string)
			paramConfigs, _ := options["ParamConfigs"].([]interface{})

			workerOptions := workers.DefaultWorkerOptions()
			workerOptions.Description = description
			workerOptions.ApiAddress = apiAddress
			workerOptions.Project = "proxyWorkers"
			workerOptions.WorkerNick = req.ID

			var configs []capability.ParamConfig
			if paramConfigs != nil {
				for _, pc := range paramConfigs {
					if pcMap, ok := pc.(map[string]interface{}); ok {
						config := capability.ParamConfig{}
						if title, ok := pcMap["Title"].(string); ok {
							config.Title = title
						} else if title, ok := pcMap["title"].(string); ok {
							config.Title = title
						}
						if nick, ok := pcMap["Nick"].(string); ok {
							config.Nick = nick
						} else if nick, ok := pcMap["nick"].(string); ok {
							config.Nick = nick
						}
						if mode, ok := pcMap["Mode"].(string); ok {
							config.Mode = mode
						} else if mode, ok := pcMap["mode"].(string); ok {
							config.Mode = mode
						}
						if must, ok := pcMap["Must"].(bool); ok {
							config.Must = must
						} else if must, ok := pcMap["must"].(bool); ok {
							config.Must = must
						}
						if placeholder, ok := pcMap["Placeholder"].(string); ok {
							config.Placeholder = placeholder
						} else if placeholder, ok := pcMap["placeholder"].(string); ok {
							config.Placeholder = placeholder
						}
						if value, ok := pcMap["Value"].(string); ok {
							config.Value = value
						} else if value, ok := pcMap["value"].(string); ok {
							config.Value = value
						}
						configs = append(configs, config)
					}
				}
				workerOptions.ParamConfigs = configs
			}

			if err := workers.RegisterWorkerWithOptions(req.ID, name, workerType, workerOptions); err != nil {
				log.Printf("Warning: failed to register proxy worker to worker registry: %v", err)
			} else {
				log.Printf("Proxy worker %s registered to worker registry successfully", req.ID)
			}

			workerID := fmt.Sprintf("proxyWorkers.%s@1.0.0", req.ID)
			cap := &capability.Capability{
				ID:               workerID,
				Name:             name,
				Address:          "localhost",
				Type:             capability.CapabilityType(workerType),
				Language:         capability.Language("go"),
				Version:          "1.0.0",
				Description:      description,
				InvokeMethod:     "local",
				Timeout:          30,
				ProxyType:        "proxy",
				VirtualIDs:       []string{},
				VirtualIDDescMap: make(map[string]string),
				Labels: map[string]string{
					"env":         "development",
					"component":   "worker",
					"api_address": apiAddress,
					"project":     "proxyWorkers",
					"workerNick":  req.ID,
				},
				Tags:         []string{"workflow", "worker", workerType},
				Owner:        "",
				Permission:   "",
				AiPermission: "",
				Quota:        0,
				MaxInstances: 0,
				ParamConfigs: configs,
				Online:       true,
				Heartbeat:    time.Now(),
			}

			if err := capClient.Register(context.Background(), cap); err != nil {
				log.Printf("Warning: failed to register proxy worker to capability center: %v", err)
			} else {
				log.Printf("Proxy worker %s registered to capability center successfully", req.ID)
			}
		}
	} else {
		workerID := fmt.Sprintf("proxyWorkers.%s@1.0.0", req.ID)
		log.Printf("Proxy worker %s is being disabled, unregistering from capability center and removing from heartbeat list", req.ID)
		if err := capClient.Unregister(context.Background(), workerID); err != nil {
			log.Printf("Warning: failed to unregister proxy worker from capability center: %v", err)
		} else {
			log.Printf("Proxy worker %s unregistered from capability center successfully", req.ID)
		}
		workers.RemoveWorkerFromHeartbeat(workerID)
	}

	updatedData, err := json.MarshalIndent(proxyWorkers, "", "  ")
	if err != nil {
		log.Printf("Warning: failed to marshal proxy worker config: %v", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to marshal proxy worker config: %v", err),
		})
		return
	}

	if err := os.WriteFile(proxyConfigFile, updatedData, 0644); err != nil {
		log.Printf("Warning: failed to write proxy worker config: %v", err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to write proxy worker config: %v", err),
		})
		return
	}

	log.Printf("Proxy worker %s enabled state toggled successfully", req.ID)

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Proxy worker %s %s successfully", req.ID, map[bool]string{true: "enabled", false: "disabled"}[req.Enabled]),
	})
}

func getProxyWorkListHandler(c *gin.Context) {
	projectRoot := getProjectRoot()
	proxyConfigFile := filepath.Join(projectRoot, "data", "proxy_worker_registrations.json")
	log.Printf("Reading proxy worker list from: %s", proxyConfigFile)

	data, err := os.ReadFile(proxyConfigFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, DiscoverResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read proxy worker config: %v", err),
		})
		return
	}

	var proxyWorkers []map[string]interface{}
	if err := json.Unmarshal(data, &proxyWorkers); err != nil {
		c.JSON(http.StatusInternalServerError, DiscoverResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse proxy worker config: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       "Proxy work list retrieved successfully from proxy_worker_registrations.json",
		"proxy_workers": proxyWorkers,
		"total":         len(proxyWorkers),
	})
}

func executeWorkerHandler(c *gin.Context) {
	var req ExecuteWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ExecuteWorkerResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   fmt.Sprintf("Failed to decode request: %v", err),
		})
		return
	}

	if req.WorkerID == "" {
		c.JSON(http.StatusBadRequest, ExecuteWorkerResponse{
			Success: false,
			Message: "Invalid request",
			Error:   "Worker ID is required",
		})
		return
	}

	apiAddress := ""

	filter := &capability.CapabilityFilter{
		OnlineOnly: true,
	}

	capabilities, err := capClient.Discover(context.Background(), req.WorkerID, filter)
	if err == nil && len(capabilities) > 0 {
		worker := capabilities[0]
		if worker.Labels != nil && worker.Labels["api_address"] != "" {
			apiAddress = worker.Labels["api_address"]
		}
	}

	if apiAddress == "" {
		c.JSON(http.StatusNotFound, ExecuteWorkerResponse{
			Success: false,
			Message: "Worker not found",
			Error:   "Worker with ID " + req.WorkerID + " not found or offline",
		})
		return
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
			c.JSON(http.StatusInternalServerError, ExecuteWorkerResponse{
				Success: false,
				Message: "Failed to marshal params",
				Error:   fmt.Sprintf("Failed to marshal parameters: %v", err),
			})
			return
		}
		log.Printf("Request Body: %s", string(requestBody))
	}

	externalReq, err := http.NewRequest(http.MethodPost, apiAddress, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, ExecuteWorkerResponse{
			Success: false,
			Message: "Failed to create request",
			Error:   fmt.Sprintf("Failed to create HTTP request: %v", err),
		})
		return
	}

	externalReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	log.Printf("Sending request to worker API...")
	externalResp, err := client.Do(externalReq)
	if err != nil {
		log.Printf("Error executing worker: %v", err)
		c.JSON(http.StatusInternalServerError, ExecuteWorkerResponse{
			Success: false,
			Message: "Failed to execute worker",
			Error:   fmt.Sprintf("Failed to call worker API: %v", err),
			Data: map[string]interface{}{
				"api_address":    apiAddress,
				"request_params": req.Params,
				"error_details":  err.Error(),
			},
		})
		return
	}
	defer externalResp.Body.Close()

	externalRespBody, err := io.ReadAll(externalResp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		c.JSON(http.StatusInternalServerError, ExecuteWorkerResponse{
			Success: false,
			Message: "Failed to read response",
			Error:   fmt.Sprintf("Failed to read worker response: %v", err),
		})
		return
	}

	log.Printf("Worker Response Status: %d", externalResp.StatusCode)
	log.Printf("Worker Response Body: %s", string(externalRespBody))

	if externalResp.StatusCode != http.StatusOK {
		log.Printf("Worker execution failed with status: %d", externalResp.StatusCode)
		c.JSON(http.StatusInternalServerError, ExecuteWorkerResponse{
			Success: false,
			Message: "Worker execution failed",
			Error:   fmt.Sprintf("Worker API returned status %d: %s", externalResp.StatusCode, string(externalRespBody)),
			Data: map[string]interface{}{
				"api_address":     apiAddress,
				"request_params":  req.Params,
				"response_status": externalResp.StatusCode,
				"response_body":   string(externalRespBody),
			},
		})
		return
	}

	var workerResult map[string]interface{}
	if err := json.Unmarshal(externalRespBody, &workerResult); err != nil {
		log.Printf("Worker executed successfully, but response parsing failed: %v", err)
		c.JSON(http.StatusOK, ExecuteWorkerResponse{
			Success: true,
			Message: "Worker executed successfully",
			Data: map[string]interface{}{
				"api_address":    apiAddress,
				"request_params": req.Params,
				"response_body":  string(externalRespBody),
				"parsing_error":  err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, ExecuteWorkerResponse{
		Success: true,
		Message: "Worker executed successfully",
		Data:    workerResult,
	})
}

func getAbilityHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing ability id",
		})
		return
	}

	cap, err := capClient.GetMetadata(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get ability: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Ability retrieved successfully",
		"ability": cap,
	})
}

type AbilityVectorSearchResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Data    []AbilityVectorSearchItem `json:"data,omitempty"`
}

type AbilityVectorSearchItem struct {
	ID          string  `json:"id"`
	Project     string  `json:"project"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"`
}

func vectorSearchHandler(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, AbilityVectorSearchResponse{
			Success: false,
			Message: "text参数不能为空",
		})
		return
	}

	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	project := c.Query("project")

	log.Printf("向量搜索请求: text=%s, limit=%d, project=%s", text, limit, project)

	// 生成查询文本的嵌入向量
	var queryEmbedding []float32
	if embeddingService != nil {
		var err error
		queryEmbedding, err = embeddingService.GetEmbedding(text)
		if err != nil {
			log.Printf("Warning: failed to generate query embedding: %v", err)
			c.JSON(http.StatusInternalServerError, AbilityVectorSearchResponse{
				Success: false,
				Message: fmt.Sprintf("生成查询向量失败: %v", err),
			})
			return
		}
		log.Printf("Successfully generated query embedding with dimension: %d", len(queryEmbedding))
		// 将向量扩展为1024维
		if len(queryEmbedding) < 1024 {
			extendedEmbedding := make([]float32, 1024)
			copy(extendedEmbedding, queryEmbedding)
			queryEmbedding = extendedEmbedding
			log.Printf("Extended query embedding to 1024 dimensions")
		}
	} else {
		log.Printf("EmbeddingService not available, using empty vector")
		c.JSON(http.StatusInternalServerError, AbilityVectorSearchResponse{
			Success: false,
			Message: "嵌入服务不可用",
		})
		return
	}

	// 构建SQL查询
	var sql string
	var args []interface{}

	if project != "" {
		sql = `
			SELECT id, project, name, description, embedding
			FROM ability_template
			WHERE project = $1 AND is_enabled = true
		`
		args = append(args, project)
	} else {
		sql = `
			SELECT id, project, name, description, embedding
			FROM ability_template
			WHERE is_enabled = true
		`
	}

	// 执行SQL查询
	rows, err := pgDB.Query(sql, args...)
	if err != nil {
		log.Printf("Database query failed: %v", err)
		c.JSON(http.StatusInternalServerError, AbilityVectorSearchResponse{
			Success: false,
			Message: fmt.Sprintf("数据库查询失败: %v", err),
		})
		return
	}
	defer rows.Close()

	// 计算相似度并排序
	type scoredItem struct {
		item     AbilityVectorSearchItem
		distance float64
	}

	var scoredItems []scoredItem

	for rows.Next() {
		var id, project, name, description string
		var embedding []byte

		err := rows.Scan(&id, &project, &name, &description, &embedding)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// 解析嵌入向量
		var dbEmbedding []float32
		if len(embedding) > 0 {
			// 尝试从JSON解析
			if err := json.Unmarshal(embedding, &dbEmbedding); err != nil {
				log.Printf("Error unmarshaling embedding: %v", err)
				// 如果解析失败，使用空的1024维向量
				dbEmbedding = make([]float32, 1024)
			} else if len(dbEmbedding) < 1024 {
				// 如果向量维度不足，扩展为1024维
				extendedEmbedding := make([]float32, 1024)
				copy(extendedEmbedding, dbEmbedding)
				dbEmbedding = extendedEmbedding
			}
		} else {
			// 如果embedding字段为空，使用空的1024维向量
			dbEmbedding = make([]float32, 1024)
		}

		// 计算相似度（欧几里得距离）
		distance := calculateDistance(queryEmbedding, dbEmbedding)

		scoredItems = append(scoredItems, scoredItem{
			item: AbilityVectorSearchItem{
				ID:          id,
				Project:     project,
				Name:        name,
				Description: description,
				Distance:    distance,
			},
			distance: distance,
		})
	}

	// 按距离排序
	sort.Slice(scoredItems, func(i, j int) bool {
		return scoredItems[i].distance < scoredItems[j].distance
	})

	// 限制返回数量
	var items []AbilityVectorSearchItem
	for i, scored := range scoredItems {
		if i >= limit {
			break
		}
		items = append(items, scored.item)
	}

	c.JSON(http.StatusOK, AbilityVectorSearchResponse{
		Success: true,
		Message: fmt.Sprintf("找到 %d 个结果", len(items)),
		Data:    items,
	})
	log.Printf("向量搜索完成: 返回 %d 个结果", len(items))
}

// textSearchHandler 文本搜索能力
func textSearchHandler(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, AbilityVectorSearchResponse{
			Success: false,
			Message: "text参数不能为空",
		})
		return
	}

	limitStr := c.Query("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	project := c.Query("project")

	log.Printf("文本搜索请求: text=%s, limit=%d, project=%s", text, limit, project)

	// 构建SQL查询 - 使用LIKE进行文本匹配
	var sql string
	var args []interface{}

	searchPattern := "%" + text + "%"

	if project != "" {
		sql = `
			SELECT id, project, name, description, embedding
			FROM ability_template
			WHERE project = $1 AND is_enabled = true AND (name LIKE $2 OR description LIKE $2)
			LIMIT $3
		`
		args = append(args, project, searchPattern, limit)
	} else {
		sql = `
			SELECT id, project, name, description, embedding
			FROM ability_template
			WHERE is_enabled = true AND (name LIKE $1 OR description LIKE $1)
			LIMIT $2
		`
		args = append(args, searchPattern, limit)
	}

	// 执行SQL查询
	rows, err := pgDB.Query(sql, args...)
	if err != nil {
		log.Printf("数据库查询失败: %v", err)
		c.JSON(http.StatusInternalServerError, AbilityVectorSearchResponse{
			Success: false,
			Message: fmt.Sprintf("数据库查询失败: %v", err),
		})
		return
	}
	defer rows.Close()

	var items []AbilityVectorSearchItem

	for rows.Next() {
		var id, project, name, description string
		var embedding []byte

		err := rows.Scan(&id, &project, &name, &description, &embedding)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		items = append(items, AbilityVectorSearchItem{
			ID:          id,
			Project:     project,
			Name:        name,
			Description: description,
			Distance:    0,
		})
	}

	c.JSON(http.StatusOK, AbilityVectorSearchResponse{
		Success: true,
		Message: fmt.Sprintf("找到 %d 个结果", len(items)),
		Data:    items,
	})
	log.Printf("文本搜索完成: 返回 %d 个结果", len(items))
}

// calculateDistance 计算两个向量之间的欧几里得距离
func calculateDistance(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range vec1 {
		diff := float64(vec1[i] - vec2[i])
		sum += diff * diff
	}

	return math.Sqrt(sum)
}
