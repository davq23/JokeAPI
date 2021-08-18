package data

import (
	"encoding/json"
	"io"
)

type QueryAllResponse struct {
	CursorNext  *string     `json:"cursor_next"`
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
