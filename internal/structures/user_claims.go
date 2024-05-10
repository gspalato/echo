package structures

import "github.com/golang-jwt/jwt"

type UserClaims struct {
	Name   string `json:"name"`
	UserId string `json:"user_id"`
	jwt.StandardClaims
}
