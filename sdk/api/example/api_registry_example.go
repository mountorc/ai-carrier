package main

import (
	"github.com/gin-gonic/gin"
	"github.com/trae/autoFlow/carriercore/sdk/api"
)

func main() {
	// 创建Gin路由器
	router := gin.Default()

	// 配置CORS
	api.SetupCORS(router)

	// 定义处理器映射
	handlerMap := map[string]gin.HandlerFunc{
		"getUsersHandler": func(c *gin.Context) {
			c.JSON(200, gin.H{
				"users": []string{"user1", "user2", "user3"},
				"total": 3,
			})
		},
		"createUserHandler": func(c *gin.Context) {
			var user struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			c.JSON(201, gin.H{
				"message": "User created successfully",
				"user":    user,
			})
		},
		"getUserByIdHandler": func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(200, gin.H{
				"id":    id,
				"name":  "User " + id,
				"email": "user" + id + "@example.com",
			})
		},
		"updateUserHandler": func(c *gin.Context) {
			id := c.Param("id")
			var user struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{
				"message": "User updated successfully",
				"id":      id,
				"user":    user,
			})
		},
		"deleteUserHandler": func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(200, gin.H{
				"message": "User deleted successfully",
				"id":      id,
			})
		},
	}

	// 从JSON文件注册路由
	configFile := "api_users.json"
	if err := api.RegisterRoutesFromJSON(router, configFile, handlerMap); err != nil {
		panic("Failed to register routes: " + err.Error())
	}

	// 启动服务器
	port := ":8080"
	print("Server running on http://localhost" + port + "\n")
	if err := router.Run(port); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}