package main

import (
	"time"

	"github.com/cen3031-team8/api/server"
	"github.com/cen3031-team8/api/server/auth"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	time.Local = time.UTC
	godotenv.Load()

	fx.New(
		fx.Provide(
			zap.NewDevelopment,
			server.NewClient,
		),

		server.Module,
		server.PokemonModule,
		auth.Module,
	).Run()
}
