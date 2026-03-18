// services/mountcore/main.go

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trae/autoFlow/common/embedding"
	"github.com/trae/autoFlow/carriercore/ability"
	"github.com/trae/autoFlow/carriercore/apiregistry"
	"github.com/trae/autoFlow/carriercore/common/sql"
	"github.com/trae/autoFlow/carriercore/scheduler"
	workflowHandlers "github.com/trae/autoFlow/carriercore/workflow"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	exePath, exeErr := os.Executable()
	if exeErr == nil {
		exeDir := filepath.Dir(exePath)
		if chdirErr := os.Chdir(exeDir); chdirErr != nil {
			log.Printf("Warning: failed to change working directory: %v", chdirErr)
		} else {
			log.Printf("Working directory set to: %s", exeDir)
		}
	}

	log.Println("Initializing MountCore Service...")

	// workflow.SetFlowExecutor(&workflow.HTTPFlowExecutor{})

	log.Println("Initializing Embedding service...")
	embeddingConfig := embedding.Config{
		ProviderType: "aliyun",
	}
	embeddingService, err := embedding.NewEmbeddingServiceFromConfig(embeddingConfig)
	if err != nil {
		log.Printf("Warning: failed to initialize Embedding service: %v", err)
	} else {
		log.Println("Embedding service initialized successfully (using Aliyun Qwen model)")
		// 设置ability包中的embeddingService
		ability.SetEmbeddingService(embeddingService)
	}

	log.Println("Initializing Workflow module...")
	log.Println("Workflow module initialized successfully")

	log.Println("Initializing Ability module...")
	if err := ability.Init(); err != nil {
		log.Printf("Warning: failed to initialize Ability: %v", err)
	} else {
		log.Println("Ability module initialized successfully")
		defer ability.Close()
	}

	log.Println("Initializing SQL config...")
	if err := scheduler.LoadSQLConfig(); err != nil {
		log.Printf("Warning: failed to initialize SQL config: %v", err)
	} else {
		log.Printf("SQL config initialized successfully, loaded %d SQLs", len(sql.GetAllSQLs()))
	}

	log.Println("Initializing Scheduler module...")
	if err := scheduler.Init(); err != nil {
		log.Printf("Warning: failed to initialize Scheduler: %v", err)
	} else {
		log.Println("Scheduler module initialized successfully")
		defer scheduler.Close()
	}

	router := gin.Default()
	apiregistry.SetupCORS(router)

	log.Println("Registering routes...")
	workflowHandlers.RegisterRoutes(router)
	ability.RegisterRoutes(router)
	scheduler.RegisterRoutes(router)

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
