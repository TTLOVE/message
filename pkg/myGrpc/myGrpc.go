package myGrpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	network  string
	address  string
	timeout  time.Duration
	listener net.Listener
	ctx      context.Context
}

func NewServer() *Server {
	srv := &Server{
		ctx:     context.Background(),
		network: "tcp",
		address: "8080",
		timeout: 300 * time.Microsecond,
		Server:  grpc.NewServer(),
	}

	return srv
}

func (srv *Server) Listener(lis net.Listener) {
	srv.listener = lis
}

func (srv *Server) GetListener() net.Listener {
	return srv.listener
}
