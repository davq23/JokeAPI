package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories"
)

type JokeRating struct {
	l                *log.Logger
	repo             repositories.JokeCRUD
	vm               *middlewares.Validation
	am               *middlewares.Auth
	getJokeRatings   http.HandlerFunc
	rateJoke         http.HandlerFunc
	deleteJokeRating http.HandlerFunc
}

func NewJokeRating(l *log.Logger, repo repositories.JokeCRUD, v *middlewares.Validation, auth *middlewares.Auth) *JokeRating {
	jr := &JokeRating{l: l, repo: repo, vm: v, am: auth}

	jr.getJokeRatings = middlewares.FetchAllQueryURL(jr.vm.OneIDURLValidation(jr.get, middlewares.JokeParamKey{}))

	jr.rateJoke = jr.vm.OneIDURLValidation(
		jr.vm.DataValidation(jr.rate, middlewares.JokeRatingParamKey{}),
		middlewares.JokeParamKey{})

	jr.deleteJokeRating = jr.am.Auth(jr.vm.OneIDURLValidation(jr.delete,
		middlewares.JokeParamKey{}), false)

	return jr
}

func (jr *JokeRating) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		jkCtx := context.WithValue(r.Context(), middlewares.JokeRatingParamKey{}, &data.JokeRating{})
		jkCtx = context.WithValue(jkCtx, middlewares.JokeParamKey{}, &data.Joke{})
		jr.rateJoke(w, r.WithContext(jkCtx))
	case http.MethodDelete:
		jkCtx := context.WithValue(r.Context(), middlewares.JokeRatingParamKey{}, &data.JokeRating{})
		jkCtx = context.WithValue(jkCtx, middlewares.JokeParamKey{}, &data.Joke{})
		jr.deleteJokeRating(w, r.WithContext(jkCtx))
	}
}

func (jr *JokeRating) get(w http.ResponseWriter, r *http.Request) {
	jrating, okRating := r.Context().Value(middlewares.JokeRatingParamKey{}).(*data.JokeRating)
	j, okID := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)

	if !okRating || !okID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, _, err := jr.repo.FetchRatings(r.Context(), j.ID, 120, "", repositories.FetchNext)

	if err != nil {
		jr.l.Println(err.Error(), j.ID)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = jrating.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (jr *JokeRating) rate(w http.ResponseWriter, r *http.Request) {
	jrating, okRating := r.Context().Value(middlewares.JokeRatingParamKey{}).(*data.JokeRating)
	j, okID := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)
	auth, okAuth := r.Context().Value(middlewares.AuthParamsKey{}).(middlewares.AuthParams)

	if !okRating || !okID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if okAuth {
		jrating.UserID = &auth.ID
	}

	err := jrating.GenerateID()

	if err != nil {
		jr.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	_, err = jr.repo.RateJoke(r.Context(), j.ID, jrating)

	if err != nil {
		jr.l.Println(err.Error(), j.ID)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = jrating.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (jr *JokeRating) delete(w http.ResponseWriter, r *http.Request) {
	jrating, okRating := r.Context().Value(middlewares.JokeRatingParamKey{}).(*data.JokeRating)
	joke, okJoke := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)
	auth, okAuth := r.Context().Value(middlewares.AuthParamsKey{}).(middlewares.AuthParams)

	if !okRating || !okJoke || !okAuth {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jr.repo.DeleteRating(r.Context(), jrating.ID, joke.ID, auth.ID)
}
