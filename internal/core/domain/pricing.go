package domain

import (
	"context"
	"time"
)

type VehicleType string

const (
	VehicleBike  VehicleType = "BIKE"
	VehicleVan   VehicleType = "VAN"
	VehicleTruck VehicleType = "TRUCK"
)

type PricingInput struct {
	DistanceMeters float64
	Vehicle        VehicleType
	Time           time.Time
}

type PricingStrategy interface {
	CalculatePrice(ctx context.Context, input PricingInput) (int, error)
}
