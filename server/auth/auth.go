package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"auth",
	fx.Decorate(func(router gin.IRouter) gin.IRouter {
		return router.Group("auth")
	}),

	fx.Invoke(RegisterHandlers),
)

func RegisterHandlers(router gin.IRouter, client *http.Client, log *zap.Logger) {
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"result": "ok"})
	})

	router.POST("/login", func(c *gin.Context) {
		type LoginPayload struct {
			User string `json:"user"`
			Pass string `json:"pass"`
		}

		var payload LoginPayload
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if (payload.User == "" || payload.Pass == "") {
			c.Status(http.StatusBadRequest)
			return
		}

		if payload.User == "notCreated" {
			c.Status(http.StatusUnauthorized)
			return
		}

		log.Info("Login request:", zap.String("user", payload.User))

		c.Status(http.StatusOK)
	})

	router.POST("/register", func(c *gin.Context) {
		type RegisterPayload struct {
			User string `json:"user"`
			Pass string `json:"pass"`
		}
	
		var payload RegisterPayload
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
	
		if payload.User == "" || payload.Pass == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing fields"})
			return
		}
	
		if payload.User == "existingUser" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
	
		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user":    payload.User,
		})
	})

	router.GET("/profile", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "Bearer valid_demo_token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": "demoUser",
			"role": "trainer",
		})
	})

	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": "1.0.0",
			"status":  "stable",
		})
	})

}
