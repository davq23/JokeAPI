package middlewares

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type Auth struct {
	l      *log.Logger
	secret []byte
}

func NewAuth(l *log.Logger, secret []byte) *Auth {
	return &Auth{
		l:      l,
		secret: secret,
	}
}

type AuthParamsKey struct{}

type AuthParams struct {
	ID    string
	Email string
	Admin bool
}

func (au *Auth) Auth(next http.HandlerFunc, admin bool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		authParts := strings.Split(authorization, " ")

		if len(authParts) != 2 || authParts[0] == "Bearer" {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(authParts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid token")
			}

			return au.secret, nil
		})

		if err != nil {
			au.l.Println(err.Error())
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		if err = claims.Valid(); err != nil {
			au.l.Println(err.Error())
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		data := claims["data"].(map[string]interface{})

		if admin && !data["admin"].(bool) {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AuthParamsKey{}, AuthParams{
			ID:    data["email"].(string),
			Email: data["id"].(string),
		})

		next(w, r.WithContext(ctx))
	})
}
