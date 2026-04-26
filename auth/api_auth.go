package auth

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init() error {
	log.Println("Initializing Auth module...")
	if err := initDB(); err != nil {
		return err
	}
	log.Println("Auth module initialized successfully")
	return nil
}

func Close() {
	// Cleanup if needed
}

func RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/auth")

	auth.POST("/register", registerHandler)
	auth.POST("/login", loginHandler)
	auth.POST("/logout", logoutHandler)
	auth.GET("/profile", authMiddleware, profileHandler)

	log.Println("Auth routes registered successfully")
}

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未登录",
		})
		c.Abort()
		return
	}

	user, err := validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "登录失效",
		})
		c.Abort()
		return
	}

	c.Set("user", user)
	c.Next()
}

func registerHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误",
		})
		return
	}

	user, err := createUser(req.Username, req.Password, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	token, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "注册成功",
		"data": gin.H{
			"user": gin.H{
				"uuid":     user.UUID,
				"username": user.Username,
				"email":    user.Email,
			},
			"token": token,
		},
	})
}

func loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误",
		})
		return
	}

	user, err := getUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户名或密码错误",
		})
		return
	}

	if !verifyPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户名或密码错误",
		})
		return
	}

	token, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成token失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登录成功",
		"data": gin.H{
			"user": gin.H{
				"uuid":         user.UUID,
				"uuid_visitor": user.UUID,
				"username":     user.Username,
				"email":        user.Email,
			},
			"token": token,
		},
	})
}

func logoutHandler(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		invalidateToken(token)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "退出成功",
	})
}

func profileHandler(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未登录",
		})
		return
	}

	userData := user.(*User)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"uuid":         userData.UUID,
			"uuid_visitor": userData.UUID,
			"username":     userData.Username,
			"email":        userData.Email,
			"createdAt":    userData.CreatedAt,
		},
	})
}
