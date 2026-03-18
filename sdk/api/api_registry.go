package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouteConfig 路由配置结构
type RouteConfig struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Handler    string `json:"handler"`
	EnableCORS bool   `json:"enableCORS,omitempty"`
	IsFunc     bool   `json:"isFunc,omitempty"`
}

// APIConfig API配置结构
type APIConfig struct {
	Description string        `json:"description,omitempty"`
	Routes      []RouteConfig `json:"routes"`
}

// RegisterRoutesFromJSON 从JSON文件注册路由
// 参数:
// - router: Gin路由器实例
// - configFile: JSON配置文件路径
// - handlerMap: 处理器函数映射
// 返回:
// - error: 注册过程中的错误
func RegisterRoutesFromJSON(router *gin.Engine, configFile string, handlerMap map[string]gin.HandlerFunc) error {
	return RegisterRoutesFromJSONData(router, []byte{}, configFile, handlerMap)
}

// RegisterRoutesFromJSONData 从JSON数据注册路由
// 参数:
// - router: Gin路由器实例
// - embeddedData: 嵌入式JSON数据
// - configFile: JSON配置文件路径（用于日志和错误信息）
// - handlerMap: 处理器函数映射
// 返回:
// - error: 注册过程中的错误
func RegisterRoutesFromJSONData(router *gin.Engine, embeddedData []byte, configFile string, handlerMap map[string]gin.HandlerFunc) error {
	var data []byte
	var err error

	// 优先使用嵌入式数据
	if len(embeddedData) > 0 {
		data = embeddedData
	} else {
		// 尝试从可执行文件目录读取
		exePath, exeErr := os.Executable()
		if exeErr == nil {
			exeDir := filepath.Dir(exePath)
			fullPath := filepath.Join(exeDir, configFile)
			data, err = os.ReadFile(fullPath)
			if err != nil {
				// 尝试从当前目录读取
				data, err = os.ReadFile(configFile)
			}
		} else {
			// 直接从当前目录读取
			data, err = os.ReadFile(configFile)
		}
		if err != nil {
			return err
		}
	}

	// 解析JSON配置
	var apiConfig APIConfig
	if err := json.Unmarshal(data, &apiConfig); err != nil {
		return err
	}

	// 注册路由
	for _, route := range apiConfig.Routes {
		handlerFunc, exists := handlerMap[route.Handler]
		if !exists {
			log.Printf("Warning: Handler %s not found for path %s", route.Handler, route.Path)
			continue
		}

		// 根据HTTP方法注册路由
		switch route.Method {
		case "GET":
			router.GET(route.Path, handlerFunc)
		case "POST":
			router.POST(route.Path, handlerFunc)
		case "PUT":
			router.PUT(route.Path, handlerFunc)
		case "DELETE":
			router.DELETE(route.Path, handlerFunc)
		default:
			log.Printf("Warning: Unsupported method %s for path %s", route.Method, route.Path)
			continue
		}

		log.Printf("Registered route: %s %s -> %s", route.Method, route.Path, route.Handler)
	}

	return nil
}

// SetupCORS 配置CORS中间件
// 参数:
// - router: Gin路由器实例
func SetupCORS(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

// ValidateAPIConfig 验证API配置
// 参数:
// - config: API配置
// 返回:
// - error: 验证错误
func ValidateAPIConfig(config *APIConfig) error {
	if config == nil {
		return nil
	}

	for i, route := range config.Routes {
		if route.Method == "" {
			return fmt.Errorf("route %d: method is required", i)
		}
		if route.Path == "" {
			return fmt.Errorf("route %d: path is required", i)
		}
		if route.Handler == "" {
			return fmt.Errorf("route %d: handler is required", i)
		}

		// 验证HTTP方法
		validMethods := map[string]bool{
			"GET":    true,
			"POST":   true,
			"PUT":    true,
			"DELETE": true,
		}
		if !validMethods[route.Method] {
			return fmt.Errorf("route %d: invalid method %s", i, route.Method)
		}
	}

	return nil
}

// LoadAPIConfig 加载API配置
// 参数:
// - configFile: 配置文件路径
// 返回:
// - *APIConfig: 加载的配置
// - error: 加载错误
func LoadAPIConfig(configFile string) (*APIConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config APIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if err := ValidateAPIConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}