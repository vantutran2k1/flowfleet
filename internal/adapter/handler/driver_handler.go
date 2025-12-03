package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vantutran2k1/flowfleet/internal/adapter/storage/postgres"
)

type DriverHandler struct {
	store postgres.Store
}

func NewDriverHandler(store postgres.Store) *DriverHandler {
	return &DriverHandler{store: store}
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

	var driver postgres.CreateDriverRow
	if err := h.store.ExecTx(c.Request.Context(), func(q postgres.Querier) error {
		createdDriver, err := q.CreateDriver(c.Request.Context(), params)
		if err != nil {
			return err
		}

		driver = createdDriver

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create driver"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         driver.ID,
		"created_at": driver.CreatedAt,
		"status":     "success",
	})
}
