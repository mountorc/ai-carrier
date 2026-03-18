package ability

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trae/autoFlow/common/capability"
	"github.com/trae/autoFlow/mounts/workflow/workers"
	"github.com/trae/autoFlow/carriercore/workflow"
)

// ProxyAbilityRegisterRequest 定义添加代理能力的请求结构
type ProxyAbilityRegisterRequest struct {
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
	Project      string                   `json:"project"`
	UUIDProject  string                   `json:"uuid_project"`
}

// addProxyAbilityHandler 添加代理能力
func addProxyAbilityHandler(c *gin.Context) {
	var req ProxyAbilityRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	// 注册到 worker 注册中心和能力中心（如果启用）
	if req.Enabled {
		options := workers.DefaultWorkerOptions()
		options.Description = req.Description
		options.ApiAddress = req.ApiAddress
		options.ParamConfigs = req.ParamConfigs
		project := req.Project
		if project == "" {
			project = "proxyWorkers"
		}
		options.Project = project
		options.WorkerNick = req.ID

		if err := workers.RegisterWorkerWithOptions(req.ID, req.Name, req.Type, options); err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: fmt.Sprintf("Failed to register proxy ability to worker registry: %v", err),
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
		project = req.Project
		if project == "" {
			project = "proxyWorkers"
		}
		abilityID := fmt.Sprintf("%s.%s@%s", project, req.ID, version)

		cap := &capability.Capability{
			ID:               abilityID,
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
				"project":     project,
				"uuidProject": req.UUIDProject,
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
			log.Printf("Warning: failed to register proxy ability to capability center: %v", err)
		} else {
			log.Printf("Proxy ability %s registered to capability center successfully", req.ID)
		}
	} else {
		log.Printf("Proxy ability %s is disabled, skipping registration", req.ID)
	}

	// 存储到数据库
	paramConfigsJSON, err := json.Marshal(req.ParamConfigs)
	if err != nil {
		paramConfigsJSON = []byte("[]")
	}

	_, err = pgDB.Exec(`
		INSERT INTO ability_proxy (id, name, worker_type, enabled, description, api_address, param_configs, project, uuid_project, last_registered)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			name = $2,
			worker_type = $3,
			enabled = $4,
			description = $5,
			api_address = $6,
			param_configs = $7,
			project = $8,
			uuid_project = $9,
			last_registered = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
	`, req.ID, req.Name, req.Type, req.Enabled, req.Description, req.ApiAddress, paramConfigsJSON, req.Project, req.UUIDProject)

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to save proxy ability to database: %v", err),
		})
		return
	}

	log.Printf("Proxy ability %s saved to database successfully", req.ID)

	// 同时将代理能力注册到数据库的ability_template表中
	if workflow.PostgresStore != nil {
		paramConfigsJSON, _ := json.Marshal(req.ParamConfigs)
		tagsJSON, _ := json.Marshal([]string{"workflow", "worker", req.Type})

		// 生成向量嵌入
		var embedding []float32
		if workflow.EmbeddingService != nil {
			// 为能力生成嵌入向量，将需要的字段合到一起
			embeddingText := req.Name + " " + req.Description + " " + req.Type + " " + req.Language + " " + strings.Join([]string{"workflow", "worker", req.Type}, " ")
			// 添加param_configs中的字段
			for _, param := range req.ParamConfigs {
				embeddingText += " " + param.Title + " " + param.Nick + " " + param.Mode
			}
			// 添加api_address
			if req.ApiAddress != "" {
				embeddingText += " " + req.ApiAddress
			}
			log.Printf("Generating embedding for proxy ability text: %s", embeddingText)
			var err error
			embedding, err = workflow.EmbeddingService.GetEmbedding(embeddingText)
			if err != nil {
				log.Printf("Warning: failed to generate embedding: %v", err)
				// 如果生成嵌入失败，使用空的1024维向量
				embedding = make([]float32, 1024)
			} else if len(embedding) != 1024 {
				// 如果嵌入维度不正确，使用空的1024维向量
				log.Printf("Warning: embedding dimension mismatch, expected 1024, got %d, using empty vector", len(embedding))
				embedding = make([]float32, 1024)
			} else {
				log.Printf("Successfully generated embedding with dimension: %d", len(embedding))
			}
		} else {
			// 如果嵌入服务不可用，使用空的1024维向量
			embedding = make([]float32, 1024)
			log.Printf("EmbeddingService not available, using empty vector")
		}

		project := req.Project
		if project == "" {
			project = "proxyWorkers"
		}
		version := req.Version
		if version == "" {
			version = "1.0.0"
		}

		err = workflow.PostgresStore.RegisterWorkerToTemplate(
			context.Background(),
			project,
			req.ID,
			req.Name,
			req.Description,
			req.Type,
			version,
			json.RawMessage("{}"), // input_schema
			json.RawMessage("{}"), // output_schema
			tagsJSON,
			paramConfigsJSON,
			"qwen",    // embedding_type
			embedding, // 向量嵌入
		)
		if err != nil {
			log.Printf("Warning: failed to register proxy ability to template: %v", err)
		} else {
			log.Printf("Proxy ability registered to ability_template successfully")
		}
	} else {
		log.Printf("Warning: workflow.PostgresStore is nil, cannot register proxy ability to template")
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Proxy ability registered successfully",
	})
}

