package structures

import "time"

type LocationClaim struct {
	Latitude float32 `json:"latitude"`

	Longitude float32 `json:"longitude"`

	Timestamp int64 `json:"timestamp"`

	StationId string `json:"station_id"`

	Age time.Duration `json:"age"`
}
