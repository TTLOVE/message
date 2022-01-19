package biz

import (
	"context"
	v1 "message/api/message/v1"
)

type MessageRepo interface {
	CreateMessage(context.Context, *v1.Message) error
	UpdateMessage(context.Context, *v1.Message) error
	GetMessage(context.Context, int32) (*v1.Message, error)
}

type MessageUsecase struct {
	repo MessageRepo
}

func NewMessageUsecase(repo MessageRepo) *MessageUsecase {
	return &MessageUsecase{repo: repo}
}

func (uc *MessageUsecase) Create(ctx context.Context, g *v1.Message) error {
	return uc.repo.CreateMessage(ctx, g)
}

func (uc *MessageUsecase) Update(ctx context.Context, g *v1.Message) error {
	return uc.repo.UpdateMessage(ctx, g)
}

func (uc *MessageUsecase) Get(ctx context.Context, id int32) (*v1.Message, error) {
	return uc.repo.GetMessage(ctx, id)
}
