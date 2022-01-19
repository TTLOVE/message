package main

import (
	"context"
	"fmt"
	"log"
	message "message/api/message/v1"
	v1 "message/api/message/v1"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/kataras/iris/v12"
	"google.golang.org/grpc"
)

var messageClient message.MessageServiceClient

func main() {
	app := iris.New()
	app.Logger().SetLevel("debug") //debug
	app.Handle("GET", "/message", getMessage)
	app.Handle("POST", "/message/add", createMessage)
	app.Run(iris.Addr("127.0.0.1:8000"))
}

func getMessage(ctx iris.Context) {
	params := message.GetMessageRequest{}
	params.Id = 12
	res, err := messageClient.GetMessage(context.Background(), &params)
	if err != nil {
		log.Fatalf("client.GetAddressBook err: %v", err)
	}
	ctx.JSON(res)
}

type create struct {
	SystemId int32  `valid:"required~请输入系统id"`
	Title    string `valid:"required~请输入标题"`
	Content  string `valid:"required~请输入消息内容"`
	Url      string `valid:"required~请输入链接"`
}

func createMessage(ctx iris.Context) {
	systemId64, err := strconv.ParseInt(ctx.FormValue("system_id"), 10, 64)
	if err != nil {
		fmt.Println("system_id is not int")
		return
	}

	params := &message.CreateMessageRequest{
		SystemId: int32(systemId64),
		Title:    ctx.FormValue("title"),
		Content:  ctx.FormValue("content"),
		Url:      ctx.FormValue("url"),
	}

	_, err = govalidator.ValidateStruct(&create{
		SystemId: params.SystemId,
		Title:    params.Title,
		Content:  params.Content,
		Url:      params.Url,
	})
	if err != nil {
		ctx.JSON(map[string]interface{}{
			"code": v1.ErrorReason_VALIDATE_FAIL.Enum(),
			"msg":  fmt.Sprintf("校验出错:%s", err),
		})
		return
	}

	res, err := messageClient.CreateMessage(context.Background(), params)
	if err != nil {
		log.Fatalf("client.GetAddressBook err: %v", err)
	}
	ctx.JSON(res)
}

func init() {
	connect, err := grpc.Dial("127.0.0.1:9527", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	messageClient = message.NewMessageServiceClient(connect)
}
