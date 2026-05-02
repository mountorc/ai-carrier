package task

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	dbconfig "github.com/trae/autoFlow/carriercore/common/db"
)

//go:embed api_task.json
var apiTaskData []byte

//go:embed sql_task.json
var sqlTaskData []byte

var (
	db        *sql.DB
	apiConfig map[string]interface{}
	sqlConfig map[string]interface{}
)

type Task struct {
	UUIDTask              string     `json:"uuid_task"`
	Description           string     `json:"description"`
	Type                  string     `json:"type"`
	UUIDAgent             string     `json:"uuid_agent"`
	UUIDIdentityPublisher string     `json:"uuid_identity_publisher"`
	StartTime             *time.Time `json:"start_time"`
	EndTime               *time.Time `json:"end_time"`
	Status                string     `json:"status"`
	UUIDIdentityExecutor  string     `json:"uuid_identity_executor"`
	Result                string     `json:"result"`
	Priority              int        `json:"priority"`
	Tags                  string     `json:"tags"`
	Attachments           string     `json:"attachments"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	StartedAt             *time.Time `json:"started_at"`
	CompletedAt           *time.Time `json:"completed_at"`
}

type TaskOperation struct {
	ID               int       `json:"id"`
	UUIDTask         string    `json:"uuid_task"`
	OperationType    string    `json:"operation_type"`
	OperatorUUID     string    `json:"operator_uuid"`
	OperationDetails string    `json:"operation_details"`
	CreatedAt        time.Time `json:"created_at"`
}

type TaskComment struct {
	ID           int       `json:"id"`
	UUIDTask     string    `json:"uuid_task"`
	UUIDIdentity string    `json:"uuid_identity"`
	Content      string    `json:"content"`
	ParentID     *int      `json:"parent_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Subtask struct {
	ID           int        `json:"id"`
	UUIDTask     string     `json:"uuid_task"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	AssigneeUUID string     `json:"assignee_uuid"`
	DueDate      *time.Time `json:"due_date"`
	CompletedAt  *time.Time `json:"completed_at"`
	SortOrder    int        `json:"sort_order"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ListTasksResponse struct {
	Total int    `json:"total"`
	List  []Task `json:"list"`
}

func InitDB(connStr string) error {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Task module: Database connection established")
	return nil
}

func LoadConfig(configPath string) error {
	if err := json.Unmarshal(apiTaskData, &apiConfig); err != nil {
		return fmt.Errorf("failed to parse api_task.json: %v", err)
	}

	if err := json.Unmarshal(sqlTaskData, &sqlConfig); err != nil {
		return fmt.Errorf("failed to parse sql_task.json: %v", err)
	}

	log.Println("Task module: Configuration loaded successfully")
	return nil
}

func Init() error {
	log.Println("Task Init: starting...")
	if err := LoadConfig(""); err != nil {
		log.Printf("Task Init: LoadConfig failed: %v", err)
		return err
	}
	log.Println("Task Init: LoadConfig succeeded")

	dsn := dbconfig.GetDSN()
	log.Println("Task Init: got DSN, initializing Postgres connection...")

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("Task Init: sql.Open failed: %v", err)
		return err
	}
	log.Println("Task Init: sql.Open succeeded")

	if err := db.Ping(); err != nil {
		log.Printf("Task Init: db.Ping failed: %v", err)
		return err
	}
	log.Println("Task Init: db.Ping succeeded")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Task Init: Postgres connected successfully")

	if err := EnsureTables(); err != nil {
		log.Printf("Task Init: EnsureTables failed: %v", err)
		return err
	}
	log.Println("Task Init: EnsureTables succeeded")

	log.Println("Task Init: complete!")
	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
}

