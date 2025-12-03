package port

import "context"

type GeoFinder interface {
	FindNearestDrivers(ctx context.Context, lat, lng float64, radiusKm float64) ([]string, error)
}
