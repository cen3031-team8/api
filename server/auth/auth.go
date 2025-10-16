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
}
