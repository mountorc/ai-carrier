package oss

import (
	"io"
	"net/http"

	"github.com/trae/oss-sdk-go/oss"
)

type Client struct {
	inner *oss.Client
}

var (
	instance *Client
)

func NewClient(configPath string) (*Client, error) {
	inner, err := oss.NewClient(configPath)
	if err != nil {
		return nil, err
	}
	return &Client{inner: inner}, nil
}

func Init(configPath string) error {
	var err error
	instance, err = NewClient(configPath)
	if err != nil {
		return err
	}
	return nil
}

func GetInstance() *Client {
	return instance
}

func (c *Client) ValidateToken(token string) *TokenInfo {
	info := c.inner.ValidateToken(token)
	if info == nil {
		return nil
	}
	return &TokenInfo{
		Token:        info.Token,
		UuidAutoAuth: info.UuidAutoAuth,
		BasePath:     info.BasePath,
		CreatedAt:    info.CreatedAt,
		ExpiresAt:    info.ExpiresAt,
	}
}

func (c *Client) GenerateToken(uuidAutoAuth, basePath string) string {
	return c.inner.GenerateToken(uuidAutoAuth, basePath)
}

func (c *Client) UploadByToken(token, fileName string, content []byte) (string, error) {
	return c.inner.UploadByToken(token, fileName, content)
}

func (c *Client) Upload(uuidAutoAuth, filePath string, content []byte) (string, error) {
	return c.inner.Upload(uuidAutoAuth, filePath, content)
}

func (c *Client) CreateFolder(uuidAutoAuth, folderPath string) error {
	return c.inner.CreateFolder(uuidAutoAuth, folderPath)
}

func (c *Client) GetTokenByPath(basePath string) ([]string, error) {
	return c.inner.GetTokenByPath(basePath)
}

func (c *Client) GetAllTokens() map[string]*TokenInfo {
	innerTokens := c.inner.GetAllTokens()
	tokens := make(map[string]*TokenInfo)
	for token, info := range innerTokens {
		tokens[token] = &TokenInfo{
			Token:        info.Token,
			UuidAutoAuth: info.UuidAutoAuth,
			BasePath:     info.BasePath,
			CreatedAt:    info.CreatedAt,
			ExpiresAt:    info.ExpiresAt,
		}
	}
	return tokens
}

func (c *Client) AddToken(tokenInfo *TokenInfo) {
	c.inner.AddToken(&oss.TokenInfo{
		Token:        tokenInfo.Token,
		UuidAutoAuth: tokenInfo.UuidAutoAuth,
		BasePath:     tokenInfo.BasePath,
		CreatedAt:    tokenInfo.CreatedAt,
		ExpiresAt:    tokenInfo.ExpiresAt,
	})
}

func (c *Client) RemoveToken(token string) {
	c.inner.RemoveToken(token)
}

func (c *Client) ListFiles(uuidAutoAuth, prefix string) ([]map[string]interface{}, error) {
	return c.inner.ListFiles(uuidAutoAuth, prefix)
}

func (c *Client) UploadByTokenWithReader(token, fileName string, reader io.Reader) (string, error) {
	return c.inner.UploadByTokenWithReader(token, fileName, reader)
}

func (c *Client) UploadFormData(token string, r *http.Request) (string, error) {
	return c.inner.UploadFormData(token, r)
}
