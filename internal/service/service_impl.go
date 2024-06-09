package service

import (
	"context" 
	// "time"

	"github.com/zde37/Jusgo/internal/models"
	"github.com/zde37/Jusgo/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type serviceImpl struct {
	repo repository.RepositoryProvider
}

func newServiceImpl(repo repository.RepositoryProvider) *serviceImpl {
	return &serviceImpl{
		repo: repo,
	}
}

func (s *serviceImpl) CreateJoke(ctx context.Context, data models.Jusgo) (models.Jusgo, error) {
	return s.repo.Create(ctx, data)
}

func (s *serviceImpl) GetJoke(ctx context.Context, id string) (models.Jusgo, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return models.Jusgo{}, err
	}

	return s.repo.Get(ctx, objectID)
}

// TODO: set default to limit -> 10, page -> 1 
func (s *serviceImpl) GetAllJokes(ctx context.Context, page, limit int) ([]models.Jusgo, error) {
	skip := (page - 1) * limit
	return s.repo.GetAll(ctx, int64(skip), int64(limit))

}

func (s *serviceImpl) UpdateJoke(ctx context.Context, id string, data models.Jusgo) (models.Jusgo, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return data, err
	}

	data.ID = objectID
	// data.UpdatedAt = time.Now()
	return s.repo.Update(ctx, data)
}

func (s *serviceImpl) DeleteJoke(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, objectID)
}
