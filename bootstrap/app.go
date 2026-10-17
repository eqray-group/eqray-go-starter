package bootstrap

import (
	contractsFoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"

	"eqray-go-starter/config"
	"eqray-go-starter/routes"
)

func Boot() contractsFoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).
		WithRouting(func() {
			routes.Web()
			routes.Grpc()
			routes.Modules()
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
