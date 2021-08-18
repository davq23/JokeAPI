package middlewares

import (
	"context"
	"log"
	"net/http"
	"regexp"

	"github.com/davq23/jokeapi/data"
	"github.com/go-playground/validator/v10"
)

type Validation struct {
	l        *log.Logger
	v        *validator.Validate
	idRegexp *regexp.Regexp
}

type JokeIDParamKey struct{}
type JokeParamKey struct{}

func NewValidation(l *log.Logger, v *validator.Validate, idRegexp *regexp.Regexp) *Validation {
	return &Validation{l, v, idRegexp}
}

func (vln *Validation) JokeIDURLValidation(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		ids := vln.idRegexp.FindAllString(path, -1)

		if len(ids) != 1 || ids[0] == "" {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		if err := vln.v.VarCtx(r.Context(), ids[0], "joke_id"); err != nil {
			vln.l.Println(err.Error())
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), JokeIDParamKey{}, ids[0])))
	})
}

func (vln *Validation) JokeValidation(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}
