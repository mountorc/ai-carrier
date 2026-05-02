package skill

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/trae/autoFlow/carriercore/oss"
)

type SkillListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type SkillCreateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Nick        string                 `json:"nick"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Download    map[string]interface{} `json:"download"`
}

type SkillRegisterRequest struct {
	Name        string   `json:"name" binding:"required"`
	Version     string   `json:"version"`
	Nick        string   `json:"nick"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	FileURL     string   `json:"file_url"`
	FileSize    int64    `json:"file_size"`
	PackageName string   `json:"package_name"`
}

type SkillUpdateRequest struct {
	UUID        string                 `json:"uuid" binding:"required"`
	Name        string                 `json:"name"`
	Nick        string                 `json:"nick"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Download    map[string]interface{} `json:"download"`
}

type SkillUploadRequest struct {
	Name         string `form:"name" binding:"required"`
	Version      string `form:"version" binding:"required"`
	Description  string `form:"description"`
	Author       string `form:"author"`
	Category     string `form:"category"`
	Tags         string `form:"tags"`
	ReleaseNotes string `form:"release_notes"`
}

func GetSkillListHandler(c *gin.Context) {
	var req SkillListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("Error parsing skill list request: %v", err)
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := GetSkillList()
	if err != nil {
		log.Printf("Error getting skill list: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get skill list: %v", err),
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
	uuid := c.Query("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill UUID is required",
		})
		return
	}

	result, err := GetSkillByUUID(uuid)
	if err != nil {
		log.Printf("Error getting skill: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get skill: %v", err),
		})
		return
	}

	if result.Count == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Skill not found: %s", uuid),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill found",
		Data:    result.Rows[0],
	})
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

	if req.Version == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill version is required",
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

	fileExt := ""
	if idx := lastDotIndex(handler.Filename); idx != -1 {
		fileExt = handler.Filename[idx:]
	}
	formattedFileName := fmt.Sprintf("%s_%s%s", sanitizeFileName(req.Name), sanitizeFileName(req.Version), fileExt)
	log.Printf("Uploading skill package with formatted name: %s", formattedFileName)

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

	existingFiles, err := ossClient.ListFilesByToken(token, basePath)
	if err == nil && existingFiles != nil {
		for _, fileInfo := range existingFiles {
			if fileName, ok := fileInfo["name"].(string); ok && fileName == formattedFileName {
				c.JSON(http.StatusConflict, Response{
					Success: false,
					Message: fmt.Sprintf("File %s already exists in OSS. Please use a different version.", formattedFileName),
				})
				return
			}
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

	fileUrl, err := ossClient.UploadByToken(token, formattedFileName, fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to upload file to OSS: %v", err),
		})
		return
	}

	log.Printf("File uploaded to OSS: %s", fileUrl)

	now := time.Now()
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

	skillUUID := ""
	existingSkillResult, err := GetSkillByName(req.Name)

	if err == nil && existingSkillResult != nil && existingSkillResult.Count > 0 {
		skillRow := existingSkillResult.Rows[0]
		if uuidVal, ok := skillRow["uuid"].(string); ok && uuidVal != "" {
			skillUUID = uuidVal
			log.Printf("Existing skill found, reusing UUID: %s", skillUUID)
		} else {
			skillUUID = uuid.New().String()
		}
	} else {
		skillUUID = uuid.New().String()
		downloadData := map[string]interface{}{
			"url":     fileUrl,
			"size":    int64(len(fileContent)),
			"package": formattedFileName,
			"version": req.Version,
		}
		downloadBytes, _ := json.Marshal(downloadData)
		downloadJSON := string(downloadBytes)

		var resultUUID string
		err = pgDB.QueryRow(`
			INSERT INTO skill_store_skills (uuid, name, nick, description, author, type, download, tags, created_at, updated_at, is_deleted)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, FALSE)
			RETURNING uuid`,
			skillUUID, req.Name, req.Category, req.Description, req.Author, req.Category, downloadJSON, tagsJSON, now, now).Scan(&resultUUID)

		if err != nil {
			log.Printf("Error creating new skill: %v", err)
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: fmt.Sprintf("Failed to create skill in database: %v", err),
			})
			return
		}

		log.Printf("New skill created successfully: %s", resultUUID)
	}

	existingVersionResult, err := GetSkillVersion(skillUUID, req.Version)
	if err == nil && existingVersionResult != nil && existingVersionResult.Count > 0 {
		c.JSON(http.StatusConflict, Response{
			Success: false,
			Message: fmt.Sprintf("Skill version %s already exists. Please use a different version number.", req.Version),
		})
		return
	}

	skillVersionUUID := uuid.New().String()
	var versionResultUUID string
	err = pgDB.QueryRow(`
		INSERT INTO skill_store_skill_versions (uuid, skill_uuid, version, download_url, file_name, file_size, release_notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING uuid`,
		skillVersionUUID, skillUUID, req.Version, fileUrl, formattedFileName, int64(len(fileContent)), req.ReleaseNotes, now, now).Scan(&versionResultUUID)

	if err != nil {
		log.Printf("Error creating skill version record: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to create skill version in database: %v", err),
		})
		return
	}

	log.Printf("Skill version record created: %s", versionResultUUID)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill uploaded and version tracked successfully",
		Data: map[string]interface{}{
			"uuid":          skillUUID,
			"version_uuid":  skillVersionUUID,
			"name":          req.Name,
			"version":       req.Version,
			"author":        req.Author,
			"file_url":      fileUrl,
			"file_name":     formattedFileName,
			"file_size":     int64(len(fileContent)),
			"release_notes": req.ReleaseNotes,
		},
	})
}

func lastDotIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

func sanitizeFileName(name string) string {
	result := strings.ReplaceAll(name, " ", "_")
	specialChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
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
		SET name = COALESCE(NULLIF($1, ''), name),
			nick = COALESCE(NULLIF($2, ''), nick),
			description = COALESCE(NULLIF($3, ''), description),
			author = COALESCE(NULLIF($4, ''), author),
			type = COALESCE(NULLIF($5, ''), type),
			download = COALESCE(NULLIF($6, ''), download),
			updated_at = $7
		WHERE uuid = $8 AND is_deleted = FALSE
		RETURNING uuid`,
		req.Name, req.Nick, req.Description, req.Author, req.Category, downloadJSON, time.Now(), req.UUID).Scan(&resultUUID)

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
			"name": req.Name,
		},
	})
}

