package repository

import (
	"context"

	"github.com/zde37/Jusgo/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type repositoryImpl struct {
	collection *mongo.Collection
}

func newRepositoryImpl(c *mongo.Collection) *repositoryImpl {
	return &repositoryImpl{
		collection: c,
	}
}

func (r *repositoryImpl) Create(ctx context.Context, data models.Jusgo) (models.Jusgo, error) {
	_, err := r.collection.InsertOne(ctx, data)
	return data, err
}

func (r *repositoryImpl) Get(ctx context.Context, id primitive.ObjectID) (models.Jusgo, error) {
	var jusgo models.Jusgo
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&jusgo)

	return jusgo, err
}

func (r *repositoryImpl) Update(ctx context.Context, data models.Jusgo) (models.Jusgo, error) {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": data.ID}, bson.M{"$set": data})
	return data, err
}

func (r *repositoryImpl) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *repositoryImpl) GetAll(ctx context.Context, skip, limit int64) ([]models.Jusgo, error) {
	options := options.Find()
	options.SetSkip(skip)
	options.SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{}, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	jokes := []models.Jusgo{} // initialize it so it will return '[]' instead of null if the list is empty
	if err = cursor.All(ctx, &jokes); err != nil {
		return nil, err
	}

	return jokes, nil
}
