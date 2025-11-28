package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vantutran2k1/flowfleet/internal/adapter/storage/postgres"
)

type DriverHandler struct {
	repo *postgres.Queries
}

func NewDriverHandler(repo *postgres.Queries) *DriverHandler {
	return &DriverHandler{repo: repo}
}

type CreateDriverRequest struct {
	FleetID string  `json:"fleet_id" binding:"required,uuid"`
	Name    string  `json:"name" binding:"required"`
	Phone   string  `json:"phone" binding:"required,e164"`
	Lat     float64 `json:"lat" binding:"required,latitude"`
	Lng     float64 `json:"lng" binding:"required,longitude"`
}

func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var req CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fleetUUID, _ := uuid.Parse(req.FleetID)

	params := postgres.CreateDriverParams{
		FleetID:       fleetUUID,
		Name:          req.Name,
		Phone:         req.Phone,
		Status:        postgres.DriverStatusOffline,
		StMakepoint:   req.Lng,
		StMakepoint_2: req.Lat,
	}

	driver, err := h.repo.CreateDriver(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create driver"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         driver.ID,
		"created_at": driver.CreatedAt,
		"status":     "success",
	})
}