// getProxyAbilityHandler 获取代理能力详情
func getProxyAbilityHandler(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing ability id",
		})
		return
	}

	// 解析完整格式的 ID（如 project.workerId@version），只保留 workerId 部分
	// 例如：proxyWorkers.worker_api_filters_62@1.0.0 -> worker_api_filters_62
	if parts := strings.Split(id, "@"); len(parts) > 1 {
		idPart := parts[0]
		if dotParts := strings.Split(idPart, "."); len(dotParts) > 1 {
			id = strings.Join(dotParts[1:], ".")
		} else {
			id = idPart
		}
	}
	// 如果只包含点号的部分，也解析
	if strings.Contains(id, ".") && !strings.Contains(id, "@") {
		if dotParts := strings.Split(id, "."); len(dotParts) > 1 {
			id = strings.Join(dotParts[1:], ".")
		}
	}

	var name, workerType, description, apiAddress, project, uuidProject string
	var enabled bool
	var lastRegistered, createdAt, updatedAt time.Time
	var paramConfigsJSON []byte

	err := pgDB.QueryRow(`
		SELECT id, name, worker_type, enabled, description, api_address, param_configs, project, uuid_project, last_registered, created_at, updated_at
		FROM ability_proxy WHERE id = $1
	`, id).Scan(&id, &name, &workerType, &enabled, &description, &apiAddress, &paramConfigsJSON, &project, &uuidProject, &lastRegistered, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Proxy ability with ID %s not found", id),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to read proxy ability from database: %v", err),
		})
		return
	}

	var paramConfigs []capability.ParamConfig
	if len(paramConfigsJSON) > 0 {
		if err := json.Unmarshal(paramConfigsJSON, &paramConfigs); err != nil {
			paramConfigs = []capability.ParamConfig{}
		}
	}

	targetAbility := map[string]interface{}{
		"ID":          id,
		"Name":        name,
		"WorkerType":  workerType,
		"Enabled":     enabled,
		"Project":     project,
		"UUIDProject": uuidProject,
		"Options": map[string]interface{}{
			"Description":  description,
			"ApiAddress":   apiAddress,
			"ParamConfigs": paramConfigs,
		},
		"LastRegistered": lastRegistered.Format(time.RFC3339),
		"CreatedAt":      createdAt.Format(time.RFC3339),
		"UpdatedAt":      updatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Proxy ability %s retrieved successfully", id),
		"proxy_ability": targetAbility,
	})
}

