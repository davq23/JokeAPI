package repositories

import (
	"context"

	"github.com/davq23/jokeapi/data"
	"github.com/go-playground/validator/v10"
)

type JokeRepository interface {
	FetchAll(ctx context.Context, limit uint64, offset string, direction FetchDirection) (data.Jokes, *string, error)
	FetchOne(ctx context.Context, id string) (*data.Joke, error)
	Insert(ctx context.Context, joke *data.Joke) (string, error)
	Update(ctx context.Context, id string, joke *data.Joke) (string, error)
	CheckValidID(fl validator.FieldLevel) bool
	Delete(ctx context.Context, id string) (string, error)
}
