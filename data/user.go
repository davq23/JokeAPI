package data

import (
	"encoding/json"
	"io"
)

type User struct {
	ID       string  `json:"user_id" bson:"_id,omitempty"`
	Email    string  `json:"email" validator:"required,email"`
	Password *string `json:"password,omitempty" validator:"password"`
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
