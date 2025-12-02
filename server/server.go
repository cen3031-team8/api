package server

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/cen3031-team8/api/database"
	_ "github.com/lib/pq"
)

var Module = fx.Module("server",
	fx.Provide(NewDatabaseConnection),
	fx.Provide(func(db *sql.DB) *database.Queries {
		return database.New(db)
	}),
	fx.Provide(gin.New),
	fx.Provide(func(engine *gin.Engine) gin.IRouter {
		return gin.IRouter(engine)
	}),
	fx.Invoke(NewHTTPServer),
)

func NewDatabaseConnection() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pokemon_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewHTTPServer(lc fx.Lifecycle, engine *gin.Engine, log *zap.Logger) {
	engine.Use(errorMiddleware(log))

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			go engine.Run(":" + port)

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
