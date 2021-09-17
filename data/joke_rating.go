package data

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

type JokeRating struct {
	ID     string  `json:"rating_id" bson:"id,omitempty"`
	UserID *string `json:"user_id" bson:"user_id,omitempty"`
	Rating float64 `json:"rating" validate:"required,gt=0,lte=5" bson:"rating"`
}

func (jr *JokeRating) SetID(id string) {
	jr.ID = id
}

func (j *JokeRating) GetID() (string, error) {
	if j.ID == "" {
		return "", ErrNoID
	}

	if _, err := uuid.Parse(j.ID); err != nil {
		return j.ID, ErrInvalidID
	}

	return j.ID, nil
}

func (j *JokeRating) GenerateID() error {
	id, err := uuid.NewRandom()

	if err != nil {
		return err
	}

	j.ID = id.String()

	return nil
}

func (j *JokeRating) CheckValidID(id string) error {
	_, err := uuid.Parse(id)

	return err
}

func (j *JokeRating) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(j)
}

func (j *JokeRating) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(j)
}

type JokeRatings []*JokeRating

func (j *JokeRatings) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(j)
}

func (j *JokeRatings) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(j)
}
