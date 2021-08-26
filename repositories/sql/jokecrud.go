package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type JokeCRUD struct {
	db *sqlx.DB
}

func NewJokeCRUD(db *sqlx.DB) *JokeCRUD {
	return &JokeCRUD{
		db: db,
	}
}

func (jr *JokeCRUD) CheckValidID(fl validator.FieldLevel) bool {
	id := fl.Field().String()

	_, err := strconv.ParseUint(id, 10, 64)

	return err == nil
}

func (jr *JokeCRUD) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Jokes, *string, error) {
	var jokes data.Jokes

	offsetID, err := strconv.ParseUint(offset, 10, 64)

	if err != nil && offset != "" {
		return jokes, nil, repositories.ErrInvalidOffset
	}

	var condition string

	if direction == repositories.FetchBack {
		condition = "<"
	} else {
		condition = ">="
	}

	rows, err := jr.db.QueryContext(ctx, "SELECT id, author_id, text, explanation, lang FROM jokes WHERE id "+condition+" ? LIMIT ?", offsetID, limit+1)

	if err != nil {
		return jokes, nil, err
	}

	defer rows.Close()

	i := uint64(0)

	jokes = make(data.Jokes, 0, limit)

	var joke *data.Joke

	for i != limit && rows.Next() {
		joke = new(data.Joke)

		if err = rows.Scan(&joke.ID, &joke.AuthorID, &joke.Text, &joke.Explanation, &joke.Language); err != nil {
			return jokes, nil, err
		}

		jokes = append(jokes, joke)
		i++
	}

	var nextID *string

	if rows.Next() && joke != nil {
		joke := new(data.Joke)

		rows.Scan(&joke.ID, &joke.AuthorID, &joke.Text, &joke.Explanation, &joke.Language)
		nextID = &joke.ID
	}

	return jokes, nextID, nil
}

func (jr *JokeCRUD) FetchOne(ctx context.Context, id string) (*data.Joke, error) {
	objectID, err := strconv.ParseUint(id, 10, 64)

	if err != nil {
		return nil, err
	}

	result := jr.db.QueryRowContext(ctx, "SELECT id, author_id, text, explanation, lang FROM jokes WHERE id = ?", objectID)

	if result.Err() != nil {
		return nil, result.Err()
	}

	joke := new(data.Joke)

	if err = result.Scan(&joke.ID, &joke.AuthorID, &joke.Text, &joke.Explanation, &joke.Language); err != nil {
		if err == sql.ErrNoRows {
			return nil, repositories.ErrUnknownID
		}

		return nil, err
	}

	return joke, nil
}

func (jr *JokeCRUD) Insert(ctx context.Context, joke *data.Joke) (string, error) {
	tx, err := jr.db.BeginTx(ctx, nil)

	if err != nil {
		return "", err
	}

	result, err := tx.ExecContext(ctx,
		"INSERT INTO jokes (author_id, text, explanation, lang) VALUES (?, ?, ?, ?)",
		joke.AuthorID, joke.Text, joke.Explanation, joke.Language)

	if err != nil {
		tx.Rollback()
		return "", err
	}

	var newID string

	newIDInt, err := result.LastInsertId()

	if err != nil {
		row := tx.QueryRowContext(ctx, "SELECT MAX(id) FROM jokes")

		err = row.Err()

		if err != nil {
			tx.Rollback()
			return "", err
		}

		row.Scan(&newID)
	} else {
		newID = fmt.Sprint(newIDInt)
	}

	tx.Commit()

	return newID, nil
}

func (jr *JokeCRUD) Update(ctx context.Context, id string, joke *data.Joke) (string, error) {
	result, err := jr.db.ExecContext(ctx,
		"UPDATE jokes SET author_id = ?, text = ?, explanation = ?, lang = ? WHERE id = ?",
		joke.AuthorID, joke.Text, joke.Explanation, joke.Language, id)

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

func (jr *JokeCRUD) Delete(ctx context.Context, id string) (string, error) {
	result, err := jr.db.ExecContext(ctx, "DELETE FROM jokes WHERE id = ?", id)

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
