package server

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var Module = fx.Module("server", 
	fx.Provide(gin.New),
	fx.Provide(func(engine *gin.Engine) gin.IRouter {
		return gin.IRouter(engine)
	}),
	fx.Invoke(NewHTTPServer),
)

func NewHTTPServer(lc fx.Lifecycle, engine *gin.Engine, log *zap.Logger){
	engine.Use(errorMiddleware(log))

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			go engine.Run(":"+port)
			
			return nil
		},
	})
}

func NewClient() *http.Client {
	return &http.Client{}
}

func errorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, ginErr := range c.Errors {
			logger.Error(ginErr.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error:": ginErr.Error()})
		}
	}
}
