package inputs

import "unreal.sh/echo/internal/structures"

type RegisterEcobucksStationInput struct {
	Location structures.LocationClaim `json:"location"`
}
