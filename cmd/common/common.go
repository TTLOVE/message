package common

import (
	"fmt"
	"io"
	"log"
	"message/internal/conf"

	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var cf *conf.ConfigData

func GetConfig() *conf.ConfigData {
	if cf != nil {
		return cf
	}

	// 读取配置信息
	viper.SetConfigName("config/config.json")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("read config failed: %v", err)
	}

	viper.Unmarshal(&cf)

	return cf
}

// initJaeger 将jaeger tracer设置为全局tracer
func InitJaeger(service string) io.Closer {
	c := GetConfig()
	cfg := jaegercfg.Configuration{
		// 将采样频率设置为1，每一个span都记录，方便查看测试结果
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
			// 将span发往jaeger-collector的服务地址
			CollectorEndpoint: fmt.Sprintf("%s:%d%s", c.GetJaeger().Host, c.GetJaeger().Port, c.GetJaeger().Path),
		},
	}
	closer, err := cfg.InitGlobalTracer(service, jaegercfg.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return closer
}
