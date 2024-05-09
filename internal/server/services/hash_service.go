package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"unreal.sh/echo/internal/structures"
)

type HashService struct {
	params *HashParams
}

type HashParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func (hs *HashService) Init(ctx context.Context) {
	params := &HashParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}

	hs.params = params
}

func (hs *HashService) HashPassword(password string) (encodedHash string, err error) {
	salt, err := hs.generateRandomBytes(hs.params.saltLength)
	if err != nil {
		return "", err
	}

	passwordBytes := []byte(password)

	hash := argon2.IDKey(passwordBytes, salt, hs.params.iterations,
		hs.params.memory, hs.params.parallelism, hs.params.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version,
		hs.params.memory, hs.params.iterations, hs.params.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func (hs *HashService) generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (hs *HashService) ComparePasswordAndHash(encodedHash string, password string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := hs.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func (hs *HashService) decodeHash(encodedHash string) (params *HashParams, alt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, structures.ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, structures.ErrIncompatibleVersion
	}

	params = &HashParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &params.memory, &params.iterations, &params.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	params.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	params.keyLength = uint32(len(hash))

	return params, salt, hash, nil
}
