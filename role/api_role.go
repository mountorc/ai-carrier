package role

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	dbconfig "github.com/xmzail/ai-carrier-dev/carriercore/common/db"
)

//go:embed api_role.json
var apiRoleData []byte

//go:embed sql_role.json
var sqlRoleData []byte

var (
	db        *sql.DB
	apiConfig map[string]interface{}
	sqlConfig map[string]interface{}
)

type Role struct {
	UUID          string                 `json:"uuid"`
	Name          string                 `json:"name"`
	AgentNaming   string                 `json:"agent_naming"`
	Slogan        string                 `json:"slogan"`
	Description   string                 `json:"description"`
	Skills        []interface{}          `json:"skills"`
	Status        string                 `json:"status"`
	Icon          string                 `json:"icon"`
	Tags          []interface{}          `json:"tags"`
	CurrentVersion int                   `json:"current_version"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type RoleVersion struct {
	ID            string    `json:"id"`
	UUID          string    `json:"uuid"`
	RoleUUID      string    `json:"role_uuid"`
	Version       int       `json:"version"`
	Name          string    `json:"name"`
	AgentNaming   string    `json:"agent_naming"`
	Slogan        string    `json:"slogan"`
	Description   string    `json:"description"`
	Skills        string    `json:"skills"`
	Status        string    `json:"status"`
	Icon          string    `json:"icon"`
	Tags          string    `json:"tags"`
	ChangeDesc    string    `json:"change_desc"`
	CreatedAt     time.Time `json:"created_at"`
}

func LoadConfig(configPath string) error {
	if err := json.Unmarshal(apiRoleData, &apiConfig); err != nil {
		return fmt.Errorf("failed to parse api_role.json: %v", err)
	}

	if err := json.Unmarshal(sqlRoleData, &sqlConfig); err != nil {
		return fmt.Errorf("failed to parse sql_role.json: %v", err)
	}

	log.Println("Role module: Configuration loaded successfully")
	return nil
}

func Init() error {
	log.Println("Role Init: starting...")
	if err := LoadConfig(""); err != nil {
		log.Printf("Role Init: LoadConfig failed: %v", err)
		return err
	}
	log.Println("Role Init: LoadConfig succeeded")

	dsn := dbconfig.GetDSN()
	log.Println("Role Init: got DSN, initializing Postgres connection...")

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("Role Init: sql.Open failed: %v", err)
		return err
	}
	log.Println("Role Init: sql.Open succeeded")

	if err := db.Ping(); err != nil {
		log.Printf("Role Init: db.Ping failed: %v", err)
		return err
	}
	log.Println("Role Init: db.Ping succeeded")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Role Init: Postgres connected successfully")

	if err := EnsureTables(); err != nil {
		log.Printf("Role Init: EnsureTables failed: %v", err)
		return err
	}
	log.Println("Role Init: EnsureTables succeeded")

	log.Println("Role Init: complete!")
	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
}

func RegisterRoutes(router *gin.Engine) {
	roleGroup := router.Group("/role")
	{
		roleGroup.GET("/list", GetRoleListHandler)
		roleGroup.GET("/get", GetRoleHandler)
		roleGroup.POST("/create", CreateRoleHandler)
		roleGroup.POST("/update", UpdateRoleHandler)
		roleGroup.DELETE("/delete", DeleteRoleHandler)
		roleGroup.GET("/get-by-name", GetRoleByNameHandler)
		roleGroup.GET("/versions", GetRoleVersionsHandler)
		roleGroup.GET("/version", GetRoleVersionDetailHandler)
		roleGroup.POST("/rollback", RollbackRoleHandler)
		roleGroup.GET("/compare", CompareRoleVersionsHandler)
	}
	log.Println("Role module: Routes registered")
}

func EnsureTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS agent_role (
			id VARCHAR(100) PRIMARY KEY,
			uuid VARCHAR(100) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			agent_naming VARCHAR(255),
			slogan VARCHAR(500),
			description TEXT,
			skills JSONB,
			status VARCHAR(50) NOT NULL,
			icon TEXT,
			tags JSONB,
			current_version INT DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS agent_role_version (
			id VARCHAR(100) PRIMARY KEY,
			uuid VARCHAR(100) NOT NULL,
			role_uuid VARCHAR(100) NOT NULL,
			version INT NOT NULL,
			name VARCHAR(255),
			agent_naming VARCHAR(255),
			slogan VARCHAR(500),
			description TEXT,
			skills JSONB,
			status VARCHAR(50),
			icon TEXT,
			tags JSONB,
			change_desc TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(role_uuid, version)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_role_version_role_uuid ON agent_role_version(role_uuid)`,
	}

	for _, tableSQL := range tables {
		if _, err := db.Exec(tableSQL); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	log.Println("Role module: Database tables ensured")
	return nil
}

