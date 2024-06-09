package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	mockproviders "github.com/zde37/Jusgo/internal/mock"
	"github.com/zde37/Jusgo/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/mock/gomock"
)

func TestCreateJoke(t *testing.T) {
	ctx := context.Background()
	joke := createJoke()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockproviders.NewMockRepositoryProvider(ctrl)

	// build stubs
	repo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Times(1).
		Return(joke, nil)

	service := NewService(repo)
	createdJoke, err := service.srv.CreateJoke(ctx, joke)
	require.NoError(t, err)
	require.NotEmpty(t, createdJoke)
	require.Equal(t, joke, createdJoke)
}

func TestGetJoke(t *testing.T) {
	ctx := context.Background()
	joke := createJoke()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockproviders.NewMockRepositoryProvider(ctrl)

	// build stubs
	repo.EXPECT().
		Get(gomock.Any(), gomock.Eq(joke.ID)).
		Times(1).
		Return(joke, nil)

	service := NewService(repo)
	joke2, err := service.srv.GetJoke(ctx, joke.ID.Hex())
	require.NoError(t, err)
	require.NotEmpty(t, joke2)
	require.Equal(t, joke, joke2)
}

func TestGetAllJokes(t *testing.T) {
	ctx := context.Background()
	jokes := []models.Jusgo{createJoke(), createJoke(), createJoke(), createJoke(), createJoke(), createJoke(), createJoke(), createJoke(), createJoke(), createJoke()}
	page, limit := 1, 10
	skip := (page - 1) * limit

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockproviders.NewMockRepositoryProvider(ctrl)

	// build stubs
	repo.EXPECT().
		GetAll(gomock.Any(), gomock.Eq(int64(skip)), gomock.Eq(int64(limit))).
		Times(1).
		Return(jokes, nil)

	service := NewService(repo)
	allJokes, err := service.srv.GetAllJokes(ctx, page, limit)
	require.NoError(t, err)
	require.NotEmpty(t, allJokes)
	require.Len(t, jokes, 10)
}

func TestUpdateJoke(t *testing.T) {
	ctx := context.Background()
	joke := createJoke()
	joke.Joke = "Yay...I love coding"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockproviders.NewMockRepositoryProvider(ctrl)

	// build stubs
	repo.EXPECT().
		Update(gomock.Any(), gomock.Eq(joke)).
		Times(1).
		Return(joke, nil)

	service := NewService(repo)
	updatedJoke, err := service.srv.UpdateJoke(ctx, joke.ID.Hex(), joke)
	require.NoError(t, err)
	require.NotEmpty(t, updatedJoke)
	require.Equal(t, joke, updatedJoke)
}

func TestDeleteJoke(t *testing.T) {
	ctx := context.Background()
	joke := createJoke()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockproviders.NewMockRepositoryProvider(ctrl)

	// build stubs
	repo.EXPECT().
		Delete(gomock.Any(), gomock.Eq(joke.ID)).
		Times(1).
		Return(nil)

	service := NewService(repo)
	err := service.srv.DeleteJoke(ctx, joke.ID.Hex())
	require.NoError(t, err)
}

func createJoke() models.Jusgo {
	return models.Jusgo{
		ID:        primitive.NewObjectID(),
		Joke:      "Roses are red violets are blue...unknown error on line 42",
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}
}
