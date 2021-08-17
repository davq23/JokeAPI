package handlers

import (
	"log"
	"net/http"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories"
)

type Joke struct {
	l    *log.Logger
	repo repositories.JokeRepository
	v    *middlewares.Validation
}

func NewJoke(l *log.Logger, repo repositories.JokeRepository, v *middlewares.Validation) *Joke {
	return &Joke{l, repo, v}
}

func (j *Joke) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		middlewares.FetchAllQueryURL(w, r, j.fetchAll)
	case http.MethodPost:
		j.v.JokeValidation(w, r, j.insert)
	case http.MethodPut:
		j.v.JokeValidation(w, r, j.update)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (j *Joke) fetchAll(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(middlewares.FetchQueryURLParamsKey{}).(*middlewares.FetchQueryURLParams)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jokes, _, err := j.repo.FetchAll(r.Context(), params.Limit, params.Offset, params.Direction)

	if err != nil {
		if err == repositories.ErrInvalidOffset {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			j.l.Println(err.Error())
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}
	}

	err = jokes.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *Joke) insert(w http.ResponseWriter, r *http.Request) {
	joke, ok := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)

	joke.ID = ""

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Insert(r.Context(), joke)

	if err != nil {
		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	joke.ID = objectID

	err = joke.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *Joke) update(w http.ResponseWriter, r *http.Request) {
	joke, ok := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)
	jokeID := joke.ID
	joke.ID = ""

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Update(r.Context(), jokeID, joke)

	if err != nil {
		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	joke.ID = objectID

	err = joke.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}