func GetRoleListHandler(c *gin.Context) {
	log.Println("GetRoleListHandler: called")

	sqlStr := `
		SELECT id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, created_at, updated_at
		FROM agent_role
		WHERE status != 'deleted'
		ORDER BY created_at DESC
	`

	rows, err := db.Query(sqlStr)
	if err != nil {
		log.Printf("Error querying roles: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role list: %v", err),
		})
		return
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var skillsJSON, tagsJSON string
		var agentNaming, slogan, description, icon sql.NullString

		err := rows.Scan(
			&role.UUID,
			&role.UUID,
			&role.Name,
			&agentNaming,
			&slogan,
			&description,
			&skillsJSON,
			&role.Status,
			&icon,
			&tagsJSON,
			&role.CreatedAt,
			&role.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning role: %v", err)
			continue
		}

		role.AgentNaming = agentNaming.String
		role.Slogan = slogan.String
		role.Description = description.String
		role.Icon = icon.String

		if skillsJSON != "" {
			if err := json.Unmarshal([]byte(skillsJSON), &role.Skills); err != nil {
				log.Printf("Error parsing skills: %v", err)
			}
		}

		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &role.Tags); err != nil {
				log.Printf("Error parsing tags: %v", err)
			}
		}

		roles = append(roles, role)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d roles", len(roles)),
		Data:    roles,
	})
}

func GetRoleHandler(c *gin.Context) {
	roleUUID := c.Query("uuid")
	if roleUUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	sqlStr := `
		SELECT id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, created_at, updated_at
		FROM agent_role
		WHERE uuid = $1 AND status != 'deleted'
	`

	var role Role
	var skillsJSON, tagsJSON string
	var agentNaming, slogan, description, icon sql.NullString

	err := db.QueryRow(sqlStr, roleUUID).Scan(
		&role.UUID,
		&role.UUID,
		&role.Name,
		&agentNaming,
		&slogan,
		&description,
		&skillsJSON,
		&role.Status,
		&icon,
		&tagsJSON,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Role not found: %s", roleUUID),
		})
		return
	}
	if err != nil {
		log.Printf("Error getting role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role: %v", err),
		})
		return
	}

	role.AgentNaming = agentNaming.String
	role.Slogan = slogan.String
	role.Description = description.String
	role.Icon = icon.String

	if skillsJSON != "" {
		if err := json.Unmarshal([]byte(skillsJSON), &role.Skills); err != nil {
			log.Printf("Error parsing skills: %v", err)
		}
	}

	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &role.Tags); err != nil {
			log.Printf("Error parsing tags: %v", err)
		}
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role found",
		Data:    role,
	})
}

