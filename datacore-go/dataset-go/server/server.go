package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/example/dataset-go/database"
	"github.com/example/dataset-go/scheduler"
	"github.com/example/datasource"
)

type Server struct {
	scheduler *scheduler.Scheduler
	database  *database.Database
	mu        sync.RWMutex
}

func NewServer(s *scheduler.Scheduler, db *database.Database) *Server {
	return &Server{
		scheduler: s,
		database:  db,
	}
}

func (s *Server) handleGetAutoSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuidDatasource := r.URL.Query().Get("uuid_datasource")
	uuidWorkflow := r.URL.Query().Get("uuid_workflow")
	uuidDataset := r.URL.Query().Get("uuid_dataset")
	uuid := r.URL.Query().Get("uuid")

	var targetUUID string
	var tableName string

	if uuidDatasource != "" {
		if uuidWorkflow != "" {
			targetUUID = uuidWorkflow
			tableName = "workflow"
		} else if uuidDataset != "" {
			targetUUID = uuidDataset
			tableName = "dataset"
		} else if uuid != "" {
			if strings.HasPrefix(uuid, "uuid_") {
				parts := strings.SplitN(uuid, "_", 3)
				if len(parts) >= 2 {
					tableName = parts[1]
					if len(parts) >= 3 {
						targetUUID = parts[2]
					}
				}
			}
			if tableName == "" || targetUUID == "" {
				http.Error(w, "invalid uuid format, should be uuid_tablename_xxx", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "uuid_workflow, uuid_dataset, or uuid parameter is required", http.StatusBadRequest)
			return
		}

		result, err := database.GetAutoSetFromDataSource(uuidDatasource, targetUUID, tableName)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get data: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    result,
		})
		return
	}

	if uuid == "" {
		http.Error(w, "uuid parameter is required", http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(uuid, "uuid_") {
		parts := strings.SplitN(uuid, "_", 3)
		if len(parts) >= 2 {
			tableName = parts[1]
		}
	}

	if tableName == "" {
		http.Error(w, "invalid uuid format, should be uuid_tablename_xxx", http.StatusBadRequest)
		return
	}

	result, err := s.database.GetAutoSet(uuid, tableName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

func (s *Server) handleGetSQL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "uuid parameter is required", http.StatusBadRequest)
		return
	}

	sqlStr, exists := s.scheduler.GetSQL(uuid)
	if !exists {
		http.Error(w, fmt.Sprintf("SQL not found for uuid: %s", uuid), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"uuid":    uuid,
		"sql":     sqlStr,
	})
}

func (s *Server) handleGetSQLItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "uuid parameter is required", http.StatusBadRequest)
		return
	}

	item, exists := s.scheduler.GetSQLItem(uuid)
	if !exists {
		http.Error(w, fmt.Sprintf("SQL not found for uuid: %s", uuid), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"item":    item,
	})
}

func (s *Server) handleListUUIDs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuids := s.scheduler.ListAllUUIDs()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(uuids),
		"uuids":   uuids,
	})
}

func (s *Server) handleGetFromDataSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uuidDatasource := r.URL.Query().Get("uuid_datasource")
	if uuidDatasource == "" {
		http.Error(w, "uuid_datasource parameter is required", http.StatusBadRequest)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		http.Error(w, "uuid parameter is required", http.StatusBadRequest)
		return
	}

	var tableName string
	if strings.HasPrefix(uuid, "uuid_") {
		parts := strings.SplitN(uuid, "_", 3)
		if len(parts) >= 2 {
			tableName = parts[1]
		}
	}

	if tableName == "" {
		http.Error(w, "invalid uuid format, should be uuid_tablename_xxx", http.StatusBadRequest)
		return
	}

	result, err := database.GetAutoSetFromDataSource(uuidDatasource, uuid, tableName)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

func (s *Server) handleAddDataSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		ConfigJSON string `json:"config_json"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %v", err), http.StatusBadRequest)
		return
	}

	manager := datasource.GetManager()
	if err := manager.AddDataSource(req.ConfigJSON); err != nil {
		http.Error(w, fmt.Sprintf("failed to add datasource: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "datasource added successfully",
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"status":  "healthy",
		"count":   s.scheduler.Count(),
	})
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/datasource/add", s.handleAddDataSource)
	mux.HandleFunc("/sql", s.handleGetSQL)
	mux.HandleFunc("/sql/item", s.handleGetSQLItem)
	mux.HandleFunc("/sql/list", s.handleListUUIDs)
	mux.HandleFunc("/getAutoSet", s.handleGetAutoSet)
	mux.HandleFunc("/getFromDataSource", s.handleGetFromDataSource)

	fmt.Printf("Server starting on %s\n", addr)
	fmt.Println("Available APIs:")
	fmt.Println("  GET  /health - Health check")
	fmt.Println("  POST /datasource/add - Add datasource")
	fmt.Println("  GET  /sql - Get SQL by uuid")
	fmt.Println("  GET  /sql/item - Get SQL item by uuid")
	fmt.Println("  GET  /sql/list - List all SQL uuids")
	fmt.Println("  GET  /getAutoSet - Get data from default or specified database")
	fmt.Println("  GET  /getFromDataSource - Get data from specified datasource")

	return http.ListenAndServe(addr, mux)
}
