package sql

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type UserCRUD struct {
	db *sqlx.DB
}

func NewUserCRUD(db *sqlx.DB) *UserCRUD {
	return &UserCRUD{
		db: db,
	}
}

func (jr *UserCRUD) CheckValidID(fl validator.FieldLevel) bool {
	id := fl.Field().String()

	_, err := strconv.ParseUint(id, 10, 64)

	return err == nil
}

func (jr *UserCRUD) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Users, *string, error) {
	var users data.Users

	offsetID, err := strconv.ParseUint(offset, 10, 64)

	if err != nil && offset != "" {
		return users, nil, repositories.ErrInvalidOffset
	}

	var condition string

	if direction == repositories.FetchBack {
		condition = "<"
	} else {
		condition = ">="
	}

	rows, err := jr.db.QueryContext(ctx, "SELECT id, author_id, text, explanation FROM users WHERE id "+condition+" $1 LIMIT $2", offsetID, limit)

	if err != nil {
		return users, nil, repositories.ErrInvalidOffset
	}

	defer rows.Close()

	i := uint64(0)

	users = make(data.Users, 0, limit)

	var user *data.User

	for i != limit && rows.Next() {
		user = new(data.User)

		if err = rows.Scan(user); err != nil {
			return users, nil, err
		}

		users = append(users, user)
		i++
	}

	var nextID *string

	if rows.Next() && user != nil {
		nextID = &user.ID
	}

	return users, nextID, nil
}

func (jr *UserCRUD) FetchOne(ctx context.Context, id string) (*data.User, error) {
	objectID, err := strconv.ParseUint(id, 10, 64)

	if err != nil {
		return nil, err
	}

	result := jr.db.QueryRowContext(ctx, "SELECT id, author_id, text, explanation FROM users WHERE id = $1", objectID)

	if result.Err() != nil {
		return nil, result.Err()
	}

	user := new(data.User)

	if err = result.Scan(user); err != nil {
		if err == sql.ErrNoRows {
			return nil, repositories.ErrUnknownID
		}

		return nil, err
	}

	return user, nil
}

func (jr *UserCRUD) Insert(ctx context.Context, user *data.User) (string, error) {
	result, err := jr.db.QueryContext(ctx,
		"INSER INTO users (author_id, text, explanation, lang) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Email, user.Password)

	if err != nil {
		return "", err
	}

	var newID string

	if result.Next() {
		err = result.Scan(&newID)

		if err != nil {
			return "", err
		}
	}

	return newID, nil
}

func (jr *UserCRUD) Update(ctx context.Context, id string, user *data.User) (string, error) {
	result, err := jr.db.ExecContext(ctx,
		"UPDATE users SET author_id = $1, text = $2, explanation = $3, lang = $4 WHERE id = $5",
		user.Email, user.Password, id)

	if err != nil {
		return "", err
	}

	n, err := result.RowsAffected()

	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", repositories.ErrUnknownID
	}

	return id, nil
}

func (jr *UserCRUD) Delete(ctx context.Context, id string) (string, error) {
	result, err := jr.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)

	if err != nil {
		return "", err
	}

	n, err := result.RowsAffected()

	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", repositories.ErrUnknownID
	}

	return id, nil
}
