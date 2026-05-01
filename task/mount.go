package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/trae/autoFlow/carriercore/common/db"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusFailed    TaskStatus = "failed"
)

type TaskType string

const (
	TaskTypeFeature      TaskType = "feature"
	TaskTypeBug          TaskType = "bug"
	TaskTypeOptimization TaskType = "optimization"
	TaskTypeResearch     TaskType = "research"
	TaskTypeOther        TaskType = "other"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

type Task struct {
	UUIDTask              string          `json:"uuid_task"`
	Description           string          `json:"description"`
	Type                  TaskType        `json:"type"`
	UUIDAgent             *string         `json:"uuid_agent,omitempty"`
	UUIDIdentityPublisher string          `json:"uuid_identity_publisher"`
	StartTime             time.Time       `json:"start_time"`
	EndTime               time.Time       `json:"end_time"`
	Status                TaskStatus      `json:"status"`
	UUIDIdentityExecutor  *string         `json:"uuid_identity_executor,omitempty"`
	Result                json.RawMessage `json:"result,omitempty"`
	Priority              TaskPriority    `json:"priority"`
	Tags                  json.RawMessage `json:"tags,omitempty"`
	Attachments           json.RawMessage `json:"attachments,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	StartedAt             *time.Time      `json:"started_at,omitempty"`
	CompletedAt           *time.Time      `json:"completed_at,omitempty"`
}

type TaskOperation struct {
	ID               int64           `json:"id"`
	UUIDTask         string          `json:"uuid_task"`
	OperationType    string          `json:"operation_type"`
	OperatorUUID     string          `json:"operator_uuid"`
	OperationDetails json.RawMessage `json:"operation_details"`
	CreatedAt        time.Time       `json:"created_at"`
}

type TaskComment struct {
	ID           int64     `json:"id"`
	UUIDTask     string    `json:"uuid_task"`
	UUIDIdentity string    `json:"uuid_identity"`
	Content      string    `json:"content"`
	ParentID     *int64    `json:"parent_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TaskSubtask struct {
	ID           string     `json:"id"`
	UUIDTask     string     `json:"uuid_task"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	Status       string     `json:"status"`
	AssigneeUUID *string   `json:"assignee_uuid,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	SortOrder    int        `json:"sort_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateTaskRequest struct {
	Description           string       `json:"description"`
	Type                  TaskType     `json:"type"`
	UUIDAgent             *string      `json:"uuid_agent,omitempty"`
	UUIDIdentityPublisher string       `json:"uuid_identity_publisher"`
	StartTime             time.Time    `json:"start_time"`
	EndTime               time.Time    `json:"end_time"`
	Priority              TaskPriority `json:"priority,omitempty"`
	Tags                  []string     `json:"tags,omitempty"`
}

type UpdateTaskRequest struct {
	Description          *string          `json:"description,omitempty"`
	Type                 *TaskType        `json:"type,omitempty"`
	UUIDAgent            *string          `json:"uuid_agent,omitempty"`
	StartTime            *time.Time       `json:"start_time,omitempty"`
	EndTime              *time.Time       `json:"end_time,omitempty"`
	Status               *TaskStatus      `json:"status,omitempty"`
	UUIDIdentityExecutor *string          `json:"uuid_identity_executor,omitempty"`
	Result               json.RawMessage  `json:"result,omitempty"`
	Priority             *TaskPriority    `json:"priority,omitempty"`
	Tags                 []string         `json:"tags,omitempty"`
}

var pgDB *sql.DB

func Init() error {
	dsn := db.GetDSN()

	log.Println("Initializing Task Center database connection...")

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

	log.Println("Task Center database connected successfully")

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func createTables() error {
	log.Println("Creating Task Center tables if not exist...")

	queries := []string{
		`CREATE TABLE IF NOT EXISTS task_center (
			uuid_task VARCHAR(100) PRIMARY KEY,
			description TEXT NOT NULL,
			type VARCHAR(50) NOT NULL,
			uuid_agent VARCHAR(100),
			uuid_identity_publisher VARCHAR(100) NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			uuid_identity_executor VARCHAR(100),
			result JSONB,
			priority VARCHAR(20) DEFAULT 'medium',
			tags JSONB,
			attachments JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS task_operations (
			id SERIAL PRIMARY KEY,
			uuid_task VARCHAR(100) NOT NULL,
			operation_type VARCHAR(50) NOT NULL,
			operator_uuid VARCHAR(100) NOT NULL,
			operation_details JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (uuid_task) REFERENCES task_center(uuid_task) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS task_comments (
			id SERIAL PRIMARY KEY,
			uuid_task VARCHAR(100) NOT NULL,
			uuid_identity VARCHAR(100) NOT NULL,
			content TEXT NOT NULL,
			parent_id INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (uuid_task) REFERENCES task_center(uuid_task) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS task_subtasks (
			id VARCHAR(100) PRIMARY KEY,
			uuid_task VARCHAR(100) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			assignee_uuid VARCHAR(100),
			due_date TIMESTAMP,
			completed_at TIMESTAMP,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (uuid_task) REFERENCES task_center(uuid_task) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_task_center_status ON task_center(status)`,
		`CREATE INDEX IF NOT EXISTS idx_task_center_type ON task_center(type)`,
		`CREATE INDEX IF NOT EXISTS idx_task_center_priority ON task_center(priority)`,
		`CREATE INDEX IF NOT EXISTS idx_task_center_uuid_identity_publisher ON task_center(uuid_identity_publisher)`,
		`CREATE INDEX IF NOT EXISTS idx_task_center_uuid_identity_executor ON task_center(uuid_identity_executor)`,
		`CREATE INDEX IF NOT EXISTS idx_task_center_uuid_agent ON task_center(uuid_agent)`,
		`CREATE INDEX IF NOT EXISTS idx_task_operations_uuid_task ON task_operations(uuid_task)`,
		`CREATE INDEX IF NOT EXISTS idx_task_comments_uuid_task ON task_comments(uuid_task)`,
		`CREATE INDEX IF NOT EXISTS idx_task_subtasks_uuid_task ON task_subtasks(uuid_task)`,
	}

	for i, query := range queries {
		_, err := pgDB.Exec(query)
		if err != nil {
			log.Printf("Warning: failed to execute query %d: %v", i+1, err)
		}
	}

	log.Println("Task Center tables created successfully")
	return nil
}

func RegisterRoutes(router *gin.Engine) {
	taskGroup := router.Group("/task")

	taskGroup.POST("/create", CreateTask)
	taskGroup.GET("/get", GetTask)
	taskGroup.POST("/update", UpdateTask)
	taskGroup.DELETE("/delete", DeleteTask)
	taskGroup.GET("/list", ListTasks)
	taskGroup.GET("/operations", GetTaskOperations)
	taskGroup.POST("/comment/add", AddTaskComment)
	taskGroup.GET("/comments", GetTaskComments)
	taskGroup.POST("/subtask/create", CreateSubtask)
	taskGroup.GET("/subtasks", GetSubtasks)
	taskGroup.POST("/subtask/update", UpdateSubtask)
}

func GenerateUUID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

func CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.Description) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "任务描述不能为空",
		})
		return
	}

	if req.StartTime.After(req.EndTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "开始时间不能晚于截止时间",
		})
		return
	}

	uuidTask := GenerateUUID()
	priority := req.Priority
	if priority == "" {
		priority = TaskPriorityMedium
	}

	tagsJSON, _ := json.Marshal(req.Tags)

	query := `
		INSERT INTO task_center (
			uuid_task, description, type, uuid_agent,
			uuid_identity_publisher, start_time, end_time,
			status, priority, tags, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := pgDB.Exec(query,
		uuidTask, req.Description, req.Type, req.UUIDAgent,
		req.UUIDIdentityPublisher, req.StartTime, req.EndTime,
		TaskStatusPending, priority, tagsJSON, time.Now(), time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create task",
			"error":   err.Error(),
		})
		return
	}

	logOperation(uuidTask, "create", req.UUIDIdentityPublisher, map[string]string{"content": "任务创建"})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task created successfully",
		"data": gin.H{
			"uuid_task":    uuidTask,
			"description":  req.Description,
			"type":         req.Type,
			"priority":     priority,
		},
	})
}

func GetTask(c *gin.Context) {
	uuidTask := c.Query("uuid_task")
	if uuidTask == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "uuid_task is required",
		})
		return
	}

	task, err := getTaskByUUID(uuidTask)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Task not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get task",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task retrieved successfully",
		"data":    task,
	})
}

func UpdateTask(c *gin.Context) {
	uuidTask := c.Query("uuid_task")
	if uuidTask == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "uuid_task is required",
		})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	task, err := getTaskByUUID(uuidTask)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Task not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get task",
			"error":   err.Error(),
		})
		return
	}

	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Type != nil {
		task.Type = *req.Type
	}
	if req.UUIDAgent != nil {
		task.UUIDAgent = req.UUIDAgent
	}
	if req.StartTime != nil {
		task.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		task.EndTime = *req.EndTime
	}
	if req.Status != nil {
		if *req.Status == TaskStatusInProgress && task.Status == TaskStatusPending {
			now := time.Now()
			task.StartedAt = &now
		}
		if *req.Status == TaskStatusCompleted && task.Status != TaskStatusCompleted {
			now := time.Now()
			task.CompletedAt = &now
		}
		task.Status = *req.Status
	}
	if req.UUIDIdentityExecutor != nil {
		task.UUIDIdentityExecutor = req.UUIDIdentityExecutor
	}
	if req.Result != nil {
		task.Result = req.Result
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Tags != nil {
		task.Tags, _ = json.Marshal(req.Tags)
	}
	task.UpdatedAt = time.Now()

	query := `
		UPDATE task_center SET
			description = $1, type = $2, uuid_agent = $3,
			start_time = $4, end_time = $5, status = $6,
			uuid_identity_executor = $7, result = $8, priority = $9,
			tags = $10, updated_at = $11, started_at = $12, completed_at = $13
		WHERE uuid_task = $14
	`

	_, err = pgDB.Exec(query,
		task.Description, task.Type, task.UUIDAgent,
		task.StartTime, task.EndTime, task.Status,
		task.UUIDIdentityExecutor, task.Result, task.Priority,
		task.Tags, task.UpdatedAt, task.StartedAt, task.CompletedAt,
		task.UUIDTask,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update task",
			"error":   err.Error(),
		})
		return
	}

	logOperation(task.UUIDTask, "update", "", map[string]string{"content": "任务更新"})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task updated successfully",
		"data":    task,
	})
}

func DeleteTask(c *gin.Context) {
	uuidTask := c.Query("uuid_task")
	if uuidTask == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "uuid_task is required",
		})
		return
	}

	query := `DELETE FROM task_center WHERE uuid_task = $1`
	result, err := pgDB.Exec(query, uuidTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete task",
			"error":   err.Error(),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task deleted successfully",
	})
}

func ListTasks(c *gin.Context) {
	status := c.Query("status")
	taskType := c.Query("type")
	priority := c.Query("priority")
	uuidIdentityExecutor := c.Query("uuid_identity_executor")
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	page := 1
	pageSize := 20
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	var conditions []string
	var args []interface{}
	argCount := 0

	if status != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, status)
	}
	if taskType != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("type = $%d", argCount))
		args = append(args, taskType)
	}
	if priority != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argCount))
		args = append(args, priority)
	}
	if uuidIdentityExecutor != "" {
		argCount++
		conditions = append(conditions, fmt.Sprintf("uuid_identity_executor = $%d", argCount))
		args = append(args, uuidIdentityExecutor)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_center %s", whereClause)
	var total int64
	if err := pgDB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to count tasks",
			"error":   err.Error(),
		})
		return
	}

	offset := (page - 1) * pageSize
	argCount++
	args = append(args, pageSize)
	argCount++
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT uuid_task, description, type, uuid_agent,
			uuid_identity_publisher, start_time, end_time,
			status, uuid_identity_executor, result, priority, tags,
			attachments, created_at, updated_at, started_at, completed_at
		FROM task_center
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount-1, argCount)

	rows, err := pgDB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list tasks",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.UUIDTask, &task.Description, &task.Type, &task.UUIDAgent,
			&task.UUIDIdentityPublisher, &task.StartTime, &task.EndTime,
			&task.Status, &task.UUIDIdentityExecutor, &task.Result, &task.Priority,
			&task.Tags, &task.Attachments, &task.CreatedAt, &task.UpdatedAt,
			&task.StartedAt, &task.CompletedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to scan task",
				"error":   err.Error(),
			})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tasks retrieved successfully",
		"data": gin.H{
			"tasks":     tasks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func getTaskByUUID(uuidTask string) (*Task, error) {
	query := `
		SELECT uuid_task, description, type, uuid_agent,
			uuid_identity_publisher, start_time, end_time,
			status, uuid_identity_executor, result, priority, tags,
			attachments, created_at, updated_at, started_at, completed_at
		FROM task_center
		WHERE uuid_task = $1
	`

	var task Task
	err := pgDB.QueryRow(query, uuidTask).Scan(
		&task.UUIDTask, &task.Description, &task.Type, &task.UUIDAgent,
		&task.UUIDIdentityPublisher, &task.StartTime, &task.EndTime,
		&task.Status, &task.UUIDIdentityExecutor, &task.Result, &task.Priority,
		&task.Tags, &task.Attachments, &task.CreatedAt, &task.UpdatedAt,
		&task.StartedAt, &task.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func logOperation(uuidTask, operationType, operatorUUID string, details map[string]string) {
	detailsJSON, _ := json.Marshal(details)
	query := `
		INSERT INTO task_operations (uuid_task, operation_type, operator_uuid, operation_details)
		VALUES ($1, $2, $3, $4)
	`
	pgDB.Exec(query, uuidTask, operationType, operatorUUID, detailsJSON)
}

func GetTaskOperations(c *gin.Context) {
	uuidTask := c.Query("uuid_task")
	if uuidTask == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "uuid_task is required",
		})
		return
	}

	query := `
		SELECT id, uuid_task, operation_type, operator_uuid, operation_details, created_at
		FROM task_operations
		WHERE uuid_task = $1
		ORDER BY created_at DESC
	`

	rows, err := pgDB.Query(query, uuidTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get operations",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var operations []TaskOperation
	for rows.Next() {
		var op TaskOperation
		if err := rows.Scan(&op.ID, &op.UUIDTask, &op.OperationType, &op.OperatorUUID, &op.OperationDetails, &op.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to scan operation",
				"error":   err.Error(),
			})
			return
		}
		operations = append(operations, op)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Operations retrieved successfully",
		"data":    operations,
	})
}

func AddTaskComment(c *gin.Context) {
	var req struct {
		UUIDTask     string `json:"uuid_task"`
		UUIDIdentity string `json:"uuid_identity"`
		Content      string `json:"content"`
		ParentID     *int64 `json:"parent_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	query := `
		INSERT INTO task_comments (uuid_task, uuid_identity, content, parent_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	var comment TaskComment
	err := pgDB.QueryRow(query, req.UUIDTask, req.UUIDIdentity, req.Content, req.ParentID).Scan(
		&comment.ID, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to add comment",
			"error":   err.Error(),
		})
		return
	}

	comment.UUIDTask = req.UUIDTask
	comment.UUIDIdentity = req.UUIDIdentity
	comment.Content = req.Content
	comment.ParentID = req.ParentID

	logOperation(req.UUIDTask, "comment", req.UUIDIdentity, map[string]string{"content": req.Content})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Comment added successfully",
		"data":    comment,
	})
}

