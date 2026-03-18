package postgres

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AliyunEmbedding struct {
	APIURL    string
	APIKey    string
	Model     string
	HTTPClient *http.Client
}

type AliyunEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type AliyunEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

const (
	DefaultAliyunAPIURL = "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings"
	DefaultAliyunModel  = "text-embedding-v4"
	DefaultAliyunAPIKey = "sk-403ca84daa9740df82ce0a1737ceccdf"
)

func NewAliyunEmbedding(apiURL, apiKey, model string) *AliyunEmbedding {
	if apiURL == "" {
		apiURL = DefaultAliyunAPIURL
	}
	if apiKey == "" {
		apiKey = DefaultAliyunAPIKey
	}
	if model == "" {
		model = DefaultAliyunModel
	}

	return &AliyunEmbedding{
		APIURL: apiURL,
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *AliyunEmbedding) EmbedText(text string) ([]float64, error) {
	reqBody := AliyunEmbeddingRequest{
		Model: a.Model,
		Input: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", a.APIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.APIKey)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embResp AliyunEmbeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embResp.Data[0].Embedding, nil
}