func CreateRoleHandler(c *gin.Context) {
	var input struct {
		Name          string                 `json:"name" binding:"required"`
		AgentNaming   string                 `json:"agent_naming"`
		Slogan        string                 `json:"slogan"`
		Description   string                 `json:"description"`
		Skills        []interface{}          `json:"skills"`
		Status        string                 `json:"status"`
		Icon          string                 `json:"icon"`
		Tags          []interface{}          `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if input.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role name is required",
		})
		return
	}

	if input.Status == "" {
		input.Status = "draft"
	}

	roleUUID := uuid.New().String()
	roleID := fmt.Sprintf("role-%s", roleUUID[:8])
	now := time.Now()

	var skillsJSON string
	if input.Skills != nil && len(input.Skills) > 0 {
		skillsBytes, _ := json.Marshal(input.Skills)
		skillsJSON = string(skillsBytes)
	} else {
		skillsJSON = "[]"
	}

	var tagsJSON string
	if input.Tags != nil && len(input.Tags) > 0 {
		tagsBytes, _ := json.Marshal(input.Tags)
		tagsJSON = string(tagsBytes)
	} else {
		tagsJSON = "[]"
	}

	sqlStr := `
		INSERT INTO agent_role (id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, current_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1, $11, $12)
	`

	_, err := db.Exec(sqlStr, roleID, roleUUID, input.Name, input.AgentNaming, input.Slogan, input.Description, skillsJSON, input.Status, input.Icon, tagsJSON, now, now)

	if err != nil {
		log.Printf("Error creating role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to create role: %v", err),
		})
		return
	}

	log.Printf("Role created successfully: %s", roleUUID)

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Role created successfully",
		Data: map[string]interface{}{
			"uuid": roleUUID,
			"name": input.Name,
		},
	})
}

func UpdateRoleHandler(c *gin.Context) {
	var input struct {
		UUID          string                 `json:"uuid" binding:"required"`
		Name          string                 `json:"name"`
		AgentNaming   string                 `json:"agent_naming"`
		Slogan        string                 `json:"slogan"`
		Description   string                 `json:"description"`
		Skills        []interface{}          `json:"skills"`
		Status        string                 `json:"status"`
		Icon          string                 `json:"icon"`
		Tags          []interface{}          `json:"tags"`
		ChangeDesc    string                 `json:"change_desc"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if input.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	changeDesc := input.ChangeDesc
	if changeDesc == "" {
		changeDesc = "Manual update"
	}

	newVersion, err := saveRoleVersion(input.UUID, changeDesc)
	if err != nil {
		log.Printf("Error saving role version before update: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to save role version: %v", err),
		})
		return
	}

	now := time.Now()

	var setClauses []string
	var args []interface{}
	argIndex := 2

	if input.Name != "" {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, input.Name)
		argIndex++
	}

	if input.AgentNaming != "" {
		setClauses = append(setClauses, fmt.Sprintf("agent_naming = $%d", argIndex))
		args = append(args, input.AgentNaming)
		argIndex++
	}

	if input.Slogan != "" {
		setClauses = append(setClauses, fmt.Sprintf("slogan = $%d", argIndex))
		args = append(args, input.Slogan)
		argIndex++
	}

	if input.Description != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, input.Description)
		argIndex++
	}

	if input.Skills != nil {
		skillsBytes, _ := json.Marshal(input.Skills)
		setClauses = append(setClauses, fmt.Sprintf("skills = $%d::jsonb", argIndex))
		args = append(args, string(skillsBytes))
		argIndex++
	}

	if input.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, input.Status)
		argIndex++
	}

	if input.Icon != "" {
		setClauses = append(setClauses, fmt.Sprintf("icon = $%d", argIndex))
		args = append(args, input.Icon)
		argIndex++
	}

	if input.Tags != nil {
		tagsBytes, _ := json.Marshal(input.Tags)
		setClauses = append(setClauses, fmt.Sprintf("tags = $%d::jsonb", argIndex))
		args = append(args, string(tagsBytes))
		argIndex++
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, now)

	// 第一个参数是 uuid
	args = append([]interface{}{input.UUID}, args...)

	sqlStr := fmt.Sprintf("UPDATE agent_role SET %s WHERE uuid = $1 AND status != 'deleted'", strings.Join(setClauses, ", "))

	result, err := db.Exec(sqlStr, args...)

	if err != nil {
		log.Printf("Error updating role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to update role: %v", err),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "Role not found",
		})
		return
	}

	log.Printf("Role updated successfully: %s", input.UUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role updated successfully",
		Data: map[string]interface{}{
			"uuid": input.UUID,
			"version": newVersion,
		},
	})
}

