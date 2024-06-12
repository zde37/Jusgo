package service

import (
	"context"

	"github.com/zde37/Jusgo/internal/models"
	"github.com/zde37/Jusgo/internal/repository"
)

type ServiceProvider interface {
	CreateJoke(ctx context.Context, data models.Jusgo) (models.Jusgo, error)
	GetJoke(ctx context.Context, id string) (models.Jusgo, error)
	UpdateJoke(ctx context.Context, data models.Jusgo) (models.Jusgo, error)
	DeleteJoke(ctx context.Context, id string) error
	GetAllJokes(ctx context.Context, page, limit int) ([]models.Jusgo, error)
}

type Service struct {
	Srvc ServiceProvider
}

func NewService(repo repository.RepositoryProvider) *Service {
	return &Service{
		Srvc: newServiceImpl(repo),
	}
}
