//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"message/internal/biz"
	"message/internal/conf"
	"message/internal/data"
	"message/internal/server"
	"message/internal/service"

	"github.com/google/wire"
	"github.com/kataras/iris/v12"
)

// initApp init application.
func initApp(c *conf.ConfigData) (*iris.Application, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
