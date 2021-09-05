package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/davq23/jokeapi/data"
	"github.com/davq23/jokeapi/middlewares"
	"github.com/davq23/jokeapi/repositories"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	am        *middlewares.Auth
	l         *log.Logger
	repo      repositories.UserCRUD
	vm        *middlewares.Validation
	loginUser http.HandlerFunc
}

func NewAuth(am *middlewares.Auth, l *log.Logger, repo repositories.UserCRUD, vm *middlewares.Validation) *Auth {
	au := &Auth{am: am, l: l, repo: repo, vm: vm}

	au.loginUser = au.vm.DataValidation(au.login, middlewares.UserParamKey{})

	return au
}

func (au *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		uCtx := context.WithValue(r.Context(), middlewares.UserParamKey{}, &data.User{})
		au.loginUser(w, r.WithContext(uCtx))
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (au *Auth) login(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middlewares.UserParamKey{}).(*data.User)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fetchUser, err := au.repo.FetchOneByEmail(r.Context(), user.Email)

	if err != nil {
		if err == repositories.ErrUnknownEmail {
			au.l.Println(err.Error())
			http.Error(w, "Invalid email or password", http.StatusBadRequest)
		} else {
			au.l.Println(err.Error())
			http.Error(w, "Database error", http.StatusInternalServerError)
		}

		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(fetchUser.Password), []byte(user.Password))

	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			au.l.Println(err.Error())
			http.Error(w, "Invalid email or password", http.StatusBadRequest)
		} else {
			au.l.Println(err.Error())
			http.Error(w, "Application error", http.StatusInternalServerError)
		}

		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    fetchUser.ID,
		"admin":      fetchUser.Admin,
		"exp":        time.Now().Add(time.Minute * 45).Unix(),
		"authorized": true,
	})

	res := data.TokenResponse{}

	if res.Token, err = token.SignedString(au.am.Secret); err != nil {
		au.l.Println(err.Error())
		http.Error(w, "Application error", http.StatusInternalServerError)
		return
	}

	if err = res.ToJSON(w); err != nil {
		au.l.Println(err.Error())
		http.Error(w, "Application error", http.StatusInternalServerError)
	}
}