func RegisterRoutes(router *gin.Engine) {
	taskGroup := router.Group("/task")
	{
		taskGroup.GET("/list", ListTasksHandler)
		taskGroup.GET("/get", GetTaskHandler)
		taskGroup.POST("/create", CreateTaskHandler)
		taskGroup.POST("/update", UpdateTaskHandler)
		taskGroup.DELETE("/delete", DeleteTaskHandler)
		taskGroup.GET("/operations", GetTaskOperationsHandler)
		taskGroup.GET("/comments", GetTaskCommentsHandler)
		taskGroup.POST("/comment/add", AddTaskCommentHandler)
		taskGroup.GET("/subtasks", GetSubtasksHandler)
		taskGroup.POST("/subtask/create", CreateSubtaskHandler)
		taskGroup.POST("/subtask/update", UpdateSubtaskHandler)
		taskGroup.GET("/debug/tables", DebugTablesHandler)
	}
	log.Println("Task module: Routes registered")
}

func DebugTablesHandler(c *gin.Context) {
	// 查询所有表名
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		log.Printf("Error querying tables: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			tables = append(tables, tableName)
		}
	}

	// 检查 task 表的数据量
	var taskCount int
	db.QueryRow("SELECT COUNT(*) FROM task").Scan(&taskCount)

	c.JSON(http.StatusOK, gin.H{
		"tables":     tables,
		"task_count": taskCount,
	})
}

func ListTasksHandler(c *gin.Context) {
	log.Println("ListTasksHandler: called")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	log.Printf("ListTasksHandler: page=%d, pageSize=%d, offset=%d", page, pageSize, offset)

	countSql := "SELECT COUNT(*) FROM task"
	var total int
	if err := db.QueryRow(countSql).Scan(&total); err != nil {
		log.Printf("Error counting tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to count tasks: %v", err)})
		return
	}

	sqlStr := fmt.Sprintf(`
		SELECT uuid_task, description, type, uuid_agent, uuid_identity_publisher,
		       start_time, end_time, status, uuid_identity_executor, CAST(result AS TEXT) as result,
		       priority, tags, attachments, created_at, updated_at, started_at, completed_at
		FROM task
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`)

	rows, err := db.Query(sqlStr, pageSize, offset)
	if err != nil {
		log.Printf("Error querying tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query tasks"})
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		var startTime, endTime, startedAt, completedAt sql.NullTime
		var uuidAgent, uuidIdentityPublisher, uuidIdentityExecutor, result, tags, attachments sql.NullString
		var priority sql.NullInt32

		err := rows.Scan(
			&task.UUIDTask,
			&task.Description,
			&task.Type,
			&uuidAgent,
			&uuidIdentityPublisher,
			&startTime,
			&endTime,
			&task.Status,
			&uuidIdentityExecutor,
			&result,
			&priority,
			&tags,
			&attachments,
			&task.CreatedAt,
			&task.UpdatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}

		task.UUIDAgent = uuidAgent.String
		task.UUIDIdentityPublisher = uuidIdentityPublisher.String
		task.UUIDIdentityExecutor = uuidIdentityExecutor.String
		task.Result = result.String
		task.Tags = tags.String
		task.Attachments = attachments.String

		if priority.Valid {
			task.Priority = int(priority.Int32)
		}

		if startTime.Valid {
			task.StartTime = &startTime.Time
		}
		if endTime.Valid {
			task.EndTime = &endTime.Time
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"list":  tasks,
		"page":  page,
		"size":  pageSize,
	})
}

