package server

import (
	"net/http"
	"strings"

	"github.com/cen3031-team8/api/server/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var PokemonModule = fx.Module("pokemon",
	fx.Invoke(RegisterPokemonHandlers),
)

func authMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		_, err := authService.VerifyJWT(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func RegisterPokemonHandlers(router gin.IRouter, authService *auth.AuthService, log *zap.Logger) {
	// Pokemon endpoints
	pokemonGroup := router.Group("pokemon")
	pokemonGroup.Use(authMiddleware(authService))

	pokemonGroup.POST("/add", authService.HandleAddPokemon)

	// Inventory endpoints
	inventoryGroup := router.Group("inventory")
	inventoryGroup.Use(authMiddleware(authService))

	inventoryGroup.PUT("/update", authService.HandleUpdateInventory)
}
