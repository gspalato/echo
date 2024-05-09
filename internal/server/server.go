package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/unrolled/render"

	"unreal.sh/echo/internal/server/routes"
	"unreal.sh/echo/internal/server/services"
)

func Start(ctx context.Context) {
	// Initialize database.
	dbService := services.DatabaseService{}
	dbService.Init(ctx)

	// Initialize services.
	hashService := services.HashService{}
	hashService.Init(ctx)

	authService := services.AuthService{}
	authService.Init(ctx, &dbService, &hashService)

	stationsService := services.StationsService{}
	stationsService.Init(ctx)

	r := chi.NewRouter()
	render := render.Render{}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(jwtauth.Verifier(authService.TokenAuth))
		r.Use(jwtauth.Authenticator(authService.TokenAuth))

		r.Mount("/me", routes.GetMeRouter(ctx, &render, &dbService))
		r.Mount("/stations", routes.GetStationsRouter(ctx, &render, &authService, &dbService, &stationsService))
	})

	r.Mount("/auth", routes.GetAuthRouter(ctx, &render, &authService))

	http.ListenAndServe(":4000", r)
}
