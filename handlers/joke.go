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
	vm   *middlewares.Validation
}

func NewJoke(l *log.Logger, repo repositories.JokeRepository, v *middlewares.Validation) *Joke {
	return &Joke{l, repo, v}
}

func (j *Joke) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/jokes" || r.URL.Path == "/jokes/" {
			middlewares.FetchAllQueryURL(j.fetchAll)(w, r)
		} else {
			j.vm.JokeIDURLValidation(j.fetchOne)(w, r)
		}

	case http.MethodPost:
		j.vm.JokeValidation(j.insert)(w, r)

	case http.MethodPut:
		j.vm.JokeIDURLValidation(j.vm.JokeValidation(j.update))(w, r)

	case http.MethodDelete:
		j.vm.JokeIDURLValidation(j.delete)(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (j *Joke) delete(w http.ResponseWriter, r *http.Request) {
	jokeID, ok := r.Context().Value(middlewares.JokeIDParamKey{}).(string)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Delete(r.Context(), jokeID)

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
	qar.Results = jokes

	err = qar.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *Joke) fetchOne(w http.ResponseWriter, r *http.Request) {
	jokeID, ok := r.Context().Value(middlewares.JokeIDParamKey{}).(string)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	joke, err := j.repo.FetchOne(r.Context(), jokeID)

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
	jokeID, okID := r.Context().Value(middlewares.JokeIDParamKey{}).(string)

	if !ok || !okID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Update(r.Context(), jokeID, joke)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown Joke ID", http.StatusNotFound)
			return
		}

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
