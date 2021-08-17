package handlers

import (
	"log"
	"net/http"

	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories"
)

type Joke struct {
	l    *log.Logger
	repo repositories.JokeRepository
}

func NewJoke(l *log.Logger, repo repositories.JokeRepository) *Joke {
	return &Joke{l, repo}
}

func (j *Joke) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		middlewares.FetchAllQueryURL(j.fetchAll)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (j *Joke) fetchAll(w http.ResponseWriter, r *http.Request) {
	params := r.Context().Value(middlewares.FetchQueryURLParamsKey{}).(*middlewares.FetchQueryURLParams)

	jokes, _, err := j.repo.FetchAll(r.Context(), params.Offset, params.Limit, params.Direction)

	if err != nil {
		if err == repositories.ErrInvalidOffset {
			http.Error(w, "Invalid offset ID", http.StatusBadRequest)
			return
		}
	}

	err = jokes.ToJSON(w)

	if err != nil {
		http.Error(w, "Unknown error", http.StatusInternalServerError)
	}
}
