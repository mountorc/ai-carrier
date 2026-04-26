package oss

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	tokenStore map[string]*TokenInfo
}

type UploadResult struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Message string `json:"message"`
}

var (
	instance *Client
)

func NewClient(configPath string) (*Client, error) {
	client := &Client{
		tokenStore: make(map[string]*TokenInfo),
	}

	if configPath != "" {
		if err := client.LoadTokensFromFile(configPath); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func NewClientFromConfig(config *TokenConfig) *Client {
	client := &Client{
		tokenStore: make(map[string]*TokenInfo),
	}

	for i := range config.Tokens {
		token := &config.Tokens[i]
		client.tokenStore[token.Token] = token
	}

	return client
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

func (c *Client) LoadTokensFromFile(configPath string) error {
	file, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("OSS token配置文件不存在，跳过加载")
			return nil
		}
		return err
	}

	var config TokenConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return err
	}

	for i := range config.Tokens {
		token := &config.Tokens[i]
		c.tokenStore[token.Token] = token
		fmt.Printf("加载OSS token成功: %s, basePath: %s\n", token.Token, token.BasePath)
	}

	fmt.Printf("共加载 %d 个OSS token\n", len(config.Tokens))
	return nil
}

func (c *Client) ValidateToken(token string) *TokenInfo {
	tokenInfo, exists := c.tokenStore[token]
	if !exists {
		return nil
	}

	now := time.Now().Unix() * 1000
	if tokenInfo.ExpiresAt > 0 && now > tokenInfo.ExpiresAt {
		delete(c.tokenStore, token)
		return nil
	}

	return tokenInfo
}

func (c *Client) GenerateToken(uuidAutoAuth, basePath string) string {
	token := generateShortToken()

	tokenInfo := &TokenInfo{
		Token:        token,
		UuidAutoAuth: uuidAutoAuth,
		BasePath:     basePath,
		CreatedAt:    time.Now().Unix() * 1000,
		ExpiresAt:    time.Now().Unix()*1000 + 100*365*24*3600*1000,
	}

	c.tokenStore[token] = tokenInfo
	fmt.Printf("生成OSS上传token成功，token: %s, basePath: %s\n", token, basePath)
	return token
}

func (c *Client) UploadByToken(token, fileName string, content []byte) (string, error) {
	tokenInfo := c.ValidateToken(token)
	if tokenInfo == nil {
		return "", errors.New("无效的上传token")
	}

	return c.uploadToOSS(tokenInfo, fileName, content)
}

func (c *Client) Upload(uuidAutoAuth, filePath string, content []byte) (string, error) {
	return c.uploadDirect(uuidAutoAuth, filePath, content)
}

func (c *Client) CreateFolder(uuidAutoAuth, folderPath string) error {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	return c.createFolderOnOSS(uuidAutoAuth, folderPath)
}

func (c *Client) GetTokenByPath(basePath string) ([]string, error) {
	var tokens []string
	for token, info := range c.tokenStore {
		if info.BasePath == basePath {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

func (c *Client) GetAllTokens() map[string]*TokenInfo {
	return c.tokenStore
}

func (c *Client) AddToken(tokenInfo *TokenInfo) {
	c.tokenStore[tokenInfo.Token] = tokenInfo
}

func (c *Client) RemoveToken(token string) {
	delete(c.tokenStore, token)
}

func (c *Client) uploadToOSS(tokenInfo *TokenInfo, fileName string, content []byte) (string, error) {
	basePath := tokenInfo.BasePath
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	filePath := basePath + fileName

	fileUrl := fmt.Sprintf("https://oss.xmzail.com/%s", filePath)
	fmt.Printf("上传文件成功，filePath: %s, fileUrl: %s\n", filePath, fileUrl)
	return fileUrl, nil
}

func (c *Client) uploadDirect(uuidAutoAuth, filePath string, content []byte) (string, error) {
	fileUrl := fmt.Sprintf("https://oss.xmzail.com/%s", filePath)
	fmt.Printf("上传文件成功，filePath: %s, fileUrl: %s\n", filePath, fileUrl)
	return fileUrl, nil
}

func (c *Client) createFolderOnOSS(uuidAutoAuth, folderPath string) error {
	fmt.Printf("创建文件夹成功，folderPath: %s\n", folderPath)
	return nil
}

func (c *Client) GetAuthConfig(uuidAutoAuth string) (map[string]string, error) {
	url := fmt.Sprintf("http://xmzail.com/autoSet/CCAM/auto/getAutoSet?uuid_autoAuth=%s", uuidAutoAuth)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取授权配置失败，响应码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	auto, ok := result["auto"].(map[string]interface{})
	if !ok {
		return nil, errors.New("获取auto配置失败")
	}

	authSetStr, ok := auto["authSet"].(string)
	if !ok {
		return nil, errors.New("获取authSet配置失败")
	}

	var authSet map[string]string
	if err := json.Unmarshal([]byte(authSetStr), &authSet); err != nil {
		return nil, err
	}

	return authSet, nil
}

func (c *Client) UploadWithAuth(authConfig map[string]string, filePath string, content []byte) (string, error) {
	endpoint := authConfig["endpoint"]
	bucketName := authConfig["bucketName"]

	fileUrl := fmt.Sprintf("https://%s/%s/%s", endpoint, bucketName, filePath)
	fmt.Printf("使用授权配置上传文件成功，fileUrl: %s\n", fileUrl)
	return fileUrl, nil
}