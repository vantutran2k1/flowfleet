package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vantutran2k1/flowfleet/internal/core/domain"
	"github.com/vantutran2k1/flowfleet/internal/core/service"
)

type OrderHandler struct {
	svc *service.DispatchService
}

func NewOrderHandler(svc *service.DispatchService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type CreateOrderRequest struct {
	FleetID   string  `json:"fleet_id" binding:"required,uuid"`
	PickupLat float64 `json:"pickup_lat" binding:"required"`
	PickupLng float64 `json:"pickup_lng" binding:"required"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fleetUUID, _ := uuid.Parse(req.FleetID)

	orderID, err := h.svc.CreateAndDispatchOrder(c.Request.Context(), fleetUUID, req.PickupLat, req.PickupLng)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"order_id": orderID, "status": "processing"})
}

func (h *OrderHandler) ArriveAtPickup(c *gin.Context) {
	h.handleTransition(c, h.svc.ArriveAtPickup)
}

func (h *OrderHandler) PickUpOrder(c *gin.Context) {
	h.handleTransition(c, h.svc.PickUpOrder)
}

func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	h.handleTransition(c, h.svc.CompleteOrder)
}

func (h *OrderHandler) handleTransition(c *gin.Context, fn func(context.Context, uuid.UUID, uuid.UUID) error) {
	orderIDStr := c.Param("id")
	orderUUID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid order id"})
		return
	}

	driverIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	driverUUID := driverIDInterface.(uuid.UUID)

	if err := fn(c.Request.Context(), driverUUID, orderUUID); err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state transition, check if order is in correct status"})
			return
		}

		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
