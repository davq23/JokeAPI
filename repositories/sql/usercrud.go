package sql

import (
	"context"
	"database/sql"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/repositories"
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

func (jr *UserCRUD) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Users, *string, error) {
	var users data.Users

	var condition string

	if direction == repositories.FetchBack {
		condition = "<"
	} else {
		condition = ">="
	}

	rows, err := jr.db.QueryContext(ctx, "SELECT id, author_id, text, explanation FROM users WHERE id "+condition+" $1 LIMIT $2", offset, limit)

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
	result := jr.db.QueryRowContext(ctx, "SELECT id, email, admin FROM users WHERE id = $1", id)
	user := new(data.User)

	if err := result.Scan(&user.ID, &user.Email, &user.Admin); err != nil {
		if err == sql.ErrNoRows {
			return nil, repositories.ErrUnknownID
		}

		return nil, err
	}

	return user, nil
}

func (jr *UserCRUD) Insert(ctx context.Context, user *data.User) (string, error) {
	tx, err := jr.db.BeginTx(ctx, nil)

	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO users (id, email, password) VALUES (?, ?, ?)",
		user.ID, user.Email, user.Password)

	if err != nil {
		tx.Rollback()

		return "", err
	}

	tx.Commit()

	return user.ID, nil
}

func (jr *UserCRUD) Update(ctx context.Context, id string, user *data.User) (string, error) {
	tx, err := jr.db.BeginTx(ctx, nil)

	if err != nil {
		return "", err
	}

	row := tx.QueryRowContext(ctx, "SELECT id FROM jokes WHERE id = ?", id)
	err = row.Scan(&id)

	if err != nil {
		tx.Rollback()

		if err == sql.ErrNoRows {
			return "", repositories.ErrUnknownID
		}

		return "", err
	}

	_, err = tx.ExecContext(ctx,
		"UPDATE users SET email = ?, password = ? WHERE id = ?",
		user.Email, user.Password, id)

	if err != nil {
		tx.Rollback()
		return "", err
	}

	tx.Commit()

	return id, nil
}

func (jr *UserCRUD) Delete(ctx context.Context, id string) (string, error) {
	tx, err := jr.db.BeginTx(ctx, nil)

	if err != nil {
		return "", err
	}

	row := tx.QueryRowContext(ctx, "SELECT id FROM jokes WHERE id = ?", id)
	err = row.Scan(&id)

	if err != nil {
		tx.Rollback()

		if err == sql.ErrNoRows {
			return "", repositories.ErrUnknownID
		}

		return "", err
	}

	_, err = jr.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)

	if err != nil {
		tx.Rollback()

		return "", err
	}

	tx.Commit()

	return id, nil
}
