package services

import (
	"context"
	"time"

	"unreal.sh/echo/internal/structures"
)

type StationsService struct {
	Locations []structures.LocationClaim
}

func (ss *StationsService) Init(ctx context.Context) {
	ss.Locations = make([]structures.LocationClaim, 0)
}

func (ss *StationsService) RegisterStation(station structures.LocationClaim) {
	ss.Locations = append(ss.Locations, station)
	time.AfterFunc(5*time.Minute, func() {
		ss.Locations = ss.Locations[1:]
	})
}
