package middlewares

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/davq23/jokeapi/data"
	"github.com/go-playground/validator/v10"
)

type Validation struct {
	l *log.Logger
	v *validator.Validate
}

type IDParamKey struct{}
type MultipleIDParamKey struct{}
type JokeParamKey struct{}
type JokeRatingParamKey struct{}
type UserParamKey struct{}

func NewValidation(l *log.Logger, v *validator.Validate) *Validation {
	return &Validation{l, v}
}

func (vln *Validation) MultipleIDURLValidation(next http.HandlerFunc, ctxKey []interface{}, ds []data.Data) http.HandlerFunc {
	if len(ctxKey) != len(ds) {
		panic("Context Type must agree with context type")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids := strings.Split(r.URL.Path, "/")

		if len(ids) != len(ds) {
			vln.l.Println("IDs don't match validate funcs")
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var newContext context.Context

		for i := 0; i < len(ds); i++ {
			if err := ds[i].CheckValidID(ids[i]); err != nil {
				vln.l.Println(err.Error())
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			newContext = context.WithValue(r.Context(), ctxKey[i], ds)
		}

		next(w, r.WithContext(newContext))
	})

}

func (vln *Validation) IDOrEmailUserValidation(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		u, ok := r.Context().Value(UserParamKey{}).(*data.User)

		if !ok {
			http.Error(w, "Invalid ID or email", http.StatusBadRequest)
			return
		}

		begin := strings.LastIndex(path, "/")

		if begin == -1 || begin+1 >= len(path) {
			vln.l.Println(path)
			http.Error(w, "Invalid ID or email", http.StatusBadRequest)
			return
		}

		id := path[begin+1:]

		if err := u.CheckValidID(id); err != nil {
			if err = vln.v.Var(id, "email"); err != nil {
				http.Error(w, "Invalid ID or email", http.StatusBadRequest)
				return
			}
			u.Email = id
		} else {
			u.ID = id
		}

		next(w, r)
	})
}

func (vln *Validation) OneIDURLValidation(next http.HandlerFunc, ctxKey interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		d, ok := r.Context().Value(ctxKey).(data.Data)

		if !ok {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		begin := strings.LastIndex(path, "/")

		if begin == -1 || begin+1 >= len(path) {
			vln.l.Println(path)
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		id := path[begin+1:]

		if err := d.CheckValidID(id); err != nil {
			vln.l.Println(err.Error())
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		d.SetID(id)

		next(w, r)
	})
}

func (vln *Validation) DataValidation(next http.HandlerFunc, ctxKey interface{}) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, ok := r.Context().Value(ctxKey).(data.Data)

		if !ok {
			http.Error(w, "Invalid payload", http.StatusUnprocessableEntity)
			return
		}

		err := d.FromJSON(r.Body)

		if err != nil {
			vln.l.Println(err.Error())
			http.Error(w, "Invalid payload", http.StatusUnprocessableEntity)
			return
		}

		if _, err = d.GetID(); err != nil && err != data.ErrNoID {
			vln.l.Println(err.Error())
			http.Error(w, "Invalid payload", http.StatusUnprocessableEntity)
			return
		}

		err = vln.v.StructCtx(r.Context(), d)

		if err != nil {
			vln.l.Println(err.Error())
			http.Error(w, "Invalid payload", http.StatusUnprocessableEntity)
			return
		}

		next(w, r)
	})
}
