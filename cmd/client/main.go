package main

import (
	"context"
	"fmt"
	"io"
	"log"
	message "message/api/message/v1"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/kataras/iris/v12"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	opLog "github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var messageClient message.MessageServiceClient
var rootCtx context.Context

// initJaeger 将jaeger tracer设置为全局tracer
func initJaeger(service string) io.Closer {
	cfg := jaegercfg.Configuration{
		// 将采样频率设置为1，每一个span都记录，方便查看测试结果
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
			// 将span发往jaeger-collector的服务地址
			CollectorEndpoint: "http://127.0.0.1:14268/api/traces",
		},
	}
	closer, err := cfg.InitGlobalTracer(service, jaegercfg.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return closer
}

func main() {

	app := iris.New()
	app.Logger().SetLevel("debug") //debug

	// jaeger / opentracing specific stuff
	{
		closer := initJaeger("client")
		defer closer.Close()
		// 获取jaeger tracer
		t := opentracing.GlobalTracer()
		// 创建root span
		sp := t.StartSpan("client-service")
		// main执行完结束这个span
		defer sp.Finish()
		// 将span传递给Foo
		rootCtx = opentracing.ContextWithSpan(context.Background(), sp)
	}

	app.Handle("GET", "/message", getMessage)
	app.Handle("POST", "/message/add", createMessage)
	app.Run(iris.Addr("127.0.0.1:8000"))
}

func getMessage(ctx iris.Context) {
	span, mCtx := opentracing.StartSpanFromContext(rootCtx, "message")
	defer span.Finish()

	fmt.Println("mCtx", mCtx)
	fmt.Println("span", opentracing.SpanFromContext(mCtx))

	params := message.GetMessageRequest{}
	params.Id = 12
	res, err := messageClient.GetMessage(mCtx, &params)
	if err != nil {
		log.Fatalf("client.GetMessage err: %v", err)
		ext.Error.Set(span, true) // Tag the span as errored
		span.LogFields(
			opLog.String("event", fmt.Sprintf("getMessage error:%s", err)),
		)
		span.LogEventWithPayload("GET service error", params) // Log the error
	}

	span.SetTag("get", "success")

	span1, _ := opentracing.StartSpanFromContext(mCtx, "new message")
	span1.Finish()

	span1.LogEventWithPayload("payload", params)

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
			"code": message.ErrorReason_VALIDATE_FAIL.Enum(),
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
	connect, err := grpc.Dial(
		"127.0.0.1:9527",
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(jaegerGrpcClientInterceptor),
	)
	if err != nil {
		log.Fatalln(err)
	}
	messageClient = message.NewMessageServiceClient(connect)
}

type TextMapWriter struct {
	metadata.MD
}

//重写TextMapWriter的Set方法，我们需要将carrier中的数据写入到metadata中，这样grpc才会携带。
func (t TextMapWriter) Set(key, val string) {
	//key = strings.ToLower(key)
	t.MD[key] = append(t.MD[key], val)
}

func jaegerGrpcClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	var parentContext opentracing.SpanContext
	//先从context中获取原始的span
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		parentContext = parentSpan.Context()
	}
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan(method, opentracing.ChildOf(parentContext))
	defer span.Finish()
	//从context中获取metadata。md.(type) == map[string][]string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		//如果对metadata进行修改，那么需要用拷贝的副本进行修改。（FromIncomingContext的注释）
		md = md.Copy()
	}
	//定义一个carrier，下面的Inject注入数据需要用到。carrier.(type) == map[string]string
	//carrier := opentracing.TextMapCarrier{}
	carrier := TextMapWriter{md}
	//将span的context信息注入到carrier中
	e := tracer.Inject(span.Context(), opentracing.TextMap, carrier)
	if e != nil {
		fmt.Println("tracer Inject err,", e)
	}
	//创建一个新的context，把metadata附带上
	ctx = metadata.NewOutgoingContext(ctx, md)

	return invoker(ctx, method, req, reply, cc, opts...)
}