func DeleteRoleHandler(c *gin.Context) {
	var input struct {
		UUID string `json:"uuid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if input.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	now := time.Now()

	sqlStr := `UPDATE agent_role SET status = 'deleted', updated_at = $2 WHERE uuid = $1 AND status != 'deleted'`

	result, err := db.Exec(sqlStr, input.UUID, now)

	if err != nil {
		log.Printf("Error deleting role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to delete role: %v", err),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "Role not found",
		})
		return
	}

	log.Printf("Role deleted successfully: %s", input.UUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role deleted successfully",
	})
}

func GetRoleByNameHandler(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role name is required",
		})
		return
	}

	sqlStr := `
		SELECT id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, created_at, updated_at
		FROM agent_role
		WHERE name = $1 AND status != 'deleted'
	`

	var role Role
	var skillsJSON, tagsJSON string
	var agentNaming, slogan, description, icon sql.NullString

	err := db.QueryRow(sqlStr, name).Scan(
		&role.UUID,
		&role.UUID,
		&role.Name,
		&agentNaming,
		&slogan,
		&description,
		&skillsJSON,
		&role.Status,
		&icon,
		&tagsJSON,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Role not found: %s", name),
		})
		return
	}
	if err != nil {
		log.Printf("Error getting role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role: %v", err),
		})
		return
	}

	role.AgentNaming = agentNaming.String
	role.Slogan = slogan.String
	role.Description = description.String
	role.Icon = icon.String

	if skillsJSON != "" {
		if err := json.Unmarshal([]byte(skillsJSON), &role.Skills); err != nil {
			log.Printf("Error parsing skills: %v", err)
		}
	}

	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &role.Tags); err != nil {
			log.Printf("Error parsing tags: %v", err)
		}
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role found",
		Data:    role,
	})
}

func saveRoleVersion(roleUUID string, changeDesc string) (int, error) {
	var currentVersion int
	var role Role
	var skillsJSON, tagsJSON string
	var agentNaming, slogan, description, icon sql.NullString

	err := db.QueryRow(`SELECT current_version FROM agent_role WHERE uuid = $1`, roleUUID).Scan(&currentVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to get current version: %v", err)
	}

	err = db.QueryRow(`
		SELECT id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags
		FROM agent_role WHERE uuid = $1
	`, roleUUID).Scan(&role.UUID, &role.UUID, &role.Name, &agentNaming, &slogan, &description, &skillsJSON, &role.Status, &icon, &tagsJSON)

	if err != nil {
		return 0, fmt.Errorf("failed to get role for version: %v", err)
	}

	versionID := fmt.Sprintf("rv-%s-%d", roleUUID[:8], currentVersion)
	now := time.Now()

	_, err = db.Exec(`
		INSERT INTO agent_role_version (id, uuid, role_uuid, version, name, agent_naming, slogan, description, skills, status, icon, tags, change_desc, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, versionID, uuid.New().String(), roleUUID, currentVersion, role.Name, agentNaming.String, slogan.String, description.String, skillsJSON, role.Status, icon.String, tagsJSON, changeDesc, now)

	if err != nil {
		return 0, fmt.Errorf("failed to save role version: %v", err)
	}

	newVersion := currentVersion + 1
	_, err = db.Exec(`UPDATE agent_role SET current_version = $1, updated_at = $2 WHERE uuid = $3`, newVersion, now, roleUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to update role version: %v", err)
	}

	return newVersion, nil
}

func GetRoleVersionsHandler(c *gin.Context) {
	roleUUID := c.Query("uuid_role")
	if roleUUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	sqlStr := `
		SELECT id, uuid, role_uuid, version, name, agent_naming, slogan, description, skills, status, icon, tags, change_desc, created_at
		FROM agent_role_version
		WHERE role_uuid = $1
		ORDER BY version DESC
	`

	rows, err := db.Query(sqlStr, roleUUID)
	if err != nil {
		log.Printf("Error querying role versions: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role versions: %v", err),
		})
		return
	}
	defer rows.Close()

	var versions []RoleVersion
	for rows.Next() {
		var v RoleVersion
		var agentNaming, slogan, description, icon, changeDesc sql.NullString

		err := rows.Scan(&v.ID, &v.UUID, &v.RoleUUID, &v.Version, &v.Name, &agentNaming, &slogan, &description, &v.Skills, &v.Status, &icon, &v.Tags, &changeDesc, &v.CreatedAt)
		if err != nil {
			log.Printf("Error scanning role version: %v", err)
			continue
		}

		v.AgentNaming = agentNaming.String
		v.Slogan = slogan.String
		v.Description = description.String
		v.Icon = icon.String
		v.ChangeDesc = changeDesc.String

		versions = append(versions, v)
	}

	var currentVersion int
	db.QueryRow(`SELECT current_version FROM agent_role WHERE uuid = $1`, roleUUID).Scan(&currentVersion)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d versions", len(versions)),
		Data: map[string]interface{}{
			"versions": versions,
			"current_version": currentVersion,
		},
	})
}

func GetRoleVersionDetailHandler(c *gin.Context) {
	roleUUID := c.Query("uuid_role")
	versionStr := c.Query("version")

	if roleUUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	if versionStr == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Version is required",
		})
		return
	}

	version := 0
	fmt.Sscanf(versionStr, "%d", &version)

	sqlStr := `
		SELECT id, uuid, role_uuid, version, name, agent_naming, slogan, description, skills, status, icon, tags, change_desc, created_at
		FROM agent_role_version
		WHERE role_uuid = $1 AND version = $2
	`

	var v RoleVersion
	var agentNaming, slogan, description, icon, changeDesc sql.NullString

	err := db.QueryRow(sqlStr, roleUUID, version).Scan(&v.ID, &v.UUID, &v.RoleUUID, &v.Version, &v.Name, &agentNaming, &v.Slogan, &v.Description, &v.Skills, &v.Status, &v.Icon, &v.Tags, &v.ChangeDesc, &v.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Role version not found: %d", version),
		})
		return
	}
	if err != nil {
		log.Printf("Error getting role version: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role version: %v", err),
		})
		return
	}

	v.AgentNaming = agentNaming.String
	v.Slogan = slogan.String
	v.Description = description.String
	v.Icon = icon.String
	v.ChangeDesc = changeDesc.String

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role version found",
		Data:    v,
	})
}

