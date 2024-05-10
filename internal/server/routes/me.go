package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/unrolled/render"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"unreal.sh/echo/internal/server/middleware"
	"unreal.sh/echo/internal/server/services"
	"unreal.sh/echo/internal/structures"
	"unreal.sh/echo/internal/structures/inputs"
	"unreal.sh/echo/internal/structures/payloads"
	"unreal.sh/echo/internal/utils"
)

type MeHandler struct {
	r         *render.Render
	dbService *services.DatabaseService
}

// GetProfile returns the profile of the currently authenticated user.
// It sends a GetEcobucksProfilePayload with Profile as nil if an error occurs.
func (mh *MeHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*structures.User)

	payload := payloads.GetEcobucksProfilePayload{Profile: user.ToProfile()}

	mh.r.JSON(w, http.StatusOK, payload)
}

func (mh *MeHandler) GetDisposals(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*structures.User)

	disposalClaims, err := mh.dbService.GetDisposalsByUserId(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := payloads.GetUserDisposalsPayload{UserDisposals: disposalClaims}

	mh.r.JSON(w, http.StatusOK, payload)
}

func (mh *MeHandler) RegisterDisposal(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*structures.User)

	var input inputs.RegisterDisposalInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input.", http.StatusBadRequest)
		return
	}

	if !user.IsOperator {
		http.Error(w, "User is not an operator.", http.StatusUnauthorized)
		return
	}

	disposal := structures.DisposalClaim{
		OperatorId: user.Id,
		Token:      uuid.New().String(),
		IsClaimed:  false,
		Disposals:  input.Disposals,
	}

	disposal.Credits = utils.Sum(disposal.Disposals, func(d structures.Disposal) float32 { return d.Credits })
	disposal.Weight = utils.Sum(disposal.Disposals, func(d structures.Disposal) float32 { return d.Weight })

	err = mh.dbService.InsertDisposal(&disposal)
	if err != nil {
		fmt.Printf("Failed to insert disposal: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := payloads.RegisterDisposalPayload{Success: true, Disposal: disposal}

	mh.r.JSON(w, http.StatusOK, payload)
}

func (mh *MeHandler) ClaimDisposal(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*structures.User)

	var input inputs.ClaimDisposalInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input.", http.StatusBadRequest)
		return
	}

	disposal, err := mh.dbService.GetDisposalByToken(*input.DisposalToken)
	if err == structures.ErrNoDisposal {
		http.Error(w, "Disposal not found.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if disposal.IsClaimed || disposal.UserId != "" {
		http.Error(w, "Disposal already claimed.", http.StatusBadRequest)
		return
	}

	err = mh.dbService.UpdateDisposal(*input.DisposalToken, primitive.M{"is_claimed": true, "user_id": user.Id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	weight, unit := getLargestUnit(disposal.Weight)

	typeToName := map[structures.DisposalType]string{
		structures.RECYCLABLE: "Recyclable",
		structures.BATTERY:    "Battery",
		structures.SPONGE:     "Sponge",
		structures.ELECTRONIC: "Electronic",
	}

	disposalList := utils.Map(disposal.Disposals, func(d structures.Disposal, i int) string {
		if i == 1 {
			return " and more."
		} else if i > 1 {
			return ""
		}

		return typeToName[d.DisposalType]
	})

	transactionDescription := fmt.Sprintf("Disposed %.2f%s of %s", weight, unit, disposalList)

	transaction := structures.Transaction{
		TransactionType: structures.CLAIM,
		UserId:          user.Id,
		ClaimId:         disposal.Id,
		Credits:         disposal.Credits,
		Timestamp:       time.Now().Unix(),
		Description:     transactionDescription,
	}

	err = mh.dbService.LinkTransactionToUserById(&transaction, user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := payloads.ClaimDisposalPayload{Success: true, Disposal: disposal}

	mh.r.JSON(w, http.StatusOK, payload)
}

func GetMeRouter(ctx context.Context, render *render.Render, db *services.DatabaseService) chi.Router {
	r := chi.NewRouter()

	meHandler := MeHandler{r: render}

	r.Get("/", meHandler.GetProfile)
	r.Get("/disposals", meHandler.GetDisposals)

	r.Put("/disposals", meHandler.RegisterDisposal)
	r.Post("/disposals", meHandler.ClaimDisposal)

	return r
}

/* Utilities */

func getLargestUnit(grams float32) (float32, string) {
	if grams < 1000 {
		return grams, "g"
	} else if grams < 1000000 {
		return grams / 1000, "kg"
	} else {
		return grams / 1000000, "t"
	}
}
