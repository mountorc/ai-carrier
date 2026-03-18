package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trae/autoFlow/mounts/workflow"
	"github.com/trae/autoFlow/mounts/workflow/types"
	"github.com/trae/autoFlow/carriercore/common/db"
	sqlutil "github.com/trae/autoFlow/carriercore/common/sql"
)

// 列表实例
func listInstances(c *gin.Context) {
	ctx := context.Background()

	sqlEntry, err := sqlutil.GetSQLByUUID("550e8400-e29b-41d4-a716-446655440001")
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取 SQL 配置失败: " + err.Error(),
		})
		return
	}

	rows, err := dbPool.Query(ctx, sqlEntry.SQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取实例列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var instances []Instance
	for rows.Next() {
		var inst Instance
		var startupParams, result []byte
		var projectID, templateID, owner, description, dataUUID, uuidStr sql.NullString

		err := rows.Scan(
			&inst.ID, &uuidStr, &inst.Name, &projectID, &templateID, &inst.Status, &owner, &description,
			&startupParams, &dataUUID, &result, &inst.CreatedAt, &inst.StartedAt, &inst.CompletedAt, &inst.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "扫描实例数据失败: " + err.Error(),
			})
			return
		}

		if uuidStr.Valid {
			inst.UUID = uuidStr.String
		}
		if projectID.Valid {
			inst.ProjectID = projectID.String
		}
		if templateID.Valid {
			inst.TemplateID = templateID.String
		}
		if owner.Valid {
			inst.Owner = owner.String
		}
		if description.Valid {
			inst.Description = description.String
		}
		if dataUUID.Valid {
			inst.DataUUID = dataUUID.String
		}

		if startupParams != nil {
			inst.StartupParams = startupParams
		}
		if result != nil {
			inst.Result = result
		}

		// Load nodes for this instance
		nodes, err := loadInstanceNodes(ctx, inst.ID)
		if err != nil {
			log.Printf("获取节点列表失败: %v", err)
		} else {
			inst.Nodes = nodes
		}

		instances = append(instances, inst)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "获取实例列表成功",
		Data:    instances,
	})
}