func RollbackRoleHandler(c *gin.Context) {
	var input struct {
		UUID     string `json:"uuid_role" binding:"required"`
		Version  int    `json:"version" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	var v RoleVersion
	var agentNaming, slogan, description, icon sql.NullString

	err := db.QueryRow(`
		SELECT name, agent_naming, slogan, description, skills, status, icon, tags
		FROM agent_role_version
		WHERE role_uuid = $1 AND version = $2
	`, input.UUID, input.Version).Scan(&v.Name, &agentNaming, &slogan, &description, &v.Skills, &v.Status, &icon, &v.Tags)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Role version not found: %d", input.Version),
		})
		return
	}
	if err != nil {
		log.Printf("Error getting role version for rollback: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role version: %v", err),
		})
		return
	}

	_, err = saveRoleVersion(input.UUID, fmt.Sprintf("Rollback to version %d", input.Version))
	if err != nil {
		log.Printf("Error saving rollback version: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to save rollback version: %v", err),
		})
		return
	}

	now := time.Now()
	_, err = db.Exec(`
		UPDATE agent_role SET name = $2, agent_naming = $3, slogan = $4, description = $5, skills = $6, status = $7, icon = $8, tags = $9, updated_at = $10
		WHERE uuid = $1
	`, input.UUID, v.Name, agentNaming.String, slogan.String, description.String, v.Skills, v.Status, icon.String, v.Tags, now)

	if err != nil {
		log.Printf("Error rolling back role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to rollback role: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Role rolled back to version %d", input.Version),
	})
}

func CompareRoleVersionsHandler(c *gin.Context) {
	roleUUID := c.Query("uuid_role")
	version1Str := c.Query("version1")
	version2Str := c.Query("version2")

	if roleUUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	if version1Str == "" || version2Str == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Both version1 and version2 are required",
		})
		return
	}

	version1 := 0
	version2 := 0
	fmt.Sscanf(version1Str, "%d", &version1)
	fmt.Sscanf(version2Str, "%d", &version2)

	var v1, v2 RoleVersion
	var agentNaming1, slogan1, desc1, icon1, changeDesc1 sql.NullString
	var agentNaming2, slogan2, desc2, icon2, changeDesc2 sql.NullString

	err := db.QueryRow(`
		SELECT name, agent_naming, slogan, description, skills, status, icon, tags, change_desc
		FROM agent_role_version
		WHERE role_uuid = $1 AND version = $2
	`, roleUUID, version1).Scan(&v1.Name, &agentNaming1, &v1.Slogan, &desc1, &v1.Skills, &v1.Status, &icon1, &v1.Tags, &changeDesc1)

	if err != nil {
		log.Printf("Error getting version1: %v", err)
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Version %d not found", version1),
		})
		return
	}

	err = db.QueryRow(`
		SELECT name, agent_naming, slogan, description, skills, status, icon, tags, change_desc
		FROM agent_role_version
		WHERE role_uuid = $1 AND version = $2
	`, roleUUID, version2).Scan(&v2.Name, &agentNaming2, &v2.Slogan, &desc2, &v2.Skills, &v2.Status, &icon2, &v2.Tags, &changeDesc2)

	if err != nil {
		log.Printf("Error getting version2: %v", err)
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Version %d not found", version2),
		})
		return
	}

	v1.AgentNaming = agentNaming1.String
	v1.Slogan = slogan1.String
	v1.Description = desc1.String
	v1.Icon = icon1.String
	v1.ChangeDesc = changeDesc1.String
	v2.AgentNaming = agentNaming2.String
	v2.Slogan = slogan2.String
	v2.Description = desc2.String
	v2.Icon = icon2.String
	v2.ChangeDesc = changeDesc2.String

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Versions compared",
		Data: map[string]interface{}{
			"version1": v1,
			"version2": v2,
		},
	})
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
