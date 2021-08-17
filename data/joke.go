package data

import (
	"encoding/json"
	"io"
)

type Joke struct {
	ID          string  `json:"joke_id" bson:"_id,omitempty"`
	AuthorID    *string `json:"author_id,omitempty" bson:"author_id,omitempty"`
	Text        string  `json:"text" validate:"required"`
	Explanation string  `json:"explanation"`
	Language    string  `json:"lang" validate:"required"`
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
