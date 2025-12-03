package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vantutran2k1/flowfleet/internal/adapter/storage/postgres"
	redis_adapter "github.com/vantutran2k1/flowfleet/internal/adapter/storage/redis"
	"github.com/vantutran2k1/flowfleet/internal/adapter/websocket"
	"github.com/vantutran2k1/flowfleet/internal/core/domain"
	"github.com/vantutran2k1/flowfleet/internal/core/service/pricing"
	"github.com/vantutran2k1/flowfleet/internal/pkg/geo"
)

type DispatchService struct {
	store  postgres.Store
	geo    *redis_adapter.GeoStore
	hub    *websocket.Hub
	pricer domain.PricingStrategy
}

func NewDispatchService(store postgres.Store, geo *redis_adapter.GeoStore, hub *websocket.Hub) *DispatchService {
	return &DispatchService{
		store:  store,
		geo:    geo,
		hub:    hub,
		pricer: pricing.NewStandardStrategy(),
	}
}

func (s *DispatchService) CreateAndDispatchOrder(ctx context.Context, fleetID uuid.UUID, pickupLat, pickupLng, dropoffLat, dropoffLng float64) (uuid.UUID, error) {
	distMeters := geo.CalculateDistance(pickupLat, pickupLng, dropoffLat, dropoffLng)

	priceCents, err := s.pricer.CalculatePrice(ctx, domain.PricingInput{
		DistanceMeters: distMeters,
		Vehicle:        domain.VehicleBike,
		Time:           time.Now(),
	})
	if err != nil {
		return uuid.Nil, err
	}

	params := postgres.CreateOrderParams{
		FleetID:       fleetID,
		AmountCents:   int32(priceCents),
		StMakepoint:   pickupLng,
		StMakepoint_2: pickupLat,
		StMakepoint_3: pickupLng,
		StMakepoint_4: pickupLat,
	}

	order, err := s.store.CreateOrder(ctx, params)
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

		driver, err := s.store.GetDriver(ctx, driverUUID)
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

	if err := s.store.ExecTx(ctx, func(q postgres.Querier) error {
		driverUUID, _ := uuid.Parse(assignedDriverID)

		if err := q.SetDriverStatus(ctx, postgres.SetDriverStatusParams{
			ID:     driverUUID,
			Status: postgres.DriverStatusEnRoute,
		}); err != nil {
			return err
		}

		if err := q.AssignDriverToOrder(ctx, postgres.AssignDriverToOrderParams{
			DriverID: pgtype.UUID{Bytes: driverUUID, Valid: true},
			ID:       order.ID,
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
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

func (s *DispatchService) AcceptAssignment(ctx context.Context, driverID uuid.UUID, orderID uuid.UUID) error {
	return s.store.ExecTx(ctx, func(q postgres.Querier) error {
		return q.SetDriverStatus(ctx, postgres.SetDriverStatusParams{
			ID:     driverID,
			Status: postgres.DriverStatusEnRoute,
		})
	})
}

func (s *DispatchService) RejectAssignment(ctx context.Context, driverID uuid.UUID, orderID uuid.UUID) error {
	return s.store.ExecTx(ctx, func(q postgres.Querier) error {
		if err := q.RejectOrderAssignment(ctx, postgres.RejectOrderAssignmentParams{
			ID:       orderID,
			DriverID: pgtype.UUID{Bytes: driverID, Valid: true},
		}); err != nil {
			return err
		}

		if err := q.SetDriverStatus(ctx, postgres.SetDriverStatusParams{
			ID:     driverID,
			Status: postgres.DriverStatusIdle,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *DispatchService) ArriveAtPickup(ctx context.Context, driverID uuid.UUID, orderID uuid.UUID) error {
	return s.store.ExecTx(ctx, func(q postgres.Querier) error {
		rows, err := q.MarkOrderArrived(ctx, postgres.MarkOrderArrivedParams{
			ID:       orderID,
			DriverID: pgtype.UUID{Bytes: driverID, Valid: true},
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrInvalidTransition
		}

		return nil
	})
}

func (s *DispatchService) PickUpOrder(ctx context.Context, driverID uuid.UUID, orderID uuid.UUID) error {
	return s.store.ExecTx(ctx, func(q postgres.Querier) error {
		rows, err := q.MarkOrderPickedUp(ctx, postgres.MarkOrderPickedUpParams{
			ID:       orderID,
			DriverID: pgtype.UUID{Bytes: driverID, Valid: true},
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrInvalidTransition
		}

		return nil
	})
}

func (s *DispatchService) CompleteOrder(ctx context.Context, driverID uuid.UUID, orderID uuid.UUID) error {
	return s.store.ExecTx(ctx, func(q postgres.Querier) error {
		rows, err := q.MarkOrderDelivered(ctx, postgres.MarkOrderDeliveredParams{
			ID:       orderID,
			DriverID: pgtype.UUID{Bytes: driverID, Valid: true},
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrInvalidTransition
		}

		if err := q.SetDriverStatus(ctx, postgres.SetDriverStatusParams{
			ID:     driverID,
			Status: postgres.DriverStatusIdle,
		}); err != nil {
			return err
		}

		return nil
	})
}
