package seat

import (
	"errors"
	"net/http"

	"github.com/cinema-booking/backend/internal/auth"
	"github.com/cinema-booking/backend/internal/booking"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	service *booking.Service
}

func NewHandler(service *booking.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListShowtimes(c *gin.Context) {
	items, err := h.service.ListShowtimes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list showtimes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"showtimes": items})
}

func (h *Handler) ListSeats(c *gin.Context) {
	showtimeID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid showtime id"})
		return
	}

	seats, err := h.service.ListSeats(c.Request.Context(), showtimeID)
	if err != nil {
		if errors.Is(err, booking.ErrShowtimeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list seats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"seats": seats})
}

type lockSeatsRequest struct {
	SeatNos []string `json:"seat_nos" binding:"required,min=1"`
}

func (h *Handler) LockSeats(c *gin.Context) {
	showtimeID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid showtime id"})
		return
	}

	var req lockSeatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "seat_nos is required"})
		return
	}

	userIDHex := c.GetString(auth.ContextUserIDKey)
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.service.LockSeats(c.Request.Context(), showtimeID, userID, req.SeatNos)
	if err != nil {
		switch {
		case errors.Is(err, booking.ErrShowtimeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
		case errors.Is(err, booking.ErrSeatNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "seat not found"})
		case errors.Is(err, booking.ErrSeatUnavailable):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lock seats"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"booking": result})
}
