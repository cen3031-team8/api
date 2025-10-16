package main

import (
	"github.com/cen3031-team8/api/server"
	"github.com/cen3031-team8/api/server/auth"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			zap.NewDevelopment,
			server.NewClient,
		),

		server.Module,
		auth.Module,
	).Run()
}