func DeleteSkillHandler(c *gin.Context) {
	var req struct {
		UUID string `json:"uuid" binding:"required"`
	}
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
			Message: "Skill UUID is required",
		})
		return
	}

	now := time.Now()

	var resultUUID string
	err := pgDB.QueryRow(`
		UPDATE skill_store_skills SET is_deleted = TRUE, deleted_at = $2 WHERE uuid = $1 AND is_deleted = FALSE RETURNING uuid`,
		req.UUID, now).Scan(&resultUUID)

	if err != nil {
		log.Printf("Error soft deleting skill: %v", err)
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

func ListSkillVersionsHandler(c *gin.Context) {
	skillUUID := c.Query("skill_uuid")
	if skillUUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill UUID is required",
		})
		return
	}

	result, err := ListSkillVersions(skillUUID)
	if err != nil {
		log.Printf("Error listing skill versions: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to list skill versions: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d versions", result.Count),
		Data:    result.Rows,
	})
}

func GetSkillVersionHandler(c *gin.Context) {
	skillUUID := c.Query("skill_uuid")
	version := c.Query("version")

	if skillUUID == "" || version == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill UUID and version are required",
		})
		return
	}

	result, err := GetSkillVersion(skillUUID, version)
	if err != nil {
		log.Printf("Error getting skill version: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get skill version: %v", err),
		})
		return
	}

	if result.Count == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("Skill version not found for skill %s and version %s", skillUUID, version),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Skill version found",
		Data:    result.Rows[0],
	})
}

func GetLatestSkillVersionHandler(c *gin.Context) {
	skillUUID := c.Query("skill_uuid")
	if skillUUID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Skill UUID is required",
		})
		return
	}

	result, err := GetLatestSkillVersion(skillUUID)
	if err != nil {
		log.Printf("Error getting latest skill version: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: fmt.Sprintf("Failed to get latest skill version: %v", err),
		})
		return
	}

	if result.Count == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: fmt.Sprintf("No versions found for skill %s", skillUUID),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Latest skill version found",
		Data:    result.Rows[0],
	})
}
