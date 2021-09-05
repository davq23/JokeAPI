package repositories

import (
	"context"
)

type TokenCRUD interface {
	Delete(ctx context.Context, id string) (string, error)
	Renew(ctx context.Context, id string) (string, error)
	Insert(ctx context.Context, id string) (string, error)
}
