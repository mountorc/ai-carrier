package role

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoleListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type RoleCreateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	AgentNaming string                `json:"agent_naming"`
	Slogan      string                `json:"slogan"`
	Description string                `json:"description"`
	Skills      []map[string]interface{} `json:"skills"`
	Status      string                `json:"status"`
	Icon        string                `json:"icon"`
	Tags        []string              `json:"tags"`
}

type RoleUpdateRequest struct {
	UUID        string                 `json:"uuid" binding:"required"`
	Name        string                 `json:"name"`
	AgentNaming string                `json:"agent_naming"`
	Slogan      string                 `json:"slogan"`
	Description string                 `json:"description"`
	Skills      []map[string]interface{} `json:"skills"`
	Status      string                 `json:"status"`
	Icon        string                 `json:"icon"`
	Tags        []string               `json:"tags"`
}

type RoleDeleteRequest struct {
	UUID string `json:"uuid" binding:"required"`
}

func GetRoleListHandler(c *gin.Context) {
	var req RoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("Error parsing role list request: %v", err)
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := GetRoleList()
	if err != nil {
		log.Printf("Error getting role list: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role list: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d roles", result.Count),
		Data:    result.Rows,
	})
}

func GetRoleHandler(c *gin.Context) {
	uuid := c.Query("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	result, err := GetRoleByUUID(uuid)
	if err != nil {
		log.Printf("Error getting role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role: %v", err),
		})
		return
	}

	if result.Count == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Role not found: %s", uuid),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role found",
		Data:    result.Rows[0],
	})
}

func CreateRoleHandler(c *gin.Context) {
	var req RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing create role request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role name is required",
		})
		return
	}

	if req.Status == "" {
		req.Status = "draft"
	}

	roleUUID := uuid.New().String()
	roleID := fmt.Sprintf("role-%s", roleUUID[:8])
	now := time.Now()

	var skillsJSON string
	if req.Skills != nil && len(req.Skills) > 0 {
		skillsBytes, _ := json.Marshal(req.Skills)
		skillsJSON = string(skillsBytes)
	} else {
		skillsJSON = "[]"
	}

	var tagsJSON string
	if req.Tags != nil && len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsJSON = string(tagsBytes)
	} else {
		tagsJSON = "[]"
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		INSERT INTO role_store (id, uuid, name, agent_naming, slogan, description, skills, status, icon, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING uuid`,
		roleID, roleUUID, req.Name, req.AgentNaming, req.Slogan, req.Description, skillsJSON, req.Status, req.Icon, tagsJSON, now, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error inserting role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to create role: %v", err),
		})
		return
	}

	log.Printf("Role created successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role created successfully",
		Data: map[string]interface{}{
			"uuid": resultUUID,
			"name": req.Name,
		},
	})
}

func UpdateRoleHandler(c *gin.Context) {
	var req RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing update role request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	now := time.Now()

	var skillsJSON string
	if req.Skills != nil && len(req.Skills) > 0 {
		skillsBytes, _ := json.Marshal(req.Skills)
		skillsJSON = string(skillsBytes)
	} else {
		skillsJSON = "[]"
	}

	var tagsJSON string
	if req.Tags != nil && len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsJSON = string(tagsBytes)
	} else {
		tagsJSON = "[]"
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		UPDATE role_store
		SET name = COALESCE(NULLIF($2, ''), name),
			agent_naming = COALESCE(NULLIF($3, ''), agent_naming),
			slogan = COALESCE(NULLIF($4, ''), slogan),
			description = COALESCE(NULLIF($5, ''), description),
			skills = COALESCE(NULLIF($6, ''), skills),
			status = COALESCE(NULLIF($7, ''), status),
			icon = COALESCE(NULLIF($8, ''), icon),
			tags = COALESCE(NULLIF($9, ''), tags),
			updated_at = $10
		WHERE uuid = $1 AND status != 'deleted'
		RETURNING uuid`,
		req.UUID, req.Name, req.AgentNaming, req.Slogan, req.Description, skillsJSON, req.Status, req.Icon, tagsJSON, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error updating role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to update role: %v", err),
		})
		return
	}

	log.Printf("Role updated successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role updated successfully",
		Data: map[string]interface{}{
			"uuid": resultUUID,
			"name": req.Name,
		},
	})
}

func DeleteRoleHandler(c *gin.Context) {
	var req RoleDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Role UUID is required",
		})
		return
	}

	now := time.Now()

	var resultUUID string
	err := pgDB.QueryRow(`
		UPDATE role_store SET status = 'deleted', updated_at = $2 WHERE uuid = $1 AND status != 'deleted' RETURNING uuid`,
		req.UUID, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error deleting role: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to delete role: %v", err),
		})
		return
	}

	log.Printf("Role deleted successfully: %s", req.UUID)

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

	result, err := GetRoleByName(name)
	if err != nil {
		log.Printf("Error getting role by name: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get role: %v", err),
		})
		return
	}

	if result.Count == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Role not found: %s", name),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Role found",
		Data:    result.Rows[0],
	})
}
