package data

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

type User struct {
	ID       string `json:"user_id" bson:"id,omitempty"`
	Email    string `json:"email" validate:"required,email" bson:"email"`
	Password string `json:"password,omitempty" validate:"required,password" bson:"password"`
	Admin    bool   `json:"admin" bson:"admin"`
}

func (u *User) SetID(id string) {
	u.ID = id
}

func (u *User) GetID() (string, error) {
	if u.ID == "" {
		return "", ErrNoID
	}

	if _, err := uuid.Parse(u.ID); err != nil {
		return u.ID, ErrInvalidID
	}

	return u.ID, nil
}

func (u *User) GenerateID() error {
	id, err := uuid.NewRandom()

	if err != nil {
		return err
	}

	u.ID = id.String()

	return nil
}

func (u *User) CheckValidID(id string) error {
	_, err := uuid.Parse(id)

	return err
}

func (j *User) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(j)
}

func (j *User) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(j)
}

type Users []*User

func (j *Users) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(j)
}

func (j *Users) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(j)
}
