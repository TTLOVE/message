package data

import (
	"context"
	"fmt"
	v1 "message/api/message/v1"
	"message/cmd/common"
	"message/internal/biz"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
)

type messageRepo struct {
	data *Data
}

// NewMessageRepo .
func NewMessageRepo(data *Data) biz.MessageRepo {
	return &messageRepo{
		data: data,
	}
}

func (r *messageRepo) CreateMessage(ctx context.Context, g *v1.Message) error {
	r.data.Create(&g)
	return nil
}

func (r *messageRepo) UpdateMessage(ctx context.Context, g *v1.Message) error {
	return nil
}

func (r *messageRepo) GetMessage(ctx context.Context, id int32) (*v1.Message, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "data")
	defer span.Finish()

	cf := common.GetConfig()
	// send message to kafka
	pid, offset, err := r.data.KafkaClient.SendMessage(
		&sarama.ProducerMessage{
			Topic: cf.Kafka.Topic,
			Key:   sarama.StringEncoder("test_key"),
			Value: sarama.StringEncoder("go send message"),
		},
	)
	if err != nil {
		fmt.Println("send msg failed, err:", err)
		span.LogKV("send kafka fail:", err)
		return nil, err
	}
	fmt.Printf("pid:%v offset:%v\n", pid, offset)

	span.SetTag("data", "get")

	var m v1.Message
	r.data.DB.First(&m)
	return &m, nil
}