func GetTaskComments(c *gin.Context) {
	uuidTask := c.Query("uuid_task")
	if uuidTask == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "uuid_task is required",
		})
		return
	}

	query := `
		SELECT id, uuid_task, uuid_identity, content, parent_id, created_at, updated_at
		FROM task_comments
		WHERE uuid_task = $1
		ORDER BY created_at ASC
	`

	rows, err := pgDB.Query(query, uuidTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get comments",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var comments []TaskComment
	for rows.Next() {
		var comment TaskComment
		if err := rows.Scan(&comment.ID, &comment.UUIDTask, &comment.UUIDIdentity, &comment.Content, &comment.ParentID, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to scan comment",
				"error":   err.Error(),
			})
			return
		}
		comments = append(comments, comment)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Comments retrieved successfully",
		"data":    comments,
	})
}

func CreateSubtask(c *gin.Context) {
	var req struct {
		UUIDTask     string  `json:"uuid_task"`
		Title        string  `json:"title"`
		Description  string  `json:"description,omitempty"`
		AssigneeUUID *string `json:"assignee_uuid,omitempty"`
		DueDate      *string `json:"due_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	id := fmt.Sprintf("subtask-%d", time.Now().UnixNano())
	var dueDate *time.Time
	if req.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			dueDate = &t
		}
	}

	query := `
		INSERT INTO task_subtasks (id, uuid_task, title, description, status, assignee_uuid, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	var subtask TaskSubtask
	err := pgDB.QueryRow(query, id, req.UUIDTask, req.Title, req.Description, "pending", req.AssigneeUUID, dueDate).Scan(
		&subtask.CreatedAt, &subtask.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create subtask",
			"error":   err.Error(),
		})
		return
	}

	subtask.ID = id
	subtask.UUIDTask = req.UUIDTask
	subtask.Title = req.Title
	subtask.Description = req.Description
	subtask.AssigneeUUID = req.AssigneeUUID
	subtask.DueDate = dueDate

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subtask created successfully",
		"data":    subtask,
	})
}

