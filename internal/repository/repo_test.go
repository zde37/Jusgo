package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zde37/Jusgo/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreate(t *testing.T) {
	createJoke(t, context.Background())
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	data := createJoke(t, ctx)

	joke, err := testRepo.repo.Get(ctx, data.ID)
	require.NoError(t, err)
	require.NotEmpty(t, joke)
	require.Equal(t, data.ID, joke.ID)
	require.Equal(t, data.Joke, joke.Joke)
	require.Equal(t, data.UpdatedAt.Unix(), joke.UpdatedAt.Unix())
	require.Equal(t, data.CreatedAt.Unix(), joke.CreatedAt.Unix())
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	data := createJoke(t, ctx)

	data.Joke = "I used to know a joke about Java...but I ran out of memory"
	data.UpdatedAt = time.Now()

	updatedJoke, err := testRepo.repo.Update(ctx, data)
	require.NoError(t, err)
	require.NotEmpty(t, updatedJoke)
	require.Equal(t, data.ID, updatedJoke.ID)
	require.Equal(t, data.Joke, updatedJoke.Joke)
	require.Equal(t, data.UpdatedAt.Unix(), updatedJoke.UpdatedAt.Unix())
	require.Equal(t, data.CreatedAt.Unix(), updatedJoke.CreatedAt.Unix())
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()
	for range 20 {
		createJoke(t, ctx)
	}

	testData := []struct {
		Name  string
		limit int64
		page  int64
		stub  func(t *testing.T, limit, page int64)
	}{
		{
			Name:  "Fetch 10 jokes from page 1",
			limit: 10,
			page:  1,
			stub: func(t *testing.T, limit, page int64) {
				skip := (page - 1) * limit
				jokes, err := testRepo.repo.GetAll(ctx, skip, limit)
				require.NoError(t, err)
				require.NotEmpty(t, jokes)
				require.Len(t, jokes, 10)

				for _, joke := range jokes {
					require.NotEmpty(t, joke)
				}
			},
		},
		{
			Name:  "Fetch 10 jokes from page 2",
			limit: 10,
			page:  2,
			stub: func(t *testing.T, limit, page int64) {
				skip := (page - 1) * limit
				jokes, err := testRepo.repo.GetAll(ctx, skip, limit)
				require.NoError(t, err)
				require.NotEmpty(t, jokes)
				require.Len(t, jokes, 10)

				for _, joke := range jokes {
					require.NotEmpty(t, joke)
				}
			},
		},
		{
			Name:  "Fetch 5 jokes from page 1",
			limit: 5,
			page:  1,
			stub: func(t *testing.T, limit, page int64) {
				skip := (page - 1) * limit
				jokes, err := testRepo.repo.GetAll(ctx, skip, limit)
				require.NoError(t, err)
				require.NotEmpty(t, jokes)
				require.Len(t, jokes, 5)

				for _, joke := range jokes {
					require.NotEmpty(t, joke)
				}
			},
		},
	}

	for _, tc := range testData {
		t.Run(tc.Name, func(t *testing.T) {
			tc.stub(t, tc.limit, tc.page)
		})
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	joke := createJoke(t, ctx)

	err := testRepo.repo.Delete(ctx, joke.ID)
	require.NoError(t, err)

	deletedJoke, err := testRepo.repo.Get(ctx, joke.ID)
	require.Error(t, err)
	require.Empty(t, deletedJoke)
}

func createJoke(t *testing.T, ctx context.Context) models.Jusgo {
	data := models.Jusgo{
		ID:        primitive.NewObjectID(),
		Joke:      "I'm declaring a war. var war",
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	joke, err := testRepo.repo.Create(ctx, data)
	require.NoError(t, err)
	require.NotEmpty(t, joke)
	require.Equal(t, data, joke)

	return joke
} 