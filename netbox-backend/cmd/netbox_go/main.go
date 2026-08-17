// Package main is the http and grpc server of the application.
package main

import (
	"github.com/go-dev-frame/sponge/pkg/app"

	"netbox-go/cmd/netbox_go/initial"
)

func main() {
	httpRuntime := initial.InitApp()
	services := initial.CreateServices(httpRuntime)
	closes := initial.Close(services)

	a := app.New(services, closes)
	a.Run()
}
