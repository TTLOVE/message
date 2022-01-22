package main

import (
	"context"
	"fmt"
	"log"
	"message/cmd/common"
	"message/pkg/myGrpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kataras/iris/v12"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var mgs *myGrpc.Server

func newApp(gs *myGrpc.Server) *iris.Application {
	app := iris.New()

	listen, err := net.Listen("tcp", "127.0.0.1:9527")
	if err != nil {
		log.Fatalf("tcp listen failed:%v", err)
	}

	mgs = gs
	mgs.Listener(listen)

	return app
}

func main() {
	closer := common.InitJaeger("service")
	defer closer.Close()

	c := common.GetConfig()
	viper.Unmarshal(&c)

	app, cleanup, err := initApp(c)
	if err != nil {
		fmt.Println("启动失败", err)
	}
	defer cleanup()

	// 服务启动
	ctx := context.Background()
	startServe(app, ctx)
}

// 启动多个服务信息
func startServe(app *iris.Application, ctx context.Context) {
	// 生成errgroup
	g, cxt := errgroup.WithContext(ctx)

	// 生成处理请求的handler
	mux := http.NewServeMux()

	// 模拟页面申请退出
	reqOut := make(chan struct{})
	// 注册路由信息
	out(app, reqOut)
	mux.HandleFunc("/out", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "shutting down")
		reqOut <- struct{}{}
	})

	g.Go(func() error {
		select {
		case <-reqOut:
			log.Println("shutdown from request")
		case <-cxt.Done():
			log.Println("shutdown from errgroup")
		}

		return stopServe(app, ctx)
	})

	// 发起http服务
	g.Go(func() error {
		log.Println("http server start")

		err := app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
		if err != nil {
			return errors.Wrap(err, "http serve fail")
		}

		return nil
	})

	// 发起grpc服务
	g.Go(func() error {
		log.Println("grpc server start")

		err := mgs.Server.Serve(mgs.GetListener())
		if err != nil {
			return errors.Wrap(err, "grpc serve fail")
		}

		return nil
	})

	done := make(chan os.Signal)
	// 创建系统信号接收器
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// 监听linux signal退出通道
	g.Go(func() error {
		select {
		case <-done:
			log.Println("shutdown from signal")
		case <-cxt.Done():
			log.Println("shutdown from errgroup")
		}

		return stopServe(app, ctx)
	})

	if err := g.Wait(); err != nil {
		log.Println(err)
	}
}

func out(app *iris.Application, reqOut chan struct{}) {
	app.Get("/out", func(c iris.Context) {
		c.ResponseWriter().Write([]byte("shutting down"))

		reqOut <- struct{}{}
	})
}

func stopServe(app *iris.Application, ctx context.Context) error {
	// 关闭http
	app.Shutdown(ctx)
	// 关闭grpc
	mgs.Server.Stop()

	return errors.New("out")
}
