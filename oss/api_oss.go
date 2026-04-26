package oss

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func RegisterRoutes(router *gin.Engine) {
	ossGroup := router.Group("/oss")

	ossGroup.POST("/upload", handleUpload)
	ossGroup.POST("/upload-by-token", handleUploadByToken)
	ossGroup.POST("/create-folder", handleCreateFolder)
	ossGroup.POST("/generate-token", handleGenerateToken)
	ossGroup.POST("/validate-token", handleValidateToken)
	ossGroup.GET("/tokens", handleListTokens)
	ossGroup.DELETE("/token/:token", handleDeleteToken)

	log.Println("OSS routes registered successfully")
}

func handleUpload(c *gin.Context) {
	var req struct {
		UuidAutoAuth string `json:"uuid_autoAuth"`
		FilePath     string `json:"filePath"`
		Content      string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	client := GetInstance()
	if client == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	fileUrl, err := client.Upload(req.UuidAutoAuth, req.FilePath, []byte(req.Content))
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Upload successful",
		Data:    fileUrl,
	})
}

func handleUploadByToken(c *gin.Context) {
	var req struct {
		Token     string `json:"token"`
		FileName  string `json:"fileName"`
		Content   string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	client := GetInstance()
	if client == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	fileUrl, err := client.UploadByToken(req.Token, req.FileName, []byte(req.Content))
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Upload successful",
		Data:    fileUrl,
	})
}

func handleCreateFolder(c *gin.Context) {
	var req struct {
		UuidAutoAuth string `json:"uuid_autoAuth"`
		FolderPath   string `json:"folderPath"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	client := GetInstance()
	if client == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	err := client.CreateFolder(req.UuidAutoAuth, req.FolderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Folder created successfully",
	})
}

func handleGenerateToken(c *gin.Context) {
	var req struct {
		UuidAutoAuth string `json:"uuid_autoAuth"`
		BasePath     string `json:"basePath"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	client := GetInstance()
	if client == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	token := client.GenerateToken(req.UuidAutoAuth, req.BasePath)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Token generated successfully",
		Data:    token,
	})
}

func handleValidateToken(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	client := GetInstance()
	if client == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	tokenInfo := client.ValidateToken(req.Token)
	if tokenInfo == nil {
		c.JSON(http.StatusUnauthorized, Response{
			Success: false,
			Message: "Invalid or expired token",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Token is valid",
		Data:    tokenInfo,
	})
}

func handleListTokens(c *gin.Context) {
	client := GetInstance()
	if client == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	tokens := client.GetAllTokens()

	var tokenList []interface{}
	for token, info := range tokens {
		tokenList = append(tokenList, map[string]interface{}{
			"token":        token,
			"uuid_autoAuth": info.UuidAutoAuth,
			"basePath":     info.BasePath,
			"createdAt":    info.CreatedAt,
			"expiresAt":    info.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d tokens", len(tokenList)),
		Data:    tokenList,
	})
}

func handleDeleteToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Token is required",
		})
		return
	}

	client := GetInstance()
	if client == nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "OSS client not initialized",
		})
		return
	}

	client.RemoveToken(token)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Token deleted successfully",
	})
}
