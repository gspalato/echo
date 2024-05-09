package payloads

import "unreal.sh/echo/internal/structures"

type GetEcobucksStationsPayload struct {
	Stations []structures.LocationClaim `json:"stations"`
}
