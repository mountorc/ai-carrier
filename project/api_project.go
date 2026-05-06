package project

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	manager *ProjectManager
}

func NewProjectHandler(configPath string) (*ProjectHandler, error) {
	manager, err := NewProjectManager(configPath)
	if err != nil {
		return nil, err
	}
	return &ProjectHandler{manager: manager}, nil
}

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects := h.manager.GetAllProjects()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    projects,
	})
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	uuid := c.Param("uuid")
	project, err := h.manager.GetProjectByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    project,
	})
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req struct {
		Project     string `json:"project" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Create not implemented",
	})
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	uuid := c.Param("uuid")
	var req struct {
		Project     string `json:"project"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Update not implemented for " + uuid,
	})
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	uuid := c.Param("uuid")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Delete not implemented for " + uuid,
	})
}

func (h *ProjectHandler) ToggleProject(c *gin.Context) {
	uuid := c.Param("uuid")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Toggle not implemented for " + uuid,
	})
}

func (h *ProjectHandler) ValidateProjectAPI(c *gin.Context) {
	var req struct {
		UUIDProject string `json:"uuid_project"`
		Project     string `json:"project"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request",
		})
		return
	}
	project, err := h.manager.ValidateUUIDProject(req.UUIDProject, req.Project)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    project,
	})
}

func (pm *ProjectManager) GetProjectByUUID(uuid string) (*Project, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.uuidMap[uuid]
	if !exists {
		return nil, fmt.Errorf("project not found")
	}
	return &p, nil
}
