package data

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

type Joke struct {
	ID          string      `json:"joke_id" db:"id,omitempty" bson:"id,omitempty"`
	AuthorID    *string     `json:"author_id,omitempty" db:"author_id" bson:"author_id,omitempty"`
	Text        string      `json:"text" db:"text" validate:"required"`
	Explanation string      `json:"explanation,omitempty" db:"explanation,omitempty" bson:"explanation,omitempty"`
	Language    string      `json:"lang" db:"language" bson:"language" validate:"required"`
	Ratings     JokeRatings `json:"ratings,omitempty" bson:"ratings,omitempty"`
	AvgRating   *float64    `json:"avg_rating" bson:"avgRating"`
}

func (j *Joke) GetID() (string, error) {
	if j.ID == "" {
		return "", ErrNoID
	}

	if _, err := uuid.Parse(j.ID); err != nil {
		return j.ID, ErrInvalidID
	}

	return j.ID, nil
}

func (j *Joke) GenerateID() error {
	id, err := uuid.NewRandom()

	if err != nil {
		return err
	}

	j.ID = id.String()

	return nil
}

func (j *Joke) SetID(id string) {
	j.ID = id
}

func (j *Joke) CheckValidID(id string) error {
	_, err := uuid.Parse(id)

	return err
}

func (j *Joke) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(j)
}

func (j *Joke) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(j)
}

type Jokes []*Joke

func (j *Jokes) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(j)
}

func (j *Jokes) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(j)
}
