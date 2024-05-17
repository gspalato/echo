package services

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt"
	"unreal.sh/echo/internal/structures"
)

type AuthService struct {
	secretKey *string

	dbService   *DatabaseService
	hashService *HashService
}

func (as *AuthService) Init(ctx context.Context, dbService *DatabaseService, hashService *HashService) error {
	secret, found := os.LookupEnv("JWT_SECURITY_KEY")
	if !found {
		return errors.New("missing JWT_SECURITY_KEY environment variable")
	}

	as.secretKey = &secret

	as.dbService = dbService
	as.hashService = hashService

	return nil
}

// Authenticate authenticates a user with the given username and password.
// It returns a profile on success, and an error on failure.
func (as *AuthService) Authenticate(username string, password string) (*structures.User, error) {
	fmt.Printf("Authenticating user %s...\n", username)

	user, err := as.dbService.GetUserByUsername(username)
	if err != nil {
		fmt.Println("User not found.")
		return nil, err
	}

	match, err := as.hashService.ComparePasswordAndHash(user.PasswordHash, password)
	if err != nil {
		fmt.Println("Error comparing password and hash.")
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

func (as *AuthService) GenerateToken(u *structures.User) (string, error) {
	claims := structures.UserClaims{
		UserId:         u.Id,
		StandardClaims: jwt.StandardClaims{},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(*as.secretKey))
}

func (as *AuthService) GenerateRefreshToken(claims jwt.StandardClaims) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return refreshToken.SignedString([]byte(*as.secretKey))
}

func (as *AuthService) ParseAccessToken(accessToken string) (*structures.User, *structures.UserClaims, error) {
	fmt.Println("ParseAccessToken reached.")

	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &structures.UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(*as.secretKey), nil
		})

	if err != nil {
		return nil, nil, err
	}

	if !parsedAccessToken.Valid {
		return nil, nil, structures.ErrInvalidToken
	}

	userClaims := parsedAccessToken.Claims.(*structures.UserClaims)
	id := userClaims.UserId

	fmt.Printf("User ID: %s\n", id)

	user, err := as.dbService.GetUserById(id)
	if err != nil {
		return nil, nil, structures.ErrNoUser
	}

	return user, userClaims, nil
}

func (as *AuthService) ParseRefreshToken(refreshToken string) *jwt.StandardClaims {
	parsedRefreshToken, _ := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(*as.secretKey), nil
		})

	return parsedRefreshToken.Claims.(*jwt.StandardClaims)
}

func (as *AuthService) IsAuthorized(token string) (bool, error) {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, structures.ErrInvalidToken
		}
		return []byte(*as.secretKey), nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
