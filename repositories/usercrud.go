package repositories

import (
	"context"

	"github.com/davq23/jokeapi/data"
	"github.com/go-playground/validator/v10"
)

type UserCRUD interface {
	FetchAll(ctx context.Context, limit uint64, offset string, direction FetchDirection) (data.Users, *string, error)
	FetchOne(ctx context.Context, id string) (*data.User, error)
	Insert(ctx context.Context, user *data.User) (string, error)
	Update(ctx context.Context, id string, user *data.User) (string, error)
	CheckValidID(fl validator.FieldLevel) bool
	Delete(ctx context.Context, id string) (string, error)
}
