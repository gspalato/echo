package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/unrolled/render"

	"unreal.sh/echo/internal/server/middleware"
	"unreal.sh/echo/internal/server/services"
	"unreal.sh/echo/internal/structures"
	"unreal.sh/echo/internal/structures/inputs"
)

type StationsHandler struct {
	r               *render.Render
	stationsService *services.StationsService
}

// GetStations returns a list of all registered stations.
// It returns a GetEcobucksStationsPayload with a list of LocationClaims.
func (sh *StationsHandler) GetStations(w http.ResponseWriter, r *http.Request) {
	sh.r.JSON(w, http.StatusOK, sh.stationsService.Locations)
}

func (sh *StationsHandler) RegisterStation(w http.ResponseWriter, r *http.Request) {
	// Get token, get user and verify if user has is_operator true
	user := r.Context().Value(middleware.UserContextKey).(*structures.User)

	if !user.IsOperator {
		http.Error(w, "User is not an operator.", http.StatusUnauthorized)
		return
	}

	// Parse station from request body
	var input inputs.RegisterEcobucksStationInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input.", http.StatusBadRequest)
		return
	}

	// Register station
	sh.stationsService.RegisterStation(input.Location)
}

func GetStationsRouter(ctx context.Context, render *render.Render,
	ss *services.StationsService) chi.Router {
	r := chi.NewRouter()

	stationsHandler := StationsHandler{
		r:               render,
		stationsService: ss,
	}

	r.Get("/", stationsHandler.GetStations)
	r.Put("/", stationsHandler.RegisterStation)

	return r
}
