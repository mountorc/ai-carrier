package skill

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	mountsql "github.com/trae/autoFlow/carriercore/mountcore/sql"
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

	result, err := pgDB.Exec(`DELETE FROM skill_store_skills WHERE uuid = $1`, req.UUID)
	if err != nil {
		log.Printf("Error deleting skill: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to delete skill: %v", err),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to delete skill",
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Skill with UUID %s not found", req.UUID),
		})
		return
	}

	log.Printf("Skill deleted successfully: %s", req.UUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill deleted successfully",
	})
}