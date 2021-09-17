package repositories

import (
	"context"

	"github.com/davq23/jokeapi/data"
)

type JokeCRUD interface {
	Delete(ctx context.Context, id string) (string, error)
	FetchAll(ctx context.Context, limit uint64, offset string, direction FetchDirection) (data.Jokes, *string, error)
	FetchRatings(ctx context.Context, jokeID string, limit uint64, offset string, direction FetchDirection) (data.JokeRatings, *string, error)
	FetchOne(ctx context.Context, id string) (*data.Joke, error)
	Insert(ctx context.Context, joke *data.Joke) (string, error)
	Update(ctx context.Context, id string, joke *data.Joke) (string, error)
	RateJoke(ctx context.Context, jokeID string, jokeRating *data.JokeRating) (string, error)
	DeleteRating(ctx context.Context, jokeID string, ratingID string, authID string) (string, error)
}
