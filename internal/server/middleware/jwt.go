package middleware

import (
	"context"
	"net/http"

	"unreal.sh/echo/internal/server/services"
)

type MiddlewareContextKey string

const (
	TokenContextKey  MiddlewareContextKey = "token"
	UserContextKey   MiddlewareContextKey = "user"
	ClaimsContextKey MiddlewareContextKey = "claims"
)

func ValidateToken(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// Get the JWT token from the Authorization header
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(rw, "No token provided", http.StatusUnauthorized)
				return
			}

			// Validate the token
			// If the token is invalid, return an error
			// If the token is valid, set the user in the context and call the next handler
			authorized, err := authService.IsAuthorized(token)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusUnauthorized)
				return
			}

			if !authorized {
				http.Error(rw, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TokenContextKey, token)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

func RequireAuthentication(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// Get the JWT token from the "token" context value.
			token := r.Context().Value("token").(string)

			// Get the user from the token
			user, claims, err := authService.ParseAccessToken(token)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, ClaimsContextKey, claims)

			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}
