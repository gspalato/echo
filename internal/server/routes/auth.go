package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/unrolled/render"

	"unreal.sh/echo/internal/server/services"
	"unreal.sh/echo/internal/structures/inputs"
	"unreal.sh/echo/internal/structures/payloads"
)

type AuthHandler struct {
	r *render.Render

	authService *services.AuthService
}

// Authenticate authenticates a user with the given username and password.
// It receives an AuthenticationInput body, and returns an AuthenticationPayload.
// It returns a profile on success, and an error on failure.
func (ah *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	var input inputs.AuthenticationInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		fmt.Printf("Failed to decode input: %v\n", err)
		ah.r.JSON(w, http.StatusBadRequest, payloads.AuthenticationPayload{Error: "Invalid input."})
		return
	}

	user, err := ah.authService.Authenticate(input.Username, input.Password)
	if err != nil {
		fmt.Printf("Failed to authenticate: %v\n", err)
		ah.r.JSON(w, http.StatusUnauthorized, payloads.AuthenticationPayload{Error: "Invalid credentials."})
		return
	}

	token, err := ah.authService.GenerateToken(user)
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		ah.r.JSON(w, http.StatusInternalServerError, payloads.AuthenticationPayload{Error: "Failed to generate token."})
		return
	}

	payload := payloads.AuthenticationPayload{Token: token, User: user.ToProfile()}

	ah.r.JSON(w, http.StatusOK, payload)
}

// CreateAccount creates a new account with the given name, username, and password.
// It receives an AccountInput body, and returns an AuthenticationPayload.
func (ah *AuthHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var input inputs.CreateAccountInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		ah.r.JSON(w, http.StatusBadRequest, payloads.AuthenticationPayload{Error: "Invalid input."})
		return
	}

	user, err := ah.authService.CreateAccount(input.Name, input.Username, input.Password)
	if err != nil {
		ah.r.JSON(w, http.StatusInternalServerError, payloads.AuthenticationPayload{Error: "Failed to create account."})
		return
	}

	token, err := ah.authService.GenerateToken(&user)
	if err != nil {
		ah.r.JSON(w, http.StatusInternalServerError, payloads.AuthenticationPayload{Error: "Failed to generate token."})
		return
	}

	payload := payloads.AuthenticationPayload{Token: token, User: user.ToProfile()}

	ah.r.JSON(w, http.StatusOK, payload)
}

func GetAuthRouter(ctx context.Context, render *render.Render, as *services.AuthService) chi.Router {
	r := chi.NewRouter()

	authHandler := AuthHandler{r: render, authService: as}

	r.Post("/", authHandler.Authenticate)
	r.Put("/", authHandler.CreateAccount)

	return r
}
