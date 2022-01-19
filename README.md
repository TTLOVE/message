# graduate
go graduate code

## 1.运行例子
```
运行服务端方法：
方法1：直接执行 go run ./cmd/message/main.go ./cmd/message/wire_gen.go 
方法2：构建服务, go build ./cmd/message，然后执行目录下方的可执行文件 message

服务监听的是8080端口，可以通过请求http://localhost:8080/out或者命令行直接停止运行

客户端代码在./cmd/client，可通过和上方服务端运行方法启动

客户端监听的是8000端口，可通过请求http://localhost:8000/message得到返回的message信息
```

