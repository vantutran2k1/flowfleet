package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vantutran2k1/flowfleet/internal/adapter/storage/postgres"
	redis_adapter "github.com/vantutran2k1/flowfleet/internal/adapter/storage/redis"
	"github.com/vantutran2k1/flowfleet/internal/adapter/websocket"
)

type DispatchService struct {
	db   *pgxpool.Pool
	repo *postgres.Queries
	geo  *redis_adapter.GeoStore
	hub  *websocket.Hub
}

func NewDispatchService(db *pgxpool.Pool, geo *redis_adapter.GeoStore, hub *websocket.Hub) *DispatchService {
	return &DispatchService{
		db:   db,
		repo: postgres.New(db),
		geo:  geo,
		hub:  hub,
	}
}

func (s *DispatchService) CreateAndDispatchOrder(ctx context.Context, fleetID uuid.UUID, pickupLat, pickupLng float64) (uuid.UUID, error) {
	params := postgres.CreateOrderParams{
		FleetID:       fleetID,
		AmountCents:   1500,
		StMakepoint:   pickupLng,
		StMakepoint_2: pickupLat,
		StMakepoint_3: pickupLng,
		StMakepoint_4: pickupLat,
	}

	order, err := s.repo.CreateOrder(ctx, params)
	if err != nil {
		return uuid.Nil, err
	}

	candidateIDs, err := s.geo.FindNearestDrivers(ctx, pickupLat, pickupLng, 5.0)
	if err != nil {
		log.Println("redis error:", err)
		return order.ID, nil
	}

	var assignedDriverID string
	for _, dID := range candidateIDs {
		driverUUID, _ := uuid.Parse(dID)

		driver, err := s.repo.GetDriver(ctx, driverUUID)
		if err != nil {
			continue
		}

		if driver.Status == postgres.DriverStatusIdle {
			assignedDriverID = dID
			break
		}
	}

	if assignedDriverID == "" {
		return order.ID, errors.New("no available drivers found")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return order.ID, err
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	driverUUID, _ := uuid.Parse(assignedDriverID)

	if err := qtx.SetDriverStatus(ctx, postgres.SetDriverStatusParams{
		ID:     driverUUID,
		Status: postgres.DriverStatusEnRoute,
	}); err != nil {
		return order.ID, err
	}

	if err := qtx.AssignDriverToOrder(ctx, postgres.AssignDriverToOrderParams{
		DriverID: pgtype.UUID{Bytes: driverUUID, Valid: true},
		ID:       order.ID,
	}); err != nil {
		return order.ID, err
	}

	if err := tx.Commit(ctx); err != nil {
		return order.ID, err
	}

	s.hub.SendToDriver(assignedDriverID, map[string]any{
		"event":    "ORDER_ASSIGNED",
		"order_id": order.ID,
		"lat":      pickupLat,
		"lng":      pickupLng,
	})

	return order.ID, nil
}
