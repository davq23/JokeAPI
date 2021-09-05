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
	Secret []byte
}

func NewAuth(l *log.Logger, secret []byte) *Auth {
	return &Auth{
		l:      l,
		Secret: secret,
	}
}

type AuthParamsKey struct{}

type AuthParams struct {
	ID    string
	Admin bool
}

func (au *Auth) Auth(next http.HandlerFunc, admin bool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		authParts := strings.Split(authorization, " ")

		if len(authParts) != 2 || authParts[0] != "Bearer" {
			http.Error(w, "Invalid authorization", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(authParts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid token")
			}

			return au.Secret, nil
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

		adminClaims, okAdmin := claims["admin"].(bool)
		uid, okUid := claims["user_id"].(string)

		if !okUid || !okAdmin {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if admin && !adminClaims {
			http.Error(w, "Unauthorized access", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AuthParamsKey{}, AuthParams{
			ID:    uid,
			Admin: adminClaims,
		})

		next(w, r.WithContext(ctx))
	})
}
