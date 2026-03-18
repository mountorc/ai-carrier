package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

type Client struct {
	config          *Config
	db              *sql.DB
	useAliyun       bool
	aliyunEmbedding *AliyunEmbedding
}

func NewClient(config *Config, useAliyun bool, aliyunAPIKey, aliyunModel string) *Client {
	client := &Client{
		config:    config,
		useAliyun: useAliyun,
	}
	if useAliyun {
		client.aliyunEmbedding = NewAliyunEmbedding("", aliyunAPIKey, aliyunModel)
	}
	return client
}

func NewClientFromURL(url string) (*Client, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	client := &Client{
		db: db,
	}

	fmt.Printf("Connected to PostgreSQL at %s\n", url)
	return client, nil
}

func (c *Client) Connect() error {
	var err error
	c.db, err = sql.Open("postgres", c.config.ConnectionString())
	if err != nil {
		return err
	}

	err = c.db.Ping()
	if err != nil {
		return err
	}

	fmt.Printf("Connected to PostgreSQL at %s:%d/%s\n", c.config.Host, c.config.Port, c.config.Database)
	return nil
}

func (c *Client) Disconnect() {
	if c.db != nil {
		c.db.Close()
		fmt.Println("Disconnected from PostgreSQL")
	}
}

func (c *Client) CreateCollection(collectionName string, dimension int, indexType string) error {
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			embedding vector(%d),
			metadata jsonb
		)
	`, collectionName, dimension)

	_, err := c.db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	if indexType != "" {
		var createIndexSQL string
		switch strings.ToUpper(indexType) {
		case "IVFFLAT":
			createIndexSQL = fmt.Sprintf(`
				CREATE INDEX IF NOT EXISTS %s_embedding_idx
				ON %s USING ivfflat (embedding vector_l2_ops)
				WITH (lists = 128)
			`, collectionName, collectionName)
		case "HNSW":
			createIndexSQL = fmt.Sprintf(`
				CREATE INDEX IF NOT EXISTS %s_embedding_idx
				ON %s USING hnsw (embedding vector_l2_ops)
				WITH (m = 16, ef_construction = 64)
			`, collectionName, collectionName)
		default:
			return fmt.Errorf("unsupported index type: %s", indexType)
		}

		_, err = c.db.Exec(createIndexSQL)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Collection %s created with dimension %d\n", collectionName, dimension)
	return nil
}

func (c *Client) DropCollection(collectionName string) error {
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", collectionName)
	_, err := c.db.Exec(dropTableSQL)
	if err != nil {
		return err
	}

	fmt.Printf("Collection %s dropped\n", collectionName)
	return nil
}

func (c *Client) ListCollections() ([]string, error) {
	rows, err := c.db.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		collections = append(collections, tableName)
	}

	return collections, nil
}

func (c *Client) InsertVectors(collectionName string, vectors [][]float64, metadata []map[string]interface{}) error {
	if metadata == nil {
		metadata = make([]map[string]interface{}, len(vectors))
		for i := range metadata {
			metadata[i] = make(map[string]interface{})
		}
	}

	for i, vector := range vectors {
		vectorStr := formatVector(vector)
		metadataJSON, err := json.Marshal(metadata[i])
		if err != nil {
			return err
		}

		insertSQL := fmt.Sprintf("INSERT INTO %s (embedding, metadata) VALUES ($1::vector, $2::jsonb)", collectionName)
		_, err = c.db.Exec(insertSQL, vectorStr, string(metadataJSON))
		if err != nil {
			return err
		}
	}

	fmt.Printf("Inserted %d vectors into %s\n", len(vectors), collectionName)
	return nil
}

type SearchResult struct {
	ID       int                    `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (c *Client) SearchVectors(collectionName string, queryVector []float64, topK int, filterCondition string) ([]SearchResult, error) {
	vectorStr := formatVector(queryVector)

	var sqlStr string
	if filterCondition != "" {
		sqlStr = fmt.Sprintf(`
			SELECT id, embedding <-> $1::vector as distance, metadata
			FROM %s
			WHERE %s
			ORDER BY distance
			LIMIT $2
		`, collectionName, filterCondition)
	} else {
		sqlStr = fmt.Sprintf(`
			SELECT id, embedding <-> $1::vector as distance, metadata
			FROM %s
			ORDER BY distance
			LIMIT $2
		`, collectionName)
	}

	rows, err := c.db.Query(sqlStr, vectorStr, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id int
		var distance float64
		var metadataJSON string

		if err := rows.Scan(&id, &distance, &metadataJSON); err != nil {
			return nil, err
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, err
		}

		results = append(results, SearchResult{
			ID:       id,
			Score:    distance,
			Metadata: metadata,
		})
	}

	return results, nil
}

func (c *Client) ExecuteSQL(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := c.db.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

func (c *Client) Query(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	return c.ExecuteSQL(sqlStr, args...)
}

func (c *Client) QueryJSON(sqlStr string, args ...interface{}) (string, error) {
	results, err := c.ExecuteSQL(sqlStr, args...)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (c *Client) GetTableFields(tableSchema, tableName string) ([]map[string]interface{}, error) {
	sqlStr := `
		SELECT column_name, data_type, character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`
	return c.ExecuteSQL(sqlStr, tableSchema, tableName)
}

func (c *Client) GetTableList(tableSchema string) ([]string, error) {
	if tableSchema == "" {
		tableSchema = "public"
	}

	sqlStr := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		ORDER BY table_name
	`
	results, err := c.ExecuteSQL(sqlStr, tableSchema)
	if err != nil {
		return nil, err
	}

	var tables []string
	for _, row := range results {
		if tableName, ok := row["table_name"].(string); ok {
			tables = append(tables, tableName)
		}
	}
	return tables, nil
}

func (c *Client) BeginTransaction() error {
	_, err := c.db.Exec("BEGIN")
	return err
}

func (c *Client) Commit() error {
	_, err := c.db.Exec("COMMIT")
	return err
}

func (c *Client) Rollback() error {
	_, err := c.db.Exec("ROLLBACK")
	return err
}

func (c *Client) GetEmbedding(text string) ([]float64, error) {
	if !c.useAliyun || c.aliyunEmbedding == nil {
		return nil, fmt.Errorf("aliyun embedding is not enabled")
	}
	return c.aliyunEmbedding.EmbedText(text)
}

func (c *Client) InsertTextWithEmbedding(collectionName string, texts []string, metadata []map[string]interface{}) error {
	if metadata == nil {
		metadata = make([]map[string]interface{}, len(texts))
		for i := range metadata {
			metadata[i] = make(map[string]interface{})
		}
	}

	var vectors [][]float64
	for _, text := range texts {
		vector, err := c.GetEmbedding(text)
		if err != nil {
			return err
		}
		vectors = append(vectors, vector)
	}

	return c.InsertVectors(collectionName, vectors, metadata)
}

func (c *Client) SearchByText(collectionName string, queryText string, topK int, filterCondition string) ([]SearchResult, error) {
	queryVector, err := c.GetEmbedding(queryText)
	if err != nil {
		return nil, err
	}
	return c.SearchVectors(collectionName, queryVector, topK, filterCondition)
}

func formatVector(vector []float64) string {
	var parts []string
	for _, v := range vector {
		parts = append(parts, fmt.Sprintf("%f", v))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
