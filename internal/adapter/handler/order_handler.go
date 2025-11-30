package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
