package domain

import (
	"errors"
	"time"
)

var (
	ErrDriverNotFound = errors.New("driver not found")
	ErrDriverOffline  = errors.New("driver is currently offline")
)

type DriverStatus string

const (
	DriverStatusOffline DriverStatus = "OFFLINE"
	DriverStatusIdle    DriverStatus = "IDLE"
	DriverStatusEnRoute DriverStatus = "EN_ROUTE"
)

type Driver struct {
	ID        string
	FleetID   string
	Name      string
	Phone     string
	Status    DriverStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (d *Driver) CanAcceptOrder() bool {
	return d.Status == DriverStatusIdle
}