// 创建实例
func createInstance(c *gin.Context) {
	ctx := context.Background()
	var inst Instance

	if err := c.ShouldBindJSON(&inst); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 生成实例ID
	if inst.ID == "" {
		inst.ID = "instance-" + uuid.New().String()[:8]
	}

	// 生成DataUUID
	if inst.DataUUID == "" {
		inst.DataUUID = uuid.New().String()
	}

	// 开始事务
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "开始事务失败: " + err.Error(),
		})
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	// 插入实例
	query := `
		INSERT INTO scheduler_instances (
			id, name, project_id, template_id, status, owner, description, 
			startup_params, data_uuid, result, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = tx.Exec(ctx, query,
		inst.ID, inst.Name, inst.ProjectID, inst.TemplateID, inst.Status, inst.Owner, inst.Description,
		inst.StartupParams, inst.DataUUID, inst.Result, time.Now(), time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "创建实例失败: " + err.Error(),
		})
		return
	}

	// 插入节点
	for i := range inst.Nodes {
		node := &inst.Nodes[i]
		node.UUIDInstances = inst.ID

		// 生成节点ID
		if node.ID == "" {
			node.ID = "node-" + uuid.New().String()[:8]
		}

		nodeQuery := `
			INSERT INTO scheduler_nodes (
				id, uuid_instances, name, type, capability_id, status, 
				input, output, result, properties, description, 
				sources, position, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`

		_, err = tx.Exec(ctx, nodeQuery,
			node.ID, node.UUIDInstances, node.Name, node.Type, node.CapabilityID, node.Status,
			node.Input, node.Output, node.Result, node.Properties, node.Description,
			node.Sources, node.Position, time.Now(), time.Now(),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "创建节点失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "创建实例成功",
		Data:    inst,
	})
}

// 获取实例
func getInstance(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("uuid_instances")

	// 获取实例基本信息
	query := `
		SELECT id, uuid, name, project_id, template_id, status, owner, description, 
		       startup_params, data_uuid, result, created_at, started_at, completed_at, updated_at
		FROM scheduler_instances
		WHERE id = $1
	`

	var inst Instance
	var startupParams, result []byte
	var projectID, templateID, owner, description, dataUUID, uuidStr sql.NullString

	err := dbPool.QueryRow(ctx, query, id).Scan(
		&inst.ID, &uuidStr, &inst.Name, &projectID, &templateID, &inst.Status, &owner, &description,
		&startupParams, &dataUUID, &result, &inst.CreatedAt, &inst.StartedAt, &inst.CompletedAt, &inst.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "实例不存在: " + err.Error(),
		})
		return
	}

	if uuidStr.Valid {
		inst.UUID = uuidStr.String
	}
	if projectID.Valid {
		inst.ProjectID = projectID.String
	}
	if templateID.Valid {
		inst.TemplateID = templateID.String
	}
	if owner.Valid {
		inst.Owner = owner.String
	}
	if description.Valid {
		inst.Description = description.String
	}
	if dataUUID.Valid {
		inst.DataUUID = dataUUID.String
	}
	inst.StartupParams = startupParams
	inst.Result = result

	// 获取实例节点
	nodes, err := loadInstanceNodes(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取节点列表失败: " + err.Error(),
		})
		return
	}

	inst.Nodes = nodes

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "获取实例成功",
		Data:    inst,
	})
}

// 更新实例
func updateInstance(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("uuid_instances")

	var inst Instance
	if err := c.ShouldBindJSON(&inst); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 开始事务
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "开始事务失败: " + err.Error(),
		})
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	// 更新实例
	query := `
		UPDATE scheduler_instances
		SET name = $1, project_id = $2, template_id = $3, status = $4, 
		    owner = $5, description = $6, startup_params = $7, data_uuid = $8, 
		    result = $9, started_at = $10, completed_at = $11, updated_at = $12
		WHERE id = $13
	`

	_, err = tx.Exec(ctx, query,
		inst.Name, inst.ProjectID, inst.TemplateID, inst.Status,
		inst.Owner, inst.Description, inst.StartupParams, inst.DataUUID,
		inst.Result, inst.StartedAt, inst.CompletedAt, time.Now(), id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新实例失败: " + err.Error(),
		})
		return
	}

	// 更新节点
	if len(inst.Nodes) > 0 {
		// 删除旧节点
		_, err = tx.Exec(ctx, "DELETE FROM scheduler_nodes WHERE uuid_instances = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "删除旧节点失败: " + err.Error(),
			})
			return
		}

		// 插入新节点
		for i := range inst.Nodes {
			node := &inst.Nodes[i]
			node.UUIDInstances = id

			// 生成节点ID
			if node.ID == "" {
				node.ID = "node-" + uuid.New().String()[:8]
			}

			nodeQuery := `
				INSERT INTO scheduler_nodes (
					id, uuid_instances, name, type, capability_id, status, 
					input, output, result, properties, description, 
					sources, position, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			`

			_, err = tx.Exec(ctx, nodeQuery,
				node.ID, node.UUIDInstances, node.Name, node.Type, node.CapabilityID, node.Status,
				node.Input, node.Output, node.Result, node.Properties, node.Description,
				node.Sources, node.Position, time.Now(), time.Now(),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Message: "创建节点失败: " + err.Error(),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "更新实例成功",
		Data:    inst,
	})
}

// 删除实例
func deleteInstance(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("uuid_instances")

	// 开始事务
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "开始事务失败: " + err.Error(),
		})
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	// 删除节点
	_, err = tx.Exec(ctx, "DELETE FROM scheduler_nodes WHERE uuid_instances = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除节点失败: " + err.Error(),
		})
		return
	}

	// 删除连接
	_, err = tx.Exec(ctx, "DELETE FROM scheduler_connections WHERE instance_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除连接失败: " + err.Error(),
		})
		return
	}

	// 删除执行记录
	_, err = tx.Exec(ctx, "DELETE FROM scheduler_executions WHERE instance_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除执行记录失败: " + err.Error(),
		})
		return
	}

	// 删除实例
	_, err = tx.Exec(ctx, "DELETE FROM scheduler_instances WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除实例失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "删除实例成功",
	})
}

// 启动实例
func startInstance(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("uuid_instances")

	now := time.Now()

	// 第一步：获取实例详情，包含nodes
	var inst Instance
	var startupParams []byte
	var projectID, templateID, owner, description, dataUUID, uuidStr sql.NullString

	// 获取实例基本信息
	instanceQuery := `
		SELECT id, uuid, name, project_id, template_id, status, owner, description, 
		       startup_params, data_uuid, created_at, started_at, completed_at, updated_at
		FROM scheduler_instances
		WHERE id = $1
	`

	err := dbPool.QueryRow(ctx, instanceQuery, id).Scan(
		&inst.ID, &uuidStr, &inst.Name, &projectID, &templateID, &inst.Status, &owner, &description,
		&startupParams, &dataUUID, &inst.CreatedAt, &inst.StartedAt, &inst.CompletedAt, &inst.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "实例不存在: " + err.Error(),
		})
		return
	}

	if uuidStr.Valid {
		inst.UUID = uuidStr.String
	}
	if projectID.Valid {
		inst.ProjectID = projectID.String
	}
	if templateID.Valid {
		inst.TemplateID = templateID.String
	}
	if owner.Valid {
		inst.Owner = owner.String
	}
	if description.Valid {
		inst.Description = description.String
	}
	if dataUUID.Valid {
		inst.DataUUID = dataUUID.String
	}
	inst.StartupParams = startupParams

	// 获取实例节点
	nodeQuery := `
		SELECT id, uuid_instances, name, type, capability_id, status, 
		       input, output, properties, description, sources, position
		FROM scheduler_nodes
		WHERE uuid_instances = $1
		ORDER BY created_at ASC
	`

	nodeRows, err := dbPool.Query(ctx, nodeQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取节点列表失败: " + err.Error(),
		})
		return
	}
	defer nodeRows.Close()

	var nodes []*types.Node
	for nodeRows.Next() {
		var node types.Node
		var name, status, description string
		var input, output, properties, sources, position []byte
		var capabilityIDNull, uuidInstances sql.NullString

		err := nodeRows.Scan(
			&node.ID, &uuidInstances, &name, &node.Type, &capabilityIDNull, &status,
			&input, &output, &properties, &description, &sources, &position,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "扫描节点数据失败: " + err.Error(),
			})
			return
		}

		// 设置节点属性
		if properties != nil {
			var props map[string]interface{}
			if err := json.Unmarshal(properties, &props); err == nil {
				node.Properties = props
			}
		} else {
			node.Properties = make(map[string]interface{})
		}

		// 添加额外属性到Properties中
		if uuidInstances.Valid {
			node.Properties["uuid_instances"] = uuidInstances.String
		}
		node.Properties["name"] = name
		if capabilityIDNull.Valid {
			node.Properties["capability_id"] = capabilityIDNull.String
		}
		node.Properties["status"] = status
		node.Properties["description"] = description
		if sources != nil {
			var srcs []interface{}
			if err := json.Unmarshal(sources, &srcs); err == nil {
				node.Properties["sources"] = srcs
			}
		}
		// 设置位置信息
		node.Position = &types.Position{
			X: 0,
			Y: 0,
		}
		if position != nil {
			var posMap map[string]interface{}
			if err := json.Unmarshal(position, &posMap); err == nil {
				if x, ok := posMap["x"].(float64); ok {
					node.Position.X = x
				}
				if y, ok := posMap["y"].(float64); ok {
					node.Position.Y = y
				}
			}
		}

		nodes = append(nodes, &node)
	}

	// 更新实例状态为运行中
	updateQuery := `
		UPDATE scheduler_instances
		SET status = 'running', started_at = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = dbPool.Exec(ctx, updateQuery, now, now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新实例状态失败: " + err.Error(),
		})
		return
	}

	// 创建执行记录
	executionID := "execution-" + uuid.New().String()[:8]
	executionQuery := `
		INSERT INTO scheduler_executions (
			id, instance_id, status, start_time, updated_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err = dbPool.Exec(ctx, executionQuery, executionID, id, "running", now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "创建执行记录失败: " + err.Error(),
		})
		return
	}

	// 第二步：把nodes给到supervisor进行执行，并携带need_save_status参数
	// 创建数据库连接
	dsn := db.GetDSN()
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("创建数据库连接失败: %v", err)
		// 继续执行，不影响实例启动
	} else {
		// 测试数据库连接
		if err := sqlDB.Ping(); err != nil {
			log.Printf("数据库连接测试失败: %v", err)
			// 继续执行，不影响实例启动
		} else {
			log.Printf("数据库连接成功")
		}
	}

	// 创建初始上下文
	initialContext := &types.Context{
		Data:    make(map[string]interface{}),
		Process: true,
	}
	if startupParams != nil {
		var params map[string]interface{}
		if err := json.Unmarshal(startupParams, &params); err == nil {
			initialContext.Data = params
		}
	}

	// 创建监督器并执行节点
	supervisor := workflow.NewSupervisor(nodes, initialContext, true, id, sqlDB)

	// 执行节点（同步执行，确保获取执行结果）
	execResult := supervisor.Execute(ctx)

	// 构建响应数据
	responseData := map[string]interface{}{
		"instanceId":    id,
		"message":       "实例启动成功",
		"instance":      inst,
		"logs":          execResult.Logs,
		"nodeDetails":   execResult.NodeDetails,
		"success":       execResult.Success,
		"executedNodes": execResult.ExecutedNodes,
		"error":         execResult.Error,
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "启动实例成功",
		Data:    responseData,
	})
}

// 停止实例
func stopInstance(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("uuid_instances")

	query := `
		UPDATE scheduler_instances
		SET status = 'stopped', updated_at = $1
		WHERE id = $2
	`

	_, err := dbPool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "停止实例失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "停止实例成功",
	})
}

// 重启实例
func restartInstance(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("uuid_instances")

	now := time.Now()

	// 第一步：获取实例详情，包含nodes
	var inst Instance
	var startupParams, result []byte
	var projectID, templateID, owner, description, dataUUID, uuidStr sql.NullString

	// 获取实例基本信息
	instanceQuery := `
		SELECT id, uuid, name, project_id, template_id, status, owner, description, 
		       startup_params, data_uuid, result, created_at, started_at, completed_at, updated_at
		FROM scheduler_instances
		WHERE id = $1
	`

	err := dbPool.QueryRow(ctx, instanceQuery, id).Scan(
		&inst.ID, &uuidStr, &inst.Name, &projectID, &templateID, &inst.Status, &owner, &description,
		&startupParams, &dataUUID, &result, &inst.CreatedAt, &inst.StartedAt, &inst.CompletedAt, &inst.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "实例不存在: " + err.Error(),
		})
		return
	}

	if uuidStr.Valid {
		inst.UUID = uuidStr.String
	}
	if projectID.Valid {
		inst.ProjectID = projectID.String
	}
	if templateID.Valid {
		inst.TemplateID = templateID.String
	}
	if owner.Valid {
		inst.Owner = owner.String
	}
	if description.Valid {
		inst.Description = description.String
	}
	if dataUUID.Valid {
		inst.DataUUID = dataUUID.String
	}
	inst.StartupParams = startupParams
	inst.Result = result

	// 获取实例节点
	nodeQuery := `
		SELECT id, uuid_instances, name, type, capability_id, status, 
		       input, output, result, properties, description, 
		       sources, position, created_at, started_at, completed_at, updated_at
		FROM scheduler_nodes
		WHERE uuid_instances = $1
		ORDER BY created_at ASC
	`

	nodeRows, err := dbPool.Query(ctx, nodeQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取节点列表失败: " + err.Error(),
		})
		return
	}
	defer nodeRows.Close()

	var nodes []*types.Node
	for nodeRows.Next() {
		var node types.Node
		var name, status, description string
		var input, output, nodeResult, properties, sources, position []byte
		var capabilityIDNull, uuidInstances sql.NullString
		var created_at, started_at, completed_at, updated_at time.Time

		err := nodeRows.Scan(
			&node.ID, &uuidInstances, &name, &node.Type, &capabilityIDNull, &status,
			&input, &output, &nodeResult, &properties, &description,
			&sources, &position, &created_at, &started_at, &completed_at, &updated_at,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "扫描节点数据失败: " + err.Error(),
			})
			return
		}

		// 设置节点属性
		if properties != nil {
			var props map[string]interface{}
			if err := json.Unmarshal(properties, &props); err == nil {
				node.Properties = props
			}
		} else {
			node.Properties = make(map[string]interface{})
		}

		// 添加额外属性到Properties中
		if uuidInstances.Valid {
			node.Properties["uuid_instances"] = uuidInstances.String
		}
		node.Properties["name"] = name
		if capabilityIDNull.Valid {
			node.Properties["capability_id"] = capabilityIDNull.String
		}
		node.Properties["status"] = status
		node.Properties["description"] = description
		if sources != nil {
			var srcs []interface{}
			if err := json.Unmarshal(sources, &srcs); err == nil {
				node.Properties["sources"] = srcs
			}
		}
		node.Properties["created_at"] = created_at
		node.Properties["started_at"] = started_at
		node.Properties["completed_at"] = completed_at
		node.Properties["updated_at"] = updated_at

		// 设置位置信息
		node.Position = &types.Position{
			X: 0,
			Y: 0,
		}
		if position != nil {
			var posMap map[string]interface{}
			if err := json.Unmarshal(position, &posMap); err == nil {
				if x, ok := posMap["x"].(float64); ok {
					node.Position.X = x
				}
				if y, ok := posMap["y"].(float64); ok {
					node.Position.Y = y
				}
			}
		}

		nodes = append(nodes, &node)
	}

	// 更新实例状态为运行中
	updateQuery := `
		UPDATE scheduler_instances
		SET status = 'running', started_at = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = dbPool.Exec(ctx, updateQuery, now, now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新实例状态失败: " + err.Error(),
		})
		return
	}

	// 创建执行记录
	executionID := "execution-" + uuid.New().String()[:8]
	executionQuery := `
		INSERT INTO scheduler_executions (
			id, instance_id, status, start_time, updated_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err = dbPool.Exec(ctx, executionQuery, executionID, id, "running", now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "创建执行记录失败: " + err.Error(),
		})
		return
	}

	// 第二步：把nodes给到supervisor进行执行，并携带need_save_status参数
	// 创建数据库连接
	dsn := db.GetDSN()
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("创建数据库连接失败: %v", err)
		// 继续执行，不影响实例启动
	}

	// 创建初始上下文
	initialContext := &types.Context{
		Data:    make(map[string]interface{}),
		Process: true,
	}
	if startupParams != nil {
		var params map[string]interface{}
		if err := json.Unmarshal(startupParams, &params); err == nil {
			initialContext.Data = params
		}
	}

	// 创建监督器并执行节点
	supervisor := workflow.NewSupervisor(nodes, initialContext, true, id, sqlDB)

	// 异步执行节点，避免阻塞API响应
	go func() {
		supervisor.Execute(ctx)
	}()

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "重启实例成功",
	})
}

// 加载实例节点
func loadInstanceNodes(ctx context.Context, instanceID string) ([]Node, error) {
	query := `
		SELECT id, uuid_instances, name, type, capability_id, status, 
		       input, output, result, properties, description, 
		       sources, position, created_at, started_at, completed_at, updated_at
		FROM scheduler_nodes
		WHERE uuid_instances = $1
		ORDER BY created_at ASC
	`

	rows, err := dbPool.Query(ctx, query, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var node Node
		var input, output, result, properties, sources, position []byte
		var capabilityID, uuidInstances sql.NullString

		err := rows.Scan(
			&node.ID, &uuidInstances, &node.Name, &node.Type, &capabilityID, &node.Status,
			&input, &output, &result, &properties, &node.Description,
			&sources, &position, &node.CreatedAt, &node.StartedAt, &node.CompletedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if uuidInstances.Valid {
			node.UUIDInstances = uuidInstances.String
		}
		if capabilityID.Valid {
			node.CapabilityID = capabilityID.String
		}

		if input != nil {
			node.Input = input
		}
		if output != nil {
			node.Output = output
		}
		if result != nil {
			node.Result = result
		}
		if properties != nil {
			node.Properties = properties
		}
		if sources != nil {
			node.Sources = sources
		}
		if position != nil {
			node.Position = position
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// 列表节点
func listNodes(c *gin.Context) {
	ctx := context.Background()
	instanceID := c.Param("instanceId")

	nodes, err := loadInstanceNodes(ctx, instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取节点列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "获取节点列表成功",
		Data:    nodes,
	})
}

// 创建节点
func createNode(c *gin.Context) {
	ctx := context.Background()

	var node Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 生成节点ID
	if node.ID == "" {
		node.ID = "node-" + uuid.New().String()[:8]
	}

	query := `
		INSERT INTO scheduler_nodes (
			id, uuid_instances, name, type, capability_id, status, 
			input, output, result, properties, description, 
			sources, position, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := dbPool.Exec(ctx, query,
		node.ID, node.UUIDInstances, node.Name, node.Type, node.CapabilityID, node.Status,
		node.Input, node.Output, node.Result, node.Properties, node.Description,
		node.Sources, node.Position, time.Now(), time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "创建节点失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "创建节点成功",
		Data:    node,
	})
}

