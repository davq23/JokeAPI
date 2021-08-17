package middlewares

import (
	"context"
	"log"
	"net/http"

	"github.com/davq23/jokeapi/data"
	"github.com/go-playground/validator/v10"
)

type Validation struct {
	l *log.Logger
	v *validator.Validate
}

type JokeParamKey struct{}

func NewValidation(l *log.Logger, v *validator.Validate) *Validation {
	return &Validation{l, v}
}

func (vln *Validation) JokeValidation(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	joke := new(data.Joke)

	err := joke.FromJSON(r.Body)

	if err != nil {
		vln.l.Println(err.Error())
		http.Error(w, "Invalid payload", http.StatusUnprocessableEntity)
		return
	}

	err = vln.v.StructCtx(r.Context(), joke)

	if err != nil {
		vln.l.Println(err.Error())
		http.Error(w, "Invalid payload", http.StatusUnprocessableEntity)
		return
	}

	ctx := context.WithValue(r.Context(), JokeParamKey{}, joke)

	next(w, r.WithContext(ctx))
}
