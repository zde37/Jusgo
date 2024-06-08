package repository

import (
	"context"
 
	"github.com/zde37/Jusgo/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepositoryProvider interface {
	Create(ctx context.Context, data models.Jusgo) (models.Jusgo, error)
	Get(ctx context.Context, id primitive.ObjectID) (models.Jusgo, error)
	Update(ctx context.Context, data models.Jusgo) (models.Jusgo, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetAll(ctx context.Context, skip, limit int64) ([]models.Jusgo, error)
}

type Repository struct {
	repo RepositoryProvider
}

func NewRepository(collection *mongo.Collection) *Repository {
	return &Repository{
		repo: newRepositoryImpl(collection),
	}
}
