package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/xmzail/ai-carrier-dev/common/embedding"
	"github.com/xmzail/ai-carrier-dev/common/seekdb"
	"github.com/xmzail/ai-carrier-dev/carriercore/common/db"
)

var PostgresStore *seekdb.PostgresStore
var EmbeddingService *embedding.EmbeddingService

func InitPostgresForHTTPProxy() error {
	dsn := db.GetPostgresURL()
	var err error
	PostgresStore, err = seekdb.NewPostgresStore(dsn)
	if err != nil {
		return err
	}
	log.Println("PostgreSQL connected successfully for HTTP Proxy")
	return nil
}

func ClosePostgresForHTTPProxy() {
	if PostgresStore != nil {
		PostgresStore.Close()
	}
}

type VectorSearchResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Data    []VectorSearchResultItem `json:"data,omitempty"`
}

type VectorSearchResultItem struct {
	ID           string      `json:"id"`
	Project      string      `json:"project"`
	WorkerType   string      `json:"worker_type"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	Version      string      `json:"version"`
	Distance     float64     `json:"distance"`
	InputSchema  interface{} `json:"input_schema,omitempty"`
	OutputSchema interface{} `json:"output_schema,omitempty"`
	Tags         interface{} `json:"tags,omitempty"`
}

func VectorSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	if PostgresStore == nil {
		response := VectorSearchResponse{
			Success: false,
			Message: "PostgreSQL连接未初始化",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	query := r.URL.Query()
	text := query.Get("text")
	if text == "" {
		response := VectorSearchResponse{
			Success: false,
			Message: "text参数不能为空",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	limitStr := query.Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	project := query.Get("project")

	log.Printf("向量搜索请求: text=%s, limit=%d, project=%s", text, limit, project)

	var queryVector []float32
	if EmbeddingService != nil {
		var err error
		queryVector, err = EmbeddingService.GetEmbedding(text)
		if err != nil || len(queryVector) == 0 {
			log.Printf("Failed to generate embedding for query: %v", err)
			response := VectorSearchResponse{
				Success: false,
				Message: fmt.Sprintf("生成嵌入向量失败: %v", err),
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
		log.Printf("Generated embedding for query: %s", text)
	} else {
		log.Printf("EmbeddingService not initialized")
		response := VectorSearchResponse{
			Success: false,
			Message: "嵌入服务未初始化",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()
	results, err := PostgresStore.VectorSearchTemplates(ctx, queryVector, limit, project)
	if err != nil {
		log.Printf("向量搜索失败: %v", err)
		response := VectorSearchResponse{
			Success: false,
			Message: fmt.Sprintf("向量搜索失败: %v", err),
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	var items []VectorSearchResultItem
	for _, r := range results {
		var inputSchema, outputSchema, tags interface{}
		json.Unmarshal(r.InputSchema, &inputSchema)
		json.Unmarshal(r.OutputSchema, &outputSchema)
		json.Unmarshal(r.Tags, &tags)

		items = append(items, VectorSearchResultItem{
			ID:           r.ID,
			Project:      r.Project,
			WorkerType:   r.WorkerType,
			Name:         r.Name,
			Description:  r.Description,
			Type:         r.Type,
			Version:      r.Version,
			Distance:     r.Distance,
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
			Tags:         tags,
		})
	}

	response := VectorSearchResponse{
		Success: true,
		Message: fmt.Sprintf("找到 %d 个结果", len(items)),
		Data:    items,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.Printf("向量搜索完成: 返回 %d 个结果", len(items))
}
