package payloads

import "unreal.sh/echo/internal/structures"

type AuthenticationPayload struct {
	Token string              `json:"token"`
	User  *structures.Profile `json:"user"`
	Error string              `json:"error"`
}
