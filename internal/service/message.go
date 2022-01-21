package service

import (
	"context"
	"fmt"

	v1 "message/api/message/v1"
	"message/internal/biz"

	errors "github.com/go-kratos/kratos/v2/errors"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// MessageService is a message service.
type MessageService struct {
	v1.UnimplementedMessageServiceServer

	uc *biz.MessageUsecase
}

// NewMessageService new a message service.
func NewMessageService(uc *biz.MessageUsecase) *MessageService {
	return &MessageService{uc: uc}
}

// GetMessage implements message.MessageServer
func (s *MessageService) GetMessage(ctx context.Context, re *v1.GetMessageRequest) (*v1.Message, error) {
	span, mCtx := opentracing.StartSpanFromContext(ctx, "service message")
	defer span.Finish()

	if re.GetId() == 0 {
		span.LogFields(
			log.String("id", "empty"),
		)

		span.LogEventWithPayload("request:", *re)

		return nil, errors.New(
			int(v1.ErrorReason_MESSAGE_NOT_FOUND),
			fmt.Sprintf("message not found :%d", re.GetId()),
			"not found",
		)
	}

	span.SetTag("service done", "go usecase")

	return s.uc.Get(mCtx, re.GetId())
}

// CreateMessage implements message.MessageServer
func (s *MessageService) CreateMessage(ctx context.Context, re *v1.CreateMessageRequest) (*v1.Message, error) {
	m := &v1.Message{
		SystemId: re.SystemId,
		Title:    re.Title,
		Content:  re.Content,
		Url:      re.Url,
	}
	return m, s.uc.Create(ctx, m)
}
