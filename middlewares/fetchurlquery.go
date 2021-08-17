package middlewares

import (
	"context"
	"net/http"
	"strconv"

	"github.com/davq23/jokeapi/repositories"
)

type FetchQueryURLParamsKey struct{}

type FetchQueryURLParams struct {
	Offset    string
	Limit     uint64
	Direction repositories.FetchDirection
}

func FetchAllQueryURL(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		limitString := r.URL.Query().Get("limit")
		directionString := r.URL.Query().Get("direction")

		var direction repositories.FetchDirection

		limit, err := strconv.ParseUint(limitString, 10, 64)

		if err != nil {
			http.Error(w, "Invalid limit type", http.StatusBadRequest)
		}

		switch directionString {
		case "next":
			direction = repositories.FetchNext
		case "last":
			direction = repositories.FetchBack
		default:
			direction = repositories.FetchNext
		}

		ctx := context.WithValue(r.Context(), FetchQueryURLParamsKey{}, &FetchQueryURLParams{
			Offset:    offset,
			Limit:     limit,
			Direction: direction,
		})

		next(w, r.WithContext(ctx))
	})
}
