package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/unrolled/render"
	"go.uber.org/zap"

	"unreal.sh/echo/internal/server/middleware"
	"unreal.sh/echo/internal/server/routes"
	"unreal.sh/echo/internal/server/services"
)

func Start(ctx context.Context, logger *zap.SugaredLogger) {
	// Initialize database.
	dbService := services.DatabaseService{}
	dbService.Init(ctx)

	// Initialize services.
	hashService := services.HashService{}
	hashService.Init(ctx)

	authService := services.AuthService{}
	err := authService.Init(ctx, &dbService, &hashService)
	if err != nil {
		panic("Failed to initialize auth service: " + err.Error())
	}

	userService := services.UserService{}
	err = userService.Init(ctx)
	if err != nil {
		panic("Failed to initialize auth service: " + err.Error())
	}

	stationsService := services.StationsService{}
	stationsService.Init(ctx)

	r := chi.NewRouter()
	render := render.Render{}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Group(func(r chi.Router) {
		r.Use(chiMiddleware.Logger)
		r.Use(middleware.ValidateToken(&authService))
		r.Use(middleware.RequireAuthentication(&authService))

		r.Mount("/me", routes.GetMeRouter(ctx, &render, &userService, &dbService))
		r.Mount("/stations", routes.GetStationsRouter(ctx, &render, &stationsService))
	})

	r.Mount("/auth", routes.GetAuthRouter(ctx, &render, &authService))

	http.ListenAndServe(":4000", r)

	logger.Info("Server started on port :4000.")
}
