package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories"
)

type User struct {
	l          *log.Logger
	repo       repositories.UserCRUD
	vm         *middlewares.Validation
	am         *middlewares.Auth
	getUser    http.HandlerFunc
	getUsers   http.HandlerFunc
	insertUser http.HandlerFunc
	updateUser http.HandlerFunc
	deleteUser http.HandlerFunc
}

func NewUser(l *log.Logger, repo repositories.UserCRUD, vm *middlewares.Validation, am *middlewares.Auth) *User {
	u := &User{l: l, repo: repo, vm: vm, am: am}

	u.getUser = u.am.Auth(u.vm.IDOrEmailUserValidation(u.fetchOne), false)

	u.getUsers = u.am.Auth(middlewares.FetchAllQueryURL(u.fetchAll), true)

	u.insertUser = u.am.Auth(
		u.vm.DataValidation(
			middlewares.BCryptPassword(u.insert, u.l),
			middlewares.UserParamKey{}),
		true)

	u.updateUser = u.am.Auth(u.vm.OneIDURLValidation(u.vm.DataValidation(
		middlewares.BCryptPassword(u.update, u.l),
		middlewares.UserParamKey{}), middlewares.UserParamKey{}), false)

	u.deleteUser = u.am.Auth(u.vm.OneIDURLValidation(u.delete, middlewares.UserParamKey{}), true)

	return u
}

func (u User) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/users" || r.URL.Path == "/users/" {
			u.getUsers(w, r)
		} else {
			uCtx := context.WithValue(r.Context(), middlewares.UserParamKey{}, &data.User{})
			u.getUser(w, r.WithContext(uCtx))
		}
	case http.MethodPost:
		uCtx := context.WithValue(r.Context(), middlewares.UserParamKey{}, &data.User{})
		u.insertUser(w, r.WithContext(uCtx))

	case http.MethodPut:
		uCtx := context.WithValue(r.Context(), middlewares.UserParamKey{}, &data.User{})
		u.updateUser(w, r.WithContext(uCtx))

	case http.MethodDelete:
		uCtx := context.WithValue(r.Context(), middlewares.UserParamKey{}, &data.User{})
		u.deleteUser(w, r.WithContext(uCtx))
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (u *User) delete(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserParamKey{}).(*data.User)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	objectID, err := u.repo.Delete(r.Context(), user.ID)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown User ID", http.StatusNotFound)
			return
		}

		u.l.Println(err.Error())
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

func (u *User) fetchAll(w http.ResponseWriter, r *http.Request) {
	params, ok := r.Context().Value(middlewares.FetchQueryURLParamsKey{}).(*middlewares.FetchQueryURLParams)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users, cursorNext, err := u.repo.FetchAll(r.Context(), params.Limit, params.Offset, params.Direction)

	if err != nil {
		if err == repositories.ErrInvalidOffset {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			u.l.Println(err.Error())
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}
	}

	var qar data.QueryAllResponse
	qar.ResultCount = uint64(len(users))
	qar.CursorNext = cursorNext
	qar.Offset = params.Offset
	qar.Limit = params.Limit
	qar.Results = users

	err = qar.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (u *User) fetchOne(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserParamKey{}).(*data.User)
	auth, ok2 := r.Context().Value(middlewares.AuthParamsKey{}).(middlewares.AuthParams)

	if !ok || !ok2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var err error

	if user.ID != "" {
		user, err = u.repo.FetchOne(r.Context(), user.ID)
	} else {
		user, err = u.repo.FetchOneByEmail(r.Context(), user.Email)
	}

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown User ID", http.StatusNotFound)
			return
		}

		u.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	if !auth.Admin && user.ID != auth.ID {
		http.Error(w, "Unknown User ID", http.StatusNotFound)
		return
	}

	err = user.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (u *User) insert(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserParamKey{}).(*data.User)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := user.GenerateID()

	if err != nil {
		u.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	_, err = u.repo.Insert(r.Context(), user)

	if err != nil {
		u.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = user.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}

func (u *User) update(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserParamKey{}).(*data.User)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := u.repo.Update(r.Context(), user.ID, user)

	if err != nil {
		if err == repositories.ErrUnknownID {
			http.Error(w, "Unknown User ID", http.StatusNotFound)
			return
		}

		u.l.Println(err.Error())
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	err = user.ToJSON(w)

	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
	}
}
