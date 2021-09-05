package sql

import (
	"context"
	"database/sql"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/repositories"
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

func (jr *JokeCRUD) Delete(ctx context.Context, id string) (string, error) {
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

	_, err = tx.ExecContext(ctx, "DELETE FROM jokes WHERE id = ?", id)

	if err != nil {
		tx.Rollback()
		return "", err
	}

	tx.Commit()

	return id, nil
}

func (jr *JokeCRUD) FetchAll(ctx context.Context, limit uint64, offset string, direction repositories.FetchDirection) (data.Jokes, *string, error) {
	var jokes data.Jokes

	var condition string

	if direction == repositories.FetchBack {
		condition = "<"
	} else {
		condition = ">="
	}

	rows, err := jr.db.QueryContext(ctx, "SELECT id, author_id, text, explanation, lang FROM jokes WHERE id "+condition+" ? LIMIT ?", offset, limit+1)

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
	row := jr.db.QueryRowContext(ctx, "SELECT id, author_id, text, explanation, lang FROM jokes WHERE id = ?", id)
	joke := new(data.Joke)

	if err := row.Scan(&joke.ID, &joke.AuthorID, &joke.Text, &joke.Explanation, &joke.Language); err != nil {

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

	_, err = tx.ExecContext(ctx,
		"INSERT INTO jokes (id, author_id, text, explanation, lang) VALUES (?, ?, ?, ?, ?)",
		joke.ID, joke.AuthorID, joke.Text, joke.Explanation, joke.Language)

	if err != nil {
		tx.Rollback()

		return "", err
	}

	tx.Commit()

	return joke.ID, nil
}

func (jr *JokeCRUD) RateJoke(ctx context.Context, jokeID string, jokeRating *data.JokeRating) (string, error) {
	tx, err := jr.db.BeginTx(ctx, nil)

	if err != nil {
		return "", err
	}

	row := tx.QueryRowContext(ctx, "SELECT id FROM jokes WHERE id = ?", jokeID)
	err = row.Scan(&jokeID)

	if err != nil {
		tx.Rollback()

		if err == sql.ErrNoRows {
			return "", repositories.ErrUnknownID
		}

		return "", err
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO joke_ratings (id, rating, joke_id) VALUES (?, ?, ?)",
		jokeRating.ID, jokeRating.Rating, jokeID)

	if err != nil {
		tx.Rollback()

		return "", err
	}

	tx.Commit()

	return jokeRating.ID, nil
}

func (jr *JokeCRUD) Update(ctx context.Context, id string, joke *data.Joke) (string, error) {
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
		"UPDATE jokes SET author_id = ?, text = ?, explanation = ?, lang = ? WHERE id = ?",
		joke.AuthorID, joke.Text, joke.Explanation, joke.Language, id)

	if err != nil {
		tx.Rollback()
		return "", err
	}

	tx.Commit()

	return id, nil
}