// updateProxyAbilityHandler 更新代理能力
func updateProxyAbilityHandler(c *gin.Context) {
	var updateRequest map[string]interface{}
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	log.Printf("=== updateProxyAbilityHandler called ===")
	log.Printf("Received update request: %+v", updateRequest)

	id, ok := updateRequest["id"].(string)
	if !ok || id == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing ability id",
		})
		return
	}

	// 解析完整格式的 ID（如 project.workerId@version），只保留 workerId 部分
	if parts := strings.Split(id, "@"); len(parts) > 1 {
		idPart := parts[0]
		if dotParts := strings.Split(idPart, "."); len(dotParts) > 1 {
			id = strings.Join(dotParts[1:], ".")
		} else {
			id = idPart
		}
	}
	// 如果只包含点号的部分，也解析
	if strings.Contains(id, ".") && !strings.Contains(id, "@") {
		if dotParts := strings.Split(id, "."); len(dotParts) > 1 {
			id = strings.Join(dotParts[1:], ".")
		}
	}

	log.Printf("Updating ability with ID: %s", id)

	// 先从数据库获取当前状态
	var currentEnabled bool
	var currentName, currentWorkerType, currentProject, currentUUIDProject string
	var currentDescription, currentApiAddress string
	var currentParamConfigsJSON []byte

	err := pgDB.QueryRow(`
		SELECT name, worker_type, enabled, description, api_address, param_configs, project, uuid_project
		FROM ability_proxy WHERE id = $1
	`, id).Scan(&currentName, &currentWorkerType, &currentEnabled, &currentDescription, &currentApiAddress, &currentParamConfigsJSON, &currentProject, &currentUUIDProject)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Proxy ability with ID %s not found", id),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to read proxy ability from database: %v", err),
		})
		return
	}

	// 更新数据库中的记录
	if name, ok := updateRequest["name"].(string); ok {
		currentName = name
	}
	if project, ok := updateRequest["project"].(string); ok {
		currentProject = project
	}
	if uuidProject, ok := updateRequest["uuid_project"].(string); ok {
		currentUUIDProject = uuidProject
	}
	if options, ok := updateRequest["options"].(map[string]interface{}); ok {
		if apiAddress, ok := options["api_address"].(string); ok {
			currentApiAddress = apiAddress
		}
		if description, ok := options["description"].(string); ok {
			currentDescription = description
		}
		if paramConfigs, ok := options["param_configs"].([]interface{}); ok {
			paramConfigsJSON, err := json.Marshal(paramConfigs)
			if err == nil {
				currentParamConfigsJSON = paramConfigsJSON
			}
		}
	}

	_, err = pgDB.Exec(`
		UPDATE ability_proxy 
		SET name = $1, description = $2, api_address = $3, param_configs = $4, project = $5, uuid_project = $6, updated_at = CURRENT_TIMESTAMP, last_registered = CURRENT_TIMESTAMP
		WHERE id = $7
	`, currentName, currentDescription, currentApiAddress, currentParamConfigsJSON, currentProject, currentUUIDProject, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update proxy ability in database: %v", err),
		})
		return
	}

	// 如果启用，则重新注册到 worker 注册中心和能力中心
	if currentEnabled {
		log.Printf("Proxy ability %s is enabled, re-registering to worker registry and capability center", id)

		var paramConfigs []capability.ParamConfig
		if len(currentParamConfigsJSON) > 0 {
			if err := json.Unmarshal(currentParamConfigsJSON, &paramConfigs); err != nil {
				paramConfigs = []capability.ParamConfig{}
			}
		}

		workerOptions := workers.DefaultWorkerOptions()
		workerOptions.Description = currentDescription
		workerOptions.ApiAddress = currentApiAddress
		project := currentProject
		if project == "" {
			project = "proxyWorkers"
		}
		workerOptions.Project = project
		workerOptions.WorkerNick = id
		workerOptions.ParamConfigs = paramConfigs

		if err := capClient.Unregister(context.Background(), id); err != nil {
			log.Printf("Warning: failed to unregister old proxy ability from capability center: %v", err)
		}

		if err := workers.RegisterWorkerWithOptions(id, currentName, currentWorkerType, workerOptions); err != nil {
			log.Printf("Warning: failed to re-register proxy ability to worker registry: %v", err)
		} else {
			log.Printf("Proxy ability %s re-registered to worker registry successfully", id)
		}

		cap := &capability.Capability{
			ID:               id,
			Name:             currentName,
			Address:          "localhost",
			Type:             capability.CapabilityType(currentWorkerType),
			Language:         capability.Language("go"),
			Version:          "1.0.0",
			Description:      currentDescription,
			InvokeMethod:     "local",
			Timeout:          30,
			ProxyType:        "proxy",
			VirtualIDs:       []string{},
			VirtualIDDescMap: make(map[string]string),
			Labels: map[string]string{
				"env":         "development",
				"component":   "worker",
				"api_address": currentApiAddress,
				"project":     project,
				"uuidProject": currentUUIDProject,
			},
			Tags:         []string{"workflow", "worker", currentWorkerType},
			Owner:        "",
			Permission:   "",
			AiPermission: "",
			Quota:        0,
			MaxInstances: 0,
			ParamConfigs: paramConfigs,
			Online:       true,
			Heartbeat:    time.Now(),
		}

		if err := capClient.Register(context.Background(), cap); err != nil {
			log.Printf("Warning: failed to re-register proxy ability to capability center: %v", err)
		} else {
			log.Printf("Proxy ability %s re-registered to capability center successfully", id)
		}
	}

	// 同时更新ability_template表中的embedding
	if workflow.PostgresStore != nil {
		paramConfigsJSON, _ := json.Marshal(currentParamConfigsJSON)
		tagsJSON, _ := json.Marshal([]string{"workflow", "worker", currentWorkerType})

		// 生成向量嵌入
		var embedding []float32
		if workflow.EmbeddingService != nil {
			// 为能力生成嵌入向量，将需要的字段合到一起
			embeddingText := currentName + " " + currentDescription + " " + currentWorkerType + " " + strings.Join([]string{"workflow", "worker", currentWorkerType}, " ")
			// 添加param_configs中的字段
			var paramConfigs []capability.ParamConfig
			if len(currentParamConfigsJSON) > 0 {
				if err := json.Unmarshal(currentParamConfigsJSON, &paramConfigs); err != nil {
					paramConfigs = []capability.ParamConfig{}
				}
			}
			for _, param := range paramConfigs {
				embeddingText += " " + param.Title + " " + param.Nick + " " + param.Mode
			}
			// 添加api_address
			if currentApiAddress != "" {
				embeddingText += " " + currentApiAddress
			}
			log.Printf("Generating embedding for updated proxy ability text: %s", embeddingText)
			var err error
			embedding, err = workflow.EmbeddingService.GetEmbedding(embeddingText)
			if err != nil {
				log.Printf("Warning: failed to generate embedding: %v", err)
				// 如果生成嵌入失败，使用空的1024维向量
				embedding = make([]float32, 1024)
			} else if len(embedding) != 1024 {
				// 如果嵌入维度不正确，使用空的1024维向量
				log.Printf("Warning: embedding dimension mismatch, expected 1024, got %d, using empty vector", len(embedding))
				embedding = make([]float32, 1024)
			} else {
				log.Printf("Successfully generated embedding with dimension: %d", len(embedding))
			}
		} else {
			// 如果嵌入服务不可用，使用空的1024维向量
			embedding = make([]float32, 1024)
			log.Printf("EmbeddingService not available, using empty vector")
		}

		project := currentProject
		if project == "" {
			project = "proxyWorkers"
		}

		err = workflow.PostgresStore.RegisterWorkerToTemplate(
			context.Background(),
			project,
			id,
			currentName,
			currentDescription,
			currentWorkerType,
			"1.0.0",
			json.RawMessage("{}"), // input_schema
			json.RawMessage("{}"), // output_schema
			tagsJSON,
			paramConfigsJSON,
			"qwen",    // embedding_type
			embedding, // 向量嵌入
		)
		if err != nil {
			log.Printf("Warning: failed to update proxy ability template: %v", err)
		} else {
			log.Printf("Proxy ability template updated successfully with new embedding")
		}
	} else {
		log.Printf("Warning: workflow.PostgresStore is nil, cannot update proxy ability template")
	}

	log.Printf("Proxy ability %s updated successfully", id)

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Proxy ability %s updated successfully", id),
	})
}

