package auth

import (
	"net/http"

	"github.com/cen3031-team8/api/database"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"auth",
	fx.Provide(func(queries *database.Queries, log *zap.Logger) *AuthService {
		return NewAuthService(queries, log)
	}),
	fx.Decorate(func(router gin.IRouter) gin.IRouter {
		return router.Group("auth")
	}),

	fx.Invoke(RegisterHandlers),
)

func RegisterHandlers(router gin.IRouter, authService *AuthService, client *http.Client, log *zap.Logger) {
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"result": "ok"})
	})

	router.POST("/login", authService.HandleLogin)
	router.POST("/register", authService.HandleRegister)
	router.GET("/profile", authService.HandleProfile)
}
