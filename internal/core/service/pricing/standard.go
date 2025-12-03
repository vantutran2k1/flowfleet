package pricing

import (
	"context"
	"errors"

	"github.com/vantutran2k1/flowfleet/internal/core/domain"
)

var (
	BaseRates = map[domain.VehicleType]int{
		domain.VehicleBike:  500,
		domain.VehicleVan:   1500,
		domain.VehicleTruck: 3000,
	}
	PerKmRates = map[domain.VehicleType]int{
		domain.VehicleBike:  50,
		domain.VehicleVan:   100,
		domain.VehicleTruck: 200,
	}
)

type StandardStrategy struct{}

func NewStandardStrategy() *StandardStrategy {
	return &StandardStrategy{}
}

func (s *StandardStrategy) CalculatePrice(ctx context.Context, input domain.PricingInput) (int, error) {
	base, ok := BaseRates[input.Vehicle]
	if !ok {
		return 0, errors.New("unsupported vehicle type")
	}

	ratePerKm, _ := PerKmRates[input.Vehicle]

	distanceKM := input.DistanceMeters / 1000.0
	variable := int(distanceKM * float64(ratePerKm))

	total := base + variable

	if total < base {
		total = base
	}

	return total, nil
}
