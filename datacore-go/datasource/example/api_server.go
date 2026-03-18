package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/example/datasource"
)

type AddDataSourceRequest struct {
	ConfigJSON string `json:"config_json"`
}

type QueryRequest struct {
	UUIDDatasource string `json:"uuid_datasource"`
	SQL            string `json:"sql"`
}

func handleAddDataSource(w http.ResponseWriter, r *http.Request) {
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

	var req AddDataSourceRequest
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

func handleQueryWithUUID(w http.ResponseWriter, r *http.Request) {
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

	var req QueryRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %v", err), http.StatusBadRequest)
		return
	}

	if req.UUIDDatasource == "" {
		http.Error(w, "uuid_datasource is required", http.StatusBadRequest)
		return
	}

	if req.SQL == "" {
		http.Error(w, "sql is required", http.StatusBadRequest)
		return
	}

	manager := datasource.GetManager()
	results, err := manager.QueryWithUUID(req.UUIDDatasource, req.SQL)
	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    results,
	})
}

func handleQueryWithUUIDGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	prefix := "/datasource/query/"
	if !strings.HasPrefix(path, prefix) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	uuidDatasource := strings.TrimPrefix(path, prefix)
	if uuidDatasource == "" {
		http.Error(w, "uuid_datasource is required", http.StatusBadRequest)
		return
	}

	sql := r.URL.Query().Get("sql")
	if sql == "" {
		http.Error(w, "sql query parameter is required", http.StatusBadRequest)
		return
	}

	manager := datasource.GetManager()
	results, err := manager.QueryWithUUID(uuidDatasource, sql)
	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    results,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"status":  "healthy",
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/datasource/add", handleAddDataSource)
	mux.HandleFunc("/datasource/query", handleQueryWithUUID)
	mux.HandleFunc("/datasource/query/", handleQueryWithUUIDGet)

	addr := ":8085"
	fmt.Printf("API Server starting on %s\n", addr)
	fmt.Println("Available APIs:")
	fmt.Println("  GET  /health - Health check")
	fmt.Println("  POST /datasource/add - Add a datasource with JSON config")
	fmt.Println("  POST /datasource/query - Query using uuid_datasource and SQL (JSON body)")
	fmt.Println("  GET  /datasource/query/{uuid_datasource}?sql=... - Query using uuid_datasource and SQL (GET)")

	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}
