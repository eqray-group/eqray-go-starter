package main

import (
	"eqray-go-starter/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
