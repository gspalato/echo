package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/unrolled/render"
	"go.mongodb.org/mongo-driver/bson"

	"unreal.sh/echo/internal/server/middleware"
	"unreal.sh/echo/internal/server/services"
	"unreal.sh/echo/internal/structures"
	"unreal.sh/echo/internal/structures/inputs"
	"unreal.sh/echo/internal/structures/payloads"
	"unreal.sh/echo/internal/utils"
)

type MeHandler struct {
	r *render.Render

	dbService   *services.DatabaseService
	userService *services.UserService
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
		fmt.Printf("Failed to decode input: %v\n", err)
		http.Error(w, "Invalid input.", http.StatusBadRequest)
		return
	}

	disposal, err := mh.dbService.GetDisposalByToken(input.DisposalToken)
	if err == structures.ErrNoDisposal {
		fmt.Printf("Disposal not found: %v\n", input.DisposalToken)
		http.Error(w, "Disposal not found.", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Printf("Failed to get disposal: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if disposal.IsClaimed || disposal.UserId != "" {
		fmt.Printf("Disposal already claimed: %v\n", disposal.Token)
		http.Error(w, "Disposal already claimed.", http.StatusBadRequest)
		return
	}

	err = mh.dbService.UpdateDisposal(input.DisposalToken, bson.M{"$set": bson.M{"is_claimed": true, "user_id": user.Id}})
	if err != nil {
		fmt.Printf("Failed to update disposal: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = mh.dbService.UpdateUserById(user.Id, bson.M{"$inc": bson.M{"credits": disposal.Credits}})
	if err != nil {
		fmt.Printf("Failed to update user credits: %v\n", err)
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
		fmt.Printf("Failed to link transaction to user: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload := payloads.ClaimDisposalPayload{Success: true, Disposal: disposal}

	mh.r.JSON(w, http.StatusOK, payload)
}

func (mh *MeHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*structures.User)

	avatarUrl, err := getAvatarUrl(user.Id)
	if err != nil {
		mh.r.JSON(w, http.StatusInternalServerError, payloads.GetAvatarPayload{
			Success: false,
			Error:   "Failed to get avatar URL.",
		})
		return
	}

	mh.r.JSON(w, http.StatusOK, payloads.GetAvatarPayload{
		Success:   true,
		AvatarUrl: avatarUrl,
	})
}

func (mh *MeHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*structures.User)

	const maxUploadSize = 5 * 1024 * 1024 // 5MB

	r.ParseMultipartForm(maxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Printf("Error retrieving the file. %s\n", err.Error())
		http.Error(w, "Error retrieving the file.", http.StatusBadRequest)
		return
	}

	defer file.Close()

	allowedMimeTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	if !slices.Contains(allowedMimeTypes, header.Header.Get("Content-Type")) {
		fmt.Println("Invalid file type.")
		mh.r.JSON(w, http.StatusBadRequest, payloads.UploadAvatarPayload{
			Success: false,
			Error:   "Invalid file type.",
		})
		return
	}

	bucket, found := os.LookupEnv("AWS_AVATAR_S3_BUCKET")
	if !found {
		fmt.Println("Missing bucket name.")
		mh.r.JSON(w, http.StatusInternalServerError, payloads.UploadAvatarPayload{
			Success: false,
			Error:   "Missing bucket name.",
		})
		return
	}

	_, err = mh.userService.S3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &user.Id,
		Body:   file,
	})

	if err != nil {
		fmt.Printf("Failed to upload file. %s\n", err.Error())
		mh.r.JSON(w, http.StatusInternalServerError, payloads.UploadAvatarPayload{
			Success: false,
			Error:   "Failed to upload file.",
		})
		return
	}

	mh.r.JSON(w, http.StatusOK, payloads.UploadAvatarPayload{Success: true})
}

func GetMeRouter(ctx context.Context, render *render.Render, us *services.UserService, db *services.DatabaseService) chi.Router {
	r := chi.NewRouter()

	meHandler := MeHandler{r: render, userService: us, dbService: db}

	r.Get("/", meHandler.GetProfile)

	r.Get("/avatar", meHandler.GetAvatar)
	r.Put("/avatar", meHandler.UploadAvatar)

	r.Get("/disposals", meHandler.GetDisposals)
	r.Put("/disposals", meHandler.RegisterDisposal)
	r.Post("/disposals", meHandler.ClaimDisposal)

	return r
}

/* Utilities */

func getAvatarUrl(userID string) (string, error) {
	format, found := os.LookupEnv("AWS_AVATAR_URL_FORMAT")
	if !found {
		return "", errors.New("missing AWS_AVATAR_URL_FORMAT")
	}

	bucket, found := os.LookupEnv("AWS_AVATAR_S3_BUCKET")
	if !found {
		return "", errors.New("missing AWS_AVATAR_S3_BUCKET")
	}

	return fmt.Sprintf(format, bucket, userID), nil
}

func getLargestUnit(grams float32) (float32, string) {
	if grams < 1000 {
		return grams, "g"
	} else if grams < 1000000 {
		return grams / 1000, "kg"
	} else {
		return grams / 1000000, "t"
	}
}
