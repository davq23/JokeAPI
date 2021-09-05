package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories"
)

type Joke struct {
	l          *log.Logger
	repo       repositories.JokeCRUD
	vm         *middlewares.Validation
	am         *middlewares.Auth
	getJoke    http.HandlerFunc
	getJokes   http.HandlerFunc
	insertJoke http.HandlerFunc
	updateJoke http.HandlerFunc
	deleteJoke http.HandlerFunc
}

func NewJoke(l *log.Logger, repo repositories.JokeCRUD, v *middlewares.Validation, auth *middlewares.Auth) *Joke {
	j := &Joke{l: l, repo: repo, vm: v, am: auth}

	j.getJoke = j.vm.OneIDURLValidation(j.fetchOne, middlewares.JokeParamKey{})
	j.getJokes = middlewares.FetchAllQueryURL(j.fetchAll)

	j.insertJoke = j.am.Auth(j.vm.DataValidation(j.insert, middlewares.JokeParamKey{}), false)
	j.updateJoke = j.am.Auth(
		j.vm.OneIDURLValidation(
			j.vm.DataValidation(j.update, middlewares.JokeParamKey{}), middlewares.JokeParamKey{}),
		true)
	j.deleteJoke = j.am.Auth(j.vm.OneIDURLValidation(j.delete, middlewares.JokeParamKey{}), false)

	return j
}

func (j *Joke) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/jokes" || r.URL.Path == "/jokes/" {
			j.getJokes(w, r)
		} else {
			jkCtx := context.WithValue(r.Context(), middlewares.JokeParamKey{}, &data.Joke{})
			j.getJoke(w, r.WithContext(jkCtx))
		}

	case http.MethodPost:
		jkCtx := context.WithValue(r.Context(), middlewares.JokeParamKey{}, &data.Joke{})
		j.insertJoke(w, r.WithContext(jkCtx))

	case http.MethodPut:
		jkCtx := context.WithValue(r.Context(), middlewares.JokeParamKey{}, &data.Joke{})
		j.updateJoke(w, r.WithContext(jkCtx))

	case http.MethodDelete:
		jkCtx := context.WithValue(r.Context(), middlewares.JokeParamKey{}, &data.Joke{})
		j.deleteJoke(w, r.WithContext(jkCtx))
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (j *Joke) delete(w http.ResponseWriter, r *http.Request) {
	joke, ok := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Delete(r.Context(), joke.ID)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown Joke ID", http.StatusNotFound)
			return
		}

		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	result := data.DeletedResponse{
		DeletedID: objectID,
	}

	err = result.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *Joke) fetchAll(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(middlewares.FetchQueryURLParamsKey{}).(*middlewares.FetchQueryURLParams)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jokes, cursorNext, err := j.repo.FetchAll(r.Context(), params.Limit, params.Offset, params.Direction)

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

	var qar data.QueryAllResponse
	qar.ResultCount = uint64(len(jokes))
	qar.CursorNext = cursorNext
	qar.Offset = params.Offset
	qar.Limit = params.Limit
	qar.Results = jokes

	err = qar.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *Joke) fetchOne(w http.ResponseWriter, r *http.Request) {
	joke, ok := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	joke, err := j.repo.FetchOne(r.Context(), joke.ID)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown Joke ID", http.StatusNotFound)
			return
		}

		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = joke.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *Joke) insert(w http.ResponseWriter, r *http.Request) {
	joke, ok := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := joke.GenerateID()

	if err != nil {
		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	_, err = j.repo.Insert(r.Context(), joke)

	if err != nil {
		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = joke.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *Joke) update(w http.ResponseWriter, r *http.Request) {
	joke, ok := r.Context().Value(middlewares.JokeParamKey{}).(*data.Joke)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := j.repo.Update(r.Context(), joke.ID, joke)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown Joke ID", http.StatusNotFound)
			return
		}

		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = joke.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}
