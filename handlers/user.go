package handlers

import (
	"log"
	"net/http"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories"
)

type User struct {
	l    *log.Logger
	repo repositories.UserCRUD
	vm   *middlewares.Validation
	am   *middlewares.Auth
}

func NewUser(l *log.Logger, repo repositories.UserCRUD, vm *middlewares.Validation, am *middlewares.Auth) *User {
	return &User{l, repo, vm, am}
}

func (u User) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/users" || r.URL.Path == "/users/" {
			u.am.Auth(middlewares.FetchAllQueryURL(u.fetchAll), true)(w, r)
		} else {
			u.vm.IDURLValidation(u.fetchOne, "user_id")(w, r)
		}
	case http.MethodPost:
		u.vm.UserValidation(u.insert)(w, r)
	case http.MethodPut:
		u.vm.IDURLValidation(u.vm.UserValidation(u.update), "user_id")(w, r)
	case http.MethodDelete:
		u.vm.IDURLValidation(u.delete, "user_id")(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (j *User) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.IDParamKey{}).(string)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Delete(r.Context(), userID)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown User ID", http.StatusNotFound)
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

func (j *User) fetchAll(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(middlewares.FetchQueryURLParamsKey{}).(*middlewares.FetchQueryURLParams)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users, cursorNext, err := j.repo.FetchAll(r.Context(), params.Limit, params.Offset, params.Direction)

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
	qar.ResultCount = uint64(len(users))
	qar.CursorNext = cursorNext
	qar.Results = users

	err = qar.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *User) fetchOne(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.IDParamKey{}).(string)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := j.repo.FetchOne(r.Context(), userID)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown User ID", http.StatusNotFound)
			return
		}

		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = user.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *User) insert(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserParamKey{}).(*data.User)

	user.ID = ""

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Insert(r.Context(), user)

	if err != nil {
		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	user.ID = objectID

	err = user.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (j *User) update(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserParamKey{}).(*data.User)
	userID, okID := r.Context().Value(middlewares.IDParamKey{}).(string)

	if !ok || !okID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := j.repo.Update(r.Context(), userID, user)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown User ID", http.StatusNotFound)
			return
		}

		j.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	user.ID = objectID

	err = user.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}
