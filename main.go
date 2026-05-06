// services/mountcore/main.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmzail/ai-carrier-dev/carriercore/ability"
	"github.com/xmzail/ai-carrier-dev/carriercore/auth"
	"github.com/xmzail/ai-carrier-dev/carriercore/mountcore/sql"
	"github.com/xmzail/ai-carrier-dev/carriercore/oss"
	"github.com/xmzail/ai-carrier-dev/carriercore/project"
	"github.com/xmzail/ai-carrier-dev/carriercore/role"
	"github.com/xmzail/ai-carrier-dev/carriercore/scheduler"
	"github.com/xmzail/ai-carrier-dev/carriercore/sdk/api"
	"github.com/xmzail/ai-carrier-dev/carriercore/skill"
	"github.com/xmzail/ai-carrier-dev/carriercore/sop"
	"github.com/xmzail/ai-carrier-dev/carriercore/task"
	workflowHandlers "github.com/xmzail/ai-carrier-dev/carriercore/workflow"
	"github.com/xmzail/ai-carrier-dev/common/embedding"
	"github.com/xmzail/ai-carrier-dev/mounts/skills"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Println("Initializing MountCore Service...")

	log.Println("Initializing Embedding service...")
	embeddingConfig := embedding.Config{
		ProviderType: "aliyun",
	}
	embeddingService, err := embedding.NewEmbeddingServiceFromConfig(embeddingConfig)
	if err != nil {
		log.Printf("Warning: failed to initialize Embedding service: %v", err)
	} else {
		log.Println("Embedding service initialized successfully (using Aliyun Qwen model)")
		ability.SetEmbeddingService(embeddingService)
		sql.SetEmbeddingService(embeddingService)
		log.Println("Embedding service set successfully in SQL module")
	}

	log.Println("Initializing Workflow module...")
	log.Println("Workflow module initialized successfully")

	log.Println("Initializing Database with MountCore SQL Executor...")
	executor, err := sql.NewExecutorFromConfig("./mount_config.json")
	if err != nil {
		log.Printf("Warning: failed to initialize SQL executor: %v", err)
	} else {
		log.Println("MountCore SQL Executor initialized successfully")
		if embeddingService != nil {
			executor.SetEmbeddingService(embeddingService)
			log.Println("Embedding service set in SQL executor")
		}
		defer executor.Close()
	}

	log.Println("Initializing Ability module...")
	if err := ability.Init(); err != nil {
		log.Printf("Warning: failed to initialize Ability: %v", err)
	} else {
		log.Println("Ability module initialized successfully")
		defer ability.Close()
	}

	log.Println("Initializing Scheduler module...")
	if err := scheduler.Init(); err != nil {
		log.Printf("Warning: failed to initialize Scheduler: %v", err)
	} else {
		log.Println("Scheduler module initialized successfully")
		defer scheduler.Close()
	}

	log.Println("Initializing Skill module...")
	if err := skill.Init(); err != nil {
		log.Printf("Warning: failed to initialize Skill: %v", err)
	} else {
		log.Println("Skill module initialized successfully")
		defer skill.Close()
	}

	log.Println("Initializing Auth module...")
	if err := auth.Init(); err != nil {
		log.Printf("Warning: failed to initialize Auth: %v", err)
	} else {
		log.Println("Auth module initialized successfully")
		defer auth.Close()
	}

	log.Println("Initializing OSS module...")
	if err := oss.Init("./data/oss_token.json"); err != nil {
		log.Printf("Warning: failed to initialize OSS: %v", err)
	} else {
		log.Println("OSS module initialized successfully")
	}

	log.Println("Initializing Task module...")
	if err := task.Init(); err != nil {
		log.Printf("Warning: failed to initialize Task: %v", err)
	} else {
		log.Println("Task module initialized successfully")
	}

	log.Println("Initializing Role module...")
	if err := role.Init(); err != nil {
		log.Printf("Warning: failed to initialize Role: %v", err)
	} else {
		log.Println("Role module initialized successfully")
		defer role.Close()
	}

	log.Println("Initializing SOP module...")
	if err := sop.Init(); err != nil {
		log.Printf("Warning: failed to initialize SOP: %v", err)
	} else {
		log.Println("SOP module initialized successfully")
		defer sop.Close()
	}

	log.Println("Initializing Project module...")
	projectHandler, err := project.NewProjectHandler("../services/ability/data/project.json")
	if err != nil {
		log.Printf("Warning: failed to initialize Project: %v", err)
	} else {
		log.Println("Project module initialized successfully")
	}

	router := gin.Default()
	api.SetupCORS(router)

	log.Println("Registering routes...")
	carrierAPI := router.Group("/carrier")
	carrierAPI.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "MountCore",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})
	log.Println("Health check route registered at /carrier/health")
	workflowHandlers.RegisterRoutes(carrierAPI)
	ability.RegisterRoutes(carrierAPI)
	scheduler.RegisterRoutes(carrierAPI)
	skill.RegisterRoutes(carrierAPI)
	role.RegisterRoutes(carrierAPI)
	sop.RegisterRoutes(carrierAPI)
	auth.RegisterRoutes(carrierAPI)
	oss.RegisterRoutes(carrierAPI)
	task.RegisterRoutes(carrierAPI)
	registerMountSQLRoutes(carrierAPI)

	if projectHandler != nil {
		projectAPI := router.Group("/project")
		projectAPI.GET("/list", projectHandler.ListProjects)
		projectAPI.GET("/:uuid", projectHandler.GetProject)
		projectAPI.POST("", projectHandler.CreateProject)
		projectAPI.PUT("/:uuid", projectHandler.UpdateProject)
		projectAPI.DELETE("/:uuid", projectHandler.DeleteProject)
		projectAPI.PUT("/:uuid/toggle", projectHandler.ToggleProject)
		projectAPI.POST("/validate", projectHandler.ValidateProjectAPI)
		log.Println("Project routes registered successfully")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "2427"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("MountCore Service starting on %s", addr)
	log.Printf("Visit http://localhost:%s to test", port)

	// 在后台启动服务器，这样我们就可以继续注册其他路由
	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// 等待一小段时间确保服务器启动
	time.Sleep(100 * time.Millisecond)

	log.Println("Now registering skills routes in background...")
	go func() {
		if err := skills.RegisterRoutes(router); err != nil {
			log.Printf("Warning: failed to register skills routes: %v", err)
		} else {
			log.Println("Skills routes registered successfully")
		}
	}()

	// 保持主 goroutine 运行
	select {}
}

