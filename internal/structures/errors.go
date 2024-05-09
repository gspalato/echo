package structures

import "errors"

var (
	// ErrInvalidCredentials is returned when the credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrNoUser is returned when the user is not found
	ErrNoUser = errors.New("user not found")

	// ErrInvalidPassword is returned when the password is invalid
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidPasswordLength is returned when the password's length is invalid (0 or > 72)
	ErrInvalidPasswordLength = errors.New("invalid password length")

	// ErrInvalidTokenClaims is returned when the token claims are invalid
	ErrInvalidTokenClaims = errors.New("invalid token claims")

	// ErrInvalidDatabaseId is returned when the database id is invalid
	ErrInvalidDatabaseId = errors.New("invalid id")

	// ErrFailedToCreateUser is returned when the user cannot be created
	ErrFailedToCreateUser = errors.New("failed to create user")

	// ErrFailedToCreateToken is returned when the token cannot be created
	ErrFailedToCreateToken = errors.New("failed to create token")

	// ErrFailedToCreateDisposal is returned when the disposal cannot be created
	ErrFailedToCreateDisposal = errors.New("failed to create disposal")

	// ErrInvalidHash is returned when the Argon2 hash is invalid
	ErrInvalidHash = errors.New("invalid hash")

	// ErrIncompatibleVersion is returned when the version is incompatible
	ErrIncompatibleVersion = errors.New("incompatible version")
)
