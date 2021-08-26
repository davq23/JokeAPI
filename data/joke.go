package data

import (
	"encoding/json"
	"io"
)

type Joke struct {
	ID          string  `json:"joke_id" db:"id,omitempty" bson:"_id,omitempty"`
	AuthorID    *string `json:"author_id,omitempty" db:"author_id" bson:"author_id,omitempty"`
	Text        string  `json:"text" db:"text" validate:"required"`
	Explanation string  `json:"explanation,omitempty" db:"explanation,omitempty" bson:"explanation,omitempty"`
	Language    string  `json:"lang" db:"language" bson:"language" validate:"required"`
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
