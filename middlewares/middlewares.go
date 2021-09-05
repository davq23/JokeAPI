package middlewares

import (
	"log"
	"net/http"

	"github.com/davq23/jokeapi/data"
	"golang.org/x/crypto/bcrypt"
)

func BCryptPassword(next http.HandlerFunc, l *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserParamKey{}).(*data.User)

		if !ok {
			http.Error(w, "Invalid payload", http.StatusUnprocessableEntity)
			return
		}

		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

		if err != nil {
			l.Println(err.Error())
			http.Error(w, "Unknown error", http.StatusInternalServerError)
			return
		}

		user.Password = string(passwordBytes)

		next(w, r)
	})
}
