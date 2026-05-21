package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const AuthKey contextKey = "authenticated"

func Auth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/cities" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				key := strings.TrimPrefix(authHeader, "Bearer ")
				if key == apiKey {
					ctx := context.WithValue(r.Context(), AuthKey, true)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			if key := r.URL.Query().Get("key"); key == apiKey {
				ctx := context.WithValue(r.Context(), AuthKey, true)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"code": 401, "message": "Unauthorized"},
			})
		})
	}
}
