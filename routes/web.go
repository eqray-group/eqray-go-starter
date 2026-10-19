package routes

import (
	"eqray-go-starter/app/facades"

	"github.com/goravel/framework/contracts/http"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().Success().String("api is running")
	})
}
