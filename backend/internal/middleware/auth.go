package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-oauth2/oauth2/v4/server"
)

type contextKey string

const userIDKey contextKey = "userID"

func OAuth2Guard(srv *server.Server) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenInfo, err := srv.ValidationBearerToken(r)
			if err != nil {
				http.Error(w, "invalid or missing token", http.StatusUnauthorized)
				return
			}
			userIDStr := tokenInfo.GetUserID()
			if userIDStr == "" {
				http.Error(w, "token missing user", http.StatusUnauthorized)
				return
			}
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid user id", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	val := ctx.Value(userIDKey)
	if val == nil {
		return 0, false
	}
	userID, ok := val.(int64)
	return userID, ok
}
