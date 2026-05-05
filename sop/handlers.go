package sop

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetSOPListHandler(c *gin.Context) {
	log.Println("Getting SOP list from database...")

	result, err := GetSOPList()
	if err != nil {
		log.Printf("Error getting SOP list: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get SOPs: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d SOPs", result.Count),
		Data:    result.Rows,
	})
}

func GetSOPHandler(c *gin.Context) {
	sopUUID := c.Query("uuid")
	if sopUUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Missing required parameter: uuid",
		})
		return
	}

	log.Printf("Getting SOP with UUID: %s", sopUUID)

	result, err := GetSOPByUUID(sopUUID)
	if err != nil {
		log.Printf("Error querying SOP: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get SOP: %v", err),
		})
		return
	}

	if result.Count == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("SOP with UUID %s not found", sopUUID),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "SOP retrieved successfully",
		Data:    result.Rows[0],
	})
}

type SOPCreateRequest struct {
	Name        string          `json:"name"`
	Nick        string          `json:"nick,omitempty"`
	Description string          `json:"description,omitempty"`
	Content     string          `json:"content,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Category    string          `json:"category,omitempty"`
	Tags        string          `json:"tags,omitempty"`
	Status      string          `json:"status,omitempty"`
	Version     string          `json:"version,omitempty"`
	Project     string          `json:"project,omitempty"`
}

func CreateSOPHandler(c *gin.Context) {
	var req SOPCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing create SOP request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "SOP name is required",
		})
		return
	}

	sopUUID := uuid.New().String()
	now := time.Now()

	var tagsJSON string
	if req.Tags != "" {
		tagsJSON = fmt.Sprintf(`["%s"]`, req.Tags)
	} else {
		tagsJSON = "[]"
	}

	if req.Status == "" {
		req.Status = "draft"
	}
	if req.Version == "" {
		req.Version = "1.0"
	}

	configBytes := []byte("{}")
	if req.Config != nil {
		configBytes = req.Config
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		INSERT INTO sops (uuid, name, nick, description, content, config, category, tags, status, version, project, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10, $11, $12, $13, FALSE)
		RETURNING uuid`,
		sopUUID, req.Name, req.Nick, req.Description, req.Content, configBytes, req.Category, tagsJSON, req.Status, req.Version, req.Project, now, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error inserting SOP: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to create SOP: %v", err),
		})
		return
	}

	log.Printf("SOP created successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "SOP created successfully",
		Data: map[string]interface{}{
			"uuid": resultUUID,
			"name": req.Name,
		},
	})
}

type SOPUpdateRequest struct {
	UUID        string          `json:"uuid"`
	Name        string          `json:"name,omitempty"`
	Nick        string          `json:"nick,omitempty"`
	Description string          `json:"description,omitempty"`
	Content     string          `json:"content,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Category    string          `json:"category,omitempty"`
	Tags        string          `json:"tags,omitempty"`
	Status      string          `json:"status,omitempty"`
	Version     string          `json:"version,omitempty"`
	Project     string          `json:"project,omitempty"`
}

func UpdateSOPHandler(c *gin.Context) {
	var req SOPUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing update SOP request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "SOP UUID is required",
		})
		return
	}

	var tagsJSON string
	if req.Tags != "" {
		tagsJSON = fmt.Sprintf(`["%s"]`, req.Tags)
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		UPDATE sops
		SET name = COALESCE(NULLIF($1, ''), name),
			nick = COALESCE(NULLIF($2, ''), nick),
			description = COALESCE(NULLIF($3, ''), description),
			content = COALESCE(NULLIF($4, ''), content),
			config = COALESCE($5::jsonb, config),
			category = COALESCE(NULLIF($6, ''), category),
			tags = COALESCE(NULLIF($7, ''), tags),
			status = COALESCE(NULLIF($8, ''), status),
			version = COALESCE(NULLIF($9, ''), version),
			project = $10,
			updated_at = $11
		WHERE uuid = $12 AND is_deleted = FALSE
		RETURNING uuid`,
		req.Name, req.Nick, req.Description, req.Content, req.Config, req.Category, tagsJSON, req.Status, req.Version, req.Project, time.Now(), req.UUID).Scan(&resultUUID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("SOP with UUID %s not found", req.UUID),
		})
		return
	}

	if err != nil {
		log.Printf("Error updating SOP: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to update SOP: %v", err),
		})
		return
	}

	log.Printf("SOP updated successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "SOP updated successfully",
		Data: map[string]interface{}{
			"uuid": resultUUID,
		},
	})
}

type SOPDeleteRequest struct {
	UUID string `json:"uuid"`
}

func DeleteSOPHandler(c *gin.Context) {
	var req SOPDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing delete SOP request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "SOP UUID is required",
		})
		return
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		UPDATE sops
		SET is_deleted = TRUE, deleted_at = $2
		WHERE uuid = $1 AND is_deleted = FALSE
		RETURNING uuid`,
		req.UUID, time.Now()).Scan(&resultUUID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("SOP with UUID %s not found or already deleted", req.UUID),
		})
		return
	}

	if err != nil {
		log.Printf("Error deleting SOP: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to delete SOP: %v", err),
		})
		return
	}

	log.Printf("SOP soft deleted successfully: %s", req.UUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "SOP deleted successfully",
	})
}
