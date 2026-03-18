package autodatasource

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type AutoDataSourceClient struct {
	baseURL string
	client  *resty.Client
}

type Response struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	DataList interface{} `json:"dataList,omitempty"`
	Message  string      `json:"message,omitempty"`
	Total    int         `json:"total,omitempty"`
}

type TransformSqlRequest struct {
	Sql        string `json:"sql"`
	SourceType string `json:"sourceType"`
	TargetType string `json:"targetType"`
}

type TransformSqlResponse struct {
	TransformedSql string `json:"transformedSql"`
}

type InsertVectorsRequest struct {
	Vectors  [][]float32        `json:"vectors"`
	Metadata []map[string]interface{} `json:"metadata"`
}

type SearchVectorsRequest struct {
	QueryVector []float32 `json:"queryVector"`
	TopK        int       `json:"topK"`
	Filter      string    `json:"filter,omitempty"`
}

func NewClient(baseURL string) *AutoDataSourceClient {
	return NewClientWithTimeout(baseURL, 30)
}

func NewClientWithTimeout(baseURL string, timeout int) *AutoDataSourceClient {
	client := resty.New()
	client.SetTimeout(timeout)
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	return &AutoDataSourceClient{
		baseURL: baseURL,
		client:  client,
	}
}

func (c *AutoDataSourceClient) GetDataSources() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/data-sources/external/list", c.baseURL)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *AutoDataSourceClient) GetLocalDataSources() (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources", c.baseURL)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) AddDataSource(dataSourceId string, properties map[string]interface{}) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources/add/%s", c.baseURL, dataSourceId)

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(properties).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) RemoveDataSource(dataSourceId string) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources/remove/%s", c.baseURL, dataSourceId)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) CheckDataSourceExists(dataSourceId string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/data-sources/%s/exists", c.baseURL, dataSourceId)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *AutoDataSourceClient) TestConnection(properties map[string]interface{}) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources/test-connection", c.baseURL)

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(properties).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) TransformSql(sql, sourceType, targetType string) (*Response, error) {
	url := fmt.Sprintf("%s/api/sql/transform", c.baseURL)

	req := TransformSqlRequest{
		Sql:        sql,
		SourceType: sourceType,
		TargetType: targetType,
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) GetDataSets() (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sets/list", c.baseURL)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) GetDataByUuid(uuidAutoData string) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sets/data/%s", c.baseURL, uuidAutoData)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) PreviewDataSetsByDataSourceId(dataSourceId string) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sets/preview/%s", c.baseURL, dataSourceId)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) GetOssFiles(prefix string) (*Response, error) {
	url := fmt.Sprintf("%s/api/oss/files", c.baseURL)
	if prefix != "" {
		url = fmt.Sprintf("%s?prefix=%s", url, prefix)
	}

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) GetPublicDocsList() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/docs-public/list", c.baseURL)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *AutoDataSourceClient) GetPublicDocContent(fileName string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/docs-public/docs?fileName=%s", c.baseURL, fileName)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *AutoDataSourceClient) GetExtractRecords() (*Response, error) {
	url := fmt.Sprintf("%s/api/extract-records/list", c.baseURL)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) ExecuteSqlQuery(dataSourceId, sql string) (*Response, error) {
	return c.ExecuteSqlQueryWithType(dataSourceId, sql, "")
}

func (c *AutoDataSourceClient) ExecuteSqlQueryWithType(dataSourceId, sql, sqlType string) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources/%s/query/sql", c.baseURL, dataSourceId)

	reqBody := map[string]string{
		"sql": sql,
	}
	if sqlType != "" {
		reqBody["sqlType"] = sqlType
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) ExecuteSqlUpdate(dataSourceId, sql string) (*Response, error) {
	return c.ExecuteSqlUpdateWithType(dataSourceId, sql, "")
}

func (c *AutoDataSourceClient) ExecuteSqlUpdateWithType(dataSourceId, sql, sqlType string) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources/%s/query/update", c.baseURL, dataSourceId)

	reqBody := map[string]string{
		"sql": sql,
	}
	if sqlType != "" {
		reqBody["sqlType"] = sqlType
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) GetTableFields(dataSourceId, tableSchema, tableName string) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources/%s/query/table-fields", c.baseURL, dataSourceId)

	reqBody := map[string]string{
		"tableSchema": tableSchema,
		"tableName":   tableName,
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) GetTableList(dataSourceId, tableSchema string) (*Response, error) {
	url := fmt.Sprintf("%s/api/data-sources/%s/query/tables", c.baseURL, dataSourceId)

	reqBody := map[string]string{}
	if tableSchema != "" {
		reqBody["tableSchema"] = tableSchema
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) TestVectorDatabaseConnection(properties map[string]interface{}) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/test-connection", c.baseURL)

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(properties).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) ConnectToVectorDatabase(properties map[string]interface{}) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/connect", c.baseURL)

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(properties).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) DisconnectFromVectorDatabase(connectionId, databaseType string) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/disconnect/%s?databaseType=%s", c.baseURL, connectionId, databaseType)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) CreateVectorCollection(connectionId, databaseType, collectionName string, dimension int, indexType string) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/collections?connectionId=%s&databaseType=%s&collectionName=%s&dimension=%d",
		c.baseURL, connectionId, databaseType, collectionName, dimension)
	if indexType != "" {
		url = fmt.Sprintf("%s&indexType=%s", url, indexType)
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) DropVectorCollection(connectionId, databaseType, collectionName string) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/collections/%s?connectionId=%s&databaseType=%s",
		c.baseURL, collectionName, connectionId, databaseType)

	resp, err := c.client.R().Delete(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) ListVectorCollections(connectionId, databaseType string) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/collections?connectionId=%s&databaseType=%s",
		c.baseURL, connectionId, databaseType)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) GetVectorCollectionInfo(connectionId, databaseType, collectionName string) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/collections/%s?connectionId=%s&databaseType=%s",
		c.baseURL, collectionName, connectionId, databaseType)

	resp, err := c.client.R().Get(url)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) InsertVectors(connectionId, databaseType, collectionName string, vectors [][]float32, metadata []map[string]interface{}) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/collections/%s/vectors?connectionId=%s&databaseType=%s",
		c.baseURL, collectionName, connectionId, databaseType)

	reqBody := InsertVectorsRequest{
		Vectors:  vectors,
		Metadata: metadata,
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *AutoDataSourceClient) SearchVectors(connectionId, databaseType, collectionName string, queryVector []float32, topK int, filter string) (*Response, error) {
	url := fmt.Sprintf("%s/api/vector-databases/collections/%s/search?connectionId=%s&databaseType=%s",
		c.baseURL, collectionName, connectionId, databaseType)

	reqBody := SearchVectorsRequest{
		QueryVector: queryVector,
		TopK:        topK,
		Filter:      filter,
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post(url)

	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, err
	}

	return &response, nil
}
