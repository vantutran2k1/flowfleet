package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type GeoStore struct {
	client *redis.Client
}

func NewGeoStore(client *redis.Client) *GeoStore {
	return &GeoStore{client: client}
}

func (r *GeoStore) FindNearestDrivers(ctx context.Context, lat, lng float64, radiusKm float64) ([]string, error) {
	locations, err := r.client.GeoSearch(ctx, "active_drivers", &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     radiusKm,
		RadiusUnit: "km",
		Sort:       "ASC",
		Count:      10,
	}).Result()
	if err != nil {
		return nil, err
	}

	drivers := make([]string, len(locations))
	for i, loc := range locations {
		drivers[i] = loc
	}

	return drivers, nil
}
