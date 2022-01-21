package data

import (
	"context"
	v1 "message/api/message/v1"
	"message/internal/biz"

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
	span.SetTag("data", "get")

	var m v1.Message
	r.data.DB.First(&m)
	return &m, nil
}
