package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vantutran2k1/flowfleet/internal/adapter/storage/postgres"
	"github.com/vantutran2k1/flowfleet/internal/core/service"
)

type AuthHandler struct {
	svc  *service.AuthService
	repo *postgres.Queries
}

func NewAuthHandler(svc *service.AuthService, db *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		svc:  svc,
		repo: postgres.New(db),
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driver, err := h.repo.GetDriverByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if !h.svc.CheckPasswordHash(req.Password, driver.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.svc.GenerateToken(driver.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
