package payloads

import "unreal.sh/echo/internal/structures"

type GetEcobucksProfilePayload struct {
	Profile *structures.Profile `json:"profile"`
}
