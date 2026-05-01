package skill

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	mountsql "github.com/trae/autoFlow/carriercore/mountcore/sql"
	"github.com/trae/autoFlow/carriercore/oss"
)

func GetSkillListHandler(c *gin.Context) {
	log.Println("Getting skill list from database...")

	carrierAgentUUID := c.Query("carrier_agent_uuid")
	if carrierAgentUUID != "" {
		valid, message := validateAgentUUID(carrierAgentUUID)
		if !valid {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: message,
			})
			return
		}
		log.Printf("Agent validated: %s", message)
	}

	queryText := c.Query("query")
	vectorText := c.Query("vectorText")

	var result *mountsql.QueryResult
	var err error

	if queryText != "" || vectorText != "" {
		searchText := queryText
		if vectorText != "" {
			searchText = vectorText
		}

		log.Printf("Performing vector search for: %s", searchText)

		result, err = SearchSkillsByVector(searchText)
	} else {
		result, err = GetSkillList()
	}

	if err != nil {
		log.Printf("Error getting skill list: %v", err)
		if strings.Contains(err.Error(), "executor not initialized") || strings.Contains(err.Error(), "embedding service") {
			c.JSON(http.StatusServiceUnavailable, Response{
				Success: false,
				Message: "Database connection not available. Skill service is temporarily unavailable.",
			})
		} else {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: fmt.Sprintf("Failed to get skills: %v", err),
			})
		}
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d skills", result.Count),
		Data:    result.Rows,
	})
}

func getSkillHandler(c *gin.Context) {
	carrierAgentUUID := c.Query("carrier_agent_uuid")
	if carrierAgentUUID != "" {
		valid, message := validateAgentUUID(carrierAgentUUID)
		if !valid {
			c.JSON(http.StatusUnauthorized, Response{
				Success: false,
				Message: message,
			})
			return
		}
		log.Printf("Agent validated: %s", message)
	}

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

	result, err := GetSkillByUUID(skillUUID)
	if err != nil {
		log.Printf("Error querying skill: %v", err)
		if strings.Contains(err.Error(), "未找到 UUID") {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Message: fmt.Sprintf("Skill with UUID %s not found", skillUUID),
			})
		} else {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: fmt.Sprintf("Failed to get skill: %v", err),
			})
		}
		return
	}

	if result.Count == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Skill with UUID %s not found", skillUUID),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill retrieved successfully",
		Data:    result.Rows[0],
	})
}

func validateAgentUUID(uuid string) (bool, string) {
	if uuid == "" {
		return false, "Agent UUID is empty"
	}

	if len(uuid) < 32 {
		return false, "Invalid agent UUID format"
	}

	return true, "Agent UUID validated successfully"
}

func insertApiLog(action, path, method string, status int, message, requestBody, responseBody, identityToken, carrierUserID, carrierAgentUUID, clientIP string) {
	log.Printf("[API Log] Action: %s, Path: %s, Method: %s, Status: %d, Message: %s",
		action, path, method, status, message)
}

type SkillCreateRequest struct {
	Name        string                 `json:"name"`
	Nick        string                 `json:"nick,omitempty"`
	Description string                 `json:"description,omitempty"`
	Download    map[string]interface{} `json:"download,omitempty"`
}

type SkillUpdateRequest struct {
	UUID        string                 `json:"uuid"`
	Name        string                 `json:"name,omitempty"`
	Nick        string                 `json:"nick,omitempty"`
	Description string                 `json:"description,omitempty"`
	Download    map[string]interface{} `json:"download,omitempty"`
}

type SkillDeleteRequest struct {
	UUID string `json:"uuid"`
}

type SkillRegisterRequest struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Description string      `json:"description,omitempty"`
	Author      string      `json:"author,omitempty"`
	Category    string      `json:"category,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	FileURL     string      `json:"file_url,omitempty"`
	FileSize    int64       `json:"file_size,omitempty"`
	PackageName string      `json:"package_name,omitempty"`
	SkillPath   string      `json:"skill_path,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
}

type SkillUploadRequest struct {
	Name        string `form:"name" binding:"required"`
	Version     string `form:"version"`
	Description string `form:"description"`
	Author      string `form:"author"`
	Category    string `form:"category"`
	Tags        string `form:"tags"`
}

func CreateSkillHandler(c *gin.Context) {
	var req SkillCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing create skill request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill name is required",
		})
		return
	}

	skillUUID := uuid.New().String()
	now := time.Now()

	var downloadJSON string
	if req.Download != nil {
		downloadJSON = fmt.Sprintf("%v", req.Download)
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		INSERT INTO skill_store_skills (uuid, name, nick, description, download, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING uuid`,
		skillUUID, req.Name, req.Nick, req.Description, downloadJSON, now, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error inserting skill: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to create skill: %v", err),
		})
		return
	}

	log.Printf("Skill created successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill created successfully",
		Data: map[string]interface{}{
			"uuid": resultUUID,
			"name": req.Name,
		},
	})
}

func RegisterSkillHandler(c *gin.Context) {
	var req SkillRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing register skill request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill name is required",
		})
		return
	}

	if req.Version == "" {
		req.Version = "1.0.0"
	}

	skillUUID := uuid.New().String()
	now := time.Now()

	var tagsJSON string
	if req.Tags != nil && len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsJSON = string(tagsBytes)
	} else {
		tagsJSON = "[]"
	}

	var downloadJSON string
	if req.FileURL != "" {
		downloadData := map[string]interface{}{
			"url":     req.FileURL,
			"size":    req.FileSize,
			"package": req.PackageName,
		}
		downloadBytes, _ := json.Marshal(downloadData)
		downloadJSON = string(downloadBytes)
	} else {
		downloadJSON = "{}"
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		INSERT INTO skill_store_skills (uuid, name, nick, description, author, type, download, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING uuid`,
		skillUUID, req.Name, req.Category, req.Description, req.Author, req.Category, downloadJSON, tagsJSON, now, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error registering skill: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to register skill: %v", err),
		})
		return
	}

	log.Printf("Skill registered successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill registered successfully",
		Data: map[string]interface{}{
			"uuid":         resultUUID,
			"name":         req.Name,
			"version":      req.Version,
			"author":       req.Author,
			"file_url":     req.FileURL,
			"package_name": req.PackageName,
		},
	})
}

