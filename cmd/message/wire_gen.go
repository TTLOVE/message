// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/kataras/iris/v12"
	"message/internal/biz"
	"message/internal/conf"
	"message/internal/data"
	"message/internal/server"
	"message/internal/service"
)

// Injectors from wire.go:

// initApp init application.
func initApp(c *conf.ConfigData) (*iris.Application, func(), error) {
	dataData, cleanup, err := data.NewData(c)
	if err != nil {
		return nil, nil, err
	}
	messageRepo := data.NewMessageRepo(dataData)
	messageUsecase := biz.NewMessageUsecase(messageRepo)
	messageService := service.NewMessageService(messageUsecase)
	myGrpcServer := server.NewGRPCServer(messageService)
	application := newApp(myGrpcServer)
	return application, func() {
		cleanup()
	}, nil
}