func registerMountSQLRoutes(router gin.IRouter) {
	mountSQL := router.Group("/mount/sql")

	mountSQL.GET("/list", func(c *gin.Context) {
		executor := sql.GetInstance()
		if executor == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "SQL executor not initialized",
			})
			return
		}

		entries := executor.ListRegisteredSQL()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Found %d registered SQL statements", len(entries)),
			"data":    entries,
		})
	})

	mountSQL.GET("/get/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		if uuid == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "UUID is required",
			})
			return
		}

		executor := sql.GetInstance()
		if executor == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "SQL executor not initialized",
			})
			return
		}

		entry, err := executor.GetSQLByUUID(uuid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "SQL config found",
			"data":    entry,
		})
	})

	mountSQL.Any("/execute/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		if uuid == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "UUID is required",
			})
			return
		}

		var args []interface{}

		if c.Request.Method == "POST" {
			var reqBody struct {
				Args []interface{} `json:"args"`
			}
			if err := c.ShouldBindJSON(&reqBody); err != nil {
				args = []interface{}{}
			} else {
				args = reqBody.Args
			}
		} else {
			argsStr := c.Query("args")
			if argsStr != "" {
				if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
					args = []interface{}{}
				}
			}
		}

		executor := sql.GetInstance()
		if executor == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "SQL executor not initialized",
			})
			return
		}

		result, err := executor.ExecuteSQLByUUID(context.Background(), uuid, args)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Query executed successfully",
			"data": gin.H{
				"columns": result.Columns,
				"rows":    result.Rows,
				"count":   result.Count,
			},
		})
	})

	log.Println("Mount SQL routes registered successfully")
}