func UploadSkillHandler(c *gin.Context) {
	var req SkillUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid form data: %v", err),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill name is required",
		})
		return
	}

	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get file from form: %v", err),
		})
		return
	}
	defer file.Close()

	fileName := handler.Filename
	log.Printf("Uploading skill package: %s", fileName)

	ossClient := oss.GetInstance()
	if ossClient == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	tokens := ossClient.GetAllTokens()
	if len(tokens) == 0 {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "No OSS tokens available",
		})
		return
	}

	token := ""
	basePath := "carrier/skills/"
	for t, info := range tokens {
		if info.BasePath == basePath {
			token = t
			break
		}
	}
	if token == "" {
		for t := range tokens {
			token = t
			break
		}
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to read file content: %v", err),
		})
		return
	}

	fileUrl, err := ossClient.UploadByToken(token, fileName, fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to upload file to OSS: %v", err),
		})
		return
	}

	log.Printf("File uploaded to OSS: %s", fileUrl)

	skillUUID := uuid.New().String()
	now := time.Now()

	if req.Version == "" {
		req.Version = "1.0.0"
	}
	if req.Category == "" {
		req.Category = "general"
	}

	var tagsJSON string
	if req.Tags != "" {
		var tagsArray []string
		if err := json.Unmarshal([]byte(req.Tags), &tagsArray); err == nil {
			tagsBytes, _ := json.Marshal(tagsArray)
			tagsJSON = string(tagsBytes)
		} else {
			tagsJSON = fmt.Sprintf(`["%s"]`, req.Tags)
		}
	} else {
		tagsJSON = "[]"
	}

	downloadData := map[string]interface{}{
		"url":     fileUrl,
		"size":    len(fileContent),
		"package": fileName,
	}
	downloadBytes, _ := json.Marshal(downloadData)
	downloadJSON := string(downloadBytes)

	var resultUUID string
	err = pgDB.QueryRow(`
		INSERT INTO skill_store_skills (uuid, name, nick, description, author, type, download, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING uuid`,
		skillUUID, req.Name, req.Category, req.Description, req.Author, req.Category, downloadJSON, tagsJSON, now, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error registering skill after OSS upload: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to register skill in database: %v", err),
		})
		return
	}

	log.Printf("Skill uploaded and registered successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill uploaded and registered successfully",
		Data: map[string]interface{}{
			"uuid":         resultUUID,
			"name":         req.Name,
			"version":      req.Version,
			"author":       req.Author,
			"file_url":     fileUrl,
			"package_name": fileName,
			"file_size":    len(fileContent),
		},
	})
}

func UpdateSkillHandler(c *gin.Context) {
	var req SkillUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing update skill request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill UUID is required",
		})
		return
	}

	var downloadJSON string
	if req.Download != nil {
		downloadJSON = fmt.Sprintf("%v", req.Download)
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		UPDATE skill_store_skills 
		SET name = COALESCE($1, name), 
			nick = COALESCE($2, nick), 
			description = COALESCE($3, description), 
			download = COALESCE($4, download), 
			updated_at = $5 
		WHERE uuid = $6
		RETURNING uuid`,
		req.Name, req.Nick, req.Description, downloadJSON, time.Now(), req.UUID).Scan(&resultUUID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Skill with UUID %s not found", req.UUID),
		})
		return
	}

	if err != nil {
		log.Printf("Error updating skill: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to update skill: %v", err),
		})
		return
	}

	log.Printf("Skill updated successfully: %s", resultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill updated successfully",
		Data: map[string]interface{}{
			"uuid": resultUUID,
		},
	})
}

func DeleteSkillHandler(c *gin.Context) {
	var req SkillDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error parsing delete skill request: %v", err)
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	if req.UUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill UUID is required",
		})
		return
	}

	var resultUUID string
	err := pgDB.QueryRow(`
		UPDATE skill_store_skills 
		SET is_deleted = TRUE, deleted_at = $2 
		WHERE uuid = $1 AND is_deleted = FALSE
		RETURNING uuid`,
		req.UUID, time.Now()).Scan(&resultUUID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Skill with UUID %s not found or already deleted", req.UUID),
		})
		return
	}

	if err != nil {
		log.Printf("Error deleting skill: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to delete skill: %v", err),
		})
		return
	}

	log.Printf("Skill soft deleted successfully: %s", req.UUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill deleted successfully",
	})
}
