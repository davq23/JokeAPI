package repositories

import (
	"context"

	"github.com/davq23/jokeapi/data"
)

type JokeRepository interface {
	FetchAll(ctx context.Context, offset string, limit uint64, direction FetchDirection) (data.Jokes, string, error)
	FetchOne(ctx context.Context, id string) (*data.Joke, error)
	Insert(ctx context.Context, joke *data.Joke) (string, error)
	Update(ctx context.Context, id string, joke *data.Joke) (*data.Joke, error)
	Delete(ctx context.Context, id string) (string, error)
}
