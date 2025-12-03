package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vantutran2k1/flowfleet/internal/adapter/storage/postgres"
	"github.com/vantutran2k1/flowfleet/internal/adapter/websocket"
)

func TestDispatchService_CreateAndDispatchOrder_WithOneDriver(t *testing.T) {
	mockRepo := new(MockQuerier)
	mockGeo := new(MockGeoFinder)
	mockHub := &websocket.Hub{}

	svc := NewDispatchService(mockRepo, mockGeo, mockHub)

	driverID := uuid.New()
	fleetID := uuid.New()
	mockRepo.On("ExecTx", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("CreateOrder", mock.Anything, mock.Anything).Return(postgres.CreateOrderRow{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}, nil)
	mockRepo.On("GetDriver", mock.Anything, driverID).Return(postgres.GetDriverRow{
		ID:      driverID,
		FleetID: fleetID,
		Status:  postgres.DriverStatusIdle,
	}, nil)
	mockRepo.On("SetDriverStatus", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("AssignDriverToOrder", mock.Anything, mock.Anything).Return(nil)

	mockGeo.On("FindNearestDrivers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]string{driverID.String()}, nil)

	_, err := svc.CreateAndDispatchOrder(context.Background(), fleetID, 40.0, -74.0, 40.1, -74.1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockGeo.AssertExpectations(t)
}

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) AssignDriverToOrder(ctx context.Context, arg postgres.AssignDriverToOrderParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) ConfirmOrderAcceptance(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQuerier) CreateDriver(ctx context.Context, arg postgres.CreateDriverParams) (postgres.CreateDriverRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(postgres.CreateDriverRow), args.Error(1)
}

func (m *MockQuerier) CreateOrder(ctx context.Context, arg postgres.CreateOrderParams) (postgres.CreateOrderRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(postgres.CreateOrderRow), args.Error(1)
}

func (m *MockQuerier) FindNearestDrivers(ctx context.Context, arg postgres.FindNearestDriversParams) ([]postgres.FindNearestDriversRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]postgres.FindNearestDriversRow), args.Error(1)
}

func (m *MockQuerier) GetDriver(ctx context.Context, id uuid.UUID) (postgres.GetDriverRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(postgres.GetDriverRow), args.Error(1)
}

func (m *MockQuerier) GetDriverByEmail(ctx context.Context, email string) (postgres.GetDriverByEmailRow, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(postgres.GetDriverByEmailRow), args.Error(1)
}

func (m *MockQuerier) ListDriversByFleet(ctx context.Context, fleetID uuid.UUID) ([]postgres.ListDriversByFleetRow, error) {
	args := m.Called(ctx, fleetID)
	return args.Get(0).([]postgres.ListDriversByFleetRow), args.Error(1)
}

func (m *MockQuerier) MarkOrderArrived(ctx context.Context, arg postgres.MarkOrderArrivedParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) MarkOrderDelivered(ctx context.Context, arg postgres.MarkOrderDeliveredParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) MarkOrderPickedUp(ctx context.Context, arg postgres.MarkOrderPickedUpParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) RejectOrderAssignment(ctx context.Context, arg postgres.RejectOrderAssignmentParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) SetDriverStatus(ctx context.Context, arg postgres.SetDriverStatusParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockQuerier) ExecTx(ctx context.Context, fn func(postgres.Querier) error) error {
	args := m.Called(ctx, fn)

	if fn != nil {
		if err := fn(m); err != nil {
			return err
		}
	}

	return args.Error(0)
}

type MockGeoFinder struct {
	mock.Mock
}

func (m *MockGeoFinder) FindNearestDrivers(ctx context.Context, lat, lng float64, radiusKm float64) ([]string, error) {
	args := m.Called(ctx, lat, lng, radiusKm)
	return args.Get(0).([]string), args.Error(1)
}
