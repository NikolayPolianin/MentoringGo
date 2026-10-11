package http

import (
	"checkgo/OneService/internal/infrastructure/jwt"
	"checkgo/OneService/internal/transport/http/response"
	"context"
	"net/http"
)

func JWTMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				response.Error(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			claims, err := jwt.ParseToken(auth, secret)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "Authorization token invalid")
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claims.Role,
				claims.UserID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