// 更新节点
func updateNode(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	var node Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求数据: " + err.Error(),
		})
		return
	}

	query := `
		UPDATE scheduler_nodes
		SET name = $1, type = $2, capability_id = $3, status = $4, 
		    input = $5, output = $6, result = $7, properties = $8, 
		    description = $9, sources = $10, position = $11, 
		    started_at = $12, completed_at = $13, updated_at = $14
		WHERE id = $15
	`

	_, err := dbPool.Exec(ctx, query,
		node.Name, node.Type, node.CapabilityID, node.Status,
		node.Input, node.Output, node.Result, node.Properties,
		node.Description, node.Sources, node.Position,
		node.StartedAt, node.CompletedAt, time.Now(), id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新节点失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "更新节点成功",
		Data:    node,
	})
}

// 更新节点状态
func updateNodeStatus(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	// 定义请求体结构
	type StatusUpdateRequest struct {
		Status string `json:"status" binding:"required"`
	}

	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 执行SQL更新操作
	query := `
		UPDATE scheduler_nodes
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := dbPool.Exec(ctx, query, req.Status, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新节点状态失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "更新节点状态成功",
		Data: map[string]interface{}{
			"nodeId": id,
			"status": req.Status,
		},
	})
}

// 删除节点
func deleteNode(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	query := "DELETE FROM scheduler_nodes WHERE id = $1"

	_, err := dbPool.Exec(ctx, query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除节点失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "删除节点成功",
	})
}

// 列表连接
func listConnections(c *gin.Context) {
	ctx := context.Background()
	instanceID := c.Param("instanceId")

	query := `
		SELECT id, instance_id, source_node_id, target_node_id, points, created_at, updated_at
		FROM scheduler_connections
		WHERE instance_id = $1
	`

	rows, err := dbPool.Query(ctx, query, instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取连接列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var connections []Connection
	for rows.Next() {
		var conn Connection
		var points []byte

		err := rows.Scan(
			&conn.ID, &conn.InstanceID, &conn.SourceNodeID, &conn.TargetNodeID, &points, &conn.CreatedAt, &conn.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "扫描连接数据失败: " + err.Error(),
			})
			return
		}

		if points != nil {
			conn.Points = points
		}

		connections = append(connections, conn)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "获取连接列表成功",
		Data:    connections,
	})
}

// 创建连接
func createConnection(c *gin.Context) {
	ctx := context.Background()

	var conn Connection
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 生成连接ID
	if conn.ID == "" {
		conn.ID = "connection-" + uuid.New().String()[:8]
	}

	query := `
		INSERT INTO scheduler_connections (
			id, instance_id, source_node_id, target_node_id, points, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := dbPool.Exec(ctx, query,
		conn.ID, conn.InstanceID, conn.SourceNodeID, conn.TargetNodeID, conn.Points,
		time.Now(), time.Now(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "创建连接失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "创建连接成功",
		Data:    conn,
	})
}

// 更新连接
func updateConnection(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	var conn Connection
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求数据: " + err.Error(),
		})
		return
	}

	query := `
		UPDATE scheduler_connections
		SET source_node_id = $1, target_node_id = $2, points = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := dbPool.Exec(ctx, query, conn.SourceNodeID, conn.TargetNodeID, conn.Points, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新连接失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "更新连接成功",
		Data:    conn,
	})
}

// 删除连接
func deleteConnection(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	query := "DELETE FROM scheduler_connections WHERE id = $1"

	_, err := dbPool.Exec(ctx, query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除连接失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "删除连接成功",
	})
}

// 列表执行记录
func listExecutions(c *gin.Context) {
	ctx := context.Background()

	query := `
		SELECT id, instance_id, status, start_time, end_time, duration, 
		       parameters, result, error_message, updated_at
		FROM scheduler_executions
		ORDER BY start_time DESC
	`

	rows, err := dbPool.Query(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取执行记录列表失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var executions []Execution
	for rows.Next() {
		var exec Execution
		var parameters, result []byte

		err := rows.Scan(
			&exec.ID, &exec.InstanceID, &exec.Status, &exec.StartTime, &exec.EndTime, &exec.Duration,
			&parameters, &result, &exec.ErrorMessage, &exec.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "扫描执行记录数据失败: " + err.Error(),
			})
			return
		}

		if parameters != nil {
			exec.Parameters = parameters
		}
		if result != nil {
			exec.Result = result
		}

		executions = append(executions, exec)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "获取执行记录列表成功",
		Data:    executions,
	})
}

// 获取执行记录
func getExecution(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	query := `
		SELECT id, instance_id, status, start_time, end_time, duration, 
		       parameters, result, error_message, updated_at
		FROM scheduler_executions
		WHERE id = $1
	`

	var exec Execution
	var parameters, result []byte
	var duration sql.NullInt32
	var errorMessage sql.NullString

	err := dbPool.QueryRow(ctx, query, id).Scan(
		&exec.ID, &exec.InstanceID, &exec.Status, &exec.StartTime, &exec.EndTime, &duration,
		&parameters, &result, &errorMessage, &exec.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "执行记录不存在: " + err.Error(),
		})
		return
	}

	if duration.Valid {
		exec.Duration = int(duration.Int32)
	}
	if parameters != nil {
		exec.Parameters = parameters
	}
	if result != nil {
		exec.Result = result
	}
	if errorMessage.Valid {
		exec.ErrorMessage = errorMessage.String
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "获取执行记录成功",
		Data:    exec,
	})
}
