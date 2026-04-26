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
	"github.com/trae/autoFlow/carriercore/ability"
	"github.com/trae/autoFlow/carriercore/apiregistry"
	"github.com/trae/autoFlow/carriercore/auth"
	"github.com/trae/autoFlow/carriercore/mountcore/sql"
	"github.com/trae/autoFlow/carriercore/oss"
	"github.com/trae/autoFlow/carriercore/scheduler"
	"github.com/trae/autoFlow/carriercore/skill"
	workflowHandlers "github.com/trae/autoFlow/carriercore/workflow"
	"github.com/trae/autoFlow/common/embedding"
	"github.com/trae/autoFlow/mounts/skills"
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

	log.Println("Initializing Ability module...")
	if err := ability.Init(); err != nil {
		log.Printf("Warning: failed to initialize Ability: %v", err)
	} else {
		log.Println("Ability module initialized successfully")
		defer ability.Close()

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
	if err := oss.Init("./mount_config.json"); err != nil {
		log.Printf("Warning: failed to initialize OSS: %v", err)
	} else {
		log.Println("OSS module initialized successfully")
	}

	router := gin.Default()
	apiregistry.SetupCORS(router)

	log.Println("Registering routes...")
	workflowHandlers.RegisterRoutes(router)
	ability.RegisterRoutes(router)
	scheduler.RegisterRoutes(router)
	skill.RegisterRoutes(router)
	auth.RegisterRoutes(router)
	oss.RegisterRoutes(router)
	if err := skills.RegisterRoutes(router); err != nil {
		log.Printf("Warning: failed to register skills routes: %v", err)
	} else {
		log.Println("Skills routes registered successfully")
	}

	registerMountSQLRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "2427"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("MountCore Service starting on %s", addr)
	log.Printf("Visit http://localhost:%s to test", port)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}

func registerMountSQLRoutes(router *gin.Engine) {
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
