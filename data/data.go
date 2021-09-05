package data

import (
	"encoding/json"
	"errors"
	"io"
)

type IDValidation func(string) error

type Data interface {
	FromJSON(r io.Reader) error
	ToJSON(w io.Writer) error
	GenerateID() error
	CheckValidID(id string) error
	GetID() (string, error)
	SetID(string)
}

var ErrInvalidID = errors.New("invalid ID")
var ErrNoID = errors.New("no ID")

type QueryAllResponse struct {
	CursorNext  *string     `json:"cursor_next"`
	Limit       uint64      `json:"limit"`
	Offset      string      `json:"offset"`
	ResultCount uint64      `json:"result_count"`
	Results     interface{} `json:"results"`
}

func (qar *QueryAllResponse) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(qar)
}

func (qar *QueryAllResponse) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(qar)
}

type DeletedResponse struct {
	DeletedID interface{} `json:"deleted_id"`
}

func (del *DeletedResponse) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(del)
}

func (del *DeletedResponse) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(del)
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (tokr *TokenResponse) FromJSON(r io.Reader) error {
	return json.NewDecoder(r).Decode(tokr)
}

func (tokr *TokenResponse) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(tokr)
}