func GetTaskHandler(c *gin.Context) {
	taskUUID := c.Query("uuid")
	if taskUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	sqlStr := `SELECT uuid_task, description, type, uuid_agent, uuid_identity_publisher,
		start_time, end_time, status, uuid_identity_executor, CAST(result AS TEXT) as result, priority,
		tags, attachments, created_at, updated_at, started_at, completed_at
		FROM task WHERE uuid_task = $1`

	var task Task
	var startTime, endTime, startedAt, completedAt sql.NullTime
	var uuidAgent, uuidIdentityPublisher, uuidIdentityExecutor, result, tags, attachments sql.NullString
	var priority sql.NullInt32

	err := db.QueryRow(sqlStr, taskUUID).Scan(
		&task.UUIDTask,
		&task.Description,
		&task.Type,
		&uuidAgent,
		&uuidIdentityPublisher,
		&startTime,
		&endTime,
		&task.Status,
		&uuidIdentityExecutor,
		&result,
		&priority,
		&tags,
		&attachments,
		&task.CreatedAt,
		&task.UpdatedAt,
		&startedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	if err != nil {
		log.Printf("Error getting task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}

	task.UUIDAgent = uuidAgent.String
	task.UUIDIdentityPublisher = uuidIdentityPublisher.String
	task.UUIDIdentityExecutor = uuidIdentityExecutor.String
	task.Result = result.String
	task.Tags = tags.String
	task.Attachments = attachments.String

	if priority.Valid {
		task.Priority = int(priority.Int32)
	}

	if startTime.Valid {
		task.StartTime = &startTime.Time
	}
	if endTime.Valid {
		task.EndTime = &endTime.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	c.JSON(http.StatusOK, task)
}

func CreateTaskHandler(c *gin.Context) {
	var input struct {
		Description           string     `json:"description"`
		Type                  string     `json:"type"`
		UUIDAgent             string     `json:"uuid_agent"`
		UUIDIdentityPublisher string     `json:"uuid_identity_publisher"`
		StartTime             *time.Time `json:"start_time"`
		EndTime               *time.Time `json:"end_time"`
		Priority              int        `json:"priority"`
		Tags                  string     `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskUUID := uuid.New().String()
	now := time.Now()

	sqlStr := `INSERT INTO task (uuid_task, description, type, uuid_agent, uuid_identity_publisher, 
		start_time, end_time, status, priority, tags, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := db.Exec(sqlStr,
		taskUUID,
		input.Description,
		input.Type,
		input.UUIDAgent,
		input.UUIDIdentityPublisher,
		input.StartTime,
		input.EndTime,
		"pending",
		input.Priority,
		input.Tags,
		now,
		now,
	)

	if err != nil {
		log.Printf("Error creating task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"uuid_task": taskUUID,
		"message":   "Task created successfully",
	})
}

func UpdateTaskHandler(c *gin.Context) {
	var input struct {
		UUIDTask             string     `json:"uuid_task"`
		Description          string     `json:"description"`
		Type                 string     `json:"type"`
		UUIDAgent            string     `json:"uuid_agent"`
		StartTime            *time.Time `json:"start_time"`
		EndTime              *time.Time `json:"end_time"`
		Status               string     `json:"status"`
		UUIDIdentityExecutor string     `json:"uuid_identity_executor"`
		Result               string     `json:"result"`
		Priority             int        `json:"priority"`
		Tags                 string     `json:"tags"`
		StartedAt            *time.Time `json:"started_at"`
		CompletedAt          *time.Time `json:"completed_at"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()

	sqlStr := `UPDATE task SET 
		description = $1, type = $2, uuid_agent = $3, start_time = $4, end_time = $5, 
		status = $6, uuid_identity_executor = $7, result = $8, priority = $9, tags = $10, 
		updated_at = $11, started_at = $12, completed_at = $13 
		WHERE uuid_task = $14`

	result, err := db.Exec(sqlStr,
		input.Description,
		input.Type,
		input.UUIDAgent,
		input.StartTime,
		input.EndTime,
		input.Status,
		input.UUIDIdentityExecutor,
		input.Result,
		input.Priority,
		input.Tags,
		now,
		input.StartedAt,
		input.CompletedAt,
		input.UUIDTask,
	)

	if err != nil {
		log.Printf("Error updating task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func DeleteTaskHandler(c *gin.Context) {
	taskUUID := c.Query("uuid")
	if taskUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	sqlStr := `DELETE FROM task WHERE uuid_task = $1`
	result, err := db.Exec(sqlStr, taskUUID)

	if err != nil {
		log.Printf("Error deleting task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func GetTaskOperationsHandler(c *gin.Context) {
	taskUUID := c.Query("uuid")
	if taskUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	sqlStr := `SELECT id, uuid_task, operation_type, operator_uuid, operation_details, created_at 
		FROM task_operations WHERE uuid_task = $1 ORDER BY created_at DESC`

	rows, err := db.Query(sqlStr, taskUUID)
	if err != nil {
		log.Printf("Error getting task operations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task operations"})
		return
	}
	defer rows.Close()

	operations := []TaskOperation{}
	for rows.Next() {
		var op TaskOperation
		var operatorUUID, operationDetails sql.NullString

		err := rows.Scan(&op.ID, &op.UUIDTask, &op.OperationType, &operatorUUID, &operationDetails, &op.CreatedAt)
		if err != nil {
			log.Printf("Error scanning operation: %v", err)
			continue
		}

		op.OperatorUUID = operatorUUID.String
		op.OperationDetails = operationDetails.String
		operations = append(operations, op)
	}

	c.JSON(http.StatusOK, gin.H{"list": operations})
}

func GetTaskCommentsHandler(c *gin.Context) {
	taskUUID := c.Query("uuid")
	if taskUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	sqlStr := `SELECT id, uuid_task, uuid_identity, content, parent_id, created_at, updated_at 
		FROM task_comments WHERE uuid_task = $1 ORDER BY created_at ASC`

	rows, err := db.Query(sqlStr, taskUUID)
	if err != nil {
		log.Printf("Error getting task comments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task comments"})
		return
	}
	defer rows.Close()

	comments := []TaskComment{}
	for rows.Next() {
		var comment TaskComment
		var uuidIdentity, content sql.NullString
		var parentID sql.NullInt64

		err := rows.Scan(&comment.ID, &comment.UUIDTask, &uuidIdentity, &content, &parentID, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning comment: %v", err)
			continue
		}

		comment.UUIDIdentity = uuidIdentity.String
		comment.Content = content.String
		if parentID.Valid {
			parentIDInt := int(parentID.Int64)
			comment.ParentID = &parentIDInt
		}
		comments = append(comments, comment)
	}

	c.JSON(http.StatusOK, gin.H{"list": comments})
}

func AddTaskCommentHandler(c *gin.Context) {
	var input struct {
		UUIDTask     string `json:"uuid_task" binding:"required"`
		UUIDIdentity string `json:"uuid_identity" binding:"required"`
		Content      string `json:"content" binding:"required"`
		ParentID     *int   `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sqlStr := `INSERT INTO task_comments (uuid_task, uuid_identity, content, parent_id) 
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	var id int
	var createdAt, updatedAt time.Time

	err := db.QueryRow(sqlStr, input.UUIDTask, input.UUIDIdentity, input.Content, input.ParentID).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		log.Printf("Error adding task comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         id,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"message":    "Comment added successfully",
	})
}

func GetSubtasksHandler(c *gin.Context) {
	taskUUID := c.Query("uuid")
	if taskUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uuid is required"})
		return
	}

	sqlStr := `SELECT id, uuid_task, title, description, status, assignee_uuid, due_date, 
		completed_at, sort_order, created_at, updated_at 
		FROM task_subtasks WHERE uuid_task = $1 ORDER BY sort_order ASC, created_at ASC`

	rows, err := db.Query(sqlStr, taskUUID)
	if err != nil {
		log.Printf("Error getting subtasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subtasks"})
		return
	}
	defer rows.Close()

	subtasks := []Subtask{}
	for rows.Next() {
		var subtask Subtask
		var title, description, status, assigneeUUID sql.NullString
		var dueDate, completedAt sql.NullTime

		err := rows.Scan(
			&subtask.ID, &subtask.UUIDTask, &title, &description, &status,
			&assigneeUUID, &dueDate, &completedAt, &subtask.SortOrder,
			&subtask.CreatedAt, &subtask.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning subtask: %v", err)
			continue
		}

		subtask.Title = title.String
		subtask.Description = description.String
		subtask.Status = status.String
		subtask.AssigneeUUID = assigneeUUID.String
		if dueDate.Valid {
			subtask.DueDate = &dueDate.Time
		}
		if completedAt.Valid {
			subtask.CompletedAt = &completedAt.Time
		}
		subtasks = append(subtasks, subtask)
	}

	c.JSON(http.StatusOK, gin.H{"list": subtasks})
}

func CreateSubtaskHandler(c *gin.Context) {
	var input struct {
		UUIDTask     string     `json:"uuid_task" binding:"required"`
		Title        string     `json:"title" binding:"required"`
		Description  string     `json:"description"`
		AssigneeUUID string     `json:"assignee_uuid"`
		DueDate      *time.Time `json:"due_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()

	sqlStr := `INSERT INTO task_subtasks (id, uuid_task, title, description, status, assignee_uuid, due_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`

	var createdAt, updatedAt time.Time

	err := db.QueryRow(sqlStr, id, input.UUIDTask, input.Title, input.Description, "pending", input.AssigneeUUID, input.DueDate).Scan(&createdAt, &updatedAt)
	if err != nil {
		log.Printf("Error creating subtask: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subtask"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         id,
		"created_at": createdAt,
		"updated_at": updatedAt,
		"message":    "Subtask created successfully",
	})
}

func UpdateSubtaskHandler(c *gin.Context) {
	var input struct {
		ID           int        `json:"id" binding:"required"`
		Title        *string    `json:"title"`
		Description  *string    `json:"description"`
		Status       *string    `json:"status"`
		AssigneeUUID *string    `json:"assignee_uuid"`
		DueDate      *time.Time `json:"due_date"`
		CompletedAt  *time.Time `json:"completed_at"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()

	sqlStr := `UPDATE task_subtasks SET 
		title = COALESCE($1, title), 
		description = COALESCE($2, description), 
		status = COALESCE($3, status), 
		assignee_uuid = COALESCE($4, assignee_uuid), 
		due_date = COALESCE($5, due_date), 
		completed_at = COALESCE($6, completed_at), 
		updated_at = $7 
		WHERE id = $8`

	result, err := db.Exec(sqlStr, input.Title, input.Description, input.Status, input.AssigneeUUID, input.DueDate, input.CompletedAt, now, input.ID)
	if err != nil {
		log.Printf("Error updating subtask: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subtask"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtask not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subtask updated successfully"})
}

func CloseDB() {
	if db != nil {
		db.Close()
	}
}

func EnsureTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS task (
			uuid_task VARCHAR(255) PRIMARY KEY,
			description TEXT,
			type VARCHAR(100),
			uuid_agent VARCHAR(255),
			uuid_identity_publisher VARCHAR(255),
			start_time TIMESTAMP,
			end_time TIMESTAMP,
			status VARCHAR(50) DEFAULT 'pending',
			uuid_identity_executor VARCHAR(255),
			result TEXT,
			priority INT DEFAULT 0,
			tags TEXT,
			attachments TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS task_operations (
			id SERIAL PRIMARY KEY,
			uuid_task VARCHAR(255),
			operation_type VARCHAR(100),
			operator_uuid VARCHAR(255),
			operation_details TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS task_comments (
			id SERIAL PRIMARY KEY,
			uuid_task VARCHAR(255),
			uuid_identity VARCHAR(255),
			content TEXT,
			parent_id INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS task_subtasks (
			id VARCHAR(255) PRIMARY KEY,
			uuid_task VARCHAR(255),
			title VARCHAR(255),
			description TEXT,
			status VARCHAR(50) DEFAULT 'pending',
			assignee_uuid VARCHAR(255),
			due_date TIMESTAMP,
			completed_at TIMESTAMP,
			sort_order INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, tableSQL := range tables {
		if _, err := db.Exec(tableSQL); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	log.Println("Task module: Database tables ensured")
	return nil
}

func IsRelevantPath(path string) bool {
	return strings.HasPrefix(path, "/task")
}

func GetSQLByUUID(uuid string) (string, bool) {
	if sqls, ok := sqlConfig["sqls"].([]interface{}); ok {
		for _, sqlItem := range sqls {
			if sqlMap, ok := sqlItem.(map[string]interface{}); ok {
				if sqlMap["uuid"] == uuid {
					if sql, ok := sqlMap["sql"].(string); ok {
						return sql, true
					}
				}
			}
		}
	}
	return "", false
}