// toggleProxyAbilityEnabledHandler 切换代理能力启用状态
func toggleProxyAbilityEnabledHandler(c *gin.Context) {
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
			Message: "Missing ability id",
		})
		return
	}

	// 解析完整格式的 ID（如 project.workerId@version），只保留 workerId 部分
	id := req.ID
	if parts := strings.Split(id, "@"); len(parts) > 1 {
		idPart := parts[0]
		if dotParts := strings.Split(idPart, "."); len(dotParts) > 1 {
			id = strings.Join(dotParts[1:], ".")
		} else {
			id = idPart
		}
	}
	// 如果只包含点号的部分，也解析
	if strings.Contains(id, ".") && !strings.Contains(id, "@") {
		if dotParts := strings.Split(id, "."); len(dotParts) > 1 {
			id = strings.Join(dotParts[1:], ".")
		}
	}
	req.ID = id

	// 先获取当前状态
	var currentName, currentWorkerType, currentDescription, currentApiAddress string
	var paramConfigsJSON []byte

	err := pgDB.QueryRow(`
		SELECT name, worker_type, description, api_address, param_configs
		FROM ability_proxy WHERE id = $1
	`, req.ID).Scan(&currentName, &currentWorkerType, &currentDescription, &currentApiAddress, &paramConfigsJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Proxy ability with ID %s not found", req.ID),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to read proxy ability from database: %v", err),
		})
		return
	}

	// 更新数据库
	_, err = pgDB.Exec(`
		UPDATE ability_proxy 
		SET enabled = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, req.Enabled, req.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to update proxy ability enabled status: %v", err),
		})
		return
	}

	// 根据状态注册或注销
	if req.Enabled {
		log.Printf("Proxy ability %s is being enabled, registering to worker registry and capability center", req.ID)

		var paramConfigs []capability.ParamConfig
		if len(paramConfigsJSON) > 0 {
			if err := json.Unmarshal(paramConfigsJSON, &paramConfigs); err != nil {
				paramConfigs = []capability.ParamConfig{}
			}
		}

		workerOptions := workers.DefaultWorkerOptions()
		workerOptions.Description = currentDescription
		workerOptions.ApiAddress = currentApiAddress
		workerOptions.Project = "proxyWorkers"
		workerOptions.WorkerNick = req.ID
		workerOptions.ParamConfigs = paramConfigs

		if err := workers.RegisterWorkerWithOptions(req.ID, currentName, currentWorkerType, workerOptions); err != nil {
			log.Printf("Warning: failed to register proxy ability to worker registry: %v", err)
		} else {
			log.Printf("Proxy ability %s registered to worker registry successfully", req.ID)
		}

		abilityID := fmt.Sprintf("proxyWorkers.%s@1.0.0", req.ID)
		cap := &capability.Capability{
			ID:               abilityID,
			Name:             currentName,
			Address:          "localhost",
			Type:             capability.CapabilityType(currentWorkerType),
			Language:         capability.Language("go"),
			Version:          "1.0.0",
			Description:      currentDescription,
			InvokeMethod:     "local",
			Timeout:          30,
			ProxyType:        "proxy",
			VirtualIDs:       []string{},
			VirtualIDDescMap: make(map[string]string),
			Labels: map[string]string{
				"env":         "development",
				"component":   "worker",
				"api_address": currentApiAddress,
				"project":     "proxyWorkers",
				"workerNick":  req.ID,
			},
			Tags:         []string{"workflow", "worker", currentWorkerType},
			Owner:        "",
			Permission:   "",
			AiPermission: "",
			Quota:        0,
			MaxInstances: 0,
			ParamConfigs: paramConfigs,
			Online:       true,
			Heartbeat:    time.Now(),
		}

		if err := capClient.Register(context.Background(), cap); err != nil {
			log.Printf("Warning: failed to register proxy ability to capability center: %v", err)
		} else {
			log.Printf("Proxy ability %s registered to capability center successfully", req.ID)
		}
	} else {
		abilityID := fmt.Sprintf("proxyWorkers.%s@1.0.0", req.ID)
		log.Printf("Proxy ability %s is being disabled, unregistering from capability center and removing from heartbeat list", req.ID)
		if err := capClient.Unregister(context.Background(), abilityID); err != nil {
			log.Printf("Warning: failed to unregister proxy ability from capability center: %v", err)
		} else {
			log.Printf("Proxy ability %s unregistered from capability center successfully", req.ID)
		}
		workers.RemoveWorkerFromHeartbeat(abilityID)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Proxy ability %s enabled status toggled successfully", req.ID),
	})
}

// getProxyAbilityListHandler 获取代理能力列表
func getProxyAbilityListHandler(c *gin.Context) {
	rows, err := pgDB.Query(`
		SELECT id, name, worker_type, enabled, description, api_address, param_configs, project, uuid_project, last_registered, created_at, updated_at
		FROM ability_proxy
		ORDER BY created_at DESC
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to query proxy ability list: %v", err),
		})
		return
	}
	defer rows.Close()

	var proxyAbilities []map[string]interface{}

	for rows.Next() {
		var id, name, workerType, description, apiAddress, project, uuidProject string
		var enabled bool
		var lastRegistered, createdAt, updatedAt time.Time
		var paramConfigsJSON []byte

		err := rows.Scan(&id, &name, &workerType, &enabled, &description, &apiAddress, &paramConfigsJSON, &project, &uuidProject, &lastRegistered, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Failed to scan proxy ability row: %v", err)
			continue
		}

		var paramConfigs []capability.ParamConfig
		if len(paramConfigsJSON) > 0 {
			if err := json.Unmarshal(paramConfigsJSON, &paramConfigs); err != nil {
				paramConfigs = []capability.ParamConfig{}
			}
		}

		proxyAbility := map[string]interface{}{
			"ID":          id,
			"Name":        name,
			"WorkerType":  workerType,
			"Enabled":     enabled,
			"Project":     project,
			"UUIDProject": uuidProject,
			"Options": map[string]interface{}{
				"Description":  description,
				"ApiAddress":   apiAddress,
				"ParamConfigs": paramConfigs,
			},
			"LastRegistered": lastRegistered.Format(time.RFC3339),
			"CreatedAt":      createdAt.Format(time.RFC3339),
			"UpdatedAt":      updatedAt.Format(time.RFC3339),
		}

		proxyAbilities = append(proxyAbilities, proxyAbility)
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":         true,
		"message":         "Proxy ability list retrieved successfully",
		"proxy_abilities": proxyAbilities,
	})
}