func GetSubtasks(c *gin.Context) {
	uuidTask := c.Query("uuid_task")
	if uuidTask == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "uuid_task is required",
		})
		return
	}

	query := `
		SELECT id, uuid_task, title, description, status, assignee_uuid, due_date, completed_at, sort_order, created_at, updated_at
		FROM task_subtasks
		WHERE uuid_task = $1
		ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := pgDB.Query(query, uuidTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get subtasks",
			"error":   err.Error(),
		})
		return
	}
	defer rows.Close()

	var subtasks []TaskSubtask
	for rows.Next() {
		var subtask TaskSubtask
		if err := rows.Scan(
			&subtask.ID, &subtask.UUIDTask, &subtask.Title, &subtask.Description,
			&subtask.Status, &subtask.AssigneeUUID, &subtask.DueDate,
			&subtask.CompletedAt, &subtask.SortOrder, &subtask.CreatedAt, &subtask.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to scan subtask",
				"error":   err.Error(),
			})
			return
		}
		subtasks = append(subtasks, subtask)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subtasks retrieved successfully",
		"data":    subtasks,
	})
}

func UpdateSubtask(c *gin.Context) {
	var req struct {
		ID           string  `json:"id"`
		Title        *string `json:"title,omitempty"`
		Description  *string `json:"description,omitempty"`
		Status       *string `json:"status,omitempty"`
		AssigneeUUID *string `json:"assignee_uuid,omitempty"`
		DueDate      *string `json:"due_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	var completedAt *time.Time
	if req.Status != nil && *req.Status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE task_subtasks SET
			title = COALESCE($1, title),
			description = COALESCE($2, description),
			status = COALESCE($3, status),
			assignee_uuid = COALESCE($4, assignee_uuid),
			due_date = COALESCE($5, due_date),
			completed_at = COALESCE($6, completed_at),
			updated_at = $7
		WHERE id = $8
	`

	_, err := pgDB.Exec(query, req.Title, req.Description, req.Status, req.AssigneeUUID, req.DueDate, completedAt, time.Now(), req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update subtask",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Subtask updated successfully",
	})
}
