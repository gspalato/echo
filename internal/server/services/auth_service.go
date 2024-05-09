package services

import (
	"context"
	"fmt"
	"os"

	"github.com/go-chi/jwtauth/v5"

	"unreal.sh/echo/internal/structures"
)

type AuthService struct {
	TokenAuth *jwtauth.JWTAuth

	dbService   *DatabaseService
	hashService *HashService
}

func (as *AuthService) Init(ctx context.Context, dbService *DatabaseService, hashService *HashService) {
	secret, found := os.LookupEnv("JWT_SECURITY_KEY")
	if !found {
		panic("No JWT security key found in environment.")
	}

	as.TokenAuth = jwtauth.New("HS256", []byte(secret), nil)

	as.dbService = dbService
	as.hashService = hashService
}

// Authenticate authenticates a user with the given username and password.
// It returns a profile on success, and an error on failure.
func (as *AuthService) Authenticate(username string, password string) (*structures.User, error) {
	user, err := as.dbService.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	match, err := as.hashService.ComparePasswordAndHash(user.PasswordHash, password)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, structures.ErrInvalidCredentials
	}

	return user, nil
}

// CreateAccount creates a new account with the given name, username, and password.
// It returns nil on success, and an error on failure.
func (as *AuthService) CreateAccount(name string, username string, password string) (structures.User, error) {
	_, err := as.dbService.GetUserByUsername(username)
	if err == nil {
		return structures.User{}, structures.ErrUserAlreadyExists
	}

	hash, err := as.hashService.HashPassword(password)
	if err != nil {
		return structures.User{}, err
	}

	fmt.Printf("Hashed password for user %s: %s...%s\n", username, string(hash[:5]), string(hash[len(hash)-5:]))

	user := structures.User{
		Name:         name,
		Username:     username,
		PasswordHash: hash,
		Credits:      0,
		Transactions: []structures.Transaction{},
		IsOperator:   false,
	}

	err = as.dbService.CreateUser(user)
	if err != nil {
		return structures.User{}, err
	}

	return user, nil
}

func (as *AuthService) VerifyToken(tokenString string) (*structures.User, error) {
	t, err := as.TokenAuth.Decode(tokenString)
	if err != nil {
		return nil, err
	}

	id, exists := t.Get("user_id")
	if !exists {
		return nil, structures.ErrInvalidTokenClaims
	}

	user, err := as.dbService.GetUserById(id.(string))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (as *AuthService) GenerateToken(u *structures.User) (string, error) {
	_, token, err := as.TokenAuth.Encode(
		map[string]interface{}{"user_id": u.Id})

	if err != nil {
		return "", err
	}

	return token, nil
}
