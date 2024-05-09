package routes

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/unrolled/render"

	"unreal.sh/echo/internal/server/services"
	"unreal.sh/echo/internal/structures"
	"unreal.sh/echo/internal/structures/payloads"
)

type MeHandler struct {
	r *render.Render

	dbService *services.DatabaseService
}

// GetProfile returns the profile of the currently authenticated user.
// It sends a GetEcobucksProfilePayload with Profile as nil if an error occurs.
func (mh *MeHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := claims["user_id"].(string)

	user, err := mh.dbService.GetUserById(userId)
	if err == structures.ErrNoUser {
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := payloads.GetEcobucksProfilePayload{Profile: user.ToProfile()}

	mh.r.JSON(w, http.StatusOK, payload)
}

func (mh *MeHandler) GetDisposals(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := claims["user_id"].(string)

	_, err := mh.dbService.GetUserById(userId)
	if err == structures.ErrNoUser {
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	disposalClaims, err := mh.dbService.GetDisposalsByUserId(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := payloads.GetUserDisposalsPayload{UserDisposals: disposalClaims}

	mh.r.JSON(w, http.StatusOK, payload)
}

func GetMeRouter(ctx context.Context, render *render.Render, db *services.DatabaseService) chi.Router {
	r := chi.NewRouter()

	meHandler := MeHandler{r: render}

	r.Get("/", meHandler.GetProfile)

	return r
}
