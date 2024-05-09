package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/unrolled/render"

	"unreal.sh/echo/internal/server/services"
	"unreal.sh/echo/internal/structures"
	"unreal.sh/echo/internal/structures/inputs"
)

type StationsHandler struct {
	r               *render.Render
	authService     *services.AuthService
	dbService       *services.DatabaseService
	stationsService *services.StationsService
}

// GetStations returns a list of all registered stations.
// It returns a GetEcobucksStationsPayload with a list of LocationClaims.
func (sh *StationsHandler) GetStations(w http.ResponseWriter, r *http.Request) {
	sh.r.JSON(w, http.StatusOK, sh.stationsService.Locations)
}

func (sh *StationsHandler) RegisterStation(w http.ResponseWriter, r *http.Request) {
	// Get token, get user and verify if user has is_operator true

	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := claims["user_id"].(string)

	user, err := sh.dbService.GetUserById(userId)
	if err == structures.ErrNoUser {
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !user.IsOperator {
		http.Error(w, "User is not an operator.", http.StatusUnauthorized)
		return
	}

	// Parse station from request body
	var input inputs.RegisterEcobucksStationInput

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input.", http.StatusBadRequest)
		return
	}

	// Register station
	sh.stationsService.RegisterStation(input.Location)
}

func GetStationsRouter(ctx context.Context, render *render.Render,
	as *services.AuthService, ds *services.DatabaseService, ss *services.StationsService) chi.Router {
	r := chi.NewRouter()

	stationsHandler := StationsHandler{
		r:               render,
		authService:     as,
		dbService:       ds,
		stationsService: ss,
	}

	r.Get("/", stationsHandler.GetStations)
	r.Put("/", stationsHandler.RegisterStation)

	return r
}
