package server

import (
	message "message/api/message/v1"
	"message/internal/service"
	"message/pkg/myGrpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(m *service.MessageService) *myGrpc.Server {
	srv := myGrpc.NewServer()
	message.RegisterMessageServiceServer(srv, m)
	return srv
}
